package internal

import (
	_ "embed"

	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// //go:embed JDBCRunner.class
// var embeddedJar []byte

const (
	className = "JDBCRunner"
)

type DISReaderService struct {
	config *types.Config
}

func NewDISReaderService(config *types.Config) *DISReaderService {
	// ensureEmbeddedJarWritten(config.JarPath)

	return &DISReaderService{
		config: config,
	}
}

func (s DISReaderService) GetConfig() *types.Config {
	return s.config
}

// // EnsureEmbeddedJarWritten checks if the embedded JAR exists at the given path,
// // and writes it if it does not exist. Returns the full path to the JAR.
// func ensureEmbeddedJarWritten(jarDir string) (string, error) {
// 	jarPath := filepath.Join(jarDir, className+".class")
// 	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
// 		if err := os.WriteFile(jarPath, embeddedJar, 0644); err != nil {
// 			return "", err
// 		}
// 	}
// 	return jarPath, nil
// }

// TestDISConnection tests the JDBC connection using the embedded Java class.
func (s DISReaderService) TestDISConnection() error {
	cmd := exec.Command(
		s.config.JavaPath,
		"-cp", fmt.Sprintf("%s:.", s.config.JarPath),
		className,
		s.config.Host,
		s.config.User,
		s.config.Password,
		"test",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute Java connection test: %w\nOutput: %s", err, output)
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal(line, &obj); err != nil {
			continue // Ignore bad lines
		}
		// Expect NDJSON format: {"status":"ok"|"error", ...}
		if status, ok := obj["status"].(string); ok {
			if status == "ok" {
				return nil
			}
			if status == "error" {
				return fmt.Errorf("connection test failed: %v", obj["message"])
			}
		}
	}
	return fmt.Errorf("connection test did not return a clear status; output: %s", output)
}
