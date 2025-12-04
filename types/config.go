package types

import "log/slog"

type DISUIConfig struct {
	AppName     string            `mapstructure:"appName"`
	AppVersion  string            `mapstructure:"appVersion"`
	Theme       string            `mapstructure:"theme"`
	DIS         *DISConfig        `mapstructure:"disConfig"`
	Bridge      *BridgeConfig     `mapstructure:"bridge"`
	DebugSearch DebugSearchConfig `mapstructure:"debugSearch"`
}

type DebugSearchConfig struct {
	DefaultOutputMode string `mapstructure:"defaultOutputMode"`
	DefaultOutputPath string `mapstructure:"defaultOutputPath"`
}

type DISConfig struct {
	LogLevel           slog.Level    `mapstructure:"logLevel"`
	JDBCConfig         *JDBCConfig   `mapstructure:"jdbcConfig"`
	Host               string        `mapstructure:"host"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	MaxIdleConnections int           `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int           `mapstructure:"maxOpenConnections"`
	Bridge             *BridgeConfig `mapstructure:"bridge"`
}

type JDBCConfig struct {
	JavaPath string `mapstructure:"javaPath"`
	JDBCPort string `mapstructure:"jdbcPort"`
	JarPath  string `mapstructure:"jarPath"`
	ClassDir string `mapstructure:"classDir"`
}

type BridgeConfig struct {
	Mode         string                  `mapstructure:"mode"` // embedded | remote
	ServerURL    string                  `mapstructure:"serverURL"`
	ClientID     string                  `mapstructure:"clientID"`
	ClientSecret string                  `mapstructure:"clientSecret"`
	TenantID     string                  `mapstructure:"tenantID"`
	TLS          BridgeTLSConfig         `mapstructure:"tls"`
	Allowed      []BridgeAgentCredential `mapstructure:"allowedAgents"`
}

type BridgeTLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecureSkipVerify"`
}

type BridgeAgentCredential struct {
	ClientID     string `mapstructure:"clientID"`
	ClientSecret string `mapstructure:"clientSecret"`
	TenantID     string `mapstructure:"tenantID"`
	AgentID      string `mapstructure:"agentID"`
}
