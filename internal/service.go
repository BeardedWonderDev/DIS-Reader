package internal

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

//go:embed JDBCRunner.class
var embeddedJar []byte

const (
	className = "JDBCRunner.class"
)

type DISReaderService struct {
	config *types.Config
}

func NewDISReaderService(config *types.Config) DISReaderService {
	ensureEmbeddedJarWritten(config.JarPath)

	return DISReaderService{
		config: config,
	}
}

// EnsureEmbeddedJarWritten checks if the embedded JAR exists at the given path,
// and writes it if it does not exist. Returns the full path to the JAR.
func ensureEmbeddedJarWritten(jarDir string) (string, error) {
	jarPath := filepath.Join(jarDir, className)
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		if err := os.WriteFile(jarPath, embeddedJar, 0644); err != nil {
			return "", err
		}
	}
	return jarPath, nil
}
