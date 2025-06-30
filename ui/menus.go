package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/rivo/tview"
)

// InitMainMenu is used to initialize and style sidebar menu list
func (u *UI) initMainMenu(modules []types.ViewModule) *tview.List {
	menu := tview.NewList().ShowSecondaryText(false)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	menu.SetTitle(" Main Menu ")
	menu.SetSelectedStyle(u.Theme.Style.ListSelectedStyle)

	// Dynamic modules
	for _, m := range modules {
		mod := m
		menu.AddItem(mod.Name(), "", mod.Shortcut(), func() {
			mod.Activate()
		})
	}
	menu.AddItem("DIS Config", "", 'c', u.Auth.ShowAuthModal)
	menu.AddItem("Quit", "", 'q', u.QuitApplication)

	return menu
}

// initSubMenu initializes the bookmark sidebar menu
func (u *UI) initSubMenu() *tview.List {
	subMenuList := tview.NewList().ShowSecondaryText(false)

	subMenuList.SetBorder(true)
	subMenuList.SetBorderPadding(1, 1, 1, 1)
	subMenuList.SetTitle(" Sub Menu ")

	return subMenuList
}
