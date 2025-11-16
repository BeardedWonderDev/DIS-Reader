package tview2

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	_ "github.com/mattn/go-sqlite3"
)

func ensureDebugOutputDir(cfg *types.DISUIConfig) (string, error) {
	base := strings.TrimSpace(cfg.DebugSearch.DefaultOutputPath)
	if base == "" {
		base = "debug-search"
	}
	if !strings.HasSuffix(strings.ToLower(base), ".db") && !strings.HasSuffix(strings.ToLower(base), ".csv") {
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
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".db") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files
}

func readSQLiteTables(path string) ([]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func readSQLiteRows(path, table string, page, pageSize int) ([]string, [][]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", table, pageSize, page*pageSize)
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var result [][]string
	for rows.Next() {
		raw := make([]interface{}, len(columns))
		rawPtrs := make([]interface{}, len(columns))
		for i := range raw {
			rawPtrs[i] = &raw[i]
		}
		if err := rows.Scan(rawPtrs...); err != nil {
			return nil, nil, err
		}
		formatted := make([]string, len(columns))
		for i, val := range raw {
			switch v := val.(type) {
			case []byte:
				formatted[i] = string(v)
			case nil:
				formatted[i] = "<nil>"
			default:
				formatted[i] = fmt.Sprint(v)
			}
		}
		result = append(result, formatted)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return columns, result, nil
}
