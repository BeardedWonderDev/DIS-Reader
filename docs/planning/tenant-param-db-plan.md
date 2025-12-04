# Tenant-parameter DB plan (single DB interface + tenant binding)

Goal: add explicit tenant parameters to the DB layer for remote mode, keep legacy DISReaderService APIs unchanged by binding the default tenant ("embedded"), and expose tenant-aware service methods for multi-tenant hosts with minimal churn.

## Interfaces (types/common.go, lines ~8-40)
- `DISReaderService`: updated (Connect/Disconnect/PingService/PingDatabase/RunDebugSearch/UnitService/InvoiceService/PartService). `TestConnection` was removed in favor of explicit PingService/PingDatabase.
- `DISReaderRemote` (already added):
  - Connect(ctx, tenant string), Disconnect(ctx, tenant string)
  - PingBridge(ctx)           // bridge connectivity
  - PingAgent(ctx, tenant string) // agent stream per tenant
  - PingService(ctx, tenant string) // JDBC process
  - PingDatabase(ctx, tenant string) // AS/400 DB
  - UnitService(tenant string), InvoiceService(tenant string), PartService(tenant string)

## DB interface (internal/database/database.go, lines ~16-55)
- Extend `DB` so every method accepts `tenant string`:
  - Connect(ctx, tenant string), Disconnect(ctx, tenant string)
  - PingService(ctx, tenant string), PingDatabase(ctx, tenant string)
  - Query(ctx, tenant string, ...), QueryRow(ctx, tenant string, ...)
  - Select(ctx, tenant string, ...), Get(ctx, tenant string, ...), QueryWithSource(ctx, tenant string, ...)
- Add `tenantBinding` adapter + `BindTenant(db DB, tenant string) DB` to inject a fixed tenant for callers that don’t pass one (domain services/UI). Adapter simply delegates, adding the tenant arg; no logic change.

## DB implementations
- RemoteDB (internal/database/remote.go)
  - Remove `tenantID` field; keep registry/logger/limits.
  - Constructor: `NewRemoteDB(registry bridge.AgentRegistry, logger *slog.Logger, maxRows int, maxBytes int64) DB` (no bound tenant).
  - All methods take `tenant string`; `sendJob` uses it in `registry.Pick(ctx, tenant)` and `ObserveQuery(tenant, agentID, latency)`.
  - Validate tenant is non-empty; return clear error if missing.
  - Logging/metrics include tenant when logger non-nil.

- IBMi400 (internal/database/database.go)
  - Update signatures to include `tenant string`; ignore tenant in behavior.
  - Preserve request-id logging and payload logic.

## Config updates
- Remove use of `BridgeConfig.TenantID` as a default tenant selector for clients; tenant defaults now come from `WithDefaultTenant` on the remote builder.
- Adjust `types/config.go`, `config.go` defaults, and sample `disreader.yaml` to drop `bridge.tenantID` as a client default (keep other bridge fields). Remote agent/bridge server configs remain, but client-side default tenant selection is no longer read from config.

## Service layer (split implementations under internal/service/)
- Remove legacy `internal/service.go`.
- Add `internal/service/embedded.go` implementing `types.DISReaderService` only:
  - Holds `db database.DB` (IBMi400 bound to defaultTenant="embedded").
  - Methods Connect/Disconnect/PingService/PingDatabase delegate to bound DB; Unit/Invoice/Part services constructed once with bound DB.
  - No tenant-aware methods are exposed here.
- Add `internal/service/remote.go` implementing `types.DISReaderRemote` only:
  - Holds `multiDB database.DB` (RemoteDB) and `bridge` components (registry/server/auth reload if needed).
  - Connect/Disconnect/PingService/PingDatabase accept tenant and call `multiDB` with that tenant.
  - PingBridge: bridge/server health check; PingAgent: per-tenant agent availability/heartbeat.
  - Domain service factories build tenant-bound services via `BindTenant(multiDB, tenant)`.
- Construction helpers (disreader/service.go)
  - Provide separate builders: `NewDISReaderEmbedded(cfg, logger)` returns `types.DISReaderService`; `NewDISReaderRemote(cfg, logger)` returns a `RemoteBuilder` (fluent only, no variadic options).
  - Builder type: `type RemoteBuilder struct { cfg *types.DISConfig; logger *slog.Logger; auth bridge.AgentAuthenticator; reg bridge.AgentRegistry; grpcServer *grpc.Server; mux *http.ServeMux; defaultTenant string }` with methods:
    - `.WithAuth(auth bridge.AgentAuthenticator)`
    - `.WithRegistry(reg bridge.AgentRegistry)`
    - `.WithGRPC(server *grpc.Server)`
    - `.WithMux(mux *http.ServeMux)`
    - `.WithDefaultTenant(t string)` (required if you want a legacy-style default tenant; config no longer supplies one)
  - `Build()` instantiates and returns `types.DISReaderRemote`; logic:
    1) Validate cfg.Bridge.mode == "remote" and required fields; require `defaultTenant` if legacy-style default calls are needed (else empty means none).
    2) Select registry: provided `reg` or `bridge.NewInMemoryRegistry()`.
    3) Select authenticator: provided `auth` or derived from cfg (file/allowed/clientID).
    4) Create tenant-aware DB: `remoteDB := database.NewRemoteDB(registry, logger, maxRows, maxBytes)`.
    5) Build remote service struct with `multiDB=remoteDB`, `defaultTenant` (may be empty), registry, bridge server (auth+registry), logger.
    6) If `grpcServer` provided, register bridge handlers; if `mux` provided, register health/metrics/pprof.
    7) Return service; remote mode does not start a local JDBC process.

## Domain services / repositories
- No signature changes. They keep taking `database.DB`. Services constructed with a bound DB receive tenant implicitly.

## Tests to adjust
- internal/database/remote_test.go: update constructor and calls with tenant arg; add missing-tenant error case.
- internal/integration/remote_smoke_test.go & parity_sqlite_test.go: adjust RemoteDB construction and method calls.
- Add a small service-level test to ensure tenant A failure doesn’t block tenant B (mock registry, two tenants).

## Docs
- Update `wiki/bridge_mode.md` and `wiki/dis_reader_service_integration.md` to explain tenant-aware calls, default tenant binding for legacy methods, and how PingBridge/PingAgent are used.
- Add a migration guide (`docs/planning/migration-to-tenant-aware-disreader.md`) outlining steps for existing integrators:
  - If using embedded mode: no code changes except re-vendoring; legacy API remains, but TestConnection is removed—use PingService+PingDatabase.
  - If using remote mode single-tenant: construct via the remote builder `NewDISReaderRemote(cfg, logger).Build()`; supply `WithDefaultTenant(...)` if you still want a default; pass tenant when calling.
  - If adding multi-tenant: replace multiple service instances with per-call tenant-aware methods or per-tenant factories (`UnitService(tenant)` etc.).
  - Update any direct references to old `internal/service.go`; point to new embedded/remote builders.
  - Note the new DB interface signature (tenant param) for any custom DB mocks; update tests accordingly.

## Migration note
- Remote callers must use tenant-aware methods. Embedded callers remain unchanged; tenant is bound internally as "embedded".
