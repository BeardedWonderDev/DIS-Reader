package internal

import "github.com/BeardedWonderDev/DIS-Reader/types"

type DISReaderPvtService interface {
	GetConfig() *types.Config
	Shutdown()
	AttachShutdownHook()
	TestDISConnection() error
	RunDebugSearch(searchTerm string, sqliteDBFile string, progressChan chan<- types.ProgressStatus, eventChan chan<- types.TableEvent)
	Query(sql string) ([]types.ResultRow, error)
}
