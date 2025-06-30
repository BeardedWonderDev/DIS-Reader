package types

type Config struct {
	AppName    string
	AppVersion string
	JavaPath   string
	JarPath    string
	ClassDir   string
	JDBCPort   string
	User       string
	Password   string
	Host       string
}

type DISReaderService interface {
	GetConfig() *Config
	Shutdown()
	AttachShutdownHook()
	TestDISConnection() error
	RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- ProgressStatus, eventChan chan<- TableEvent)
}
