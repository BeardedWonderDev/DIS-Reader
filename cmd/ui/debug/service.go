package debugUI

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
)

type Debug struct {
	UI typesUI.UI
}

func NewDebugService() *Debug {
	return &Debug{}
}
