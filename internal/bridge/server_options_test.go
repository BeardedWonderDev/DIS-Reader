package bridge

import (
	"context"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
)

// dummy authenticator implements AgentAuthenticator but does nothing.
type dummyAuth struct{}

func (dummyAuth) Authenticate(ctx context.Context, clientID, clientSecret, tenantID string) (string, string, error) {
	return tenantID, clientID, nil
}

// dummy registry implements AgentRegistry but stores nothing.
type dummyRegistry struct{}

func (dummyRegistry) Register(ctx context.Context, tenantID string, agentID string, conn AgentConnection) error {
	return nil
}
func (dummyRegistry) Unregister(ctx context.Context, tenantID string, agentID string) {}
func (dummyRegistry) Pick(ctx context.Context, tenantID string) (AgentConnection, error) {
	return nil, nil
}
func (dummyRegistry) Stats() RegistryStats                                         { return RegistryStats{} }
func (dummyRegistry) ObserveQuery(tenantID, agentID string, latency time.Duration) {}

func TestWithAutoConnectOnRegisterOption(t *testing.T) {
	s := NewServer(dummyAuth{}, dummyRegistry{}, nil, WithAutoConnectOnRegister(false))
	if s.autoConnectOnRegister {
		t.Fatalf("expected autoConnectOnRegister to be disabled")
	}
}

func TestWithAgentConfigOption(t *testing.T) {
	cfg := &proto.AgentConfig{}
	s := NewServer(dummyAuth{}, dummyRegistry{}, nil, WithAgentConfig(cfg))
	if s.agentConfig != cfg {
		t.Fatalf("expected agent config to be set on server")
	}
}

func TestSetAgentConfigNilNoBroadcast(t *testing.T) {
	s := NewServer(dummyAuth{}, dummyRegistry{}, nil)
	if sent := s.SetAgentConfig(nil, true, "", ""); sent != 0 {
		t.Fatalf("expected zero when cfg is nil, got %d", sent)
	}
}
