# DIS Reader Service Integration Guide

This guide explains how to embed the DIS Reader backend service into any Go project (CLI, API, or worker) without the TUI layer. It covers configuration, lifecycle management, and every public service/method exposed by the library.

---

## 1. Getting Started

### 1.1 Install / Upgrade the Module

```bash
go get github.com/BeardedWonderDev/DIS-Reader@latest
```

or add the module to your `go.mod` and run `go mod tidy`.

### 1.2 Packages of Interest

| Package | Purpose |
|---------|---------|
| `github.com/BeardedWonderDev/DIS-Reader/disreader` | Factory entrypoint for the service. |
| `github.com/BeardedWonderDev/DIS-Reader/types` | Shared configuration structs, service interfaces, DTOs, and helpers. |
| `github.com/BeardedWonderDev/DIS-Reader/internal/...` | Concrete implementations (import only if you control the repository; external consumers should stick to exported packages). |

---

## 2. Configuration Requirements

The service expects a fully populated `types.DISConfig` (or `types.DISUIConfig` if you reuse the UI config helper). Mandatory fields:

| Field | Description |
|-------|-------------|
| `Host` | IP or hostname of the DIS / AS400 server. |
| `User`, `Password` | Credentials with access to the required FILEC tables. |
| `JDBCConfig.JavaPath` | Path to the Java executable (`java`). Defaults to `java` if unset. |
| `JDBCConfig.JDBCPort` | Local TCP port for the embedded JDBC runner (e.g., `8888`). |

Optional tuning:
| Field | Purpose |
|-------|---------|
| `LogLevel` | Sets `slog.Level` for runner + service logs (default `INFO`). |
| `MaxIdleConnections`, `MaxOpenConnections` | Reserved for future DB pooling knobs. |
| `JDBCConfig.ClassDir`, `JDBCConfig.JarPath` | Normally auto-populated; override only for advanced deployments. |

### 2.1 Using the Built-in Config Loader (Optional)

The CLI’s `NewConfig()` helper (`config.go`) reads `disreader.yaml`, applies defaults, and merges any `DISREADER_*` env vars. You can reuse it:

```go
cfg := NewConfig()        // returns *types.DISUIConfig
disCfg := cfg.DIS         // *types.DISConfig
```

Or construct `types.DISConfig` manually if you already have a configuration system.

---

## 3. Creating and Managing the Service

### 3.1 Basic Initialization

```go
package main

import (
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func main() {
	cfg := &types.DISConfig{
		Host:     "10.0.0.5",
		User:     "DISUSER",
		Password: "secret",
		LogLevel: slog.LevelInfo,
		JDBCConfig: &types.JDBCConfig{
			JavaPath: "java",
			JDBCPort: "8888",
		},
	}

	disSvc, err := disreader.NewDISReaderService(cfg, nil) // nil logger -> pretty slog handler
	if err != nil {
		panic(err)
	}
	defer disSvc.Shutdown()

	// Optional: attach your own slog logger
	// disSvc.SetLogger(myLogger)

	// Test connectivity before issuing queries
	if err := disSvc.TestConnection(context.Background()); err != nil {
		panic(err)
	}

	// Use unit/invoice services (see Sections 4 & 5)
}
```

### 3.2 Lifecycle Methods

| Method | Usage |
|--------|-------|
| `Connect(ctx)` | Explicitly establish AS400 connectivity. Usually called automatically if host/user/password are set during initialization. |
| `Disconnect(ctx)` | Close the TCP connection to the Java runner. |
| `TestConnection(ctx)` | Verifies both the Java runner and the AS400 connection. |
| `AttachShutdownHook()` | Spawns a SIGINT/SIGTERM listener that gracefully stops the runner (useful inside long-lived daemons). |
| `Shutdown()` | Stops the JDBC runner, cleans up temp directories, and releases resources. Always call this (or `defer`) when finished. |
| `SetLogger(logger)` / `GetLogger()` | Replace or fetch the slog logger the service uses. |
| `GetConfig()` | Returns the underlying `*types.DISConfig` (mutable). |

---

## 4. Unit Service API

For a field-by-field breakdown, sorting/filtering matrix, and troubleshooting tips, see [wiki/unit_service.md](./unit_service.md).

Access via:

```go
unitSvc := disSvc.UnitService()
```

### 4.1 Fetch a Single Unit

```go
spec, err := unitSvc.GetByUnitNumber(ctx, "123456")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Unit %s located at %s\n", spec.UnitID, spec.Location)
```

`UnitSpec` (defined in `types/units.go`) includes identity, make/model, financials, flooring/rental info, and customer metadata. All fields are JSON-tagged for API responses.

### 4.2 List Units with Filters

The service expects a fully parsed `types.ListParams`. You can build one manually or via helper functions in `types/list.go`. Recommended flow:

```go
parser := types.UnitListParamParser{}

sortBy, _ := types.ParseSortBy("unit", parser)
sortOrder, _ := types.ParseSortOrder("ASC")

lp := types.ListParams{
	Limit:     100,
	SortBy:    sortBy,
	SortOrder: sortOrder,
	Query:     "TRACTOR", // full-text search across UNIT/MAKE/MODEL
}

specs, err := unitSvc.ListUnits(ctx, lp)
```

To add precise filters, populate `lp.Filters` with `types.Filter` values. Use `types.ParseFilter` alongside `UnitListParamParser` to validate and map logical field names to SQL columns. Special handling:
- Date filters (`indate`, `soldat`) must use `time.Time`.
- Numeric filters (cost, amount) should use numeric types; they’re interpolated directly.

