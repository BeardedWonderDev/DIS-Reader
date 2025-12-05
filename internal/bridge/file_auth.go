package bridge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// FileAuthenticator reads agent credentials from a YAML/JSON file.
// Schema: list of {clientID, clientSecret, tenantID, agentID}
type FileAuthenticator struct {
	path    string
	mu      sync.RWMutex
	secrets map[string]StaticAgentSecret
}

type fileAgent struct {
	ClientID     string `yaml:"clientID" json:"clientID"`
	ClientSecret string `yaml:"clientSecret" json:"clientSecret"`
	TenantID     string `yaml:"tenantID" json:"tenantID"`
	AgentID      string `yaml:"agentID" json:"agentID"`
}

func NewFileAuthenticator(path string) (*FileAuthenticator, error) {
	f := &FileAuthenticator{path: path, secrets: map[string]StaticAgentSecret{}}
	if err := f.Reload(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *FileAuthenticator) Authenticate(_ context.Context, clientID, clientSecret string, tenantID string) (string, string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	secret, ok := f.secrets[clientID]
	if !ok {
		return "", "", errors.New("invalid client id")
	}
	if secret.ClientSecret != clientSecret {
		return "", "", errors.New("invalid secret")
	}
	if tenantID != "" && tenantID != secret.TenantID {
		return "", "", errors.New("tenant mismatch")
	}
	return secret.TenantID, secret.AgentID, nil
}

// Upsert updates the in-memory secret map (not persisted to disk). Useful for
// live rotations coordinated by the bridge controller.
func (f *FileAuthenticator) Upsert(clientID string, secret StaticAgentSecret) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.secrets == nil {
		f.secrets = map[string]StaticAgentSecret{}
	}
	f.secrets[clientID] = secret
}

func (f *FileAuthenticator) Reload() error {
	b, err := os.ReadFile(filepath.Clean(f.path))
	if err != nil {
		return err
	}
	var agents []fileAgent
	if err := yaml.Unmarshal(b, &agents); err != nil {
		return err
	}
	m := make(map[string]StaticAgentSecret, len(agents))
	for _, a := range agents {
		m[a.ClientID] = StaticAgentSecret{
			ClientSecret: a.ClientSecret,
			TenantID:     a.TenantID,
			AgentID:      a.AgentID,
		}
	}
	f.mu.Lock()
	f.secrets = m
	f.mu.Unlock()
	return nil
}

// Ensure FileAuthenticator satisfies ReloadableAuthenticator.
var _ ReloadableAuthenticator = (*FileAuthenticator)(nil)
