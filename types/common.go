package types

import "context"

type DISReaderService interface {
	GetConfig() *DISConfig
	Shutdown() error
	AttachShutdownHook()
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	TestConnection(ctx context.Context) error
	RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- ProgressStatus, eventChan chan<- TableEvent)
}
