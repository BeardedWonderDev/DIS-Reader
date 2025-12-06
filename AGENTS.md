# AGENTS (Local Rules)

## Global Rules and Required Reading
- Always read and apply ~/.codex/AGENTS.md before these local rules.

## Repository Shape
- Go 1.24 project with an embedded Java JDBC runner (Maven project in `java-runner/`, embedded jar in `internal/runnerjar/`).
- Entry points: `main.go` (embedded DIS reader + tview UI), library package `disreader/` (embedded vs. remote builder), and `cmd/agent` (remote JDBC agent daemon).
- Domain code lives in `internal/{unit,parts,invoices}/` with repository + service layers; DTOs, interfaces, config, and themes live in `types/`.
- Infra/support: `internal/service/` (embedded + remote orchestrators, debug search), `internal/database/` (DB interfaces, JDBC runner glue, remote MultiTenant DB, tenant binding), `internal/bridge/` (gRPC bridge server, registry, auth, generated proto), `internal/logging/`, `internal/utils/`.
- UI is terminal-only, built with `rivo/tview` under `ui/`; debug search outputs land under `debug-output/` or the configured path.

## Dependency & Architecture Boundaries
- Domain services depend on `database.DB` (single-tenant abstraction). Remote mode wraps `database.MultiTenantDB` via `database.BindTenant`; keep domain/UI code unaware of multi-tenant plumbing.
- `disreader.NewDISReaderEmbedded` returns `types.DISReaderService` (backed by `EmbeddedService`). `disreader.NewDISReaderRemote(...).Build()` returns tenant-aware `types.DISReaderRemote` backed by `RemoteService` + `bridge.Server` + `database.RemoteDB`.
- `internal/bridge` owns agent auth, registry, heartbeats, and gRPC protocol; do not bypass it for DB calls—route remote DB work through the registry/RemoteDB.
- Logging is `slog` throughout; prefer helpers in `internal/logging` and `types.LogError`. Avoid fmt prints for structured logs.
- UI changes should respect the theme/look via `types.Theme` and `viewerApp` helpers; wire new widgets into the existing navigation/focus handling.

## Configuration & Environment
- UI/embedded config loader (`config.go`) reads `disreader.yaml` then merges `DISREADER_*` env vars (dot-flattened, e.g., `DISREADER_DISCONFIG_HOST`).
- Agent config (`cmd/agent`) reads `agent.yaml` then `DISAGENT_*` env vars; TLS defaults to enabled. Use `tls.insecureSkipVerify` only for development.
- Bridge settings (`types.BridgeConfig`): `mode` (`embedded`|`remote`), `serverURL` + credentials, `maxRowsPerQuery`, `maxResultBytes`, `credentialFile` or `allowedAgents`, `pprofEnabled`/`pprofPath`, Loki block, default tenant, and auto-connect on register.
- Default ports when the remote builder self-hosts servers: gRPC `DISREADER_BRIDGE_PORT` (default 8443) and HTTP/health/metrics `DISREADER_BRIDGE_HTTP_PORT` (default 8080).
- Agents cap rows via `DISAGENT_MAX_ROWS` (default 1000); remote DB also enforces `maxRowsPerQuery` and `maxResultBytes`.

## Data Access & Domain Conventions
- Build SQL in repositories; keep table names/constants near the repo (e.g., `partsTable`), and use quoting/format helpers (`quoteLiteral`, `formatFilterValue`, `ListParamParser` implementations) to avoid injection.
- Use `types.ListParams` + parser structs (`InvoiceListParamParser`, `WholeGoodsInvoiceListParamParser`, etc.) for filter/sort validation; do not accept raw column names from callers.
- AS/400 date handling is centralized in `database.as400DateHook` / `As400DateExpr`; reuse these for new date fields instead of ad hoc parsing.
- Debug Search queries live in `internal/service/queries.sql` with `{{SEARCH}}` placeholders; keep that format and the batch/SQLite-or-CSV assumptions.
- Tenant identity must be supplied on remote calls; only omit when a default tenant was set via the remote builder. Do not rely on `bridge.tenantID` as an implicit default.

