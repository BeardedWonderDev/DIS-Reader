package ui

import (
	"fmt"
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	authUI "github.com/BeardedWonderDev/DIS-Reader/internal/ui/auth"
	logUI "github.com/BeardedWonderDev/DIS-Reader/internal/ui/log"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
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

	pendingAction func()
	shortcuts     map[rune]func()
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

func (u *UI) RunPendingAction() {
	if u.pendingAction != nil {
		u.pendingAction()
		u.pendingAction = nil
	}
}

func (u *UI) SetPendingAction(f func()) {
	u.pendingAction = f
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
	raw := disreader.NewDISReaderService(cfg, ui.GetLogger())
	ui.DIS = authUI.NewAuthGuard(raw, &ui)
	ui.Auth.UI = &ui
	ui.Log.UI = &ui

	// Initialize each module
	for _, m := range modules {
		m.Init(&ui)
	}

	// Build main and sub menus
	mainMenuList := ui.InitMainMenu(modules)
	subMenuList := ui.initSubMenu()

	splitSidebar := tview.NewFlex()
	splitSidebar.SetDirection(tview.FlexRow)
	splitSidebar.AddItem(mainMenuList, 15, 1, false)
	splitSidebar.AddItem(subMenuList, 0, 1, false)

	// Build global shortcut map
	ui.shortcuts = make(map[rune]func())
	ui.shortcuts['c'] = ui.Auth.ShowAuthModal
	ui.shortcuts['q'] = ui.QuitApplication
	// Module main-menu and sub-menu shortcuts
	for _, m := range modules {
		mod := m
		ui.shortcuts[mod.Shortcut()] = func() { mod.Activate() }
		// register submenu items
		for _, item := range mod.SubMenu().Items {
			itm := item
			ui.shortcuts[itm.Shortcut] = itm.Selected
		}
	}

	splitSidebar.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if fn, ok := ui.shortcuts[event.Rune()]; ok {
			fn()
			return nil
		}

		switch event.Key() {
		case tcell.KeyTAB:
			ui.SetFocus(ui.Layout.OutputPanel)
			return nil
		}
		return event
	})

	ui.Layout = &types.ComponentLayout{
		SplitSidebar: splitSidebar,
		MainMenu:     mainMenuList,
		SubMenuList:  subMenuList,
		LogList:      ui.Log.View,
		OutputPanel:  ui.InitOutputPanel(),
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
	splitMainPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.Layout.OutputPanel, 0, 3, false).
		AddItem(u.Layout.LogList, 0, 1, false)

	childLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(u.Layout.SplitSidebar, 35, 1, true).
		AddItem(splitMainPanel, 0, 4, false)

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(u.setupAppTitle(), 3, 1, false).
		AddItem(childLayout, 0, 1, true)

	return layout
}
