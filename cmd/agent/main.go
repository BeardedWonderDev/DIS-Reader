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
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/agentcore"
	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/controlapi"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/logging"
	"github.com/BeardedWonderDev/DIS-Reader/internal/runnerjar"
	"github.com/BeardedWonderDev/DIS-Reader/internal/servicectl"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/dusted-go/logging/prettylog"
	kitlog "github.com/go-kit/log"
	"github.com/grafana/loki-client-go/loki"
	"github.com/prometheus/common/config"
	slogloki "github.com/samber/slog-loki/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var errRestartRequired = errors.New("restart requested by bridge")

const (
	heartbeatInterval = 30 * time.Second
)

// heartbeatIntervalVar is test-overridable; default heartbeatInterval.
var heartbeatIntervalVar = heartbeatInterval

type AgentConfig = agentcore.Config

type agentStatus struct {
	bridgeConnected atomic.Bool
	lastHeartbeat   atomic.Value
	lastError       atomic.Value
}

func loadConfig() (*AgentConfig, error) {
	return agentcore.Load(agentcore.DefaultAgentConfigFile)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	baseHandler := prettylog.NewHandler(&slog.HandlerOptions{
		Level: slog.Level(cfg.DIS.LogLevel),
	})
	logger := slog.New(baseHandler)

	var loggerValue atomic.Value
	loggerValue.Store(logger)

	status := &agentStatus{}

	logger.Info("agent starting",
		slog.String("server", logging.RedactURLHost(cfg.ServerURL)),
		slog.String("tenant_id", cfg.TenantID),
		slog.String("agent_id", cfg.ClientID),
		slog.Bool("auto_connect_on_start", cfg.AutoConnectOnStart),
		slog.Bool("tls_enabled", cfg.TLS.Enabled),
		slog.Bool("tls_insecure_skip_verify", cfg.TLS.InsecureSkipVerify),
	)

	if cfg.Control.Enabled {
		if cfg.Control.Token == "" {
			logger.Warn("control API enabled without token; consider setting control.token for security")
		}
		statusFn := func() controlapi.Status {
			st := controlapi.Status{
				Running:         true,
				BridgeConnected: status.bridgeConnected.Load(),
				Version:         "agent-0.1.0",
				AgentID:         cfg.ClientID,
				ServerURL:       cfg.ServerURL,
				ConfigPath:      agentcore.DefaultAgentConfigFile,
			}
			if v := status.lastHeartbeat.Load(); v != nil {
				if t, ok := v.(time.Time); ok {
					st.LastHeartbeat = &t
				}
			}
			if v := status.lastError.Load(); v != nil {
				st.LastError = fmt.Sprint(v)
			}
			return st
		}
		effFn := func() *agentcore.EffectiveConfig {
			return agentcore.ToEffective(copyConfig(cfg))
		}
		applyFn := func(c *agentcore.Config) error {
			merged := copyConfig(cfg)
			merged.ServerURL = c.ServerURL
			merged.ClientID = c.ClientID
			merged.ClientSecret = c.ClientSecret
			merged.AutoConnectOnStart = c.AutoConnectOnStart
			if merged.DIS.JDBCConfig == nil {
				merged.DIS.JDBCConfig = &types.JDBCConfig{}
			}
			if c.DIS.JDBCConfig != nil {
				if c.DIS.JDBCConfig.JavaPath != "" {
					merged.DIS.JDBCConfig.JavaPath = c.DIS.JDBCConfig.JavaPath
				}
				if c.DIS.JDBCConfig.JDBCPort != "" {
					merged.DIS.JDBCConfig.JDBCPort = c.DIS.JDBCConfig.JDBCPort
				}
			}
			merged.DIS.Host = c.DIS.Host
			merged.DIS.User = c.DIS.User
			merged.DIS.Password = c.DIS.Password
			merged.TLS = c.TLS
			merged.Control = c.Control
			*cfg = *merged
			return agentcore.Save(agentcore.DefaultAgentConfigFile, merged)
		}
		if err := controlapi.Start(ctx, controlapi.Options{
			Addr:        cfg.Control.Addr,
			Token:       cfg.Control.Token,
			StatusFn:    statusFn,
			EffectiveFn: effFn,
			ApplyFn:     applyFn,
			ServiceCtl:  servicectl.New(),
		}); err != nil {
			logger.Warn("control api failed to start", slog.Any("err", err))
		} else {
			logger.Info("control api listening", slog.String("addr", cfg.Control.Addr))
		}
	}

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

	runAgent(ctx, cfg, db, baseHandler, &loggerValue, status)
}

