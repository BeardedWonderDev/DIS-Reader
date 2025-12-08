package main

import (
	"context"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/agentcore"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"google.golang.org/protobuf/types/known/structpb"
	"log/slog"
)

// stubDB implements database.DB with in-memory flags for lifecycle assertions.
type stubDB struct {
	startCalled bool
	stopCalled  bool
}

func (s *stubDB) StartJDBCRunner() error                 { s.startCalled = true; return nil }
func (s *stubDB) StopJDBCRunner() error                  { s.stopCalled = true; return nil }
func (s *stubDB) Connect(ctx context.Context) error      { return nil }
func (s *stubDB) Disconnect(ctx context.Context) error   { return nil }
func (s *stubDB) PingService(ctx context.Context) error  { return nil }
func (s *stubDB) PingDatabase(ctx context.Context) error { return nil }
func (s *stubDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (s *stubDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return nil, nil
}
func (s *stubDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	return nil, nil
}
func (s *stubDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (s *stubDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return nil, nil
}

func TestExecuteJob_StartStopJDBC(t *testing.T) {
	db := &stubDB{}
	ctx := context.Background()
	cfg := &AgentConfig{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var loggerVal atomic.Value
	loggerVal.Store(logger)

	startRes := executeJob(ctx, cfg, db, &bridgeproto.JobRequest{JobId: "1", Kind: bridgeproto.JobKind_JOB_KIND_START_JDBC}, &loggerVal)
	if startRes.Status != bridgeproto.Status_STATUS_OK || !db.startCalled {
		t.Fatalf("expected start to succeed and flag to be set, res=%v", startRes)
	}

	stopRes := executeJob(ctx, cfg, db, &bridgeproto.JobRequest{JobId: "2", Kind: bridgeproto.JobKind_JOB_KIND_STOP_JDBC}, &loggerVal)
	if stopRes.Status != bridgeproto.Status_STATUS_OK || !db.stopCalled {
		t.Fatalf("expected stop to succeed and flag to be set, res=%v", stopRes)
	}
}

func TestExecuteJob_ReadConfigSanitized(t *testing.T) {
	db := &stubDB{}
	ctx := context.Background()
	cfg := &AgentConfig{
		DIS: types.DISConfig{
			Host:     "host",
			User:     "user",
			Password: "secretpw",
			JDBCConfig: &types.JDBCConfig{
				JDBCPort: "9999",
				JavaPath: "/usr/bin/java",
			},
		},
		ClientSecret: "client-secret",
		TenantID:     "tenant-1",
		AppliedLoki: &bridgeproto.LokiConfig{
			Url:      "https://loki.example",
			ApiKey:   "apikey",
			TenantId: "lokitenant",
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	var loggerVal atomic.Value
	loggerVal.Store(logger)

	res := executeJob(ctx, cfg, db, &bridgeproto.JobRequest{JobId: "3", Kind: bridgeproto.JobKind_JOB_KIND_READ_CONFIG}, &loggerVal)
	if res.Status != bridgeproto.Status_STATUS_OK {
		t.Fatalf("expected OK status, got %v", res.Status)
	}
	if res.ConfigStatus == nil || res.ConfigStatus.Runtime == nil {
		t.Fatalf("expected config status in response")
	}
	rt := res.ConfigStatus.Runtime
	if !rt.HasPassword || !rt.HasClientSecret {
		t.Fatalf("expected password and client secret flags to be true")
	}
	if rt.DisHost != "host" || rt.DisUser != "user" || rt.JdbcPort != "9999" || rt.JavaPath != "/usr/bin/java" || rt.TenantId != "tenant-1" {
		t.Fatalf("unexpected runtime snapshot: %+v", rt)
	}
	if res.ConfigStatus.Loki == nil {
		t.Fatalf("expected loki snapshot")
	}
	if res.ConfigStatus.Loki.ApiKey != "" {
		t.Fatalf("expected loki api key to be sanitized")
	}
}

func TestDialCredentials(t *testing.T) {
	cases := []struct {
		name string
		cfg  *AgentConfig
		want string
	}{
		{"insecure when tls disabled", &AgentConfig{}, "insecure"},
		{"tls insecure skip verify", &AgentConfig{TLS: struct {
			Enabled            bool `mapstructure:"enabled" yaml:"enabled"`
			InsecureSkipVerify bool `mapstructure:"insecureSkipVerify" yaml:"insecureSkipVerify"`
		}{Enabled: true, InsecureSkipVerify: true}}, "tls"},
		{"tls strict", &AgentConfig{TLS: struct {
			Enabled            bool `mapstructure:"enabled" yaml:"enabled"`
			InsecureSkipVerify bool `mapstructure:"insecureSkipVerify" yaml:"insecureSkipVerify"`
		}{Enabled: true, InsecureSkipVerify: false}}, "tls"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			creds := dialCredentials(tc.cfg)
			if got := creds.Info().SecurityProtocol; got != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestLabelAttrsSorted(t *testing.T) {
	attrs := labelAttrs(map[string]string{"b": "2", "a": "1"})
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attrs, got %d", len(attrs))
	}
	if attrs[0].Key != "a" || attrs[1].Key != "b" {
		t.Fatalf("expected keys sorted, got %v", []slog.Attr{attrs[0], attrs[1]})
	}
	if labelAttrs(nil) != nil {
		t.Fatalf("expected nil for empty labels")
	}
}

func TestCopyConfigDeepCopiesJDBC(t *testing.T) {
	orig := &AgentConfig{DIS: types.DISConfig{JDBCConfig: &types.JDBCConfig{JavaPath: "/bin/java"}}}
	dup := copyConfig(orig)
	if dup == orig || dup.DIS.JDBCConfig == orig.DIS.JDBCConfig {
		t.Fatalf("expected deep copy of JDBC config")
	}
	orig.DIS.JDBCConfig.JavaPath = "changed"
	if dup.DIS.JDBCConfig.JavaPath != "/bin/java" {
		t.Fatalf("expected copy to retain original values, got %s", dup.DIS.JDBCConfig.JavaPath)
	}
}

func TestResultRowsToProto(t *testing.T) {
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	rows := []types.ResultRow{{
		"ts": ts,
		"n":  int64(5),
		"x":  struct{ A int }{A: 1},
	}}
	protoRows := resultRowsToProto(rows)
	if len(protoRows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(protoRows))
	}
	fields := protoRows[0].Fields
	if got := fields["ts"].GetNumberValue(); got != float64(ts.UnixMilli()) {
		t.Fatalf("expected millis for time, got %v", got)
	}
	if fields["n"].GetNumberValue() != 5 {
		t.Fatalf("expected numeric value preserved")
	}
	if fields["x"].Kind != (*structpb.Value_StringValue)(nil) && fields["x"].GetStringValue() == "" {
		t.Fatalf("expected fallback string value for struct, got %v", fields["x"])
	}
}

func TestApplyAgentConfigRuntimeRestart(t *testing.T) {
	origDir, _ := os.Getwd()
	tempDir := t.TempDir()
	_ = os.Chdir(tempDir)
	defer os.Chdir(origDir)

	base := slog.NewTextHandler(io.Discard, nil)
	var loggerVal atomic.Value
	loggerVal.Store(slog.New(base))

	agentCfg := &AgentConfig{
		DIS: types.DISConfig{JDBCConfig: &types.JDBCConfig{JavaPath: "java", JDBCPort: "8888"}},
	}
	rt := &bridgeproto.AgentRuntimeConfig{
		JdbcPort: "9999", JavaPath: "newjava", ClientSecret: "secret", ForceRestart: true,
	}
	cfg := &bridgeproto.AgentConfig{Runtime: rt}
	db := &stubDB{}

	err := applyAgentConfig(base, &loggerVal, cfg, agentCfg, db)
	if err != errRestartRequired {
		t.Fatalf("expected errRestartRequired, got %v", err)
	}
	if !db.startCalled || !db.stopCalled {
		t.Fatalf("expected JDBC runner restart, got start=%v stop=%v", db.startCalled, db.stopCalled)
	}
	if agentCfg.DIS.JDBCConfig.JDBCPort != "9999" || agentCfg.DIS.JDBCConfig.JavaPath != "newjava" || agentCfg.ClientSecret != "secret" {
		t.Fatalf("runtime values not applied to agent config: %+v", agentCfg.DIS.JDBCConfig)
	}
}

func TestApplyAgentConfigSkipsLokiWhenNil(t *testing.T) {
	origDir, _ := os.Getwd()
	tempDir := t.TempDir()
	_ = os.Chdir(tempDir)
	defer os.Chdir(origDir)

	base := slog.NewTextHandler(io.Discard, nil)
	var loggerVal atomic.Value
	loggerVal.Store(slog.New(base))

	agentCfg := &AgentConfig{DIS: types.DISConfig{JDBCConfig: &types.JDBCConfig{}}}
	if err := applyAgentConfig(base, &loggerVal, &bridgeproto.AgentConfig{}, agentCfg, &stubDB{}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, err := os.Stat(agentcore.DefaultAgentConfigFile); err != nil {
		t.Fatalf("expected config file written, err=%v", err)
	}
}

func TestNewLokiHandlerRequiresURL(t *testing.T) {
	if _, err := newLokiHandler(&bridgeproto.LokiConfig{Url: ""}); err == nil {
		t.Fatalf("expected error for empty url")
	}
}

type recordingHandler struct {
	enabled bool
	handled int
	attrs   []slog.Attr
	groups  []string
}

func (r *recordingHandler) Enabled(ctx context.Context, level slog.Level) bool { return r.enabled }
func (r *recordingHandler) Handle(ctx context.Context, record slog.Record) error {
	r.handled++
	return nil
}
func (r *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	r.attrs = append(r.attrs, attrs...)
	return r
}
func (r *recordingHandler) WithGroup(name string) slog.Handler {
	r.groups = append(r.groups, name)
	return r
}

func TestFanoutHandlerForwards(t *testing.T) {
	h1 := &recordingHandler{enabled: true}
	h2 := &recordingHandler{enabled: false}
	fan := fanoutHandler{handlers: []slog.Handler{h1, h2}}

	if !fan.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatalf("expected fanout enabled when any child is enabled")
	}

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	if err := fan.Handle(context.Background(), rec); err != nil {
		t.Fatalf("handle err: %v", err)
	}
	if h1.handled != 1 {
		t.Fatalf("expected first handler to handle record")
	}
	if h2.handled != 0 {
		t.Fatalf("expected disabled handler not to handle record")
	}

	withAttrs := fan.WithAttrs([]slog.Attr{slog.String("k", "v")}).(fanoutHandler)
	withGroup := withAttrs.WithGroup("g").(fanoutHandler)
	if len(h1.attrs) == 0 || len(h1.groups) == 0 || len(withGroup.handlers) != 2 {
		t.Fatalf("expected attrs/groups to propagate to handlers")
	}
}
