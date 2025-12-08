package disreader

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRemoteBuilderSetters(t *testing.T) {
	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "remote"}}
	builder := NewDISReaderRemote(cfg, nil)
	auth := fakeAuth{}
	reg := fakeRegistry{}
	grpcServer := buildTestGRPCServer()
	mux := http.NewServeMux()

	builder.WithAuth(auth).WithRegistry(reg).WithGRPC(grpcServer).WithMux(mux).WithDefaultTenant("t1")
	if builder.auth != auth || builder.reg != reg || builder.grpcServer != grpcServer || builder.mux != mux || builder.defaultTenant != "t1" {
		t.Fatalf("setters did not populate builder fields")
	}
}

func TestRemoteBuilderRejectsNonRemoteMode(t *testing.T) {
	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "embedded"}}
	builder := NewDISReaderRemote(cfg, nil)
	if _, err := builder.Build(); err == nil {
		t.Fatalf("expected error when bridge mode is not remote")
	}
}

type stubHandler struct{ records []slog.Record }

func (h *stubHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *stubHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *stubHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *stubHandler) WithGroup(name string) slog.Handler       { return h }

func TestGrpcLoggingUnaryLogsErrorLevel(t *testing.T) {
	h := &stubHandler{}
	logger := slog.New(h)
	info := &grpc.UnaryServerInfo{FullMethod: "/demo.Method"}

	_, _ = grpcLoggingUnary(logger)(context.Background(), nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.InvalidArgument, "bad")
	})

	if len(h.records) == 0 {
		t.Fatalf("expected logs to be recorded")
	}
	last := h.records[len(h.records)-1]
	if last.Level != slog.LevelError {
		t.Fatalf("expected error level, got %v", last.Level)
	}
}

type stubStream struct{ grpc.ServerStream }

func (s stubStream) Context() context.Context { return context.Background() }

func TestGrpcLoggingStreamLogsInfo(t *testing.T) {
	h := &stubHandler{}
	logger := slog.New(h)
	info := &grpc.StreamServerInfo{FullMethod: "/demo.Stream", IsServerStream: true}

	_ = grpcLoggingStream(logger)(nil, stubStream{}, info, func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	})

	if len(h.records) == 0 || h.records[len(h.records)-1].Level != slog.LevelInfo {
		t.Fatalf("expected info level log for successful stream")
	}
}

// minimal fakes to satisfy interfaces without spinning servers
type fakeAuth struct{}

func (fakeAuth) Authenticate(ctx context.Context, clientID, clientSecret, tenantID string) (string, string, error) {
	return "", "", errors.New("nope")
}

type fakeRegistry struct{}

func (fakeRegistry) Register(ctx context.Context, tenantID, agentID string, conn types.AgentConnection) error {
	return nil
}
func (fakeRegistry) Unregister(ctx context.Context, tenantID, agentID string) {}
func (fakeRegistry) Pick(ctx context.Context, tenantID string) (types.AgentConnection, error) {
	return nil, errors.New("none")
}
func (fakeRegistry) Stats() types.RegistryStats                                   { return types.RegistryStats{} }
func (fakeRegistry) ObserveQuery(tenantID, agentID string, latency time.Duration) {}

// buildTestGRPCServer returns a non-nil placeholder; real server wiring is covered elsewhere.
func buildTestGRPCServer() *grpc.Server {
	return grpc.NewServer()
}

func TestBuildAgentConfigIncludesRuntimeAndLoki(t *testing.T) {
	disCfg := &types.DISConfig{
		Host:     "h1",
		User:     "u1",
		Password: "p1",
		JDBCConfig: &types.JDBCConfig{
			JDBCPort: "9999",
			JavaPath: "/usr/bin/java",
		},
		Bridge: &types.BridgeConfig{
			TenantID:     "tenant-x",
			ClientSecret: "secret-x",
		},
	}
	bridgeCfg := &types.BridgeConfig{
		Loki: &types.LokiConfig{
			URL:        "http://loki",
			TenantID:   "loki-tenant",
			APIKey:     "apikey",
			AuthHeader: "X-Auth",
			Labels:     map[string]string{"app": "dis"},
			MinLevel:   slog.LevelWarn,
		},
	}

	cfg := buildAgentConfig(disCfg, bridgeCfg)
	if cfg == nil || cfg.Runtime == nil || cfg.Loki == nil {
		t.Fatalf("expected runtime and loki to be populated, got %+v", cfg)
	}
	if cfg.Runtime.GetDisHost() != "h1" || cfg.Runtime.GetJdbcPort() != "9999" || cfg.Runtime.GetTenantId() != "tenant-x" {
		t.Fatalf("runtime fields not set correctly: %+v", cfg.Runtime)
	}
	if cfg.Loki.GetUrl() != "http://loki" || cfg.Loki.GetLabels()["app"] != "dis" || cfg.Loki.GetMinLevel() != "warn" {
		t.Fatalf("loki fields not set correctly: %+v", cfg.Loki)
	}
}

func TestBuildAgentConfigReturnsNilWhenEmpty(t *testing.T) {
	if cfg := buildAgentConfig(&types.DISConfig{}, &types.BridgeConfig{}); cfg != nil {
		t.Fatalf("expected nil config when no runtime or loki data, got %+v", cfg)
	}
}
