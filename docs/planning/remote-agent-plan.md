# Remote Agent Architecture Plan

Owner: DIS Reader  
Branch: `remote-agent`  
Scope: Move the JDBC runner into a LAN “agent” that dials out to the cloud-hosted DIS Reader service (no inbound ports required), while keeping the existing DIS Reader API and UI intact.

## Objectives
- Preserve the current `database.DB` contract so `DISReaderService`, domain services, and UI modules remain unchanged.
- Add a remote mode where database operations are proxied over a secure, outbound-only channel.
- Ship a deployable agent binary that runs the embedded JDBC bridge inside the LAN and communicates with the cloud via gRPC over TLS.
- Keep configuration backwards compatible; default to embedded mode.

## Architecture Overview
- **Bridge Protocol (gRPC, bidirectional stream):** Agents connect to the cloud service, authenticate, send heartbeats, receive job requests (SQL), and stream results back.
- **Agent Registry:** Tracks connected agents per tenant; selects an available agent for each job.
- **Remote DB Adapter:** Implements `database.DB` by sending `JobRequest` messages over the registry to a live agent; mirrors `PingService`, `PingDatabase`, `Connect`, `Disconnect`, and query methods.
- **Agent Daemon:** Headless Go binary that embeds the existing JDBC runner (`internal/database` IBMi400) and serves requests from the cloud stream.
- **Config Toggle:** `bridge.mode: embedded|remote` with server URL, credentials, and TLS options. Defaults to `embedded`.

## Deliverables by Phase

### Phase 0 — Inventory & Boundaries
- Document current data paths and interfaces that must stay stable (`database.DB`, `DISReaderService`, domain services).
- Identify new packages and files to introduce (see below).
- No functional change; commit documentation of boundaries.

### Phase 1 — Protocol & Models
- Add `proto/bridge.proto` defining:
  - `AgentService.Connect` (bi-directional stream)
  - Messages: `AgentHello`, `JobRequest`, `JobResult`, `Heartbeat`, `LogEntry`
  - `ResultRow` map for row data; `Status` enum for ok/error/done
- Generate Go stubs into `internal/bridge/proto` via `protoc` (tracked output).

### Phase 2 — Bridge Server + Remote DB
- New package `internal/bridge/server`:
  - gRPC handler implementing `AgentService`.
  - `AgentRegistry` (in-memory): register/unregister, choose available agent, track heartbeats.
  - Minimal auth interface `AgentAuthenticator` (tenant resolution).
- New package `internal/database/remote.go`:
  - `remoteDB` struct implementing `database.DB`; uses registry to dispatch `JobRequest` and await `JobResult`.
  - Supports `Query`, `QueryRow`, `Select`, `Get`, `QueryWithSource`, `PingService`, `PingDatabase`, `Connect`, `Disconnect`.
- Logging via existing `slog` helpers.

### Phase 3 — Agent Daemon
- New command `cmd/agent/main.go`:
  - Reads `agent.yaml` (or env) for `serverURL`, `clientID`, `clientSecret`, `tenantID`, TLS opts.
  - Runs embedded JDBC (`IBMi400`) locally; connects to cloud gRPC with backoff.
  - Handles `JobRequest` → executes local DB call → streams `JobResult`.
  - Sends heartbeats; reconnects on failure; graceful shutdown on signals.
- Provide sample config at `configs/agent.yaml.example`.

### Phase 4 — Config Wiring & Selection
- Extend `types.DISConfig` / `types.DISUIConfig` with `BridgeConfig`:
  - `Mode string` (embedded|remote), `ServerURL`, `ClientID`, `ClientSecret`, `TenantID`, `TLS` options.
- Update `config.go` defaults and env mapping (`DISREADER_BRIDGE_*`).
- In `internal.NewDISReaderService`, choose DB implementation:
  - `embedded` → current `IBMi400`
  - `remote` → `remoteDB` (requires registry client)
- Add helper to start bridge server when running in “cloud” role (optional flag/env).

### Phase 5 — Security & Ops
- TLS required on bridge gRPC; support `insecureSkipVerify` only for testing.
- Health/metrics endpoints for bridge: `/healthz` (readiness based on agent count) and `/metrics` (Prometheus gauges).
- Heartbeat timeout handling and reconnection in agent.
- Auth: pluggable `AgentAuthenticator`; default static allow-list with clientID/secret → tenant/agent.

### Phase 6 — Compatibility & Migration
- Keep embedded as default; document bridge config in `disreader.yaml` sample.
- Migration guide (`wiki/bridge_mode.md`) covering agent deploy (systemd example), bridge server wiring, and network expectations (outbound 443 only).
- Ensure no breaking changes to UI/services; DB backend swap only.

### Phase 7 — Testing & Validation
- Unit tests for remote DB happy/error paths using mock registry/agent.
- Integration smoke: in-memory gRPC bridge + agent using SQLite fixture; assert a simple query matches embedded output.
- Parity check for `PingService`/`PingDatabase` via remote path.

### Phase 8 — Operational Entry Points
- Optional standalone `bridge-server` command to host gRPC + health/metrics for quick deployment.
- Document how to mount `RegisterBridge`/`RegisterHealth` on an existing parent server (code snippet).

## New/Modified Paths
- New: `proto/bridge.proto`, `internal/bridge/proto/*`, `internal/bridge/{server.go,registry.go,auth.go}`, `internal/database/remote.go`, `cmd/agent/main.go`, `configs/agent.yaml.example`, `docs/planning/remote-agent-plan.md`.
- Modified: `types/config.go`, `config.go`, `internal/service.go`, `main.go`, `go.mod` (grpc/proto deps), `disreader.yaml` (sample), possibly `Makefile`/`README` later (not in scope phases 0–4).

## Testing Strategy
- Unit: mock `AgentConnection` to verify `remoteDB` satisfies `database.DB` semantics and error propagation.
- Integration: in-memory gRPC server + agent using SQLite fixture; assert parity of `UnitService` queries between embedded and remote modes.
- Smoke: CLI agent connects to local bridge server; run `PingService`/`PingDatabase` and a simple query.

## Risks & Mitigations
- **Transport mismatch:** Ensure `ResultRow` maps keep type info (use `structpb.Value` in proto).
- **Backwards compatibility:** Default to embedded; guard remote code paths behind config.
- **Agent leaks:** Heartbeat expiry + `defer Unregister` + context cancellations.
- **Large result sets:** Stream rows with `JobResult` chunks; cap payload size via gRPC defaults.

## Done Criteria for Phases 0–4
- Branch exists and builds.
- Plan documented (this file).
- Proto compiled and vendored stubs present.
- Bridge server & registry compile; `remoteDB` passes `go vet`/`go test` basic compilation.
- Agent daemon builds and can connect to server stub.
- Config toggle selects DB implementation without breaking embedded default.
