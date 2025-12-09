package ui

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	_ "github.com/mattn/go-sqlite3"
)

func TestEnsureDebugOutputDir(t *testing.T) {
	cfg := &types.DISUIConfig{DebugSearch: types.DebugSearchConfig{DefaultOutputPath: ""}}
	dir, err := ensureDebugOutputDir(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(dir) != "debug-output" {
		t.Fatalf("expected debug-output dir, got %s", dir)
	}

	cfg2 := &types.DISUIConfig{DebugSearch: types.DebugSearchConfig{DefaultOutputPath: "custom/data"}}
	dir2, err := ensureDebugOutputDir(cfg2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(dir2) != "custom" {
		t.Fatalf("expected custom dir, got %s", dir2)
	}
}

func TestScanSQLiteOutputs(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "one.db"), "")
	mustWriteFile(t, filepath.Join(dir, "two.txt"), "")

	files := scanSQLiteOutputs(dir)
	if len(files) != 1 || filepath.Base(files[0]) != "one.db" {
		t.Fatalf("expected one db file, got %v", files)
	}
}

func TestReadSQLiteTablesAndRows(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE demo(id INTEGER, name TEXT);"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO demo(id, name) VALUES (1,'alice'), (2,'bob')"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	tables, err := readSQLiteTables(dbPath)
	if err != nil || len(tables) != 1 || tables[0] != "demo" {
		t.Fatalf("unexpected tables: %v err=%v", tables, err)
	}

	cols, rows, err := readSQLiteRows(dbPath, "demo", 0, 10)
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	if len(cols) != 2 || cols[0] != "id" || cols[1] != "name" {
		t.Fatalf("unexpected columns: %v", cols)
	}
	if len(rows) != 2 || rows[0][1] != "alice" || rows[1][1] != "bob" {
		t.Fatalf("unexpected rows: %v", rows)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
