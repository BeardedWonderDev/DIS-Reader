package agentcore

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"strings"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	DefaultAgentConfigFile = "agent.yaml"
	defaultJavaPath        = "java"
	defaultJDBCPort        = "8888"
	agentEnvPrefix         = "disagent"
)

// Config is the shared agent configuration (local + effective).
type Config struct {
	ServerURL          string          `mapstructure:"serverURL" yaml:"serverURL"`
	ClientID           string          `mapstructure:"clientID" yaml:"clientID"`
	ClientSecret       string          `mapstructure:"clientSecret" yaml:"clientSecret"`
	TenantID           string          `mapstructure:"tenantID" yaml:"tenantID"`
	AutoConnectOnStart bool            `mapstructure:"autoConnectOnStart" yaml:"autoConnectOnStart"`
	DIS                types.DISConfig `mapstructure:"dis" yaml:"dis"`
	TLS                struct {
		Enabled            bool `mapstructure:"enabled" yaml:"enabled"`
		InsecureSkipVerify bool `mapstructure:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	} `mapstructure:"tls" yaml:"tls"`
	Control     ControlConfig           `mapstructure:"control" yaml:"control"`
	AppliedLoki *bridgeproto.LokiConfig `mapstructure:"-" yaml:"-"`
}

// ControlConfig governs the local control API.
type ControlConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Addr    string `mapstructure:"addr" yaml:"addr"`
	Token   string `mapstructure:"token" yaml:"token"`
}

// EffectiveConfig pairs the config with provenance metadata.
type EffectiveConfig struct {
	Config     *Config           `json:"config"`
	Provenance map[string]string `json:"provenance"`
}

// Load reads config file + env overrides and returns a Config.
func Load(path string) (*Config, error) {
	v := viper.New()
	return loadWith(v, path)
}

func loadWith(v *viper.Viper, path string) (*Config, error) {
	if path == "" {
		path = DefaultAgentConfigFile
	}
	v.SetConfigFile(path)
	v.SetDefault("dis.logLevel", slog.LevelInfo)
	v.SetDefault("dis.jdbcConfig.javaPath", defaultJavaPath)
	v.SetDefault("dis.jdbcConfig.jdbcPort", defaultJDBCPort)
	v.SetDefault("autoConnectOnStart", false)
	v.SetDefault("tls.enabled", true)
	v.SetDefault("tls.insecureSkipVerify", false)
	v.SetDefault("control.enabled", false)
	v.SetDefault("control.addr", "127.0.0.1:7777")
	v.SetDefault("control.token", "")

	if _, err := os.ReadFile(path); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
	} else {
		if os.IsNotExist(err) {
			log.Printf("Could not find %s. Attempting to use environment variables.\n", path)
		} else {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	for _, fieldName := range flattenedStructFields(reflect.TypeOf(cfg)) {
		envKey := strings.ToUpper(fmt.Sprintf("%s_%s", agentEnvPrefix, strings.ReplaceAll(fieldName, ".", "_")))
		if envVar, ok := os.LookupEnv(envKey); ok {
			v.Set(fieldName, envVar)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.DIS.JDBCConfig == nil {
		cfg.DIS.JDBCConfig = &types.JDBCConfig{JavaPath: defaultJavaPath, JDBCPort: defaultJDBCPort}
	}

	if cfg.ServerURL == "" {
		return nil, errors.New("serverURL is required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("clientID is required")
	}

	return &cfg, nil
}

// Save writes the provided config to the given path as YAML.
func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultAgentConfigFile
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, b, 0o600)
}

// ApplyRuntimeOverrides mutates the config with runtime values and reports whether restart/reconnect are needed.
func ApplyRuntimeOverrides(cfg *Config, rt *bridgeproto.AgentRuntimeConfig) (restart bool, reconnect bool) {
	if cfg == nil || rt == nil {
		return false, false
	}
	trim := func(s string) string { return strings.TrimSpace(s) }
	if h := trim(rt.GetDisHost()); h != "" {
		cfg.DIS.Host = h
	}
	if u := trim(rt.GetDisUser()); u != "" {
		cfg.DIS.User = u
	}
	if p := trim(rt.GetDisPassword()); p != "" {
		cfg.DIS.Password = p
	}
	if jp := trim(rt.GetJdbcPort()); jp != "" {
		if cfg.DIS.JDBCConfig == nil {
			cfg.DIS.JDBCConfig = &types.JDBCConfig{}
		}
		if cfg.DIS.JDBCConfig.JDBCPort != jp {
			cfg.DIS.JDBCConfig.JDBCPort = jp
			restart = true
		}
	}
	if jp := trim(rt.GetJavaPath()); jp != "" {
		if cfg.DIS.JDBCConfig == nil {
			cfg.DIS.JDBCConfig = &types.JDBCConfig{}
		}
		if cfg.DIS.JDBCConfig.JavaPath != jp {
			cfg.DIS.JDBCConfig.JavaPath = jp
			restart = true
		}
	}
	if t := trim(rt.GetTenantId()); t != "" {
		cfg.TenantID = t
	}
	if s := trim(rt.GetClientSecret()); s != "" {
		if cfg.ClientSecret != s {
			cfg.ClientSecret = s
			reconnect = true
		}
	}
	if rt.GetForceRestart() {
		restart = true
		reconnect = true
	}
	return restart, reconnect
}

func flattenedStructFields(t reflect.Type) []string {
	return flattenedStructFieldsHelper(t, []string{})
}

func flattenedStructFieldsHelper(t reflect.Type, prefixes []string) []string {
	unwrapped := t
	if t.Kind() == reflect.Pointer {
		unwrapped = t.Elem()
	}
	fields := make([]string, 0)
	for i := 0; i < unwrapped.NumField(); i++ {
		f := unwrapped.Field(i)
		name := f.Tag.Get("mapstructure")
		switch f.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			fields = append(fields, flattenedStructFieldsHelper(f.Type, append(prefixes, name))...)
		default:
			flat := name
			if len(prefixes) > 0 {
				flat = fmt.Sprintf("%s.%s", strings.Join(prefixes, "."), name)
			}
			fields = append(fields, flat)
		}
	}
	return fields
}

// ToEffective wraps the config with provenance defaults (local) for API responses.
func ToEffective(cfg *Config) *EffectiveConfig {
	prov := map[string]string{}
	for _, f := range flattenedStructFields(reflect.TypeOf(*cfg)) {
		prov[f] = "local"
	}
	return &EffectiveConfig{Config: cfg, Provenance: prov}
}

// MarkRuntimeProvenance marks runtime-managed fields as remote in provenance map.
func MarkRuntimeProvenance(eff *EffectiveConfig, rt *bridgeproto.AgentRuntimeConfig) {
	if eff == nil || eff.Provenance == nil || rt == nil {
		return
	}
	set := func(field string) { eff.Provenance[field] = "remote" }
	if strings.TrimSpace(rt.GetDisHost()) != "" {
		set("dis.host")
	}
	if strings.TrimSpace(rt.GetDisUser()) != "" {
		set("dis.user")
	}
	if strings.TrimSpace(rt.GetDisPassword()) != "" {
		set("dis.password")
	}
	if strings.TrimSpace(rt.GetJdbcPort()) != "" {
		set("dis.jdbcConfig.jdbcPort")
	}
	if strings.TrimSpace(rt.GetJavaPath()) != "" {
		set("dis.jdbcConfig.javaPath")
	}
	if strings.TrimSpace(rt.GetTenantId()) != "" {
		set("tenantID")
	}
	if strings.TrimSpace(rt.GetClientSecret()) != "" {
		set("clientSecret")
	}
}
