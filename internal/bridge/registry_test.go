package bridge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
)

type stubConn struct {
	tenant string
	agent  string
}

func (s stubConn) TenantID() string { return s.tenant }
func (s stubConn) AgentID() string  { return s.agent }
func (s stubConn) SendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	return nil, nil
}
func (s stubConn) Close() error { return nil }

func TestRegistryRegisterPickStats(t *testing.T) {
	reg := NewInMemoryRegistry()
	conn1 := stubConn{tenant: "t1", agent: "a1"}
	if err := reg.Register(context.Background(), "t1", "a1", conn1); err != nil {
		t.Fatalf("register err: %v", err)
	}
	if err := reg.Register(context.Background(), "t1", "a2", stubConn{tenant: "t1", agent: "a2"}); err != nil {
		t.Fatalf("register err: %v", err)
	}
	if _, err := reg.Pick(context.Background(), "t1"); err != nil {
		t.Fatalf("pick err: %v", err)
	}
	reg.ObserveQuery("t1", "a1", 50*time.Millisecond)
	stats := reg.Stats()
	if stats.TotalAgents != 2 || stats.Tenants["t1"] != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if stats.QueryCounts[RegistryKey{Tenant: "t1", Agent: "a1"}] != 1 {
		t.Fatalf("expected query count recorded")
	}
	reg.Unregister(context.Background(), "t1", "a1")
	stats = reg.Stats()
	if stats.TotalAgents != 1 {
		t.Fatalf("expected agent removed")
	}
	if _, err := reg.Pick(context.Background(), "missing"); err == nil {
		t.Fatalf("expected error for missing tenant")
	}
}

func TestHealthAndMetricsHandlers(t *testing.T) {
	reg := NewInMemoryRegistry()
	handler := HealthHandler(reg)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/healthz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no agents, got %d", rr.Code)
	}

	// add agent and test ready
	_ = reg.Register(context.Background(), "t1", "a1", stubConn{tenant: "t1", agent: "a1"})
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when agents present, got %d", rr.Code)
	}

	metrics := MetricsHandler(reg)
	rr = httptest.NewRecorder()
	metrics.ServeHTTP(rr, httptest.NewRequest("GET", "/metrics", nil))
	if rr.Code != http.StatusOK || len(rr.Body.String()) == 0 {
		t.Fatalf("expected metrics output, code=%d body=%q", rr.Code, rr.Body.String())
	}
}
