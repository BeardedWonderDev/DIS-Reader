package service

import (
	"context"
	"database/sql"
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

func TestRunDebugSearchCSVWritesRows(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "out.csv")
	progress := make(chan types.ProgressStatus, 10)
	events := make(chan types.TableEvent, 10)

	svc := EmbeddedService{
		db:     &fakeDB{rows: []types.ResultRow{{"A": "1", "B": "two"}}},
		logger: newTestLogger(),
	}

	svc.runDebugSearchCSV("run-1", "term", outputPath, []string{`SELECT * FROM parts`}, progress, events)

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(content), "1") || !strings.Contains(string(content), "two") {
		t.Fatalf("expected csv content to include data rows, got %s", string(content))
	}
	select {
	case evt := <-events:
		if evt.EventType != "completed" {
			t.Fatalf("expected completed event, got %s", evt.EventType)
		}
	default:
		t.Fatalf("expected at least one event emitted")
	}
}

func TestRunDebugSearchCSVEmitsFailure(t *testing.T) {
	tmp := t.TempDir()
	// directory doesn't exist; os.Create will fail
	outputPath := filepath.Join(tmp, "missing", "out.csv")
	events := make(chan types.TableEvent, 1)

	svc := EmbeddedService{
		db:     &fakeDB{rows: []types.ResultRow{{"A": "1"}}},
		logger: newTestLogger(),
	}

	svc.runDebugSearchCSV("run-err", "term", outputPath, []string{`SELECT * FROM parts`}, nil, events)

	select {
	case evt := <-events:
		if evt.EventType != "run_failed" {
			t.Fatalf("expected run_failed event, got %s", evt.EventType)
		}
	default:
		t.Fatalf("expected failure event")
	}
}

func TestLoadQueryTemplatesParsesNonComments(t *testing.T) {
	templates := loadQueryTemplates()
	if len(templates) == 0 {
		t.Fatalf("expected embedded query templates")
	}
	for _, tmpl := range templates {
		if strings.HasPrefix(strings.TrimSpace(tmpl), "--") || strings.TrimSpace(tmpl) == "" {
			t.Fatalf("unexpected comment/blank in templates: %q", tmpl)
		}
	}
}

func TestBatchStatusPersistence(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	svc := &EmbeddedService{}
	svc.ensureBatchStatusTable(db)

	runID := "run-1"
	if err := svc.markCompletedInDB(runID, "term", 2, db); err != nil {
		t.Fatalf("mark completed: %v", err)
	}
	if err := svc.markCompletedInDB(runID, "term", 5, db); err != nil {
		t.Fatalf("mark completed: %v", err)
	}

	completed := svc.loadCompletedFromDB(runID, db)
	if _, ok := completed[2]; !ok {
		t.Fatalf("expected idx 2 to be completed")
	}
	if _, ok := completed[5]; !ok {
		t.Fatalf("expected idx 5 to be completed")
	}
	if _, ok := completed[1]; ok {
		t.Fatalf("did not expect idx 1 to be completed")
	}
}
