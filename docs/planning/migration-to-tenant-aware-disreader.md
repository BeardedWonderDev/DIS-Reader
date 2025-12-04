# Migration: tenant-aware DIS Reader

Audience: teams already integrating DIS Reader (embedded or remote) who are adopting the tenant-aware refactor.

## What changed
- DB layer gained `MultiTenantDB`; domain/UI still use `DB` via `BindTenant`.
- Embedded and remote services are split: `NewDISReaderEmbedded` vs. `NewDISReaderRemote(...).With...().Build()`.
- Default tenant is no longer read from `bridge.tenantID`; supply via builder (`WithDefaultTenant`) or per call.
- `TestConnection` removed; use `PingService` + `PingDatabase`.
- Debug search moved under `internal/service` (embedded only) and still uses the DISReaderService interface.

## Steps
1) Embedded users
   - Replace `disreader.NewDISReaderService` with `disreader.NewDISReaderEmbedded`.
   - Remove `TestConnection` calls; use `PingService` + `PingDatabase` if needed.
   - No tenant changes required.

2) Remote single-tenant users
   - Replace constructor with fluent builder: `disreader.NewDISReaderRemote(cfg, logger).WithDefaultTenant("your-tenant").Build()`.
   - Call tenant-aware methods (Connect/Ping/UnitService etc.) with the tenant; if you set `WithDefaultTenant`, you can bind services via `UnitService(tenant)` using that default.

3) Remote multi-tenant users
   - Use `NewDISReaderRemote(...).Build()` and pass tenant per call (`Connect(ctx, tenant)`, `UnitService(tenant)`, etc.).
   - Drop any per-tenant service instances; instead bind via `UnitService(tenant)` or `BindTenant` if you construct services yourself.

4) Config
   - Remove reliance on `bridge.tenantID` as a client default; keep bridge auth fields intact for agents.

5) Tests/mocks
   - If you mocked `database.DB`, keep as-is. To mock tenant-aware paths, use `MultiTenantDB` or wrap with `BindTenant`.

6) Bridge/agent
   - No change to agent behavior. Bridge server now constructed via remote builder in `cmd/bridge-server`.

## Verification checklist
- `go test ./...` passes.
- Remote callers supply tenant (or set a default via builder).
- No remaining references to `NewDISReaderService` or `TestConnection`.
