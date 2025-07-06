package types

type Config struct {
	AppName    string     `mapstructure:"appName"`
	AppVersion string     `mapstructure:"appVersion"`
	LogLevel   string     `mapstructure:"logLevel"`
	DIS        *DISConfig `mapstructure:"disConfig"`
}

type DISConfig struct {
	JDBCConfig         *JDBCConfig `mapstructure:"jdbcConfig"`
	Host               string      `mapstructure:"host"`
	User               string      `mapstructure:"user"`
	Password           string      `mapstructure:"password"`
	MaxIdleConnections int         `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int         `mapstructure:"maxOpenConnections"`
}

type JDBCConfig struct {
	JavaPath string `mapstructure:"javaPath"`
	JDBCPort string `mapstructure:"jdbcPort"`
	JarPath  string `mapstructure:"jarPath"`
	ClassDir string `mapstructure:"classDir"`
}
