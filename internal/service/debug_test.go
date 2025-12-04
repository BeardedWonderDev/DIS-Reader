package service

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"path/filepath"
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

	svc := EmbeddedService{
		db:     &fakeDB{rows: []types.ResultRow{{"src_table": "FOO", "A": "1"}}},
		logger: newTestLogger(),
	}

	err := svc.runDebugSearchSQLite("run-1", "term", outputPath, []string{`SELECT 1 AS A`}, nil, nil)
	if err != nil {
		t.Fatalf("runDebugSearchSQLite returned error: %v", err)
	}

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
