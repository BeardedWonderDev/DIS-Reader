package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// stubMultiTenantDB records the last tenant passed to lifecycle/health methods.
type stubMultiTenantDB struct {
	last string
}

func (s *stubMultiTenantDB) StartJDBCRunner(tenant string) error { s.last = tenant; return nil }
func (s *stubMultiTenantDB) StopJDBCRunner(tenant string) error  { s.last = tenant; return nil }
func (s *stubMultiTenantDB) Connect(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) Disconnect(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) PingService(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) PingDatabase(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) Get(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) Query(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	s.last = tenant
	return nil, nil
}
func (s *stubMultiTenantDB) QueryRow(ctx context.Context, query string, tenant string, args ...interface{}) (types.ResultRow, error) {
	s.last = tenant
	return nil, nil
}
func (s *stubMultiTenantDB) Select(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) QueryWithSource(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	s.last = tenant
	return nil, nil
}

func TestResolveTenantFallsBackToDefault(t *testing.T) {
	mt := &stubMultiTenantDB{}
	s := NewRemoteService(&types.DISConfig{}, nil, mt, bridge.NewInMemoryRegistry(), nil)
	s.WithDefaultTenant("tenant-default")

	if err := s.Connect(context.Background(), ""); err != nil {
		t.Fatalf("expected nil error with default tenant, got %v", err)
	}
	if mt.last != "tenant-default" {
		t.Fatalf("expected default tenant to be used, got %q", mt.last)
	}
}

func TestResolveTenantPrefersExplicit(t *testing.T) {
	mt := &stubMultiTenantDB{}
	s := NewRemoteService(&types.DISConfig{}, nil, mt, bridge.NewInMemoryRegistry(), nil)
	s.WithDefaultTenant("tenant-default")

	if err := s.Connect(context.Background(), "tenant-explicit"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if mt.last != "tenant-explicit" {
		t.Fatalf("expected explicit tenant to win, got %q", mt.last)
	}
}

func TestResolveTenantErrorWhenMissing(t *testing.T) {
	mt := &stubMultiTenantDB{}
	s := NewRemoteService(&types.DISConfig{}, nil, mt, bridge.NewInMemoryRegistry(), nil)

	if err := s.Connect(context.Background(), ""); err == nil {
		t.Fatalf("expected error when no tenant and no default")
	}
}

// fakeAgentConn implements AgentConnection for tests.
type fakeAgentConn struct {
	lastReq *proto.JobRequest
	res     *proto.JobResult
}

func (f *fakeAgentConn) TenantID() string { return "tenant-a" }
func (f *fakeAgentConn) AgentID() string  { return "agent-a" }
func (f *fakeAgentConn) SendJob(ctx context.Context, req *proto.JobRequest) (*proto.JobResult, error) {
	f.lastReq = req
	return f.res, nil
}
func (f *fakeAgentConn) Close() error { return nil }

// fakeRegistry is a minimal AgentRegistry used for tests.
type fakeRegistry struct {
	conn  bridge.AgentConnection
	stats bridge.RegistryStats
}

func (r *fakeRegistry) Register(ctx context.Context, tenantID string, agentID string, conn bridge.AgentConnection) error {
	r.conn = conn
	return nil
}

func (r *fakeRegistry) Unregister(ctx context.Context, tenantID string, agentID string) {}

func (r *fakeRegistry) Pick(ctx context.Context, tenantID string) (bridge.AgentConnection, error) {
	if r.conn == nil {
		return nil, fmt.Errorf("no agent")
	}
	return r.conn, nil
}

func (r *fakeRegistry) Stats() bridge.RegistryStats { return r.stats }

func (r *fakeRegistry) ObserveQuery(tenantID, agentID string, latency time.Duration) {}

func TestPingAgentErrorsWhenNoAgent(t *testing.T) {
	reg := &fakeRegistry{}
	s := NewRemoteService(&types.DISConfig{}, nil, nil, reg, nil)
	s.WithDefaultTenant("tenant-x")
	if err := s.PingAgent(context.Background(), ""); err == nil {
		t.Fatalf("expected error when no agent is registered")
	}
}

func TestRegisterHealthRegistersHealthMetricsAndPprof(t *testing.T) {
	reg := &fakeRegistry{stats: bridge.RegistryStats{TotalAgents: 1, Tenants: map[string]int{"t1": 1}}}
	mux := http.NewServeMux()
	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{PprofEnabled: true, PprofPath: "/pprof/"}}
	s := NewRemoteService(cfg, nil, nil, reg, nil)
	s.RegisterHealth(mux)

	// healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from healthz, got %d", rec.Code)
	}
	// metrics
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from metrics, got %d", rec.Code)
	}
	// pprof index under custom path
	req = httptest.NewRequest(http.MethodGet, "/pprof/", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code == 0 {
		t.Fatalf("expected pprof handler to be registered")
	}
}

func TestReadAgentConfig(t *testing.T) {
	mt := &stubMultiTenantDB{}
	agent := &fakeAgentConn{
		res: &proto.JobResult{
			Status: proto.Status_STATUS_OK,
			ConfigStatus: &proto.AgentConfigStatus{
				Runtime: &proto.AgentRuntimeStatus{DisHost: "host-a"},
			},
		},
	}
	reg := &fakeRegistry{conn: agent}

	s := NewRemoteService(&types.DISConfig{}, nil, mt, reg, nil)
	s.WithDefaultTenant("tenant-a")

	status, err := s.ReadAgentConfig(context.Background(), "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if status.Runtime.DisHost != "host-a" {
		t.Fatalf("expected runtime host to match, got %+v", status.Runtime)
	}
	if agent.lastReq == nil || agent.lastReq.Kind != proto.JobKind_JOB_KIND_READ_CONFIG {
		t.Fatalf("expected read config job to be sent")
	}
}

func TestRegisterPprofRegistersHandlers(t *testing.T) {
	mux := http.NewServeMux()
	registerPprof(mux, "")

	paths := []string{"", "cmdline", "profile", "symbol", "trace"}
	for _, p := range paths {
		path := "/debug/pprof/" + p
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h, _ := mux.Handler(req)
		if h == nil {
			t.Fatalf("expected handler for %s", path)
		}
		h.ServeHTTP(rec, req)
		if rec.Code == 0 {
			t.Fatalf("handler for %s did not run", path)
		}
	}
}
