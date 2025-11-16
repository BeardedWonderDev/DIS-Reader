package bubbletea

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func ensureDebugOutputDir(cfg *types.DISUIConfig) (string, error) {
	base := strings.TrimSpace(cfg.DebugSearch.DefaultOutputPath)
	if base == "" {
		base = "debug-search"
	}
	if !strings.HasSuffix(base, ".db") && !strings.HasSuffix(base, ".csv") {
		base = base + ".db"
	}
	dir := filepath.Dir(base)
	if dir == "." || dir == "" {
		dir = "debug-output"
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", err
	}
	return abs, nil
}

func scanSQLiteOutputs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".db") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	return files
}
