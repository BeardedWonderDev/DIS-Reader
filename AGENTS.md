# Repository Guidelines

## Project Structure & Module Organization
- `main.go` creates the config, DIS reader service, and root UI; hook new features here.
- `cmd/debug/` houses opt-in diagnostic view modules that plug into the UI via `types.ViewModule`.
- `internal/` holds production logic: `service/` orchestrates JDBC + parsing, `database/` handles SQLite helpers, `unit/` models invoices/items, and `ui/` keeps shared keybindings/layouts.
- `ui/` renders the TUI; `ui/utils/` offers reusable primitives for styling and modal control.
- `disreader/` and `types/` expose shared constructors/models, while `images/`, `wiki/`, and the checked-in `.db` fixtures serve docs and disposable local data.

## Build, Test, and Development Commands
- `go run ./...` runs the interactive DIS Reader with the current config.
- `go build ./cmd/...` outputs binaries per command; set `GOOS`/`GOARCH` for cross-builds.
- `go test ./...` executes all suites; keep unit tests fast and deterministic.
- `go vet ./...` (or `staticcheck`) surfaces obvious mistakes and must pass before reviews that touch parsing or database code.

## Coding Style & Naming Conventions
- Use standard Go formatting: tabs, `gofmt` (or `go fmt ./...`) pre-commit, camelCase identifiers, and short lowercase package names.
- Organize imports as stdlib / third-party / internal, wrap errors with `%w`, and log only through `slog` or injected logger interfaces.
- Keep files focused; split widgets, database accessors, and config helpers instead of growing monoliths.

## Testing Guidelines
- Co-locate tests as `*_test.go` files using table-driven subtests; mock JDBC responses via the sqlite fixtures.
- Share setup via `t.Helper()` utilities and document any new sample data under `wiki/fixtures.md`.
- Cover parsing, conversion, and UI interaction boundaries before merging.

## Commit & Pull Request Guidelines
- Mirror the existing short, imperative history (`fix parsing`, `started invoice items`), keeping subject lines ≤72 chars and describing observable behavior.
- Each PR needs a concise summary, linked issue (if any), manual test notes or recordings for UI work, and explicit mention of schema/config changes.
- Request review only after `go test ./...` plus lint checks pass; attach screenshots whenever anything under `ui/` changes.

## Configuration & Data Notes
- Runtime settings live in `disreader.yaml`; override any field via env vars prefixed with `DISREADER_` (e.g., `DISREADER_DISCONFIG_HOST`).
- `disConfig.jdbcConfig` controls the Java bridge; update `javaPath` when pointing at a custom JDK.
- Treat bundled `.db` files as disposable fixtures—never commit production data and refresh docs whenever new seeds are added.
