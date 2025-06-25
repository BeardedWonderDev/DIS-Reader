package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InitMainMenu is used to initialize and populate sidebar menu
func (u *UI) InitMainMenu() *MainMenuDefinition {
	menuList := tview.NewList().ShowSecondaryText(false)
	menuList.SetBorder(true)
	menuList.SetBorderPadding(1, 1, 1, 1)

	mainMenu := MainMenuDefinition{
		MenuList: menuList,
		rootMenu: MenuDefinition{
			Title: " Main Menu ",
			Items: []MenuItem{
				{
					MainText:      "Auth Config",
					SecondaryText: "",
					Shortcut:      'a',
					Selected:      u.showAuthModal,
				},
				{
					MainText:      "Debug Search",
					SecondaryText: "",
					Shortcut:      'd',
					Selected:      u.InitDebugViews,
				},
				{
					MainText:      "Quit",
					SecondaryText: "",
					Shortcut:      'q',
					Selected:      u.QuitApplication,
				},
			},
		},
		debugMenu: MenuDefinition{
			Title: " Debug Menu ",
			Items: []MenuItem{
				{
					MainText:      "Main Menu",
					SecondaryText: "",
					Shortcut:      'm',
					Selected:      u.revetToMainMenu,
				},
			},
		},
	}

	// Handle keypress on menu list
	u.initMainMenu_SetInputCapture(mainMenu.MenuList)

	return &mainMenu
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
