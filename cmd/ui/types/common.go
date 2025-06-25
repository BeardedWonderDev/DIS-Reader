package typesUI

import (
	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

type UI interface {
	SetFocus(p tview.Primitive)
	Run() error
	QuitApplication()

	GetLayout() *ComponentLayout
	GetDIS() types.DISReaderService
	GetTheme() *entity.Theme
	GetApp() *tview.Application
	GetWinMan() *winman.Manager

	CreateModalDialog(param CreateModalDialogParam) *winman.WindowBase
	CloseModalDialog(wnd *winman.WindowBase, focus tview.Primitive)

	PrintLog(param entity.Log)
}
