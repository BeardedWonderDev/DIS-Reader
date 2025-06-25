package ui

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/rivo/tview"
)

func buildMenu(menuType typesUI.MenuDefinition, menuList *tview.List) {
	menuList.Clear()
	menuList.SetTitle(menuType.Title)
	for _, item := range menuType.Items {
		menuList.AddItem(item.MainText, item.SecondaryText, item.Shortcut, item.Selected)
	}
}

func (u *UI) revertToMainMenu() {
	buildMenu(u.Layout.MainMenu.RootMenu, u.Layout.MainMenu.MenuList)
}
