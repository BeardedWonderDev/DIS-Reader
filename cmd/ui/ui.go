package ui

import (
	"fmt"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	authUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/auth"
	debugUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/debug"
	logUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/log"
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

type UI struct {
	App    *tview.Application
	WinMan *winman.Manager
	Layout *typesUI.ComponentLayout

	DIS   types.DISReaderService
	Auth  *authUI.Auth
	Debug *debugUI.Debug
	Log   *logUI.Log

	Theme *entity.Theme
}

// GetApp implements typesUI.UI.
func (u *UI) GetApp() *tview.Application {
	return u.App
}

// GetDIS implements typesUI.UI.
func (u *UI) GetDIS() types.DISReaderService {
	return u.DIS
}

// GetLayout implements typesUI.UI.
func (u *UI) GetLayout() *typesUI.ComponentLayout {
	return u.Layout
}

// GetTheme implements typesUI.UI.
func (u *UI) GetTheme() *entity.Theme {
	return u.Theme
}

// GetWinMan implements typesUI.UI.
func (u *UI) GetWinMan() *winman.Manager {
	return u.WinMan
}

func (u *UI) GetAuth() typesUI.Auth {
	return u.Auth
}

func (u *UI) SetFocus(p tview.Primitive) {
	go u.App.QueueUpdateDraw(func() {
		u.App.SetFocus(p)
	})
}

func (u *UI) Run() error {
	return u.App.Run()
}

func (u *UI) QuitApplication() {
	u.App.Stop()
}

func NewUI(cfg *types.Config) *UI {
	app := tview.NewApplication()
	wm := winman.NewWindowManager()

	ui := UI{
		App:    app,
		WinMan: wm,
		DIS:    disreader.NewDISReaderService(cfg),
		Theme:  &entity.TerminalTheme,
		Auth:   authUI.NewAuthService(),
		Debug:  debugUI.NewDebugService(),
		Log:    logUI.NewLogService(),
	}

	ui.Auth.UI = &ui
	ui.Debug.UI = &ui
	ui.Log.UI = &ui

	ui.Layout = &typesUI.ComponentLayout{
		MainMenu:    ui.InitMainMenu(),
		SubMenuList: ui.initSubMenu(),
		LogList:     ui.Log.InitLogList(),
		OutputPanel: ui.InitOutputPanel(),
	}

	window := wm.NewWindow().
		Show().
		SetRoot(ui.setupAppLayout()).
		SetBorder(false)

	window.Maximize()

	ui.startupSequence()
	return &ui
}

// setupTitle sets up the header title of the application.
// containing the application name and version.
func setupAppTitle() *tview.TextView {
	title := tview.NewTextView()
	title.SetBorder(true)
	title.SetText(fmt.Sprintf("%s v%s", entity.APP_NAME, entity.APP_VERSION))
	title.SetTextAlign(tview.AlignCenter)

	return title
}

// setupAppLayout sets up the main grid layout of the application.
func (u *UI) setupAppLayout() *tview.Flex {

	// Setup the main layout
	splitSidebar := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.MainMenu.MenuList, 15, 1, true).
		AddItem(u.Layout.SubMenuList, 0, 1, false)

	splitMainPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.OutputPanel.Layout, 0, 3, false).
		AddItem(u.Layout.LogList, 0, 1, false)

	childLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(splitSidebar, 35, 1, true).
		AddItem(splitMainPanel, 0, 4, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(setupAppTitle(), 3, 1, false).
		AddItem(childLayout, 0, 1, true)

	return layout
}
