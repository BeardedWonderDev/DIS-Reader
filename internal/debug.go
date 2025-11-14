package internal

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	_ "embed"

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

type queryItem struct {
	Index int
	Query string
}

func loadQueryTemplates() []string {
	var templates []string
	scanner := bufio.NewScanner(strings.NewReader(queriesSQL))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			templates = append(templates, line)
		}
	}
	return templates
}

func (s DISReaderService) ensureBatchStatusTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS batch_status (
        run_id TEXT NOT NULL,
        search_term TEXT NOT NULL,
        query_index INTEGER NOT NULL,
        completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (run_id, query_index)
    )`)
	if err != nil {
		types.LogError(s.logger, "Failed to ensure batch_status table", err)
	}
}

func (s DISReaderService) RunDebugSearch(searchTerm string, opts types.DebugSearchOptions, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
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
		s.runDebugSearchSQLite(runID, searchTerm, opts.OutputPath, queryTemplates, progressChan, eventChan)
	}
}

func (s DISReaderService) runDebugSearchSQLite(runID string, searchTerm string, sqliteDBFile string, queryTemplates []string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
	// Open or create the SQLite database
	db, err := sql.Open("sqlite3", sqliteDBFile)
	if err != nil {
		types.LogError(s.logger, "Failed to open SQLite DB", err)
		return
	}
	defer db.Close()

	s.ensureBatchStatusTable(db)

	completed := s.loadCompletedFromDB(runID, db)
	s.emitProgress(progressChan, types.ProgressStatus{
		RunID:            runID,
		TotalQueries:     len(queryTemplates),
		CompletedQueries: len(completed),
		PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
	})

	var batch []queryItem
	tableCols := make(map[string]map[string]struct{}) // Table -> set of columns

	for i, tmpl := range queryTemplates {
		if _, ok := completed[i]; ok {
			continue // Already completed, skip
		}
		query := strings.ReplaceAll(tmpl, "'{{SEARCH}}'", fmt.Sprintf("'%s'", searchTerm))
		batch = append(batch, queryItem{Index: i, Query: query})
		if len(batch) >= batchSize {
			s.runBatchToSQLite(runID, searchTerm, batch, db, tableCols, eventChan)
			batch = []queryItem{}
			time.Sleep(delay)
			// Re-load completed from DB after each batch
			completed = s.loadCompletedFromDB(runID, db)
			s.emitProgress(progressChan, types.ProgressStatus{
				RunID:            runID,
				TotalQueries:     len(queryTemplates),
				CompletedQueries: len(completed),
				PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
			})
		}
	}

	if len(batch) > 0 {
		s.runBatchToSQLite(runID, searchTerm, batch, db, tableCols, eventChan)
		completed = s.loadCompletedFromDB(runID, db)
		s.emitProgress(progressChan, types.ProgressStatus{
			RunID:            runID,
			TotalQueries:     len(queryTemplates),
			CompletedQueries: len(completed),
			PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
		})
	}

	completed = s.loadCompletedFromDB(runID, db)
	if len(completed) == len(queryTemplates) {
		s.logger.Info("All queries completed", slog.String("mode", "sqlite"))
	} else {
		remaining := len(queryTemplates) - len(completed)
		s.logger.Info("Batch processing incomplete", slog.Int("remaining_queries", remaining))
	}
}

func (s DISReaderService) runDebugSearchCSV(runID string, searchTerm string, csvFile string, queryTemplates []string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
	file, err := os.Create(csvFile)
	if err != nil {
		types.LogError(s.logger, "Failed to create CSV file", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"run_id", "search_term", "table_name", "row_index", "row_json"}); err != nil {
		types.LogError(s.logger, "Failed to write CSV header", err)
		return
	}

	completed := 0
	total := len(queryTemplates)
	s.emitProgress(progressChan, types.ProgressStatus{
		RunID:            runID,
		TotalQueries:     total,
		CompletedQueries: completed,
		PercentComplete:  0,
	})

	tableSeen := make(map[string]bool)
	rowCounters := make(map[string]int)

	for idx, tmpl := range queryTemplates {
		query := strings.ReplaceAll(tmpl, "'{{SEARCH}}'", fmt.Sprintf("'%s'", searchTerm))
		rows, err := s.db.QueryWithSource(context.TODO(), query)
		if err != nil {
			types.LogError(s.logger, "Query failed", err, slog.Int("index", idx))
			s.emitEvent(eventChan, types.TableEvent{
				TableName: "query_errors",
				EventType: "query_failed",
				SampleRow: map[string]interface{}{
					"query_index":   idx,
					"query":         query,
					"error_message": err.Error(),
				},
			})
			continue
		}

		for _, row := range rows {
			tableName := getTableName(row)
			if !tableSeen[tableName] {
				tableSeen[tableName] = true
				s.emitEvent(eventChan, types.TableEvent{
					TableName:   tableName,
					EventType:   "table_created",
					ColumnCount: len(row),
				})
			}
			rowIndex := rowCounters[tableName]
			rowCounters[tableName] = rowIndex + 1

			payload, err := json.Marshal(row)
			if err != nil {
				types.LogError(s.logger, "Failed to marshal row", err)
				continue
			}

			record := []string{runID, searchTerm, tableName, fmt.Sprintf("%d", rowIndex), string(payload)}
			if err := writer.Write(record); err != nil {
				types.LogError(s.logger, "Failed to write CSV row", err)
				continue
			}

			s.emitEvent(eventChan, types.TableEvent{
				TableName:   tableName,
				EventType:   "row_inserted",
				RowCount:    1,
				ColumnCount: len(row),
				SampleRow:   row,
			})
		}

		completed++
		percent := 0.0
		if total > 0 {
			percent = float64(completed) / float64(total) * 100
		}
		s.emitProgress(progressChan, types.ProgressStatus{
			RunID:            runID,
			TotalQueries:     total,
			CompletedQueries: completed,
			PercentComplete:  percent,
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		types.LogError(s.logger, "CSV writer error", err)
	}
}

func (s DISReaderService) runBatchToSQLite(runID string, searchTerm string, batch []queryItem, db *sql.DB, tableCols map[string]map[string]struct{}, eventChan chan<- types.TableEvent) {
	s.logger.Info("Running batch", slog.Int("query_count", len(batch)))

	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)
	var mu sync.Mutex

	for _, item := range batch {
		wg.Add(1)
		sem <- struct{}{}
		go func(item queryItem) {
			defer wg.Done()
			defer func() { <-sem }()
			query := strings.TrimSuffix(item.Query, " UNION ALL")
			query = strings.TrimSuffix(query, "UNION ALL")
			query = strings.TrimSpace(query)
			s.logger.Debug("Executing query", slog.Int("query_index", item.Index), slog.String("query", query))

			rows, err := s.db.QueryWithSource(context.TODO(), query)
			if err != nil {
				errMsg := fmt.Sprintf("Query %d failed: %v", item.Index, err)
				types.LogError(s.logger, "Query failed", err, slog.Int("index", item.Index))

				_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS query_errors (
          run_id TEXT,
          query_index INTEGER,
          query TEXT,
          error_message TEXT,
          occurred_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )`)

				_, _ = db.Exec(`INSERT INTO query_errors (run_id, query_index, query, error_message) VALUES (?, ?, ?, ?)`,
					runID, item.Index, query, errMsg)

				s.emitEvent(eventChan, types.TableEvent{
					TableName:   "query_errors",
					EventType:   "query_failed",
					RowCount:    0,
					ColumnCount: 0,
					SampleRow: map[string]interface{}{
						"query_index":   item.Index,
						"query":         query,
						"error_message": err.Error(),
					},
				})
				return
			}

			if len(rows) == 0 {
				// No rows returned, mark completed and skip
				mu.Lock()
				s.markCompletedInDB(runID, searchTerm, item.Index, db)
				mu.Unlock()
				return
			}

			for _, row := range rows {
				data := row

				// Table name from SRC_TABLE, fallback "unknown"
				tableName, ok := data["SRC_TABLE"].(string)
				if !ok || tableName == "" {
					tableName = "unknown"
				}

				mu.Lock()
				// Track columns for this table
				if _, exists := tableCols[tableName]; !exists {
					tableCols[tableName] = make(map[string]struct{})
				}

				// Check for new columns and alter table if needed
				for col := range data {
					if _, seen := tableCols[tableName][col]; !seen {
						addColumnIfNotExists(db, tableName, col)
						tableCols[tableName][col] = struct{}{}
					}
				}

				// Ensure table exists
				createTableIfNotExists(db, tableName, row, eventChan)
				// Insert row
				insertRow(db, tableName, row, eventChan)
				mu.Unlock()
			}
			s.markCompletedInDB(runID, searchTerm, item.Index, db)
		}(item)
	}
	wg.Wait()
}

