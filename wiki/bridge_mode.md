# Bridge Mode (Remote JDBC Agent)

This document explains how to run DIS Reader in **remote** mode so the JDBC runner lives on a LAN “agent” while the cloud-hosted DIS Reader service proxies queries over gRPC.

## Components
- **Agent** (`cmd/agent`): runs the embedded JDBC runner locally, dials out to the cloud over TLS, executes SQL, and returns rows.
- **Bridge server**: the cloud process that registers the `AgentService` gRPC handler and routes requests to connected agents. It shares the same registry used by the remote DB adapter.
- **Remote DB adapter**: an implementation of `database.DB` that forwards queries to an available agent.

## Configuration
Add to `disreader.yaml` or envs (`DISREADER_BRIDGE_*`):
```yaml
bridge:
  mode: remote            # embedded | remote
  credentialFile: "bridge_agents.yaml"   # optional; overrides static list
  credentialReloadSeconds: 0             # >0 to auto-reload credentialFile periodically
  maxRowsPerQuery: 1000                  # 0 = unlimited (applied on remote DB proxy)
  maxResultBytes: 0                      # 0 = unlimited; drop rows beyond byte budget
  pprofEnabled: false                    # enable /debug/pprof/* on bridge HTTP mux
  pprofPath: /debug/pprof/               # custom pprof base path
  serverURL: "bridge.example.com:443"   # agent uses this
  clientID: "agent-001"
  clientSecret: "replace-me"
  tenantID: "default"
  tls:
    insecureSkipVerify: false
  allowedAgents:
    - clientID: "agent-001"
      clientSecret: "replace-me"
      tenantID: "default"
      agentID: "agent-001"
```

Defaults:
- `bridge.mode` = `embedded` (current behavior).
- TLS is required; set `insecureSkipVerify: true` only for testing.
- Auth order of precedence:
  1) Custom authenticator passed via remote builder `.WithAuth(...)`

**Heads-up:** The bridge is now embedded via the remote builder. If you don’t supply your own `WithGRPC`/`WithMux`, the builder auto-starts gRPC on `:8443` and health/metrics on `:8080` (overridable via env). For a full wiring example with custom auth/registry/servers, see `wiki/remote_builder_example.md`.
  2) `credentialFile` (YAML/JSON list of clientID/secret/tenantID/agentID)
  3) Static allow-list in `allowedAgents` / top-level clientID+secret
- Result limits:
  - `maxRowsPerQuery` trims row count on the remote DB proxy; agents also honor `DISAGENT_MAX_ROWS` (default 1000).
  - `maxResultBytes` caps total marshaled response size (best-effort).
- Debugging: enable `pprofEnabled` and protect the bridge HTTP port via network ACLs if exposing pprof.

## Hosting the Bridge Server
1. Construct `DISReaderService` with a config where `bridge.mode=remote`.
2. Create a `grpc.Server` and call `svc.RegisterBridge(grpcServer)`. Optionally register health/metrics/pprof via `svc.RegisterHealth(mux)`.
3. Start serving on your chosen port (typically 443/8443 behind TLS termination).
4. Your parent app continues to call domain services (`UnitService`, `InvoiceService`, `PartService`); all DB calls will flow through the agent registry.
5. Packaged installs are produced by `make release-agent` (deb/rpm/pkg/zip) or via the GitHub Actions `release-agent` workflow.

## Running the Agent
1. Create `agent.yaml` (see `configs/agent.yaml.example`).
2. Set `serverURL` to the public address of your bridge server.
3. Provide `clientID`, `clientSecret`, and the DIS host/user/pass for the LAN.
4. Run `go run ./cmd/agent` or build/install the binary.
5. The agent will maintain a streaming connection, send heartbeats, and execute jobs.

## Operational Notes
- Auth is static allow-list for now (clientID/secret + tenantID); rotate by updating bridge config on the server.
- Heartbeats time out after ~90s; idle agents are dropped and must reconnect.
- Result rows send values as protobuf `structpb.Value`; timestamps are epoch millis.
- Remote mode keeps the existing DIS Reader APIs unchanged; only the DB backend swaps.
