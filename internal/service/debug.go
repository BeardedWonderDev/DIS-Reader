package service

import (
	"bufio"
	"context"
	"database/sql"
	_ "embed"
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed queries.sql
var queriesSQL string

const (
	batchSize   = 20
	delay       = 2 * time.Second
	parallelism = 5
)

// RunDebugSearch executes the debug search batch and streams progress/events.
func (s *EmbeddedService) RunDebugSearch(searchTerm string, opts types.DebugSearchOptions, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
	defer func() {
		if progressChan != nil {
			close(progressChan)
		}
		if eventChan != nil {
			close(eventChan)
		}
	}()

	runID := uuid.New().String()
	if opts.OutputMode == "" {
		opts.OutputMode = types.DebugSearchOutputSQLite
	}
	queryTemplates := loadQueryTemplates()
	if len(queryTemplates) == 0 {
		s.logger.Warn("No debug search queries loaded; aborting")
		return
	}
	if opts.OutputPath == "" {
		ext := ".db"
		if opts.OutputMode == types.DebugSearchOutputCSV {
			ext = ".csv"
		}
		opts.OutputPath = fmt.Sprintf("debug-search-%s%s", runID, ext)
	}

	s.logger.Info("Starting search",
		slog.String("search_term", searchTerm),
		slog.String("mode", string(opts.OutputMode)),
		slog.String("output", opts.OutputPath),
	)

	switch opts.OutputMode {
	case types.DebugSearchOutputCSV:
		s.runDebugSearchCSV(runID, searchTerm, opts.OutputPath, queryTemplates, progressChan, eventChan)
	default:
		if err := s.runDebugSearchSQLite(runID, searchTerm, opts.OutputPath, queryTemplates, progressChan, eventChan); err != nil {
			types.LogError(s.logger, "Debug search failed", err)
		}
	}
}

func loadQueryTemplates() []string {
	lines := []string{}
	scanner := bufio.NewScanner(strings.NewReader(queriesSQL))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "--") {
			lines = append(lines, line)
		}
	}
	return lines
}

func emitProgress(progressChan chan<- types.ProgressStatus, status types.ProgressStatus) {
	if progressChan == nil {
		return
	}
	progressChan <- status
}

func emitEvent(eventChan chan<- types.TableEvent, evt types.TableEvent) {
	if eventChan == nil {
		return
	}
	eventChan <- evt
}

func emitRunFailed(eventChan chan<- types.TableEvent, err error) {
	if eventChan == nil {
		return
	}
	eventChan <- types.TableEvent{
		TableName: "debug_search",
		EventType: "run_failed",
		SampleRow: map[string]interface{}{"error": err.Error()},
	}
}

func (s *EmbeddedService) runDebugSearchSQLite(runID string, searchTerm string, sqliteDBFile string, queryTemplates []string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) error {
	if err := os.MkdirAll(filepath.Dir(sqliteDBFile), 0o755); err != nil {
		emitRunFailed(eventChan, err)
		return fmt.Errorf("ensure output dir: %w", err)
	}

	db, err := sql.Open("sqlite3", sqliteDBFile)
	if err != nil {
		emitRunFailed(eventChan, err)
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`PRAGMA busy_timeout=5000`); err != nil {
		emitRunFailed(eventChan, err)
		return fmt.Errorf("set busy_timeout: %w", err)
	}

	s.ensureBatchStatusTable(db)

	queryItems := buildQueryItems(queryTemplates, searchTerm)
	tableCols := map[string]map[string]struct{}{}
	batches := chunkQueryItems(queryItems, batchSize)

	emitProgress(progressChan, types.ProgressStatus{TotalQueries: len(queryItems), CompletedQueries: 0})

	for _, batch := range batches {
		if err := s.runBatchToSQLite(runID, searchTerm, batch, db, tableCols, eventChan); err != nil {
			emitRunFailed(eventChan, err)
			return err
		}
		emitProgress(progressChan, types.ProgressStatus{TotalQueries: len(queryItems), CompletedQueries: batch[len(batch)-1].Index + 1})
		time.Sleep(delay)
	}

	return nil
}

