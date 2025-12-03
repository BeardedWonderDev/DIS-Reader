package internal

import (
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// fakeDB implements database.DB with just enough behavior for debug search tests.
type fakeDB struct {
	rows []types.ResultRow
}

func (f *fakeDB) StartJDBCRunner() error                 { return nil }
func (f *fakeDB) StopJDBCRunner() error                  { return nil }
func (f *fakeDB) Connect(ctx context.Context) error      { return nil }
func (f *fakeDB) Disconnect(ctx context.Context) error   { return nil }
func (f *fakeDB) PingService(ctx context.Context) error  { return nil }
func (f *fakeDB) PingDatabase(ctx context.Context) error { return nil }
func (f *fakeDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (f *fakeDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return f.QueryWithSource(ctx, query, args...)
}
func (f *fakeDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	rows, err := f.Query(ctx, query, args...)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}
func (f *fakeDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}
func (f *fakeDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return f.rows, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func TestRunDebugSearchSQLiteCreatesDirAndPersistsRows(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "nested", "results.db")

	svc := DISReaderService{
		db:     &fakeDB{rows: []types.ResultRow{{"SRC_TABLE": "FOO", "A": "1"}}},
		logger: newTestLogger(),
	}

	err := svc.runDebugSearchSQLite("run-1", "term", outputPath, []string{`SELECT 1 AS A`}, nil, nil)
	if err != nil {
		t.Fatalf("runDebugSearchSQLite returned error: %v", err)
	}

	// The DB file should exist and contain one row in table FOO.
	db, err := sql.Open("sqlite3", outputPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT count(*) FROM FOO`).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestRunDebugSearchCSVWritesFile(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "out.csv")

	svc := DISReaderService{
		db:     &fakeDB{rows: []types.ResultRow{{"SRC_TABLE": "FOO", "A": "1"}}},
		logger: newTestLogger(),
	}

	svc.runDebugSearchCSV("run-1", "needle", out, []string{"SELECT 1"}, nil, nil)

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (header + data), got %d", len(rows))
	}
	if rows[1][2] != "FOO" {
		t.Fatalf("expected table_name FOO, got %s", rows[1][2])
	}
}

func TestRunDebugSearchSQLiteEmitsFailureEvent(t *testing.T) {
	tmp := t.TempDir()
	roDir := filepath.Join(tmp, "ro")
	if err := os.MkdirAll(roDir, 0o500); err != nil {
		t.Fatalf("mkdir ro dir: %v", err)
	}
	out := filepath.Join(roDir, "subdir", "out.db")

	events := make(chan types.TableEvent, 1)

	svc := DISReaderService{
		db:     &fakeDB{rows: []types.ResultRow{{"SRC_TABLE": "FOO", "A": "1"}}},
		logger: newTestLogger(),
	}

	err := svc.runDebugSearchSQLite("run-err", "term", out, []string{"SELECT 1"}, nil, events)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	select {
	case evt := <-events:
		if evt.EventType != "run_failed" {
			t.Fatalf("expected run_failed event, got %s", evt.EventType)
		}
	default:
		t.Fatalf("expected run_failed event")
	}
}
