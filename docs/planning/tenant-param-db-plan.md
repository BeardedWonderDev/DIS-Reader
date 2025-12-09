# Tenant-aware DIS Reader plan (MultiTenantDB + binding adapter)

Goal: support multi-tenant remote mode explicitly while keeping existing domain/UI code unchanged by introducing a tenant-aware DB interface plus a binding adapter. Embedded mode remains unchanged; remote mode becomes explicit per tenant.

## Interfaces
- `types/common.go`
  - `DISReaderService`: (Connect/Disconnect/PingService/PingDatabase/RunDebugSearch/UnitService/InvoiceService/PartService); `TestConnection` removed.
  - `DISReaderRemote`: tenant-aware: Connect/Disconnect/PingBridge/PingAgent/PingService/PingDatabase and Unit/Invoice/Part service factories that accept a tenant string.
  - Optional combined marker: `type DISReader interface { DISReaderService; DISReaderRemote }`.

- `internal/database/database.go`
  - Keep existing `DB` interface unchanged (no tenant parameter) for domain/repo/UI usage.
  - Add `MultiTenantDB` interface: same methods as `DB` but with `tenant string` arg.
  - Add `BindTenant(mtdb MultiTenantDB, tenant string) DB` adapter (new file if cleaner) that implements `DB` by forwarding with the bound tenant.

## DB implementations
- `internal/database/remote.go`
  - Implement `MultiTenantDB` (no stored tenantID).
  - Methods accept `tenant string`; `sendJob` uses it for `registry.Pick(ctx, tenant)` and `ObserveQuery(tenant, agentID, latency)`; validate tenant non-empty; tag logs with tenant when logger present.
  - Constructor: `NewRemoteDB(registry bridge.AgentRegistry, logger *slog.Logger, maxRows int, maxBytes int64) MultiTenantDB`.

- `internal/database/database.go` (IBMi400)
  - Keep implementing `DB` (no tenant param). No behavior change.
  - If needed, provide a thin wrapper to satisfy `MultiTenantDB` by ignoring tenant (only if a caller needs IBMi400 behind MultiTenantDB; not expected for embedded flow).

## Service layer split (under `internal/service/`)
- Remove legacy `internal/service.go`.
- Add `embedded.go` implementing `types.DISReaderService`:
  - Holds `db database.DB` (IBMi400). Tenant binding fixed as "embedded" via `BindTenant` if ever needed.
  - Methods delegate to db; domain services constructed once with this db.

- Add `remote.go` implementing `types.DISReaderRemote`:
  - Holds `multiDB database.MultiTenantDB`, bridge registry/server/auth as needed.
  - Tenant-aware Connect/Disconnect/PingService/PingDatabase call `multiDB` with provided tenant.
  - PingBridge: bridge/server health; PingAgent: per-tenant agent availability (registry pick + heartbeat/job).
  - Domain service factories: `UnitService(tenant)` etc. build via `BindTenant(multiDB, tenant)`.

## Builders (`disreader/service.go`)
- `NewDISReaderEmbedded(cfg, logger) (types.DISReaderService, error)` — wires IBMi400.
- `NewDISReaderRemote(cfg, logger) *RemoteBuilder` — returns fluent builder only.
- `RemoteBuilder` fields: cfg, logger, auth, reg, grpcServer, mux, defaultTenant (string).
- Builder methods: `.WithAuth(...)`, `.WithRegistry(...)`, `.WithGRPC(...)`, `.WithMux(...)`, `.WithDefaultTenant(t string)`.
- `Build() (types.DISReaderRemote, error)` logic:
  1) Validate cfg.Bridge.mode == "remote" and required fields; set defaultTenant from builder (may be empty if caller will always pass tenant).
  2) Registry: provided or `bridge.NewInMemoryRegistry()`.
  3) Auth: provided or derived from cfg (file/allowed/clientID).
  4) multiDB := `NewRemoteDB(registry, logger, maxRows, maxBytes)`.
  5) Construct remote service with multiDB, defaultTenant, registry, bridge server (auth+registry), logger.
  6) If grpcServer provided, register bridge handlers. If mux provided, register health/metrics/pprof.
  7) Return service (no local JDBC process in remote mode).

## Config updates
- Remove client-side use of `BridgeConfig.TenantID` as a default tenant. Tenants are supplied via `WithDefaultTenant` or per-call.
- Update `types/config.go`, `config.go` defaults, and sample `disreader.yaml` to drop `bridge.tenantID` for client defaulting (keep other bridge fields).

## Domain/services/UI
- No signature changes. They continue to accept `database.DB` and stay tenant-agnostic; tenant is injected by service factories via `BindTenant`.

## Tests
- Update `internal/database/remote_test.go` to use `MultiTenantDB` signatures and include missing-tenant error case.
- Update integration tests referencing RemoteDB constructors and method signatures.
- Optionally add a remote service test ensuring tenant A failure doesn’t block tenant B (mock registry with two tenants).

## Docs
- Update `wiki/bridge_mode.md` and `wiki/dis_reader_service_integration.md` for tenant-aware usage and builders.
- Add migration guide `docs/planning/migration-to-tenant-aware-disreader.md`: embedded unchanged except TestConnection removal; remote uses builder + tenant-aware calls; note DB mock interface change (introduce MultiTenantDB or use BindTenant in tests).

## Migration note
- Remote callers must use tenant-aware methods (or set a default via builder). Embedded callers remain unchanged; tenant bound internally as "embedded".
