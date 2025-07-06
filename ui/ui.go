package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/internal/ui"
	authUI "github.com/BeardedWonderDev/DIS-Reader/internal/ui/auth"
	logUI "github.com/BeardedWonderDev/DIS-Reader/internal/ui/log"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewUI(cfg *types.DISUIConfig, dis types.DISReaderService, screen *tcell.Screen, modules []types.ViewModule) types.UI {
	app := tview.NewApplication()
	if screen != nil {
		app.SetScreen(*screen)
	}
	wm := winman.NewWindowManager()

	u := ui.UI{
		Config: cfg,
		App:    app,
		WinMan: wm,
		Theme:  &types.DefualtTerminalTheme,
		Auth:   authUI.NewAuthService(),
	}

	u.Log = logUI.InitLogPanel(&u)
	u.DIS = authUI.NewAuthGuard(dis, &u)
	u.Auth.UI = &u
	u.Log.UI = &u

	// Initialize each module
	for _, m := range modules {
		m.Init(&u)
	}

	window := wm.NewWindow().
		Show().
		SetRoot(u.SetupAppLayout(modules)).
		SetBorder(false)

	window.Maximize()

	u.StartupSequence()
	return &u
}