func createTableIfNotExists(db *sql.DB, table string, row types.ResultRow, eventChan chan<- types.TableEvent) {
	cols := []string{}
	for col := range row {
		cols = append(cols, fmt.Sprintf("%q TEXT", col))
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" (%s)`, table, strings.Join(cols, ","))
	res, err := db.Exec(query)
	if err == nil {
		// Check if table was created by querying sqlite_master
		var count int
		err2 := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
		if err2 == nil && count == 1 && eventChan != nil {
			// Send event that table is created
			eventChan <- types.TableEvent{
				TableName:   table,
				EventType:   "table_created",
				RowCount:    0,
				ColumnCount: len(row),
				SampleRow:   nil,
			}
		}
	} else {
		_ = res
	}
}

func (s DISReaderService) emitProgress(ch chan<- types.ProgressStatus, status types.ProgressStatus) {
	if ch == nil {
		return
	}
	ch <- status
}

func (s DISReaderService) emitEvent(ch chan<- types.TableEvent, evt types.TableEvent) {
	if ch == nil {
		return
	}
	ch <- evt
}

func getTableName(row types.ResultRow) string {
	if tableName, ok := row["SRC_TABLE"].(string); ok && tableName != "" {
		return tableName
	}
	return "unknown"
}

func addColumnIfNotExists(db *sql.DB, table, col string) {
	_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" TEXT`, table, col))
}

func insertRow(db *sql.DB, table string, row types.ResultRow, eventChan chan<- types.TableEvent) {
	cols := []string{}
	vals := []interface{}{}
	holders := []string{}
	for col, v := range row {
		cols = append(cols, fmt.Sprintf("%q", col))
		vals = append(vals, fmt.Sprintf("%v", v))
		holders = append(holders, "?")
	}
	stmt := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`, table, strings.Join(cols, ","), strings.Join(holders, ","))
	_, err := db.Exec(stmt, vals...)
	if err == nil && eventChan != nil {
		eventChan <- types.TableEvent{
			TableName:   table,
			EventType:   "row_inserted",
			RowCount:    1,
			ColumnCount: len(row),
			SampleRow:   row,
		}
	}
}

func (s DISReaderService) loadCompletedFromDB(runID string, db *sql.DB) map[int]struct{} {
	rows, err := db.Query(`SELECT query_index FROM batch_status WHERE run_id = ?`, runID)
	if err != nil {
		s.logger.Warn("Failed to load completed queries", "error", err)
		return map[int]struct{}{}
	}
	defer rows.Close()

	completed := make(map[int]struct{})
	for rows.Next() {
		var idx int
		if err := rows.Scan(&idx); err == nil {
			completed[idx] = struct{}{}
		}
	}
	return completed
}

func (s DISReaderService) markCompletedInDB(runID string, searchTerm string, idx int, db *sql.DB) {
	_, err := db.Exec(`INSERT OR IGNORE INTO batch_status (run_id, search_term, query_index) VALUES (?, ?, ?)`, runID, searchTerm, idx)
	if err != nil {
		s.logger.Warn("Failed to mark query as completed", "index", idx, "error", err)
	}
}
