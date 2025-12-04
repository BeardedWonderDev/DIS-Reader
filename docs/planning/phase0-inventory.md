# Phase 0 — Inventory & Boundaries

Date: 2025-12-04  
Branch: `remote-agent`

## Stability Contracts
- `database.DB` interface (`internal/database/database.go`) must remain source-compatible; callers expect:
  - lifecycle: `StartJDBCRunner`, `StopJDBCRunner`, `Connect`, `Disconnect`, `PingService`, `PingDatabase`
  - data: `Get`, `Query`, `QueryRow`, `Select`, `QueryWithSource`
- `DISReaderService` (`internal/service.go`) must continue to expose `UnitService`, `InvoiceService`, `PartService` via `types.DISReaderService`.
- UI and domain services must be oblivious to transport choice (embedded vs remote).

## New Components (added, not replacing)
- `remoteDB` implements `database.DB` by proxying calls to a bridge server.
- Bridge server/registry layers live under `internal/bridge/`.
- Agent daemon is a new binary; embedded mode remains default.

## Footguns to Avoid
- Do not alter public structs in `types/` except to add optional bridge config with defaults.
- Do not change existing SQL or repository behavior.
- Ensure mapstructure/date decode logic is unchanged for remote rows (reuse `types.ResultRow`).
- Keep logging via `slog`; avoid introducing global loggers.

## Acceptance for Phase 0
- No functional code changes.
- Boundaries and constraints captured for downstream phases.
