package ui

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	utilsUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InitMainMenu is used to initialize and populate sidebar menu
func (u *UI) InitMainMenu() typesUI.MainMenuDefinition {
	menuList := tview.NewList().ShowSecondaryText(false)
	menuList.SetBorder(true)
	menuList.SetBorderPadding(1, 1, 1, 1)

	mainMenu := typesUI.MainMenuDefinition{
		MenuList: menuList,
		RootMenu: typesUI.MenuDefinition{
			Title: " Main Menu ",
			Items: []typesUI.MenuItem{
				{
					MainText:      "Auth Config",
					SecondaryText: "",
					Shortcut:      'a',
					Selected:      u.Auth.ShowAuthModal,
				},
				{
					MainText:      "Debug Search",
					SecondaryText: "",
					Shortcut:      'd',
					Selected:      u.Debug.InitDebugViews,
				},
				{
					MainText:      "Quit",
					SecondaryText: "",
					Shortcut:      'q',
					Selected:      u.QuitApplication,
				},
			},
		},
	}

	// Handle keypress on menu list
	u.initMainMenu_SetInputCapture(mainMenu.MenuList)

	return mainMenu
}

func (u *UI) initMainMenu_SetInputCapture(menuList *tview.List) {
	menuList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			u.SetFocus(u.Layout.SubMenuList)
			return nil
		}
		return event
	})
}

func (u UI) RevertToMainMenu() {
	utilsUI.BuildMenu(u.Layout.MainMenu.RootMenu, u.Layout.MainMenu.MenuList)
}
