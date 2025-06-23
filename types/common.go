package types

type Config struct {
	JavaPath string
	JarPath  string
	User     string
	Password string
	Host     string
}

type DISReaderService interface {
	RunDebugSearch(searchTerm string, outName string)
}
