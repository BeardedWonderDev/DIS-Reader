package types

import (
	"context"
	"log/slog"
)

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
	Connect(ctx context.Context, tenant string) error
	Disconnect(ctx context.Context, tenant string) error
	PingBridge(ctx context.Context) error                  // PingBridge checks if the communication with the gRPC bridge server is alive
	PingAgent(ctx context.Context, tenant string) error    // PingAgent checks if the communication with the remote agent is alive
	PingService(ctx context.Context, tenant string) error  // PingService checks if the Java JDBC service is responsive
	PingDatabase(ctx context.Context, tenant string) error // PingDatabase checks if the DIS AS/400 database is reachable via the JDBC service
	UnitService(tenant string) UnitService
	InvoiceService(tenant string) InvoiceService
	PartService(tenant string) PartService
}
