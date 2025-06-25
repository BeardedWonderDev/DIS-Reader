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
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed queries.sql
var queriesSQL string

const (
	batchSize     = 20
	delay         = 2 * time.Second
	parallelism   = 5
	completedFile = "completed.txt"
)

type queryItem struct {
	Index int
	Query string
}

var completedMu sync.Mutex

func (s DISReaderService) RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent) {
	// Load query templates from embedded file
	var queryTemplates []string
	scanner := bufio.NewScanner(strings.NewReader(queriesSQL))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			queryTemplates = append(queryTemplates, line)
		}
	}

	completed := loadCompleted()
	var batch []queryItem

	// Open or create the SQLite database
	db, err := sql.Open("sqlite3", sqliteDBFile)
	if err != nil {
		log.Fatalf("Failed to open SQLite DB: %v", err)
	}
	defer db.Close()

	tableCols := make(map[string]map[string]struct{}) // Table -> set of columns

	for i, tmpl := range queryTemplates {
		if _, ok := completed[i]; ok {
			continue // Already completed, skip
		}
		query := strings.ReplaceAll(tmpl, "'{{SEARCH}}'", fmt.Sprintf("'%s'", searchTerm))
		batch = append(batch, queryItem{Index: i, Query: query})
		if len(batch) >= batchSize {
			s.runBatchToSQLite(batch, db, tableCols, eventChan)
			batch = []queryItem{}
			time.Sleep(delay)
			// Re-load completed.txt after each batch in case another process is marking completions
			completed = loadCompleted()
			progressChan <- types.ProgressStatus{
				TotalQueries:     len(queryTemplates),
				CompletedQueries: len(completed),
				PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
			}
		}
	}

	// Final batch, if any remain
	if len(batch) > 0 {
		s.runBatchToSQLite(batch, db, tableCols, eventChan)
		completed = loadCompleted()
		progressChan <- types.ProgressStatus{
			TotalQueries:     len(queryTemplates),
			CompletedQueries: len(completed),
			PercentComplete:  float64(len(completed)) / float64(len(queryTemplates)) * 100,
		}
	}

	// After all batches, check if all queries are completed
	completed = loadCompleted()
	if len(completed) == len(queryTemplates) {
		err := os.Remove(completedFile)
		if err != nil {
			log.Printf("Warning: could not remove completed file: %v", err)
		} else {
			fmt.Println("All queries completed. Progress file deleted.")
		}
	} else {
		remaining := len(queryTemplates) - len(completed)
		fmt.Printf("Batch processing done. %d queries remain incomplete.\n", remaining)
	}
}

func (s DISReaderService) runBatchToSQLite(batch []queryItem, db *sql.DB, tableCols map[string]map[string]struct{}, eventChan chan<- types.TableEvent) {
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
				fmt.Fprintf(os.Stderr, "Error: %v\nOutput:\n%s\n", err, output)
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
			markCompleted(item.Index)
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

func loadCompleted() map[int]struct{} {
	data, err := os.ReadFile(completedFile)
	completed := make(map[int]struct{})
	if err != nil {
		return completed
	}
	for _, line := range strings.Split(string(data), "\n") {
		val := strings.TrimSpace(line)
		if val == "" {
			continue
		}
		if idx, err := strconv.Atoi(val); err == nil {
			completed[idx] = struct{}{}
		}
	}
	return completed
}

func markCompleted(idx int) {
	completedMu.Lock()
	defer completedMu.Unlock()
	f, err := os.OpenFile(completedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: could not mark query %d as completed: %v", idx, err)
		return
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", idx))
}
