package ui

import "github.com/rivo/tview"

type MainMenuDefinition struct {
	MenuList *tview.List

	rootMenu  MenuDefinition
	debugMenu MenuDefinition
}

type MenuDefinition struct {
	Title string
	Items []MenuItem
}

type MenuItem struct {
	MainText      string // The main text of the list item.
	SecondaryText string // A secondary text to be shown underneath the main text.
	Shortcut      rune   // The key to select the list item directly, 0 if there is no shortcut.
	Selected      func() // The optional function which is called when the item is selected.
}

func buildMenu(menuType MenuDefinition, menuList *tview.List) {
	menuList.Clear()
	menuList.SetTitle(menuType.Title)
	for _, item := range menuType.Items {
		menuList.AddItem(item.MainText, item.SecondaryText, item.Shortcut, item.Selected)
	}
}

func (u *UI) revetToMainMenu() {
	buildMenu(u.Layout.MainMenu.rootMenu, u.Layout.MainMenu.MenuList)
}
