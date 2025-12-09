# Remote Builder Example

This example shows how to wire `DISReaderRemote` with custom auth, registry, gRPC server, and HTTP mux inside your application (no standalone bridge binary required).

```go
package main

import (
    "context"
    "log"
    "log/slog"
    "net/http"

    "github.com/BeardedWonderDev/DIS-Reader/disreader"
    "github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
    "github.com/BeardedWonderDev/DIS-Reader/types"
    "google.golang.org/grpc"
)

func main() {
    cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "remote", MaxRowsPerQuery: 1000}}
    logger := slog.Default()

    // Custom auth map (replace with your own auth source)
    auth := &bridge.StaticAuthenticator{Secrets: map[string]bridge.StaticAgentSecret{
        "agent-tenant-a": {ClientSecret: "secret-a", TenantID: "tenant-a", AgentID: "agent-a"},
    }}

    // Custom registry (use default in-memory if you prefer)
    reg := bridge.NewInMemoryRegistry()

    grpcServer := grpc.NewServer()
    mux := http.NewServeMux()

    remote, err := disreader.NewDISReaderRemote(cfg, logger).
        WithAuth(auth).
        WithRegistry(reg).
        WithGRPC(grpcServer).
        WithMux(mux).
        WithDefaultTenant("tenant-a").
        Build()
    if err != nil {
        log.Fatalf("build remote: %v", err)
    }

    // Start servers (you manage lifecycle)
    go func() { log.Fatal(grpcServer.Serve(mustListen(":8443"))) }()
    go func() { log.Fatal(http.ListenAndServe(":8080", mux)) }()

    // Use tenant-aware calls
    ctx := context.Background()
    if err := remote.PingAgent(ctx, "tenant-a"); err != nil { log.Fatal(err) }
    if err := remote.Connect(ctx, "tenant-a"); err != nil { log.Fatal(err) }
    units, err := remote.UnitService("tenant-a").ListUnits(ctx, types.ListParams{Limit: 10})
    if err != nil { log.Fatal(err) }
    log.Printf("units: %d", len(units))

    // Fetch a sanitized snapshot of the agent's runtime config (secrets are never returned)
    status, err := remote.ReadAgentConfig(ctx, "tenant-a")
    if err != nil { log.Fatal(err) }
    log.Printf("agent host=%s jdbcPort=%s hasPassword=%v hasClientSecret=%v",
        status.GetRuntime().GetDisHost(),
        status.GetRuntime().GetJdbcPort(),
        status.GetRuntime().GetHasPassword(),
        status.GetRuntime().GetHasClientSecret(),
    )
}

// helper to panic on listen error
func mustListen(addr string) net.Listener {
    lis, err := net.Listen("tcp", addr)
    if err != nil { panic(err) }
    return lis
}
```

Notes:
- If you omit `WithGRPC`/`WithMux`, the builder auto-starts servers on ports `DISREADER_BRIDGE_PORT` (default 8443) and `DISREADER_BRIDGE_HTTP_PORT` (default 8080).
- You must still run at least one agent per tenant; agents register with the bridge using their `tenantID`.
- Tenants must be passed per call (`Connect(ctx, tenant)`, `UnitService(tenant)`), unless you set `WithDefaultTenant` for your own convenience wrappers.
