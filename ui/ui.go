package ui

import (
	"fmt"
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	authUI "github.com/BeardedWonderDev/DIS-Reader/internal/ui/auth"
	logUI "github.com/BeardedWonderDev/DIS-Reader/internal/ui/log"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

type UI struct {
	Config *types.Config
	App    *tview.Application
	WinMan *winman.Manager
	Layout *types.ComponentLayout

	DIS  types.DISReaderService
	Auth *authUI.Auth
	Log  *logUI.LogPanelWriter

	Theme *types.Theme
}

// GetApp implements types.UI.
func (u *UI) GetApp() *tview.Application {
	return u.App
}

// GetDIS implements types.UI.
func (u *UI) GetDIS() types.DISReaderService {
	return u.DIS
}

// GetLayout implements types.UI.
func (u *UI) GetLayout() *types.ComponentLayout {
	return u.Layout
}

// GetTheme implements types.UI.
func (u *UI) GetTheme() *types.Theme {
	return u.Theme
}

// GetWinMan implements types.UI.
func (u *UI) GetWinMan() *winman.Manager {
	return u.WinMan
}

func (u *UI) GetAuth() types.Auth {
	return u.Auth
}

func (u *UI) GetLogger() *slog.Logger {
	return u.Log.Logger
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
	u.DIS.Shutdown()
	u.App.Stop()
}

func NewUI(cfg *types.Config, modules []types.ViewModule) *UI {
	app := tview.NewApplication()
	wm := winman.NewWindowManager()

	ui := UI{
		Config: cfg,
		App:    app,
		WinMan: wm,
		Theme:  &types.DefualtTerminalTheme,
		Auth:   authUI.NewAuthService(),
	}

	ui.Log = logUI.InitLogPanel(&ui)
	ui.DIS = disreader.NewDISReaderService(cfg, ui.GetLogger())
	ui.Auth.UI = &ui
	ui.Log.UI = &ui

	// Initialize each module with the UI and shared logger
	for _, m := range modules {
		m.Init(&ui)
	}

	ui.Layout = &types.ComponentLayout{
		MainMenu:    ui.InitMainMenu(modules),
		SubMenuList: ui.initSubMenu(),
		LogList:     ui.Log.View,
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
func (u *UI) setupAppTitle() *tview.TextView {
	title := tview.NewTextView()
	title.SetBorder(true)
	title.SetText(fmt.Sprintf("%s v%s", u.Config.AppName, u.Config.AppVersion))
	title.SetTextAlign(tview.AlignCenter)

	return title
}

// setupAppLayout sets up the main grid layout of the application.
func (u *UI) setupAppLayout() *tview.Flex {

	// Setup the main layout
	splitSidebar := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.MainMenu, 15, 1, true).
		AddItem(u.Layout.SubMenuList, 0, 1, false)

	splitMainPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.OutputPanel, 0, 3, false).
		AddItem(u.Layout.LogList, 0, 1, false)

	childLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(splitSidebar, 35, 1, true).
		AddItem(splitMainPanel, 0, 4, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.setupAppTitle(), 3, 1, false).
		AddItem(childLayout, 0, 1, true)

	return layout
}
