package bridge

import (
	"context"
	"errors"
	"sync"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// AgentAuthenticator validates an incoming agent hello and returns tenant and agent IDs.
// Implementations may look up secrets in a datastore or config map.
type AgentAuthenticator = types.AgentAuthenticator

// MutableAuthenticator allows in-memory secret updates (e.g., secret rotation without restart).
type MutableAuthenticator interface {
	AgentAuthenticator
	Upsert(clientID string, secret StaticAgentSecret)
}

// ReloadableAuthenticator optionally supports live reload of credentials.
type ReloadableAuthenticator interface {
	AgentAuthenticator
	Reload() error
}

// StaticAuthenticator is a simple map-based authenticator for bootstrapping and tests.
type StaticAuthenticator struct {
	Secrets map[string]StaticAgentSecret
	mu      sync.RWMutex
}

type StaticAgentSecret struct {
	ClientSecret string
	TenantID     string
	AgentID      string
}

func (a *StaticAuthenticator) Authenticate(ctx context.Context, clientID, clientSecret string, tenantID string) (string, string, error) {
	if a == nil {
		return "", "", errors.New("authenticator not configured")
	}

	a.mu.RLock()
	secret, ok := a.Secrets[clientID]
	a.mu.RUnlock()
	if !ok {
		return "", "", errors.New("invalid client id")
	}
	if secret.ClientSecret != clientSecret {
		return "", "", errors.New("invalid secret")
	}
	tID := secret.TenantID
	if tenantID != "" && tenantID != tID {
		return "", "", errors.New("tenant mismatch")
	}
	return tID, secret.AgentID, nil
}

// Upsert adds or replaces a credential in-memory; callers must persist if desired.
func (a *StaticAuthenticator) Upsert(clientID string, secret StaticAgentSecret) {
	if a == nil {
		return
	}
	if a.Secrets == nil {
		a.Secrets = map[string]StaticAgentSecret{}
	}
	a.mu.Lock()
	a.Secrets[clientID] = secret
	a.mu.Unlock()
}