func runAgent(ctx context.Context, cfg *AgentConfig, db database.DB, baseHandler slog.Handler, loggerVal *atomic.Value, status *agentStatus) {
	backoff := time.Second
	if status != nil {
		status.bridgeConnected.Store(false)
	}
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		currentLogger(loggerVal).Info("bridge connect attempt",
			append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"),
				slog.String("server", logging.RedactURLHost(cfg.ServerURL)),
				slog.Duration("backoff", backoff),
			)...)

		if err := runOnceFunc(ctx, cfg, db, baseHandler, loggerVal, status); err != nil {
			currentLogger(loggerVal).Error("agent loop error", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.Any("err", err))...)
			if status != nil {
				status.bridgeConnected.Store(false)
				status.lastError.Store(err.Error())
			}
			if errors.Is(err, errRestartRequired) {
				time.Sleep(backoff)
				backoff = time.Second
				continue
			}
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

// runOnceFunc is patchable for tests.
var runOnceFunc = runOnce

func runOnce(ctx context.Context, cfg *AgentConfig, db database.DB, baseHandler slog.Handler, loggerVal *atomic.Value, status *agentStatus) error {
	creds := dialCredentials(cfg)
	conn, err := grpc.DialContext(ctx, cfg.ServerURL, grpc.WithTransportCredentials(creds))
	if err != nil {
		currentLogger(loggerVal).Error("bridge dial failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.Any("err", err))...)
		return fmt.Errorf("dial bridge: %w", err)
	}
	defer conn.Close()
	currentLogger(loggerVal).Info("bridge dialed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.String("server", logging.RedactURLHost(cfg.ServerURL)))...)

	client := proto.NewAgentServiceClient(conn)
	stream, err := client.Connect(ctx)
	if err != nil {
		currentLogger(loggerVal).Error("agent stream connect failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "connect"), slog.Any("err", err))...)
		return fmt.Errorf("connect stream: %w", err)
	}
	if status != nil {
		status.bridgeConnected.Store(true)
	}

	hello := &proto.AgentHello{
		ClientId:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Version:      "agent-0.1.0",
		TenantId:     cfg.TenantID,
	}
	if err := stream.Send(&proto.AgentToServer{Payload: &proto.AgentToServer_Hello{Hello: hello}}); err != nil {
		currentLogger(loggerVal).Error("send hello failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "hello"), slog.Any("err", err))...)
		return fmt.Errorf("send hello: %w", err)
	}
	currentLogger(loggerVal).Info("hello sent", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "hello"), slog.String("version", hello.Version))...)

	hbCtx, cancelHB := context.WithCancel(ctx)
	defer cancelHB()
	go sendHeartbeats(hbCtx, stream, cfg, loggerVal, status)

	for {
		msg, err := stream.Recv()
		if err != nil {
			currentLogger(loggerVal).Warn("agent stream recv error", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "recv"), slog.Any("err", err))...)
			if status != nil {
				status.bridgeConnected.Store(false)
				status.lastError.Store(err.Error())
			}
			return err
		}

		switch payload := msg.Payload.(type) {
		case *proto.ServerToAgent_JobRequest:
			if payload.JobRequest == nil {
				continue
			}
			res := executeJob(ctx, cfg, db, payload.JobRequest, loggerVal)
			if err := stream.Send(&proto.AgentToServer{Payload: &proto.AgentToServer_JobResult{JobResult: res}}); err != nil {
				return err
			}
		case *proto.ServerToAgent_Config:
			if payload.Config != nil {
				if err := applyAgentConfig(baseHandler, loggerVal, payload.Config, cfg, db); err != nil {
					currentLogger(loggerVal).Warn("apply agent config failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "bridge_connect", "config"), slog.Any("err", err))...)
				}
			}
		default:
			continue
		}
	}
}

func executeJob(ctx context.Context, cfg *AgentConfig, db database.DB, req *proto.JobRequest, loggerVal *atomic.Value) *proto.JobResult {
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
	currentLogger(loggerVal).Debug("job started", attrs...)

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
	case proto.JobKind_JOB_KIND_READ_CONFIG:
		res.ConfigStatus = buildConfigStatus(cfg)
		res.Status = proto.Status_STATUS_OK
	default:
		res.Status = proto.Status_STATUS_ERROR
		res.Message = "unknown job kind"
	}

	attrs = append(attrs, logging.DurationAttr(time.Since(start)))

	if res.Status == proto.Status_STATUS_ERROR {
		currentLogger(loggerVal).Error("job failed", append(attrs, slog.String("message", res.Message))...)
	} else {
		currentLogger(loggerVal).Debug("job completed", append(attrs, slog.String("status", res.Status.String()))...)
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

func sendHeartbeats(ctx context.Context, stream proto.AgentService_ConnectClient, cfg *AgentConfig, loggerVal *atomic.Value, status *agentStatus) {
	ticker := time.NewTicker(heartbeatIntervalVar)
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
				currentLogger(loggerVal).Warn("heartbeat send failed", append(logging.CommonAttrs(cfg.TenantID, cfg.ClientID, "", "heartbeat", "send"), slog.Any("err", err))...)
				return
			}
			if status != nil {
				now := time.Now()
				status.lastHeartbeat.Store(now)
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

func applyAgentConfig(baseHandler slog.Handler, loggerVal *atomic.Value, cfg *proto.AgentConfig, agentCfg *AgentConfig, db database.DB) error {
	if cfg == nil {
		return nil
	}

	// Apply runtime overrides first so downstream jobs use remote-configured values.
	if rt := cfg.GetRuntime(); rt != nil {
		restart, reconnect := agentcore.ApplyRuntimeOverrides(agentCfg, rt)

		if restart && db != nil {
			// Best-effort restart so new ports/paths take effect before connect jobs.
			if err := db.StopJDBCRunner(); err != nil {
				currentLogger(loggerVal).Warn("failed to stop jdbc runner for remote config", slog.Any("err", err))
			}
			if err := db.StartJDBCRunner(); err != nil {
				return fmt.Errorf("restart jdbc runner: %w", err)
			}
		}
		if reconnect {
			_ = agentcore.Save(agentcore.DefaultAgentConfigFile, agentCfg)
			return errRestartRequired
		}
	}

	// Loki config remains optional; only apply when provided.
	if cfg.Loki == nil || strings.TrimSpace(cfg.Loki.Url) == "" {
		// Persist effective config even if Loki absent to capture runtime overrides.
		_ = agentcore.Save(agentcore.DefaultAgentConfigFile, agentCfg)
		return nil
	}

	handler, err := newLokiHandler(cfg.Loki)
	if err != nil {
		return err
	}

	combined := slog.New(fanoutHandler{handlers: []slog.Handler{baseHandler, handler}})
	loggerVal.Store(combined)
	combined.Info("loki logging enabled",
		slog.String("endpoint", logging.RedactURLHost(cfg.Loki.Url)),
		slog.String("tenant_id", cfg.Loki.TenantId),
		slog.String("agent_id", agentCfg.ClientID),
	)
	agentCfg.AppliedLoki = sanitizeLokiConfig(cfg.Loki)
	_ = agentcore.Save(agentcore.DefaultAgentConfigFile, agentCfg)
	return nil
}

func buildConfigStatus(cfg *AgentConfig) *proto.AgentConfigStatus {
	if cfg == nil {
		return nil
	}

	status := &proto.AgentConfigStatus{
		Runtime: &proto.AgentRuntimeStatus{
			DisHost:         cfg.DIS.Host,
			DisUser:         cfg.DIS.User,
			JdbcPort:        "",
			JavaPath:        "",
			TenantId:        cfg.TenantID,
			HasPassword:     cfg.DIS.Password != "",
			HasClientSecret: cfg.ClientSecret != "",
		},
	}

	if cfg.DIS.JDBCConfig != nil {
		status.Runtime.JdbcPort = cfg.DIS.JDBCConfig.JDBCPort
		status.Runtime.JavaPath = cfg.DIS.JDBCConfig.JavaPath
	}

	if cfg.AppliedLoki != nil {
		status.Loki = sanitizeLokiConfig(cfg.AppliedLoki)
	}

	return status
}

func sanitizeLokiConfig(in *proto.LokiConfig) *proto.LokiConfig {
	if in == nil {
		return nil
	}
	out := *in
	out.ApiKey = ""
	return &out
}

func newLokiHandler(cfg *proto.LokiConfig) (slog.Handler, error) {
	endpoint := strings.TrimSpace(cfg.GetUrl())
	if endpoint == "" {
		return nil, fmt.Errorf("loki endpoint is required")
	}

	clientCfg, err := loki.NewDefaultConfig(endpoint)
	if err != nil {
		return nil, fmt.Errorf("loki config: %w", err)
	}

	if key := strings.TrimSpace(cfg.GetApiKey()); key != "" {
		clientCfg.Client.Authorization = &config.Authorization{
			Type:        "Bearer",
			Credentials: config.Secret(key),
		}
	}

	if t := strings.TrimSpace(cfg.GetTenantId()); t != "" {
		clientCfg.TenantID = t
	}

	client, err := loki.NewWithLogger(clientCfg, kitlog.NewNopLogger())
	if err != nil {
		return nil, fmt.Errorf("loki client: %w", err)
	}

	minLevel := logging.MapLogLevel(cfg.GetMinLevel())
	opt := slogloki.Option{Client: client, Level: minLevel}
	handler := opt.NewLokiHandler()
	attrs := labelAttrs(cfg.GetLabels())
	if cfg.GetTenantId() != "" {
		attrs = append(attrs, slog.String("tenant_id", cfg.GetTenantId()))
	}
	if len(attrs) > 0 {
		handler = handler.WithAttrs(attrs)
	}
	return handler, nil
}

func labelAttrs(labels map[string]string) []slog.Attr {
	if len(labels) == 0 {
		return nil
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attrs := make([]slog.Attr, 0, len(keys))
	for _, k := range keys {
		attrs = append(attrs, slog.String(k, labels[k]))
	}
	return attrs
}

func copyConfig(in *AgentConfig) *AgentConfig {
	if in == nil {
		return nil
	}
	out := *in
	if in.DIS.JDBCConfig != nil {
		jc := *in.DIS.JDBCConfig
		out.DIS.JDBCConfig = &jc
	}
	return &out
}

func currentLogger(loggerVal *atomic.Value) *slog.Logger {
	if loggerVal == nil {
		return slog.Default()
	}
	if l := loggerVal.Load(); l != nil {
		if logPtr, ok := l.(*slog.Logger); ok && logPtr != nil {
			return logPtr
		}
	}
	return slog.Default()
}

// fanoutHandler duplicates records to multiple handlers.
type fanoutHandler struct {
	handlers []slog.Handler
}

func (h fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (h fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	copyHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		copyHandlers[i] = handler.WithAttrs(attrs)
	}
	return fanoutHandler{handlers: copyHandlers}
}

func (h fanoutHandler) WithGroup(name string) slog.Handler {
	copyHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		copyHandlers[i] = handler.WithGroup(name)
	}
	return fanoutHandler{handlers: copyHandlers}
}
