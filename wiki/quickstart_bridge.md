# Bridge & Agent Quickstart (Non-Disruptive)

This complements the existing README without overwriting it. Steps to run DIS Reader with a remote agent.

## Prerequisites
- Go toolchain
- `protoc` + plugins already vendored
- TLS termination (or set `DISREADER_BRIDGE_TLS_INSECURESKIPVERIFY=true` for testing only)

## 1) Start Bridge Server (cloud/hosted)
```sh
# bridge mode must be remote in disreader.yaml or envs
DISREADER_BRIDGE_MODE=remote \
DISREADER_BRIDGE_CLIENTID=agent \
DISREADER_BRIDGE_CLIENTSECRET=secret \
DISREADER_BRIDGE_TENANTID=t1 \
go run ./cmd/bridge-server
```
- gRPC on `:8443` (override `DISREADER_BRIDGE_PORT`)
- Health/metrics on `:8080` (`/healthz`, `/metrics`; override `DISREADER_BRIDGE_HTTP_PORT`)
- Attach TLS/ingress as appropriate for your environment.

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

## 3) Use DIS Reader normally
- With `bridge.mode=remote` set in `disreader.yaml`, `DISReaderService` routes DB calls through the agent.
- UI and domain services remain unchanged.

## Health & Metrics
- `GET /healthz` returns `status`, `total_agents`, `tenants`.
- `GET /metrics` exposes `bridge_agents_total` and `bridge_agents_per_tenant`.

## Notes
- Auth is static allow-list (clientID/secret). Rotate by updating bridge config and restarting.
- Agent reconnects with exponential backoff and sends heartbeats every 30s.
- Result values send timestamps as epoch millis; other values use protobuf `structpb.Value`.
