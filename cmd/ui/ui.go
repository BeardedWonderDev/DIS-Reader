package ui

import (
	"fmt"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

type ComponentLayout struct {
	MainMenu    *MainMenuDefinition
	SubMenuList *tview.List
	LogList     *tview.TextView
	OutputPanel InitOutputPanelComponents
}

type UI struct {
	App    *tview.Application
	WinMan *winman.Manager
	Layout *ComponentLayout

	DIS types.DISReaderService

	Theme *entity.Theme
}

func (u *UI) SetFocus(p tview.Primitive) {
	go u.App.QueueUpdateDraw(func() {
		u.App.SetFocus(p)
	})
}

func (u UI) Run() error {
	u.App.EnableMouse(true)
	return u.App.Run()
}

func (u UI) QuitApplication() {
	u.App.Stop()
}

func NewUI(cfg *types.Config) UI {
	app := tview.NewApplication()
	wm := winman.NewWindowManager()

	ui := UI{
		App:    app,
		WinMan: wm,
		DIS:    disreader.NewDISReaderService(cfg),
		Theme:  &entity.TerminalTheme,
	}

	ui.Layout = &ComponentLayout{
		MainMenu:    ui.InitMainMenu(),
		SubMenuList: ui.initSubMenu(),
		LogList:     ui.InitLogList(),
		OutputPanel: ui.InitOutputPanel(),
	}

	window := wm.NewWindow().
		Show().
		SetRoot(ui.setupAppLayout()).
		SetBorder(false)

	window.Maximize()

	ui.startupSequence()
	return ui
}

// setupTitle sets up the header title of the application.
// containing the application name and version.
func setupAppTitle() *tview.TextView {
	title := tview.NewTextView()
	title.SetBorder(true)
	title.SetText(fmt.Sprintf("DIS Reader v%s", entity.APP_VERSION))
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
