package internal

import (
	"bufio"
	"bytes"
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

func (s DISReaderService) RunDebugSearch(searchTerm string, outputFile string) {
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

	for i, tmpl := range queryTemplates {
		if _, ok := completed[i]; ok {
			continue // Already completed, skip
		}
		query := strings.ReplaceAll(tmpl, "'{{SEARCH}}'", fmt.Sprintf("'%s'", searchTerm))
		batch = append(batch, queryItem{Index: i, Query: query})
		if len(batch) >= batchSize {
			s.runBatch(batch, outputFile)
			batch = []queryItem{}
			time.Sleep(delay)
			// Re-load completed.txt after each batch in case another process is marking completions
			completed = loadCompleted()
		}
	}

	// Final batch, if any remain
	if len(batch) > 0 {
		s.runBatch(batch, outputFile)
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

func (s DISReaderService) runBatch(batch []queryItem, outputFile string) {
	fmt.Printf("\nRunning batch of %d queries...\n", len(batch))
	outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open output JSON: %v", err)
	}
	defer outFile.Close()

	var wg sync.WaitGroup
	sem := make(chan struct{}, parallelism)
	var mu sync.Mutex
	var batchResults [][]map[string]interface{}

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
			var records []map[string]interface{}
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
				records = append(records, obj)
			}
			if len(records) > 0 {
				mu.Lock()
				batchResults = append(batchResults, records)
				mu.Unlock()
				markCompleted(item.Index)
			} else {
				fmt.Printf("Line %d: No data returned.\n", item.Index)
			}
		}(item)
	}
	wg.Wait()
	mu.Lock()
	if len(batchResults) > 0 {
		for _, records := range batchResults {
			if len(records) == 0 {
				continue
			}
			b, _ := json.Marshal(records)
			outFile.Write(b)
			outFile.Write([]byte("\n"))
		}
	}
	mu.Unlock()
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
