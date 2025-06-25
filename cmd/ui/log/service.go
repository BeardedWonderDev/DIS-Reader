package logUI

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
)

type Log struct {
	UI typesUI.UI
}

func NewLogService() *Log {
	return &Log{}
}
