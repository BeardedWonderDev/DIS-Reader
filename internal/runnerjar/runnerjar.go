package runnerjar

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

const jarFileName = "dis-runner-0.1.1.jar"

//go:embed dis-runner-0.1.1.jar
var jarBytes []byte

// Extract writes the embedded JDBC runner jar to a temp directory and returns its path.
// Call Cleanup on the returned Extracted to remove the temp files when done.
type Extracted struct {
	JarPath string
	Dir     string
}

func (e *Extracted) Cleanup() error {
	if e == nil || e.Dir == "" {
		return nil
	}
	return os.RemoveAll(e.Dir)
}

func Extract() (*Extracted, error) {
	dir, err := os.MkdirTemp("", "disrunner-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	jarPath := filepath.Join(dir, jarFileName)
	if err := os.WriteFile(jarPath, jarBytes, 0644); err != nil {
		os.RemoveAll(dir)
		return nil, fmt.Errorf("write jar: %w", err)
	}
	return &Extracted{JarPath: jarPath, Dir: dir}, nil
}