## Bridge, Agent, and Remote Mode
- Remote mode requires `bridge.mode=remote`; embedded mode runs JDBC locally via `internal/runnerjar`.
- Agent auth precedence: custom authenticator passed to the remote builder → `credentialFile` (YAML/JSON list) → `allowedAgents` / top-level `clientID+clientSecret`.
- Health/metrics endpoints are registered via `RemoteService.RegisterHealth` onto an HTTP mux; pprof is gated by `bridge.pprofEnabled`.
- `UpdateAgentConfig` pushes `proto.AgentConfig` (runtime + Loki) to connected agents and updates defaults; if you include a new client secret, the bridge mutates auth via `MutableAuthenticator`.

## Java Runner & Protobuf
- JDBC runner source: `java-runner/src/main/java/JDBCRunner.java`; embedded artifact: `internal/runnerjar/dis-runner-0.1.1.jar`. If runner code changes, rebuild the jar (Maven `java-runner/pom.xml`), replace the embedded file, and keep versions aligned.
- Bridge proto source: `proto/bridge.proto`; generated code lives in `internal/bridge/proto/` and is re-exported via `bridgeproto/alias.go`. After proto edits, run `protoc --go_out=. --go-grpc_out=. proto/bridge.proto` and commit regenerated files.

## Testing Expectations
- Gates: `go test ./...` must pass; run `go vet ./...` (or `staticcheck`) for DB/bridge changes.
- Keep tests fast and hermetic: prefer fakes (`fakeDB` in `internal/service/debug_test.go`), in-process gRPC with the in-memory registry, and temporary directories. Avoid requiring a real DIS system or Java process.
- Table-driven tests are standard in domain packages; use `testify/require` where already in use. When parsing env/config, isolate Viper instances to avoid global bleed.
- Integration tests that open sockets should bind to `127.0.0.1:0` or set port envs to `0` to avoid collisions.

## Build & Run
- `go run ./...` starts the embedded DIS reader with the tview UI (uses `disreader.NewDISReaderEmbedded`).
- `go build ./cmd/...` builds the binaries (UI + agent). Set `GOOS/GOARCH` for cross-builds; CGO is required for SQLite.
- `go test ./...` for all suites; keep `go fmt ./...` clean before sending changes.

## Extending Features Safely
- New domain features: add DTOs/interfaces in `types/`, create `internal/<domain>/model.go` + `repository.go` (SQL + validation) + `service.go` (mapping to specs), and wire into the appropriate `types` interface. Ensure remote mode works by relying solely on `database.DB`.
- UI additions: compose within `viewerApp` (pages, focusables, key handling) and respect `types.Theme` colors; place new debug outputs under the existing debug-output path logic.
- Bridge/agent changes: keep registry and auth boundaries intact; if adding new job kinds, update proto, server dispatch, agent `executeJob`, and adjust tests in `internal/integration` and `cmd/agent`.
- Config changes: add Viper defaults + env bindings in the relevant loader (`config.go` or `cmd/agent` loader) and document in `configs/*.example` plus the wiki where appropriate.
- Observability: prefer slog with `logging.CommonAttrs`, `logging.SQLHint`, and `logging.DurationAttr`; avoid logging secrets (credentials, DIS passwords, client secrets).

## Documentation & Artifacts
- Update wiki docs (`wiki/bridge_mode.md`, `wiki/remote_builder_example.md`, etc.) when changing bridge/agent behavior, config keys, or tenancy expectations.
- Packaging manifests live under `packaging/` (systemd, launchd, Windows installers); keep them in sync when adding binaries or flags.
- Example configs: `configs/agent.yaml.example`, `packaging/examples/agent.yaml`, `packaging/examples/bridge_agents.yaml`; refresh them when new config fields are introduced.
