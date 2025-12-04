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
   - Bridge server: run `cmd/bridge-server` (or embed) with `bridge.mode=remote`.
   - Agent: run one agent near DIS with its `tenantID` set. No inbound ports to DIS required.
   - Client: construct via `disreader.NewDISReaderRemote(cfg, logger).WithDefaultTenant("<tenantID>").Build()`.
   - Calls: use tenant-aware methods (Connect/Ping/UnitService etc.) with that tenant. With a default tenant set, you may wrap your own helpers to avoid passing it each time.

3) Remote multi-tenant users
   - Bridge server: same as above; can host multiple agents/tenants concurrently.
   - Agents: one per tenant (or more for HA); each configured with its own `tenantID`.
   - Client: `NewDISReaderRemote(...).Build()`; prefer passing tenant per call (`Connect(ctx, tenant)`, `UnitService(tenant)`, etc.). Use `WithDefaultTenant` only if you want a fallback for legacy wrappers.
   - Drop per-tenant DISReader instances; bind per request via `UnitService(tenant)` or `BindTenant` if constructing services yourself.

4) Config
   - Remove reliance on `bridge.tenantID` as a client default; keep bridge auth fields intact for agents.

5) Tests/mocks
   - If you mocked `database.DB`, keep as-is. To mock tenant-aware paths, use `MultiTenantDB` or wrap with `BindTenant`.

6) Bridge/agent
   - Bridge server is now auto-started by the remote builder unless you supply your own gRPC/mux. The standalone `cmd/bridge-server` binary has been removed.
   - Agent config still requires `tenantID`; that is the authoritative tenant identity on the agent side.

## Verification checklist
- `go test ./...` passes.
- Remote callers supply tenant (or set a default via builder).
- No remaining references to `NewDISReaderService` or `TestConnection`.
