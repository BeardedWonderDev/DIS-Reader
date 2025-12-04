package bridge

import (
	"context"
	"errors"
)

// AgentAuthenticator validates an incoming agent hello and returns tenant and agent IDs.
// Implementations may look up secrets in a datastore or config map.
type AgentAuthenticator interface {
	Authenticate(ctx context.Context, clientID, clientSecret string, tenantID string) (string, string, error)
}

// ReloadableAuthenticator optionally supports live reload of credentials.
type ReloadableAuthenticator interface {
	AgentAuthenticator
	Reload() error
}

// StaticAuthenticator is a simple map-based authenticator for bootstrapping and tests.
type StaticAuthenticator struct {
	Secrets map[string]StaticAgentSecret
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

	secret, ok := a.Secrets[clientID]
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
