package main

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

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
