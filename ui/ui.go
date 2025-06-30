package ui

import (
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

	window := wm.NewWindow().
		Show().
		SetRoot(ui.setupAppLayout(modules)).
		SetBorder(false)

	window.Maximize()

	ui.startupSequence()
	return &ui
}
