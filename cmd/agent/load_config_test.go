package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigUsesFileAndEnvDefaults(t *testing.T) {
	tmp := t.TempDir()
	// create minimal config file
	cfgPath := filepath.Join(tmp, "agent.yaml")
	if err := os.WriteFile(cfgPath, []byte("serverURL: https://svc\nclientID: cid\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("DISAGENT_CLIENTSECRET", "secret-env")
	t.Setenv("DISAGENT_DIS_JDBCCONFIG_JAVAPATH", "/custom/java")

	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	_ = os.Chdir(tmp)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if cfg.ServerURL != "https://svc" || cfg.ClientID != "cid" {
		t.Fatalf("expected values from file, got %+v", cfg)
	}
	if cfg.ClientSecret != "secret-env" {
		t.Fatalf("expected env client secret override, got %s", cfg.ClientSecret)
	}
	if cfg.DIS.JDBCConfig.JavaPath != "/custom/java" {
		t.Fatalf("expected env javaPath override, got %s", cfg.DIS.JDBCConfig.JavaPath)
	}
}
