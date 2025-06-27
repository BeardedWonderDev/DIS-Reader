package internal

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func ensureBatchStatusTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS batch_status (
        run_id TEXT NOT NULL,
        search_term TEXT NOT NULL,
        query_index INTEGER NOT NULL,
        completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (run_id, query_index)
    )`)
	if err != nil {
		log.Fatalf("Failed to ensure batch_status table: %v", err)
	}
}

func (s DISReaderService) RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
	runID := uuid.New().String()
	// Load query templates from embedded file
	var queryTemplates []string
	scanner := bufio.NewScanner(strings.NewReader(queriesSQL))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			queryTemplates = append(queryTemplates, line)
		}
	}

	// Open or create the SQLite database
	db, err := sql.Open("sqlite3", sqliteDBFile)
	if err != nil {
		log.Fatalf("Failed to open SQLite DB: %v", err)
	}
	defer db.Close()

	ensureBatchStatusTable(db)

	completed := loadCompletedFromDB(runID, db)

	progressChan <- types.ProgressStatus{
		RunID:            runID,
		TotalQueries:     len(queryTemplates),
		CompletedQueries: len(completed),
		PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
	}

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
			// Re-load completed from DB after each batch in case another process is marking completions
			completed = loadCompletedFromDB(runID, db)
			progressChan <- types.ProgressStatus{
				RunID:            runID,
				TotalQueries:     len(queryTemplates),
				CompletedQueries: len(completed),
				PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
			}
		}
	}

	// Final batch, if any remain
	if len(batch) > 0 {
		s.runBatchToSQLite(runID, searchTerm, batch, db, tableCols, eventChan)
		completed = loadCompletedFromDB(runID, db)
		progressChan <- types.ProgressStatus{
			RunID:            runID,
			TotalQueries:     len(queryTemplates),
			CompletedQueries: len(completed),
			PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
		}
	}

	// After all batches, check if all queries are completed
	completed = loadCompletedFromDB(runID, db)
	if len(completed) == len(queryTemplates) {
		fmt.Println("All queries completed.")
	} else {
		remaining := len(queryTemplates) - len(completed)
		fmt.Printf("Batch processing done. %d queries remain incomplete.\n", remaining)
	}
}

func (s DISReaderService) runBatchToSQLite(runID string, searchTerm string, batch []queryItem, db *sql.DB, tableCols map[string]map[string]struct{}, eventChan chan<- types.TableEvent) {
	fmt.Printf("\nRunning batch of %d queries...\n", len(batch))

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
			fmt.Printf("Executing line %d: %s\n", item.Index, query)
			cmd := exec.Command(
				s.config.JavaPath,
				"-cp", fmt.Sprintf("%s:.", s.config.JarPath),
				className,
				s.config.Host,
				s.config.User,
				s.config.Password,
				query,
			)
			output, err := cmd.CombinedOutput()
			if err != nil {
				errMsg := fmt.Sprintf("Query %d failed: %v\nOutput:\n%s", item.Index, err, output)
				fmt.Fprintln(os.Stderr, errMsg)

				// Insert into a query_errors table
				_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS query_errors (
          run_id TEXT,
          query_index INTEGER,
          query TEXT,
          error_message TEXT,
          occurred_at DATETIME DEFAULT CURRENT_TIMESTAMP
      )`)

				_, _ = db.Exec(`INSERT INTO query_errors (run_id, query_index, query, error_message) VALUES (?, ?, ?, ?)`,
					runID, item.Index, query, errMsg)

				// Send over event channel
				eventChan <- types.TableEvent{
					TableName:   "query_errors",
					EventType:   "query_failed",
					RowCount:    0,
					ColumnCount: 0,
					SampleRow: map[string]interface{}{
						"query_index":   item.Index,
						"query":         query,
						"error_message": err.Error(),
					},
				}
				return
			}

			scanner := bufio.NewScanner(bytes.NewReader(output))
			for scanner.Scan() {
				line := scanner.Bytes()
				if len(bytes.TrimSpace(line)) == 0 {
					continue
				}
				var obj map[string]interface{}
				if err := json.Unmarshal(line, &obj); err != nil {
					log.Printf("Line %d: bad JSON: %s", item.Index, string(line))
					continue
				}

				if len(obj) == 2 && obj["status"] == "ok" && obj["message"] == "Query completed" {
					// Query succeeded but returned no rows; mark completed and skip
					mu.Lock()
					markCompletedInDB(runID, searchTerm, item.Index, db)
					mu.Unlock()
					return
				}

				// Table name from SRC_TABLE, fallback "unknown"
				tableName, ok := obj["SRC_TABLE"].(string)
				if !ok || tableName == "" {
					tableName = "unknown"
				}

				mu.Lock()
				// Track columns for this table
				if _, exists := tableCols[tableName]; !exists {
					tableCols[tableName] = make(map[string]struct{})
				}

				// Check for new columns and alter table if needed
				for col := range obj {
					if _, seen := tableCols[tableName][col]; !seen {
						addColumnIfNotExists(db, tableName, col)
						tableCols[tableName][col] = struct{}{}
					}
				}

				// Ensure table exists
				createTableIfNotExists(db, tableName, obj, eventChan)
				// Insert row
				insertRow(db, tableName, obj, eventChan)
				mu.Unlock()
			}
			markCompletedInDB(runID, searchTerm, item.Index, db)
		}(item)
	}
	wg.Wait()
}

func createTableIfNotExists(db *sql.DB, table string, row map[string]interface{}, eventChan chan<- types.TableEvent) {
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
		if err2 == nil && count == 1 {
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

func addColumnIfNotExists(db *sql.DB, table, col string) {
	_, _ = db.Exec(fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN "%s" TEXT`, table, col))
}

func insertRow(db *sql.DB, table string, row map[string]interface{}, eventChan chan<- types.TableEvent) {
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
	if err == nil {
		eventChan <- types.TableEvent{
			TableName:   table,
			EventType:   "row_inserted",
			RowCount:    1,
			ColumnCount: len(row),
			SampleRow:   row,
		}
	}
}

func loadCompletedFromDB(runID string, db *sql.DB) map[int]struct{} {
	rows, err := db.Query(`SELECT query_index FROM batch_status WHERE run_id = ?`, runID)
	if err != nil {
		log.Printf("Warning: failed to load completed queries: %v", err)
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

func markCompletedInDB(runID string, searchTerm string, idx int, db *sql.DB) {
	_, err := db.Exec(`INSERT OR IGNORE INTO batch_status (run_id, search_term, query_index) VALUES (?, ?, ?)`, runID, searchTerm, idx)
	if err != nil {
		log.Printf("Warning: failed to mark query %d as completed: %v", idx, err)
	}
}
