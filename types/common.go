package types

import "context"

type DISReaderService interface {
	GetConfig() *DISConfig
	Shutdown() error
	AttachShutdownHook()
	TestConnection(ctx context.Context) error
	RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- ProgressStatus, eventChan chan<- TableEvent)
}
