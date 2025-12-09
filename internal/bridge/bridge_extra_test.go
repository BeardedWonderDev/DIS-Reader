package bridge

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
)

// fakeConn implements AgentConnection minimally for tests.
type fakeConn string

func (f fakeConn) AgentID() string { return string(f) }
func (f fakeConn) TenantID() string { return "t1" }
func (f fakeConn) SendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	return &bridgeproto.JobResult{}, nil
}
func (f fakeConn) Close() error { return nil }

func TestStaticAuthenticator_UpsertAndErrors(t *testing.T) {
	auth := &StaticAuthenticator{Secrets: map[string]StaticAgentSecret{
		"id1": {ClientSecret: "sec1", TenantID: "t1", AgentID: "a1"},
	}}

	tenant, agent, err := auth.Authenticate(context.Background(), "id1", "sec1", "")
	if err != nil || tenant != "t1" || agent != "a1" {
		t.Fatalf("expected success, got tenant=%s agent=%s err=%v", tenant, agent, err)
	}

	if _, _, err := auth.Authenticate(context.Background(), "missing", "sec1", ""); err == nil {
		t.Fatalf("expected invalid client id error")
	}
	if _, _, err := auth.Authenticate(context.Background(), "id1", "wrong", ""); err == nil {
		t.Fatalf("expected invalid secret error")
	}
	if _, _, err := auth.Authenticate(context.Background(), "id1", "sec1", "other"); err == nil {
		t.Fatalf("expected tenant mismatch error")
	}

	auth = &StaticAuthenticator{}
	auth.Upsert("id2", StaticAgentSecret{ClientSecret: "sec2", TenantID: "t2", AgentID: "a2"})
	tenant, agent, err = auth.Authenticate(context.Background(), "id2", "sec2", "t2")
	if err != nil || tenant != "t2" || agent != "a2" {
		t.Fatalf("upserted secret failed: tenant=%s agent=%s err=%v", tenant, agent, err)
	}
}

func TestFileAuthenticator_ReloadAndErrors(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "agents.yaml")
	yaml := `
- clientID: c1
  clientSecret: s1
  tenantID: t1
  agentID: a1
`
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	fa, err := NewFileAuthenticator(path)
	if err != nil {
		t.Fatalf("NewFileAuthenticator error: %v", err)
	}

	tenant, agent, err := fa.Authenticate(context.Background(), "c1", "s1", "")
	if err != nil || tenant != "t1" || agent != "a1" {
		t.Fatalf("authenticate failed: tenant=%s agent=%s err=%v", tenant, agent, err)
	}
	if _, _, err := fa.Authenticate(context.Background(), "c1", "wrong", ""); err == nil {
		t.Fatalf("expected invalid secret error")
	}
	if _, _, err := fa.Authenticate(context.Background(), "c1", "s1", "other"); err == nil {
		t.Fatalf("expected tenant mismatch")
	}

	fa.Upsert("c2", StaticAgentSecret{ClientSecret: "s2", TenantID: "t2", AgentID: "a2"})
	tenant, agent, err = fa.Authenticate(context.Background(), "c2", "s2", "t2")
	if err != nil || tenant != "t2" || agent != "a2" {
		t.Fatalf("upsert auth failed: tenant=%s agent=%s err=%v", tenant, agent, err)
	}

	// Reload error path
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove yaml: %v", err)
	}
	if err := fa.Reload(); err == nil {
		t.Fatalf("expected reload error after file removal")
	}
}

func TestInMemoryRegistry_StatsAndObserve(t *testing.T) {
	reg := NewInMemoryRegistry()

	// Register two agents for one tenant, one for another
	reg.Register(context.Background(), "t1", "a1", fakeConn("a1"))
	reg.Register(context.Background(), "t1", "a2", fakeConn("a2"))
	reg.Register(context.Background(), "t2", "a3", fakeConn("a3"))

	if _, err := reg.Pick(context.Background(), "missing"); err == nil {
		t.Fatalf("expected pick error for missing tenant")
	}
	conn, err := reg.Pick(context.Background(), "t1")
	if err != nil || conn == nil {
		t.Fatalf("pick failed: %v", err)
	}

	reg.Unregister(context.Background(), "t1", "a1")
	reg.Unregister(context.Background(), "t1", "a2") // removes tenant map
	reg.Unregister(context.Background(), "t2", "a3")

	stats := reg.Stats()
	if stats.TotalAgents != 0 || len(stats.Tenants) != 0 {
		t.Fatalf("expected empty stats, got %+v", stats)
	}

	reg.Register(context.Background(), "t1", "a1", fakeConn("a1"))
	reg.ObserveQuery("t1", "a1", 200*time.Millisecond)
	reg.ObserveQuery("t1", "a1", 100*time.Millisecond)
	stats = reg.Stats()
	if stats.QueryCounts[RegistryKey{Tenant: "t1", Agent: "a1"}] != 2 {
		t.Fatalf("query count mismatch: %+v", stats.QueryCounts)
	}
	lat := stats.Latency[RegistryKey{Tenant: "t1", Agent: "a1"}]
	if lat.Count != 2 || lat.Max <= 0 || lat.Sum <= 0 {
		t.Fatalf("latency agg missing: %+v", lat)
	}
}

func TestHealthHandler_Readiness(t *testing.T) {
	reg := NewInMemoryRegistry()
	ts := httptest.NewServer(HealthHandler(reg))
	defer ts.Close()

	resp, _ := http.Get(ts.URL)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when no agents, got %d", resp.StatusCode)
	}

	reg.Register(context.Background(), "t1", "a1", fakeConn("a1"))
	resp, _ = http.Get(ts.URL)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 when agents present, got %d", resp.StatusCode)
	}
}

func TestMetricsHandler_Outputs(t *testing.T) {
	reg := NewInMemoryRegistry()
	reg.Register(context.Background(), "t1", "a1", fakeConn("a1"))
	reg.ObserveQuery("t1", "a1", 150*time.Millisecond)

	ts := httptest.NewServer(MetricsHandler(reg))
	defer ts.Close()

	body, _ := http.Get(ts.URL)
	defer body.Body.Close()
	data, _ := io.ReadAll(body.Body)
	text := string(data)
	if !strings.Contains(text, "bridge_agents_total 1") {
		t.Fatalf("missing agent gauge: %s", text)
	}
	if !strings.Contains(text, `bridge_agents_per_tenant{tenant="t1"} 1`) {
		t.Fatalf("missing per-tenant gauge: %s", text)
	}
	if !strings.Contains(text, `bridge_queries_total{tenant="t1",agent="a1"} 1`) {
		t.Fatalf("missing query counter: %s", text)
	}
	if !strings.Contains(text, `bridge_query_latency_seconds_count{tenant="t1",agent="a1"} 1`) {
		t.Fatalf("missing latency count: %s", text)
	}
}