### 4.3 Notes

- The repository internally transforms AS/400 numeric date columns via `database.As400DateExpr`.
- `UnitRepository.ListUnits` currently applies `LIMIT` directly; cursor pagination helpers exist in `types/list.go` if you need them.

---

## 5. Invoice Service API

For exhaustive field descriptions and filter options, refer to [wiki/invoice_service.md](./invoice_service.md).

Access via:

```go
invoiceSvc := disSvc.InvoiceService()
```

### 5.1 Fetch a Specific Invoice Line

```go
item, err := invoiceSvc.GetItem(ctx, "INVOICE123", "PC0070")
if err != nil {
	log.Fatal(err)
}
fmt.Printf("%s x %.0f @ %0.2f\n", item.PartNumber, item.Quantity, item.Price)
```

`InvoiceItemSpec` (from `types/invoices.go`) exposes every decoded field from `FILEC.IAH`, including audit columns, classification, pricing, tax flags, and optional user fields.

### 5.2 List Invoice Items

Similar to units, but use `types.InvoiceListParamParser` to validate filters/sorts.

```go
parser := types.InvoiceListParamParser{}
sortBy, _ := types.ParseSortBy("postingDate", parser)

filter, _ := types.ParseFilter(types.Filter{
	Field:    "customer",
	Operator: "=",
	Value:    "C01153",
}, parser)

lp := types.ListParams{
	Limit:   200,
	SortBy:  sortBy,
	Filters: []types.Filter{filter},
}

items, err := invoiceSvc.ListItems(ctx, lp)
```

Special handling:
- `postingDate` filters require `time.Time`.
- Monetary fields (`cost`, `price`) should be numeric.
- Free-text `Query` searches document, part number, and customer number simultaneously (wildcards handled internally with proper escaping).

---

## 6. Debug Search API

Detailed behavior (batching, SQLite schema, events) is documented in [wiki/debug_search_service.md](./debug_search_service.md).

`DISReaderService` also exposes a long-running batch search that fans out curated SQL templates across DIS tables.

```go
progressCh := make(chan types.ProgressStatus)
eventCh := make(chan types.TableEvent)

opts := types.DebugSearchOptions{
	OutputMode: types.DebugSearchOutputSQLite,
	OutputPath: "debug_results.db",
}

go disSvc.RunDebugSearch("C01153", opts, progressCh, eventCh)

for {
	select {
	case prog, ok := <-progressCh:
		if !ok {
			progressCh = nil
			continue
		}
		fmt.Printf("Run %s: %d/%d (%0.1f%%)\n", prog.RunID, prog.CompletedQueries, prog.TotalQueries, prog.PercentComplete)
	case evt, ok := <-eventCh:
		if !ok {
			eventCh = nil
			continue
		}
		fmt.Printf("[%s] %s rows=%d sample=%v\n", evt.TableName, evt.EventType, evt.RowCount, evt.SampleRow)
	}
	if progressCh == nil && eventCh == nil {
		break
	}
}
```

Details:
- `searchTerm` replaces `{{SEARCH}}` placeholders inside every query defined in `internal/queries.sql`.
- `opts.OutputPath` is created/updated with result tables. Use SQLite clients for the default mode or open the CSV export in spreadsheets/ETL tools when `OutputMode` is `csv`.
- `opts.OutputPath` is created/updated with result tables. Use SQLite clients for the default mode or open the CSV export in spreadsheets/ETL tools when `OutputMode` is `csv`.
- Default values for both fields can be set globally via `debugSearch.defaultOutputMode` / `debugSearch.defaultOutputPath` in `disreader.yaml` so the TUI modal starts with sensible choices.
- `ProgressStatus` reports aggregate progress per run ID.
- `TableEvent` describes new tables, inserted rows, or query failures (the latter populate a `query_errors` table with context).
- The engine batches queries (20 at a time) and throttles them (2s delay) to avoid overwhelming DIS.

---

## 7. Logging

- The service uses `slog`. If you pass `nil` to `NewDISReaderService`, it defaults to `prettylog.NewHandler`.
- Provide your own logger to integrate with existing observability pipelines:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
disSvc, _ := disreader.NewDISReaderService(cfg, logger)
```

- The embedded JDBC runner writes JSON lines to stdout/stderr; these are parsed and re-emitted via `slog` with attributes like `status`, `event`, `request_id`.

---

## 8. Error Handling & Cleanup Tips

1. Always `Shutdown()` (or `defer` it) to terminate the Java process and delete temporary files.
2. Wrap long-running calls with contexts that carry deadlines/timeouts to prevent stuck TCP connections.
3. If you override `JDBCConfig.JDBCPort`, ensure no other process uses the same port.
4. When updating credentials at runtime, mutate `disSvc.GetConfig().Host/User/Password` and call `Connect()` again.
5. On authentication failures, inspect the slog logs for exact error messages returned by the JDBC runner/AS400.

---

## 9. Summary Checklist

- [ ] Provide a valid `types.DISConfig` (host/user/password/java path/port).
- [ ] Call `disreader.NewDISReaderService`.
- [ ] Optionally run `TestConnection` or `Connect`.
- [ ] Use `UnitService` / `InvoiceService` for structured data access.
- [ ] (Optional) Use `RunDebugSearch` for exploratory discovery.
- [ ] Stop the service with `Shutdown()` or `AttachShutdownHook`.

With these steps, you can leverage DIS Reader’s battle-tested JDBC orchestration, mapstructure decoding, and domain services inside any Go application—without depending on the TUI front-end. Update this document whenever new services or public methods become available.
