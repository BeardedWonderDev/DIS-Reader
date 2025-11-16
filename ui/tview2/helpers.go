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

func formatTable(columns []string, rows [][]string, width int) string {
	if len(columns) == 0 {
		return "No columns"
	}
	colWidths := make([]int, len(columns))
	maxWidth := width
	if maxWidth <= 0 {
		maxWidth = 80
	}
	for i, col := range columns {
		colWidths[i] = len(col)
	}
	for _, row := range rows {
		for i, val := range row {
			if l := len(val); l > colWidths[i] {
				colWidths[i] = l
			}
		}
	}
	padding := 3 * (len(colWidths) - 1)
	total := padding
	for i, w := range colWidths {
		if w > 40 {
			w = 40
		}
		colWidths[i] = w
		total += w
	}
	if total > maxWidth {
		scale := float64(maxWidth-padding) / float64(total-padding)
		for i := range colWidths {
			colWidths[i] = int(float64(colWidths[i]) * scale)
			if colWidths[i] < 4 {
				colWidths[i] = 4
			}
		}
	}
	formatRow := func(cells []string) string {
		parts := make([]string, len(cells))
		for i, cell := range cells {
			w := colWidths[i]
			parts[i] = fmt.Sprintf("%-*s", w, truncate(cell, w))
		}
		return strings.Join(parts, " │ ")
	}
	header := formatRow(columns)
	separator := strings.Repeat("─", len(header))
	lines := []string{header, separator}
	for _, row := range rows {
		lines = append(lines, formatRow(row))
	}
	return strings.Join(lines, "\n")
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 1 {
		return text[:max]
	}
	return text[:max-1] + "…"
}