func (s *EmbeddedService) runDebugSearchCSV(runID string, searchTerm string, csvFile string, queryTemplates []string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
	f, err := os.Create(csvFile)
	if err != nil {
		emitRunFailed(eventChan, err)
		return
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	queryItems := buildQueryItems(queryTemplates, searchTerm)
	emitProgress(progressChan, types.ProgressStatus{TotalQueries: len(queryItems), CompletedQueries: 0})

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)

	for _, item := range queryItems {
		wg.Add(1)
		sem <- struct{}{}
		go func(it queryItem) {
			defer wg.Done()
			defer func() { <-sem }()

			rows, err := s.db.Query(context.Background(), it.Query)
			if err != nil {
				emitEvent(eventChan, types.TableEvent{TableName: it.Table, EventType: "error", SampleRow: map[string]interface{}{"error": err.Error()}})
				return
			}

			mu.Lock()
			if len(rows) > 0 {
				headers := make([]string, 0, len(rows[0]))
				for k := range rows[0] {
					headers = append(headers, k)
				}
				_ = writer.Write(headers)
				for _, row := range rows {
					record := make([]string, 0, len(headers))
					for _, h := range headers {
						record = append(record, fmt.Sprint(row[h]))
					}
					_ = writer.Write(record)
				}
			}
			mu.Unlock()

			emitProgress(progressChan, types.ProgressStatus{TotalQueries: len(queryItems), CompletedQueries: it.Index + 1})
			emitEvent(eventChan, types.TableEvent{TableName: it.Table, EventType: "completed"})
		}(item)
	}

	wg.Wait()
}

func (s *EmbeddedService) runBatchToSQLite(runID string, searchTerm string, batch []queryItem, db *sql.DB, tableCols map[string]map[string]struct{}, eventChan chan<- types.TableEvent) error {
	for _, item := range batch {
		query := strings.ReplaceAll(item.Query, "{{searchTerm}}", searchTerm)
		if err := s.getSQLiteIterative(tableCols, db, query, nil, eventChan); err != nil {
			return err
		}
	}
	return nil
}

// getSQLiteIterative and helpers below

func (s *EmbeddedService) ensureBatchStatusTable(db *sql.DB) {
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS batch_status (run_id TEXT, idx INTEGER, PRIMARY KEY (run_id, idx))`)
}

func (s *EmbeddedService) loadCompletedFromDB(runID string, db *sql.DB) map[int]struct{} {
	completed := make(map[int]struct{})
	rows, err := db.Query(`SELECT idx FROM batch_status WHERE run_id = ?`, runID)
	if err != nil {
		return completed
	}
	defer rows.Close()
	for rows.Next() {
		var idx int
		if err := rows.Scan(&idx); err == nil {
			completed[idx] = struct{}{}
		}
	}
	return completed
}

func (s *EmbeddedService) markCompletedInDB(runID string, searchTerm string, idx int, db *sql.DB) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO batch_status (run_id, idx) VALUES (?, ?)`, runID, idx)
	return err
}

func (s *EmbeddedService) getSQLiteIterative(tableCols map[string]map[string]struct{}, db *sql.DB, query string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) error {
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		emitEvent(eventChan, types.TableEvent{TableName: "debug_search", EventType: "error", SampleRow: map[string]interface{}{"error": err.Error()}})
		return err
	}

	for _, row := range rows {
		table, _ := row["src_table"].(string)
		delete(row, "src_table")
		if _, ok := tableCols[table]; !ok {
			tableCols[table] = map[string]struct{}{}
		}
		cols := tableCols[table]
		for col := range row {
			cols[col] = struct{}{}
		}

		if err := s.writeRow(db, table, row); err != nil {
			emitEvent(eventChan, types.TableEvent{TableName: table, EventType: "error", SampleRow: map[string]interface{}{"error": err.Error()}})
			return err
		}
		emitEvent(eventChan, types.TableEvent{TableName: table, EventType: "row", SampleRow: row})
	}

	return nil
}

func (s *EmbeddedService) writeRow(db *sql.DB, table string, row map[string]interface{}) error {
	cols := make([]string, 0, len(row))
	placeholders := make([]string, 0, len(row))
	args := make([]interface{}, 0, len(row))
	for k, v := range row {
		cols = append(cols, k)
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}
	createCols := make([]string, 0, len(row))
	for k := range row {
		createCols = append(createCols, fmt.Sprintf("%s TEXT", k))
	}
	createStmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, strings.Join(createCols, ","))
	if _, err := db.Exec(createStmt); err != nil {
		return err
	}
	insertStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ","), strings.Join(placeholders, ","))
	_, err := db.Exec(insertStmt, args...)
	return err
}

type queryItem struct {
	Query string
	Table string
	Index int
}

func buildQueryItems(templates []string, searchTerm string) []queryItem {
	items := make([]queryItem, 0, len(templates))
	for idx, tmpl := range templates {
		items = append(items, queryItem{Query: strings.ReplaceAll(tmpl, "{{searchTerm}}", searchTerm), Table: tableFromQuery(tmpl), Index: idx})
	}
	return items
}

func tableFromQuery(q string) string {
	// naive extractor: after FROM first token
	parts := strings.Split(strings.ToUpper(q), "FROM")
	if len(parts) < 2 {
		return "unknown"
	}
	rest := strings.TrimSpace(parts[1])
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return "unknown"
	}
	return strings.Trim(fields[0], "\n\r	 ;")
}

func chunkQueryItems(items []queryItem, size int) [][]queryItem {
	var batches [][]queryItem
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}
	return batches
}
