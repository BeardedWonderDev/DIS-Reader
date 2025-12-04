# Debug Search Service Reference

`RunDebugSearch` is DIS Reader’s exploratory search engine. It executes a curated set of SQL templates across DIS tables, streaming progress and events while persisting results as either an SQLite database or a CSV file. This page explains how to configure, call, and consume the service.

---

## 1. When to Use It

- Investigating whether a value (invoice number, unit number, serial, customer) exists anywhere in DIS.
- Building local artifacts (`.db` files) for offline queries.
- Surfacing example rows/columns for reverse engineering unknown tables.

It is **not** a replacement for structured services; use `UnitService` or `InvoiceService` for targeted reads.

---

## 2. API Signature

```go
func (s *DISReaderService) RunDebugSearch(
	searchTerm string,
	opts types.DebugSearchOptions,
	progressChan chan<- types.ProgressStatus,
	eventChan chan<- types.TableEvent,
)
```

### Parameters

| Name | Description |
|------|-------------|
| `searchTerm` | Plain string inserted into every query template (`'{{SEARCH}}'` placeholder). Escape/quote handled internally. |
| `opts` | Output preferences. Provide a `types.DebugSearchOptions` (see below). |
| `progressChan` | Caller-owned channel receiving progress updates; service closes it when done. |
| `eventChan` | Caller-owned channel receiving table/row events; also closed on completion. |

Channels are optional but recommended. Pass buffered channels to avoid blocking.

### `types.DebugSearchOptions`

| Field | Description |
|-------|-------------|
| `OutputMode` | `types.DebugSearchOutputSQLite` (default) or `types.DebugSearchOutputCSV`. Determines whether results are written to an SQLite database or a long-form CSV file. |
| `OutputPath` | Destination file path. Optional; when empty, the service picks `debug-search-<runID>.db` or `.csv` based on the chosen mode. |

---

## 3. Internals Overview

Source: `internal/debug.go`.

1. Loads `internal/queries.sql` (one SQL template per line).
2. Splits queries into batches of 20.
3. For each batch:
   - Replaces `'{{SEARCH}}'` with the user-provided term, quoted.
   - Executes queries concurrently (max goroutines = 5).
   - Streams rows via `database.DB.QueryWithSource` to capture `SRC_TABLE`.
   - Creates/extends SQLite tables matching source names, adding columns on the fly.
   - Inserts each row, recording metadata in `batch_status` to support resumability.
4. Sleeps 2 seconds between batches to reduce load (`delay` constant).
5. Writes each row to the configured sink: SQLite tables (schema created on the fly) or a CSV file with JSON-encoded rows.

---

## 4. Output Modes

### 4.1 SQLite (default)

When `OutputMode` is `sqlite`, the service keeps the existing behavior:

- `batch_status` – tracks completed queries per `run_id`.
  ```sql
  run_id TEXT, search_term TEXT, query_index INTEGER, completed_at DATETIME
  ```
- `query_errors` – created only if errors occur.
  ```sql
  run_id TEXT, query_index INTEGER, query TEXT, error_message TEXT, occurred_at DATETIME
  ```
- For each source table (value of `SRC_TABLE`), a like-named table is created/extended with `TEXT` columns as new fields appear.

This mode is resumable: reusing the same DB file allows the service to skip already completed queries.

### 4.2 CSV (long-form)

When `OutputMode` is `csv`, the service writes a single CSV document with the following columns:

| Column | Meaning |
|--------|---------|
| `run_id` | UUID for the current run (matches `ProgressStatus.RunID`). |
| `search_term` | The literal search term used to expand templates. |
| `table_name` | Source table reported by `SRC_TABLE` (or `unknown`). |
| `row_index` | Zero-based counter per table indicating row order. |
| `row_json` | JSON object containing the raw row data/columns.

CSV mode favors portability (open in spreadsheets or ship to other systems) at the cost of resumability—each run rewrites the file from scratch. Table and row events continue to stream so the UI/logs behave the same way.

---

## 5. Progress & Event Channels

### `types.ProgressStatus`

| Field | Meaning |
|-------|---------|
| `RunID` | UUID for the current run. |
| `TotalQueries` | Number of templates loaded. |
| `CompletedQueries` | Count of successfully executed templates (tracked via SQLite). |
| `PercentComplete` | Convenience percentage (`completed/total * 100`). |

### `types.TableEvent`

| Field | Meaning |
|-------|---------|
| `TableName` | Source table (from `SRC_TABLE`) or helper table name. |
| `EventType` | `table_created`, `row_inserted`, or `query_failed`. |
| `RowCount` | Rows inserted during the event (if applicable). |
| `ColumnCount` | Column count after schema update. |
| `SampleRow` | A map containing example row data or error details. |

Use these channels to update UI components, send metrics, or log to observability pipelines.

---

## 6. Example Usage

```go
progressCh := make(chan types.ProgressStatus, 1)
eventCh := make(chan types.TableEvent, 10)

opts := types.DebugSearchOptions{
	OutputMode: types.DebugSearchOutputSQLite,
	OutputPath: "debug_results.db",
}

go disSvc.RunDebugSearch("CASHC", opts, progressCh, eventCh)

for progressCh != nil || eventCh != nil {
	select {
	case prog, ok := <-progressCh:
		if !ok {
			progressCh = nil
			break
		}
		log.Printf("Run %s: %d/%d (%.1f%%)", prog.RunID, prog.CompletedQueries, prog.TotalQueries, prog.PercentComplete)
	case evt, ok := <-eventCh:
		if !ok {
			eventCh = nil
			break
		}
		log.Printf("[%s] %s rows=%d sample=%v", evt.TableName, evt.EventType, evt.RowCount, evt.SampleRow)
	}
}
```

Once the run completes, inspect `debug_results.db`:

```bash
sqlite3 debug_results.db '.tables'
sqlite3 debug_results.db 'SELECT * FROM FILEC_DMITEM LIMIT 10;'
```

To emit CSV instead, change `OutputMode` to `types.DebugSearchOutputCSV` and point `OutputPath` at something like `debug-results.csv`.

---

## 7. Configuration Tips

| Need | Approach |
|------|----------|
| Change batch size/parallelism | Adjust constants in `internal/debug.go` (`batchSize`, `parallelism`). |
| Persist output elsewhere | Set `opts.OutputPath` (or configure `debugSearch.defaultOutputPath`) to an absolute path. |
| Change default mode | Adjust `debugSearch.defaultOutputMode` in `disreader.yaml` (values: `sqlite`, `csv`) so the TUI preselects your preference. |
| Resume interrupted run | The `batch_status` table prevents duplicate work; rerun with the same DB file to pick up remaining queries. |
| Reduce throttling | Tweak the `delay` constant (2s by default). Ensure your DIS environment tolerates the load. |

---

## 8. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Channels never receive events | Caller forgot to drain/close channels | Ensure the select loop continues until both channels close. |
| SQLite DB locked errors | Multiple concurrent runs writing to same file | Use unique DB file per run or serialize executions. |
| Frequent `query_failed` events | Some templates invalid for your DIS schema | Check `SampleRow["error_message"]` and adjust `internal/queries.sql`. |
| High disk usage | Search hits many tables | Delete or archive old `.db` files after use. |

---

By combining progress updates, SQLite artifacts, and error tracking, the debug search service provides a safe sandbox for data discovery without writing ad-hoc SQL clients. Update this document as new features (e.g., configurable template files) become available.***
