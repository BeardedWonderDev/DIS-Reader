package types

type Config struct {
	JavaPath string
	JarPath  string
	User     string
	Password string
	Host     string
}

type DISReaderService interface {
	GetConfig() *Config
	TestDISConnection() error
}
