package bridge

import (
	"context"
	"os"
	"testing"
)

func TestStaticAuthenticator(t *testing.T) {
	auth := &StaticAuthenticator{Secrets: map[string]StaticAgentSecret{
		"id1": {ClientSecret: "secret", TenantID: "t1", AgentID: "a1"},
	}}
	tenant, agent, err := auth.Authenticate(context.Background(), "id1", "secret", "t1")
	if err != nil || tenant != "t1" || agent != "a1" {
		t.Fatalf("expected success, got tenant=%s agent=%s err=%v", tenant, agent, err)
	}

	if _, _, err := auth.Authenticate(context.Background(), "id1", "bad", "t1"); err == nil {
		t.Fatalf("expected secret mismatch error")
	}
	if _, _, err := auth.Authenticate(context.Background(), "id1", "secret", "other"); err == nil {
		t.Fatalf("expected tenant mismatch error")
	}

	auth.Upsert("id2", StaticAgentSecret{ClientSecret: "s2", TenantID: "t2", AgentID: "a2"})
	if _, _, err := auth.Authenticate(context.Background(), "id2", "s2", "t2"); err != nil {
		t.Fatalf("expected upserted secret to work: %v", err)
	}
}

func TestFileAuthenticator(t *testing.T) {
	yaml := `- clientID: c1
  clientSecret: s1
  tenantID: t1
  agentID: a1
- clientID: c2
  clientSecret: s2
  tenantID: t2
  agentID: a2
`
	tmp, err := os.CreateTemp("", "auth-*.yaml")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(yaml); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	tmp.Close()

	fa, err := NewFileAuthenticator(tmp.Name())
	if err != nil {
		t.Fatalf("NewFileAuthenticator: %v", err)
	}
	tenant, agent, err := fa.Authenticate(context.Background(), "c1", "s1", "t1")
	if err != nil || tenant != "t1" || agent != "a1" {
		t.Fatalf("expected success, got tenant=%s agent=%s err=%v", tenant, agent, err)
	}
	if _, _, err := fa.Authenticate(context.Background(), "c1", "bad", "t1"); err == nil {
		t.Fatalf("expected secret mismatch error")
	}

	// Upsert should allow new secret without touching file
	fa.Upsert("c3", StaticAgentSecret{ClientSecret: "s3", TenantID: "t3", AgentID: "a3"})
	if _, _, err := fa.Authenticate(context.Background(), "c3", "s3", "t3"); err != nil {
		t.Fatalf("expected upserted secret to work: %v", err)
	}

	// Reload should restore original set (drops c3)
	if err := fa.Reload(); err != nil {
		t.Fatalf("reload: %v", err)
	}
	if _, _, err := fa.Authenticate(context.Background(), "c3", "s3", "t3"); err == nil {
		t.Fatalf("expected c3 to be absent after reload")
	}
}
