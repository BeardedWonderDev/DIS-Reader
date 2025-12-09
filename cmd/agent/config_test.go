package main

import (
	"os"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/internal/agentcore"
)

func TestLoadConfigWithEnv(t *testing.T) {
	env := map[string]string{
		"DISAGENT_SERVERURL":               "bridge.example.com:443",
		"DISAGENT_CLIENTID":                "agent-123",
		"DISAGENT_CLIENTSECRET":            "super-secret",
		"DISAGENT_TENANTID":                "tenant-a",
		"DISAGENT_DIS_HOST":                "10.0.0.5",
		"DISAGENT_DIS_USER":                "DISUSER",
		"DISAGENT_DIS_PASSWORD":            "password",
		"DISAGENT_DIS_JDBCCONFIG_JAVAPATH": "/usr/bin/java",
		"DISAGENT_DIS_JDBCCONFIG_JDBCPORT": "9999",
		"DISAGENT_TLS_ENABLED":             "false",
		"DISAGENT_TLS_INSECURESKIPVERIFY":  "true",
	}

	original := make(map[string]string, len(env))
	for k, v := range env {
		original[k] = os.Getenv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("set env %s: %v", k, err)
		}
	}
	defer func() {
		for k, v := range original {
			if v == "" {
				_ = os.Unsetenv(k)
				continue
			}
			_ = os.Setenv(k, v)
		}
	}()

	cfg, err := agentcore.Load(agentcore.DefaultAgentConfigFile)
	if err != nil {
		t.Fatalf("loadConfigWith returned error: %v", err)
	}

	if cfg.ServerURL != env["DISAGENT_SERVERURL"] {
		t.Fatalf("serverURL = %s, want %s", cfg.ServerURL, env["DISAGENT_SERVERURL"])
	}
	if cfg.ClientID != env["DISAGENT_CLIENTID"] {
		t.Fatalf("clientID = %s, want %s", cfg.ClientID, env["DISAGENT_CLIENTID"])
	}
	if cfg.ClientSecret != env["DISAGENT_CLIENTSECRET"] {
		t.Fatalf("clientSecret = %s, want %s", cfg.ClientSecret, env["DISAGENT_CLIENTSECRET"])
	}
	if cfg.TenantID != env["DISAGENT_TENANTID"] {
		t.Fatalf("tenantID = %s, want %s", cfg.TenantID, env["DISAGENT_TENANTID"])
	}

	if cfg.DIS.Host != env["DISAGENT_DIS_HOST"] || cfg.DIS.User != env["DISAGENT_DIS_USER"] || cfg.DIS.Password != env["DISAGENT_DIS_PASSWORD"] {
		t.Fatalf("DIS config not populated from env: %+v", cfg.DIS)
	}

	if cfg.DIS.JDBCConfig == nil {
		t.Fatalf("JDBCConfig is nil")
	}
	if cfg.DIS.JDBCConfig.JavaPath != env["DISAGENT_DIS_JDBCCONFIG_JAVAPATH"] {
		t.Fatalf("JavaPath = %s, want %s", cfg.DIS.JDBCConfig.JavaPath, env["DISAGENT_DIS_JDBCCONFIG_JAVAPATH"])
	}
	if cfg.DIS.JDBCConfig.JDBCPort != env["DISAGENT_DIS_JDBCCONFIG_JDBCPORT"] {
		t.Fatalf("JDBCPort = %s, want %s", cfg.DIS.JDBCConfig.JDBCPort, env["DISAGENT_DIS_JDBCCONFIG_JDBCPORT"])
	}
	if cfg.TLS.Enabled {
		t.Fatalf("expected tls.enabled to be false")
	}
	if !cfg.TLS.InsecureSkipVerify {
		t.Fatalf("expected tls.insecureSkipVerify to be true")
	}
}
