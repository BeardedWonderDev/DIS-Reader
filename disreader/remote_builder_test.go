package disreader

import (
	"context"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Ensures Build succeeds and auto-starts listeners when none are provided.
func TestRemoteBuilderAutoServers(t *testing.T) {
	t.Setenv("DISREADER_BRIDGE_PORT", "0")
	t.Setenv("DISREADER_BRIDGE_HTTP_PORT", "0")

	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "remote", MaxRowsPerQuery: 1000}}

	remote, err := NewDISReaderRemote(cfg, nil).Build()
	if err != nil {
		t.Fatalf("build remote: %v", err)
	}
	if remote == nil {
		t.Fatalf("expected remote service")
	}
}

func TestRemoteBuilderValidatesConfig(t *testing.T) {
	if _, err := NewDISReaderRemote(nil, nil).Build(); err == nil {
		t.Fatalf("expected error for nil config")
	}
	if _, err := NewDISReaderRemote(&types.DISConfig{}, nil).Build(); err == nil {
		t.Fatalf("expected error for missing bridge config")
	}
	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "embedded"}}
	if _, err := NewDISReaderRemote(cfg, nil).Build(); err == nil {
		t.Fatalf("expected error for non-remote mode")
	}
}

type stubAgentConn struct{}

func (s *stubAgentConn) TenantID() string                                      { return "t" }
func (s *stubAgentConn) AgentID() string                                       { return "a" }
func (s *stubAgentConn) SendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	return &bridgeproto.JobResult{Status: bridgeproto.Status_STATUS_OK}, nil
}
func (s *stubAgentConn) Close() error { return nil }

type stubRegistry struct{ lastTenant string }

func (r *stubRegistry) Register(ctx context.Context, tenantID string, agentID string, conn bridge.AgentConnection) error {
	return nil
}
func (r *stubRegistry) Unregister(ctx context.Context, tenantID string, agentID string) {}
func (r *stubRegistry) Pick(ctx context.Context, tenantID string) (bridge.AgentConnection, error) {
	r.lastTenant = tenantID
	return &stubAgentConn{}, nil
}
func (r *stubRegistry) Stats() bridge.RegistryStats { return bridge.RegistryStats{} }
func (r *stubRegistry) ObserveQuery(tenantID, agentID string, latency time.Duration)    {}

func TestRemoteBuilderDefaultTenantUsed(t *testing.T) {
	t.Setenv("DISREADER_BRIDGE_PORT", "0")
	t.Setenv("DISREADER_BRIDGE_HTTP_PORT", "0")

	reg := &stubRegistry{}
	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "remote", DefaultTenant: "default-tenant"}}

	remote, err := NewDISReaderRemote(cfg, nil).
		WithRegistry(reg).
		Build()
	if err != nil {
		t.Fatalf("build remote: %v", err)
	}

	if err := remote.PingAgent(context.Background(), ""); err != nil {
		t.Fatalf("expected default tenant to satisfy empty input: %v", err)
	}
	if reg.lastTenant != "default-tenant" {
		t.Fatalf("expected registry pick to use default tenant, got %q", reg.lastTenant)
	}
}
