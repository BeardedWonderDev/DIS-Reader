package debugUI

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type Debug struct {
	UI           typesUI.UI
	ProgressChan chan types.ProgressStatus
	EventChan    chan types.TableEvent
}

func NewDebugService() *Debug {
	return &Debug{
		ProgressChan: make(chan types.ProgressStatus),
		EventChan:    make(chan types.TableEvent),
	}
}
