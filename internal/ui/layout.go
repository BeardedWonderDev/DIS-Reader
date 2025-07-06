package ui

import (
	"fmt"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/rivo/tview"
)

// setupAppLayout sets up the main grid layout of the application.
func (u *UI) SetupAppLayout(modules []types.ViewModule) *tview.Flex {
	u.Layout = &types.ComponentLayout{
		LogList:     u.Log.View,
		OutputPanel: u.InitOutputPanel(),
	}
	u.initShortcuts(modules)
	u.buildSidebar(modules)

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

func (u *UI) buildSidebar(modules []types.ViewModule) {
	// Build main and sub menus
	main := u.initMainMenu(modules)
	sub := u.initSubMenu()

	splitSidebar := tview.NewFlex()
	splitSidebar.SetDirection(tview.FlexRow)
	splitSidebar.AddItem(main, 15, 1, false)
	splitSidebar.AddItem(sub, 0, 1, false)

	splitSidebar.SetInputCapture(u.handleShortcutEvent)

	u.Layout.SplitSidebar = splitSidebar
	u.Layout.MainMenu = main
	u.Layout.SubMenuList = sub
}

func (u *UI) setupAppTitle() *tview.TextView {
	title := tview.NewTextView()
	title.SetBorder(true)
	title.SetText(fmt.Sprintf("%s v%s", u.Config.AppName, u.Config.AppVersion))
	title.SetTextAlign(tview.AlignCenter)

	return title
}
