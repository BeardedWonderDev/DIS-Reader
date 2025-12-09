# Bridge & Agent Quickstart (Non-Disruptive)

This complements the existing README without overwriting it. Steps to run DIS Reader with a remote agent.

## Prerequisites
- Go toolchain
- `protoc` + plugins already vendored
- TLS termination (or set `DISREADER_BRIDGE_TLS_INSECURESKIPVERIFY=true` for testing only)

## 1) Start bridge inside your app (remote builder auto-starts servers)
- Set `bridge.mode=remote` in config or env.
- If you don’t provide your own servers, the remote builder will start:
  - gRPC on `:8443` (override `DISREADER_BRIDGE_PORT`)
  - HTTP /healthz,/metrics (and pprof if enabled) on `:8080` (override `DISREADER_BRIDGE_HTTP_PORT`)
- To register handlers on your own servers instead, call `.WithGRPC(server)` / `.WithMux(mux)` on the builder.

Example (app side):
```go
remote, err := disreader.NewDISReaderRemote(cfg, logger).
    WithAuth(customAuth).        // optional
    WithRegistry(customReg).     // optional
    WithDefaultTenant("t1").    // optional fallback
    Build()
if err != nil { panic(err) }
```

If you pass your own `*grpc.Server` / `*http.ServeMux`, you are responsible for calling `Serve`/`ListenAndServe`.

## 2) Run Agent (LAN side)
```sh
cat > agent.yaml <<'YAML'
serverURL: "bridge-host:8443"
clientID: "agent"
clientSecret: "secret"
tenantID: "t1"
dis:
  host: "10.0.0.5"
  user: "DISUSER"
  password: "secret"
  jdbcConfig:
    javaPath: "java"
    jdbcPort: "8888"
tls:
  insecureSkipVerify: true   # only for testing
YAML

go run ./cmd/agent
```

## 3) Use DIS Reader in remote mode
- Call tenant-aware methods: `remote.Connect(ctx, "t1")`, `remote.UnitService("t1")`, etc.
- If you set `WithDefaultTenant`, you can wrap helpers that omit the tenant.

## Health & Metrics
- `GET /healthz` returns `status`, `total_agents`, `tenants`.
- `GET /metrics` exposes `bridge_agents_total`, `bridge_agents_per_tenant`, per-tenant/agent query counts, and latency summary (mean/max).
- Optional: enable pprof by setting `bridge.pprofEnabled=true` (or `DISREADER_BRIDGE_PPROFENABLED=true`) — only do this on trusted networks.

## Notes
- Auth options: custom authenticator, credentialFile (with optional reload), or static allow-list (clientID/secret). Rotate by updating creds and reloading; file auth can auto-reload with `credentialReloadSeconds`.
- Agent reconnects with exponential backoff and sends heartbeats every 30s.
- Result values send timestamps as epoch millis; other values use protobuf `structpb.Value`.
