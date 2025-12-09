package types

import (
	"context"
	"log/slog"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/bridgeproto"
)

// AgentConnection abstracts an active agent stream (bridge → agent).
// Exposed for applications that want to provide custom registries.
type AgentConnection interface {
	TenantID() string
	AgentID() string
	SendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error)
	Close() error
}

// AgentRegistry tracks agent connections per tenant.
type AgentRegistry interface {
	Register(ctx context.Context, tenantID string, agentID string, conn AgentConnection) error
	Unregister(ctx context.Context, tenantID string, agentID string)
	Pick(ctx context.Context, tenantID string) (AgentConnection, error)
	Stats() RegistryStats
	ObserveQuery(tenantID, agentID string, latency time.Duration)
}

// RegistryStats is a snapshot of connected agents per tenant.
type RegistryStats struct {
	TotalAgents int
	Tenants     map[string]int
	QueryCounts map[RegistryKey]int64
	Latency     map[RegistryKey]LatencyAgg
}

type RegistryKey struct {
	Tenant string
	Agent  string
}

type LatencyAgg struct {
	Count int64
	Sum   float64
	Max   float64
}

// AgentAuthenticator validates agent credentials/tenant binding.
type AgentAuthenticator interface {
	Authenticate(ctx context.Context, clientID, clientSecret string, tenantID string) (string, string, error)
}

type DISReaderService interface {
	GetConfig() *DISConfig
	GetLogger() *slog.Logger
	SetLogger(logger *slog.Logger)
	Shutdown() error
	AttachShutdownHook()
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	PingService(ctx context.Context) error  // PingService checks if the Java JDBC service is responsive
	PingDatabase(ctx context.Context) error // PingDatabase checks if the DIS AS/400 database is reachable via the JDBC service
	RunDebugSearch(searchTerm string, opts DebugSearchOptions, progressChan chan<- ProgressStatus, eventChan chan<- TableEvent)
	UnitService() UnitService
	InvoiceService() InvoiceService
	PartService() PartService
}

type DISReaderRemote interface {
	// Tenant is optional when a default was configured; otherwise required.
	StartJDBCRunner(ctx context.Context, tenant string) error
	StopJDBCRunner(ctx context.Context, tenant string) error
	Connect(ctx context.Context, tenant string) error
	Disconnect(ctx context.Context, tenant string) error
	PingBridge(ctx context.Context) error                  // PingBridge checks if the communication with the gRPC bridge server is alive
	PingAgent(ctx context.Context, tenant string) error    // PingAgent checks if the communication with the remote agent is alive
	PingService(ctx context.Context, tenant string) error  // PingService checks if the Java JDBC service is responsive
	PingDatabase(ctx context.Context, tenant string) error // PingDatabase checks if the DIS AS/400 database is reachable via the JDBC service
	UnitService(tenant string) UnitService
	InvoiceService(tenant string) InvoiceService
	PartService(tenant string) PartService
	// UpdateAgentConfig pushes a new AgentConfig to connected agents (optionally filtered by tenant)
	// and updates the server defaults used for future connections. If broadcast is false, only future
	// connections will see the update.
	UpdateAgentConfig(ctx context.Context, cfg *bridgeproto.AgentConfig, tenant string, broadcast bool) error
	// ReadAgentConfig requests a sanitized view of the agent's current configuration.
	ReadAgentConfig(ctx context.Context, tenant string) (*bridgeproto.AgentConfigStatus, error)
}
