package debugUI

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type Debug struct {
	UI           types.UI
	ProgressChan chan types.ProgressStatus
	EventChan    chan types.TableEvent
}

func NewDebugService() *Debug {
	return &Debug{
		ProgressChan: make(chan types.ProgressStatus),
		EventChan:    make(chan types.TableEvent),
	}
}
