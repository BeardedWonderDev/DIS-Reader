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
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/logging"
	"github.com/BeardedWonderDev/DIS-Reader/internal/runnerjar"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/dusted-go/logging/prettylog"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultAgentConfigFile = "agent.yaml"
	defaultJavaPath        = "java"
	defaultJDBCPort        = "8888"
	heartbeatInterval      = 30 * time.Second
	agentEnvPrefix         = "disagent"
)

type AgentConfig struct {
	ServerURL          string          `mapstructure:"serverURL"`
	ClientID           string          `mapstructure:"clientID"`
	ClientSecret       string          `mapstructure:"clientSecret"`
	TenantID           string          `mapstructure:"tenantID"`
	AutoConnectOnStart bool            `mapstructure:"autoConnectOnStart"`
	DIS                types.DISConfig `mapstructure:"dis"`
	TLS                struct {
		Enabled            bool `mapstructure:"enabled"`
		InsecureSkipVerify bool `mapstructure:"insecureSkipVerify"`
	} `mapstructure:"tls"`
}

func loadConfig() (*AgentConfig, error) {
	return loadConfigWith(viper.New())
}

func loadConfigWith(v *viper.Viper) (*AgentConfig, error) {
	v.SetConfigFile(defaultAgentConfigFile)
	v.SetDefault("dis.logLevel", slog.LevelInfo)
	v.SetDefault("dis.jdbcConfig.javaPath", defaultJavaPath)
	v.SetDefault("dis.jdbcConfig.jdbcPort", defaultJDBCPort)
	v.SetDefault("autoConnectOnStart", false)
	v.SetDefault("tls.enabled", true)
	v.SetDefault("tls.insecureSkipVerify", false)

	if _, err := os.ReadFile(defaultAgentConfigFile); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
	} else {
		if os.IsNotExist(err) {
			log.Printf("Could not find %s. Attempting to use environment variables.\n", defaultAgentConfigFile)
		} else {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg AgentConfig
	for _, fieldName := range getFlattenedStructFields(reflect.TypeOf(cfg)) {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", agentEnvPrefix, strings.ReplaceAll(fieldName, ".", "_")))
		if envVar, ok := os.LookupEnv(envKey); ok {
			v.Set(fieldName, envVar)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.ServerURL == "" {
		return nil, errors.New("serverURL is required")
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("clientID and clientSecret are required")
	}

	if cfg.DIS.JDBCConfig == nil {
		cfg.DIS.JDBCConfig = &types.JDBCConfig{
			JavaPath: defaultJavaPath,
			JDBCPort: defaultJDBCPort,
		}
	}

	return &cfg, nil
}

func getFlattenedStructFields(t reflect.Type) []string {
	return getFlattenedStructFieldsHelper(t, []string{})
}

func getFlattenedStructFieldsHelper(t reflect.Type, prefixes []string) []string {
	unwrappedT := t
	if t.Kind() == reflect.Pointer {
		unwrappedT = t.Elem()
	}

	flattenedFields := make([]string, 0)
	for i := 0; i < unwrappedT.NumField(); i++ {
		field := unwrappedT.Field(i)
		fieldName := field.Tag.Get("mapstructure")
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			flattenedFields = append(flattenedFields, getFlattenedStructFieldsHelper(field.Type, append(prefixes, fieldName))...)
		default:
			flattenedField := fieldName
			if len(prefixes) > 0 {
				flattenedField = fmt.Sprintf("%s.%s", strings.Join(prefixes, "."), fieldName)
			}
			flattenedFields = append(flattenedFields, flattenedField)
		}
	}

	return flattenedFields
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
	logger.Info("agent starting",
		slog.String("server", logging.RedactURLHost(cfg.ServerURL)),
		slog.String("tenant_id", cfg.TenantID),
		slog.String("agent_id", cfg.ClientID),
		slog.Bool("auto_connect_on_start", cfg.AutoConnectOnStart),
		slog.Bool("tls_enabled", cfg.TLS.Enabled),
		slog.Bool("tls_insecure_skip_verify", cfg.TLS.InsecureSkipVerify),
	)

	var cleanupJar func() error
	if cfg.DIS.JDBCConfig != nil && cfg.DIS.JDBCConfig.JarPath == "" {
		extracted, err := runnerjar.Extract()
		if err != nil {
			log.Fatalf("prepare runner jar: %v", err)
		}
		cfg.DIS.JDBCConfig.JarPath = extracted.JarPath
		cleanupJar = extracted.Cleanup
	}
	if cleanupJar != nil {
		defer cleanupJar()
	}

	db := database.NewIBMi400(&cfg.DIS, logger)
	if err := db.StartJDBCRunner(); err != nil {
		log.Fatalf("start jdbc runner: %v", err)
	}
	logger.Info("jdbc runner started", slog.String("phase", "jdbc_start"))
	defer db.StopJDBCRunner()

	if cfg.AutoConnectOnStart && cfg.DIS.Host != "" && cfg.DIS.User != "" && cfg.DIS.Password != "" {
		ctxConnect, cancel := context.WithTimeout(ctx, 15*time.Second)
		start := time.Now()
		if err := db.Connect(ctxConnect); err != nil {
			logger.Error("auto-connect failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "dis_connect", "connect"), logging.DurationAttr(time.Since(start)))...)
			log.Fatalf("auto-connect failed: %v", err)
		}
		logger.Info("auto-connect succeeded", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "dis_connect", "connect"), logging.DurationAttr(time.Since(start)))...)
		cancel()
	}

	runAgent(ctx, cfg, db, logger)
}

func runAgent(ctx context.Context, cfg *AgentConfig, db database.DB, logger *slog.Logger) {
	backoff := time.Second
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		logger.Info("bridge connect attempt",
			append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"),
				slog.String("server", logging.RedactURLHost(cfg.ServerURL)),
				slog.Duration("backoff", backoff),
			)...)

		if err := runOnce(ctx, cfg, db, logger); err != nil {
			logger.Error("agent loop error", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.Any("err", err))...)
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
		logger.Error("bridge dial failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.Any("err", err))...)
		return fmt.Errorf("dial bridge: %w", err)
	}
	defer conn.Close()
	logger.Info("bridge dialed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.String("server", logging.RedactURLHost(cfg.ServerURL)))...)

	client := proto.NewAgentServiceClient(conn)
	stream, err := client.Connect(ctx)
	if err != nil {
		logger.Error("agent stream connect failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.Any("err", err))...)
		return fmt.Errorf("connect stream: %w", err)
	}

	hello := &proto.AgentHello{
		ClientId:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Version:      "agent-0.1.0",
		TenantId:     cfg.TenantID,
	}
	if err := stream.Send(&proto.AgentToServer{Payload: &proto.AgentToServer_Hello{Hello: hello}}); err != nil {
		logger.Error("send hello failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "hello"), slog.Any("err", err))...)
		return fmt.Errorf("send hello: %w", err)
	}
	logger.Info("hello sent", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "hello"), slog.String("version", hello.Version))...)

	hbCtx, cancelHB := context.WithCancel(ctx)
	defer cancelHB()
	go sendHeartbeats(hbCtx, stream, cfg, logger)

	for {
		msg, err := stream.Recv()
		if err != nil {
			logger.Warn("agent stream recv error", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "recv"), slog.Any("err", err))...)
			return err
		}
		req := msg.GetJobRequest()
		if req == nil {
			continue
		}

		res := executeJob(ctx, cfg, db, req, logger)
		if err := stream.Send(&proto.AgentToServer{Payload: &proto.AgentToServer_JobResult{JobResult: res}}); err != nil {
			return err
		}
	}
}

