package agentcore

import (
	"os"
	"testing"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func TestLoadDefaultsAndEnvOverride(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/agent.yaml"
	if err := os.WriteFile(cfgPath, []byte("serverURL: https://example\nclientID: abc\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("DISAGENT_DIS_JDBCCONFIG_JAVAPATH", "custom-java")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if cfg.DIS.JDBCConfig.JavaPath != "custom-java" {
		t.Fatalf("expected env override java path, got %s", cfg.DIS.JDBCConfig.JavaPath)
	}
	if cfg.TLS.Enabled != true { // default
		t.Fatalf("expected TLS enabled default true")
	}
}

func TestLoadRequiresFields(t *testing.T) {
	dir := t.TempDir()
	cfgPath := dir + "/empty.yaml"
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := Load(cfgPath); err == nil {
		t.Fatalf("expected error for missing required fields")
	}
}

func TestSaveAndApplyRuntimeOverrides(t *testing.T) {
	cfg := &Config{ServerURL: "http://x", ClientID: "id", DIS: types.DISConfig{JDBCConfig: &types.JDBCConfig{JavaPath: "java", JDBCPort: "8888"}}}
	file := t.TempDir() + "/out.yaml"
	if err := Save(file, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("expected file saved: %v", err)
	}

	runtimeCfg := &bridgeproto.AgentRuntimeConfig{JdbcPort: "9999", JavaPath: "newjava", ClientSecret: "sec", ForceRestart: true}
	restart, reconnect := ApplyRuntimeOverrides(cfg, runtimeCfg)
	if !restart || !reconnect {
		t.Fatalf("expected restart and reconnect true, got %v %v", restart, reconnect)
	}
	if cfg.DIS.JDBCConfig.JDBCPort != "9999" || cfg.ClientSecret != "sec" {
		t.Fatalf("runtime overrides not applied")
	}
}

func TestToEffectiveSetsLocalProvenance(t *testing.T) {
	cfg := &Config{
		ServerURL: "https://example",
		ClientID:  "client-1",
		DIS:       types.DISConfig{Host: "h", JDBCConfig: &types.JDBCConfig{JavaPath: "java", JDBCPort: "8888"}},
		Control:   ControlConfig{Enabled: true, Addr: "127.0.0.1:7777"},
	}
	eff := ToEffective(cfg)
	if eff == nil || eff.Provenance == nil {
		t.Fatalf("expected provenance map")
	}
	checks := []string{"serverURL", "clientID", "dis.host", "dis.jdbcConfig.javaPath", "control.enabled", "tls.enabled"}
	for _, key := range checks {
		if eff.Provenance[key] != "local" {
			t.Fatalf("expected %s provenance local, got %q", key, eff.Provenance[key])
		}
	}
}

func TestMarkRuntimeProvenance(t *testing.T) {
	cfg := &Config{ServerURL: "https://example", ClientID: "cid", DIS: types.DISConfig{JDBCConfig: &types.JDBCConfig{}}}
	eff := ToEffective(cfg)
	rt := &bridgeproto.AgentRuntimeConfig{
		DisHost:     "remote-host",
		DisUser:     "user",
		DisPassword: "pw",
		JdbcPort:    "9999",
		JavaPath:    "/opt/java",
		TenantId:    "tenant-1",
		ClientSecret:"secret",
	}
	MarkRuntimeProvenance(eff, rt)
	expectRemote := []string{
		"dis.host", "dis.user", "dis.password",
		"dis.jdbcConfig.jdbcPort", "dis.jdbcConfig.javaPath",
		"tenantID", "clientSecret",
	}
	for _, key := range expectRemote {
		if eff.Provenance[key] != "remote" {
			t.Fatalf("expected %s to be marked remote, got %q", key, eff.Provenance[key])
		}
	}
	// Unchanged field should remain local
	if eff.Provenance["serverURL"] != "local" {
		t.Fatalf("expected serverURL to remain local, got %q", eff.Provenance["serverURL"])
	}
}
