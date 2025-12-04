package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/dusted-go/logging/prettylog"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultAgentConfigFile = "agent.yaml"
	defaultJavaPath        = "java"
	defaultJDBCPort        = "8888"
	heartbeatInterval      = 30 * time.Second
)

type AgentConfig struct {
	ServerURL    string          `mapstructure:"serverURL"`
	ClientID     string          `mapstructure:"clientID"`
	ClientSecret string          `mapstructure:"clientSecret"`
	TenantID     string          `mapstructure:"tenantID"`
	DIS          types.DISConfig `mapstructure:"dis"`
	TLS          struct {
		InsecureSkipVerify bool `mapstructure:"insecureSkipVerify"`
	} `mapstructure:"tls"`
}

func loadConfig() (*AgentConfig, error) {
	viper.SetConfigFile(defaultAgentConfigFile)
	viper.SetDefault("dis.logLevel", slog.LevelInfo)
	viper.SetDefault("dis.jdbcConfig.javaPath", defaultJavaPath)
	viper.SetDefault("dis.jdbcConfig.jdbcPort", defaultJDBCPort)

	viper.SetEnvPrefix("DISAGENT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Config file is optional; fall back to env-only.
	if _, err := os.Stat(defaultAgentConfigFile); err == nil {
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg AgentConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.ServerURL == "" {
		return nil, errors.New("serverURL is required")
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("clientID and clientSecret are required")
	}

	// Ensure JDBCConfig exists
	if cfg.DIS.JDBCConfig == nil {
		cfg.DIS.JDBCConfig = &types.JDBCConfig{
			JavaPath: defaultJavaPath,
			JDBCPort: defaultJDBCPort,
		}
	}

	return &cfg, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := slog.New(prettylog.NewHandler(&slog.HandlerOptions{
		Level: slog.Level(cfg.DIS.LogLevel),
	}))

	db := database.NewIBMi400(&cfg.DIS, logger)
	if err := db.StartJDBCRunner(); err != nil {
		log.Fatalf("start jdbc runner: %v", err)
	}
	defer db.StopJDBCRunner()

	runAgent(ctx, cfg, db, logger)
}

func runAgent(ctx context.Context, cfg *AgentConfig, db database.DB, logger *slog.Logger) {
	backoff := time.Second
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		if err := runOnce(ctx, cfg, db, logger); err != nil {
			logger.Error("agent loop error", slog.Any("err", err))
			time.Sleep(backoff)
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}

		// Successful run; reset backoff for future interruptions.
		backoff = time.Second
	}
}

func runOnce(ctx context.Context, cfg *AgentConfig, db database.DB, logger *slog.Logger) error {
	creds := dialCredentials(cfg)
	conn, err := grpc.DialContext(ctx, cfg.ServerURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("dial bridge: %w", err)
	}
	defer conn.Close()

	client := proto.NewAgentServiceClient(conn)
	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connect stream: %w", err)
	}

	hello := &proto.AgentHello{
		ClientId:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Version:      "agent-0.1.0",
		TenantId:     cfg.TenantID,
	}
	if err := stream.Send(&proto.AgentToServer{Payload: &proto.AgentToServer_Hello{Hello: hello}}); err != nil {
		return fmt.Errorf("send hello: %w", err)
	}

	hbCtx, cancelHB := context.WithCancel(ctx)
	defer cancelHB()
	go sendHeartbeats(hbCtx, stream, cfg)

	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		req := msg.GetJobRequest()
		if req == nil {
			continue
		}

		res := executeJob(ctx, db, req, logger)
		if err := stream.Send(&proto.AgentToServer{Payload: &proto.AgentToServer_JobResult{JobResult: res}}); err != nil {
			return err
		}
	}
}

func executeJob(ctx context.Context, db database.DB, req *proto.JobRequest, logger *slog.Logger) *proto.JobResult {
	res := &proto.JobResult{JobId: req.JobId}
	const defaultMaxRows = 1000
	maxRows := defaultMaxRows
	if v := os.Getenv("DISAGENT_MAX_ROWS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxRows = n
		}
	}

	switch req.Kind {
	case proto.JobKind_JOB_KIND_QUERY:
		var rows []types.ResultRow
		var err error
		if req.IncludeSrc {
			rows, err = db.QueryWithSource(ctx, req.Sql)
		} else {
			rows, err = db.Query(ctx, req.Sql)
		}
		if err != nil {
			res.Status = proto.Status_STATUS_ERROR
			res.Message = err.Error()
			return res
		}
		if len(rows) > maxRows {
			rows = rows[:maxRows]
			res.Message = "truncated rows to max limit"
		}
		res.Rows = resultRowsToProto(rows)
		res.Status = proto.Status_STATUS_OK
	case proto.JobKind_JOB_KIND_PING_SERVICE:
		res.Status, res.Message = statusFromError(db.PingService(ctx))
	case proto.JobKind_JOB_KIND_PING_DATABASE:
		res.Status, res.Message = statusFromError(db.PingDatabase(ctx))
	case proto.JobKind_JOB_KIND_CONNECT:
		res.Status, res.Message = statusFromError(db.Connect(ctx))
	case proto.JobKind_JOB_KIND_DISCONNECT:
		res.Status, res.Message = statusFromError(db.Disconnect(ctx))
	case proto.JobKind_JOB_KIND_START_JDBC:
		res.Status, res.Message = statusFromError(db.StartJDBCRunner())
	case proto.JobKind_JOB_KIND_STOP_JDBC:
		res.Status, res.Message = statusFromError(db.StopJDBCRunner())
	default:
		res.Status = proto.Status_STATUS_ERROR
		res.Message = "unknown job kind"
	}

	if logger != nil && res.Status == proto.Status_STATUS_ERROR {
		logger.Error("job failed", slog.String("job_id", req.JobId), slog.String("message", res.Message))
	}

	return res
}

func statusFromError(err error) (proto.Status, string) {
	if err != nil {
		return proto.Status_STATUS_ERROR, err.Error()
	}
	return proto.Status_STATUS_OK, ""
}

func dialCredentials(cfg *AgentConfig) credentials.TransportCredentials {
	if cfg.TLS.InsecureSkipVerify {
		return credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}
	return credentials.NewTLS(&tls.Config{})
}

func sendHeartbeats(ctx context.Context, stream proto.AgentService_ConnectClient, cfg *AgentConfig) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			stream.Send(&proto.AgentToServer{
				Payload: &proto.AgentToServer_Heartbeat{
					Heartbeat: &proto.Heartbeat{
						At:       timestamppb.New(t),
						TenantId: cfg.TenantID,
						AgentId:  cfg.ClientID,
					},
				},
			})
		}
	}
}

func resultRowsToProto(rows []types.ResultRow) []*proto.Row {
	out := make([]*proto.Row, 0, len(rows))
	for _, r := range rows {
		fields := make(map[string]*structpb.Value, len(r))
		for k, v := range r {
			if t, ok := v.(time.Time); ok {
				fields[k] = structpb.NewNumberValue(float64(t.UnixMilli()))
				continue
			}
			val, err := structpb.NewValue(v)
			if err != nil {
				val = structpb.NewStringValue(fmt.Sprint(v))
			}
			fields[k] = val
		}
		out = append(out, &proto.Row{Fields: fields})
	}
	return out
}
