package bridge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// FileAuthenticator reads agent credentials from a YAML/JSON file.
// Schema: list of {clientID, clientSecret, tenantID, agentID}
type FileAuthenticator struct {
	path    string
	mu      sync.RWMutex
	secrets map[string]StaticAgentSecret
	stopCh  chan struct{}
}

type fileAgent struct {
	ClientID     string `yaml:"clientID" json:"clientID"`
	ClientSecret string `yaml:"clientSecret" json:"clientSecret"`
	TenantID     string `yaml:"tenantID" json:"tenantID"`
	AgentID      string `yaml:"agentID" json:"agentID"`
}

func NewFileAuthenticator(path string) (*FileAuthenticator, error) {
	f := &FileAuthenticator{path: path, secrets: map[string]StaticAgentSecret{}, stopCh: make(chan struct{})}
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

// StartAutoReload triggers periodic reload. Pass interval <=0 to skip.
func (f *FileAuthenticator) StartAutoReload(interval time.Duration) {
	if interval <= 0 {
		return
	}
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_ = f.Reload()
			case <-f.stopCh:
				return
			}
		}
	}()
}

func (f *FileAuthenticator) StopAutoReload() {
	select {
	case <-f.stopCh:
		return
	default:
		close(f.stopCh)
	}
}