func executeJob(ctx context.Context, cfg *AgentConfig, db database.DB, req *proto.JobRequest, logger *slog.Logger) *proto.JobResult {
	res := &proto.JobResult{JobId: req.JobId}
	const defaultMaxRows = 1000
	maxRows := defaultMaxRows
	if v := os.Getenv("DISAGENT_MAX_ROWS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxRows = n
		}
	}

	start := time.Now()
	attrs := logging.CommonAttrs(cfg.TenantID, cfg.ClientID, req.JobId, "job", req.Kind.String())
	attrs = append(attrs, slog.String("sql_hint", logging.SQLHint(req.Sql)))
	logger.Debug("job started", attrs...)

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
		attrs = append(attrs, logging.RowsAttrs(len(rows), res.Message != "")...)
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

	attrs = append(attrs, logging.DurationAttr(time.Since(start)))

	if logger != nil {
		if res.Status == proto.Status_STATUS_ERROR {
			logger.Error("job failed", append(attrs, slog.String("message", res.Message))...)
		} else {
			logger.Debug("job completed", append(attrs, slog.String("status", res.Status.String()))...)
		}
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
	if !cfg.TLS.Enabled {
		return insecure.NewCredentials()
	}
	if cfg.TLS.InsecureSkipVerify {
		return credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}
	return credentials.NewTLS(&tls.Config{})
}

func sendHeartbeats(ctx context.Context, stream proto.AgentService_ConnectClient, cfg *AgentConfig, logger *slog.Logger) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if err := stream.Send(&proto.AgentToServer{
				Payload: &proto.AgentToServer_Heartbeat{
					Heartbeat: &proto.Heartbeat{
						At:       timestamppb.New(t),
						TenantId: cfg.TenantID,
						AgentId:  cfg.ClientID,
					},
				},
			}); err != nil {
				logger.Warn("heartbeat send failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "heartbeat", "send"), slog.Any("err", err))...)
				return
			}
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
