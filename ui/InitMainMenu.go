package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/rivo/tview"
)

// InitMainMenu is used to initialize and style sidebar menu list
func (u *UI) InitMainMenu(modules []types.ViewModule) *tview.List {
	menu := tview.NewList().ShowSecondaryText(false)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)
	menu.SetTitle(" Main Menu ")

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
