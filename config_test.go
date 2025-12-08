package main

import (
	"reflect"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/spf13/viper"
)

// Ensure environment variables override defaults and are mapped using flattened keys.
func TestNewConfigUsesEnvOverrides(t *testing.T) {
	t.Setenv("DISREADER_DISCONFIG_HOST", "env-host.example")
	t.Setenv("DISREADER_DISCONFIG_JDBCCONFIG_JAVAPATH", "/custom/java")
	t.Setenv("DISREADER_DEBUGSEARCH_DEFAULTOUTPUTMODE", "csv")

	viper.Reset()
	t.Cleanup(viper.Reset)

	cfg := NewConfig()

	if cfg.DIS == nil {
		t.Fatalf("expected DIS config to be populated")
	}
	if cfg.DIS.Host != "env-host.example" {
		t.Fatalf("expected host from env, got %q", cfg.DIS.Host)
	}
	if cfg.DIS.JDBCConfig == nil || cfg.DIS.JDBCConfig.JavaPath != "/custom/java" {
		t.Fatalf("expected JDBC javaPath from env, got %#v", cfg.DIS.JDBCConfig)
	}
	if cfg.DebugSearch.DefaultOutputMode != "csv" {
		t.Fatalf("expected debug search mode from env, got %q", cfg.DebugSearch.DefaultOutputMode)
	}
	if cfg.Bridge == nil || cfg.Bridge.Mode != "embedded" {
		t.Fatalf("expected default bridge mode 'embedded', got %#v", cfg.Bridge)
	}
}

// Verify nested mapstructure tags are flattened correctly for env lookups.
func TestGetFlattenedStructFields(t *testing.T) {
	fields := getFlattenedStructFields(reflect.TypeOf(types.DISUIConfig{}))

	contains := func(target string) bool {
		for _, f := range fields {
			if f == target {
				return true
			}
		}
		return false
	}

	expected := []string{
		"appName",
		"disConfig.host",
		"disConfig.jdbcConfig.javaPath",
		"bridge.mode",
		"debugSearch.defaultOutputMode",
	}
	for _, key := range expected {
		if !contains(key) {
			t.Fatalf("expected flattened keys to include %q; got %v", key, fields)
		}
	}
}
