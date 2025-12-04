# Migration: tenant-aware DIS Reader

Audience: teams already integrating DIS Reader (embedded or remote) who are adopting the tenant-aware refactor.

## What changed
- DB layer gained `MultiTenantDB`; domain/UI still use `DB` via `BindTenant`.
- Embedded and remote services are split: `NewDISReaderEmbedded` vs. `NewDISReaderRemote(...).With...().Build()`.
- Default tenant is no longer read from `bridge.tenantID`; supply via builder (`WithDefaultTenant`) or per call.
- `TestConnection` removed; use `PingService` + `PingDatabase`.
- Debug search moved under `internal/service` (embedded only) and still uses the DISReaderService interface.

## Modes & migration steps

### Embedded (single-tenant, local JDBC)
- Replace `disreader.NewDISReaderService` with `disreader.NewDISReaderEmbedded`.
- `TestConnection` removed; use `PingService` + `PingDatabase` if needed.
- No tenant handling required.

### Remote single-tenant (bridge + one agent)
- Agent: run one agent near DIS with its `tenantID` set. No inbound ports to DIS required.
- Bridge/server: the remote builder auto-starts gRPC/HTTP listeners unless you pass your own `WithGRPC`/`WithMux`. Defaults: gRPC :8443, HTTP :8080 (override via env). If you provide servers, start them yourself.
- Client: `disreader.NewDISReaderRemote(cfg, logger).WithDefaultTenant("<tenantID>").Build()`.
- Calls: use tenant-aware methods (Connect/Ping/UnitService etc.) with that tenant; if you set `WithDefaultTenant`, you can wrap helpers that omit the tenant.

### Remote multi-tenant
- Agents: one per tenant (or more for HA); each configured with its own `tenantID`.
- Bridge/server: same auto-start defaults; override with `WithGRPC`/`WithMux` or env ports.
- Client: `NewDISReaderRemote(...).Build()`; pass tenant per call (`Connect(ctx, tenant)`, `UnitService(tenant)`, etc.). Use `WithDefaultTenant` only for a fallback.
- Drop per-tenant service instances; bind per request via `UnitService(tenant)` or `BindTenant` if constructing services yourself.

### Config
- Stop using `bridge.tenantID` as a client default; supply a default via `WithDefaultTenant` or pass tenant explicitly. Keep bridge auth fields for agents (`clientID/secret`, allowedAgents, credentialFile).

### Tests/mocks
- Existing `database.DB` mocks still work for embedded paths. For tenant-aware code, mock `MultiTenantDB` or wrap with `BindTenant`.

### Bridge/agent notes
- Bridge listeners are auto-started by the remote builder (unless you provide servers). The standalone `cmd/bridge-server` binary has been removed.
- Agent config still requires `tenantID`; that is the authoritative tenant identity on the agent side.

## Verification checklist
- `go test ./...` passes.
- Remote callers supply tenant (or set a default via builder).
- No remaining references to `NewDISReaderService` or `TestConnection`.
