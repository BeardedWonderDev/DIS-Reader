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
	TestConnection(ctx context.Context) error
	RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- ProgressStatus, eventChan chan<- TableEvent)
	UnitService() UnitService
}
