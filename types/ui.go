package types

import (
	"log/slog"

	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

// ViewModule represents a pluggable UI component.
type ViewModule interface {
	// Name is the display text for the main menu.
	Name() string
	// Shortcut is the rune key to activate this module.
	Shortcut() rune
	// Init is called once after UI and Logger are ready.
	Init(ui UI)
	// Activate is called when the user selects this module.
	Activate()
}

type UI interface {
	SetFocus(p tview.Primitive)
	Run() error
	QuitApplication()

	GetLayout() *ComponentLayout
	GetDIS() DISReaderService
	GetTheme() *Theme
	GetApp() *tview.Application
	GetWinMan() *winman.Manager
	GetAuth() Auth
	RunPendingAction()
	SetPendingAction(f func())
	GetLogger() *slog.Logger

	CreateModalDialog(param CreateModalDialogParam) *winman.WindowBase
	CloseModalDialog(wnd *winman.WindowBase, focus tview.Primitive)
	// RevertToMainMenu()
}

type Auth interface {
	IsAuthenticated() bool
	Authenticate(onComplete func(success bool, err error))
	ShowAuthModal()
}

type ComponentLayout struct {
	MainMenu    *tview.List
	SubMenuList *tview.List
	LogList     *tview.TextView
	OutputPanel *tview.Flex
}

type WinSize struct {
	X      int
	Y      int
	Width  int
	Height int
}

type CreateModalDialogParam struct {
	Title         string
	RootView      tview.Primitive
	Draggable     bool
	Resizeable    bool
	Size          WinSize
	FallbackFocus tview.Primitive
}

type MenuDefinition struct {
	Title string
	Items []MenuItem
}

type MenuItem struct {
	MainText      string // The main text of the list item.
	SecondaryText string // A secondary text to be shown underneath the main text.
	Shortcut      rune   // The key to select the list item directly, 0 if there is no shortcut.
	Selected      func() // The optional function which is called when the item is selected.
}
