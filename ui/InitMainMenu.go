package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InitMainMenu is used to initialize and style sidebar menu list
func (u *UI) InitMainMenu(modules []types.ViewModule) *tview.List {
	menu := tview.NewList().ShowSecondaryText(false)
	menu.SetBorder(true)
	menu.SetBorderPadding(1, 1, 1, 1)

	// Dynamic modules
	for _, m := range modules {
		mod := m
		menu.AddItem(mod.Name(), "", mod.Shortcut(), func() {
			mod.Activate()
		})
	}
	menu.AddItem("Auth Config", "", 'a', u.Auth.ShowAuthModal)
	menu.AddItem("Quit", "", 'q', u.QuitApplication)

	menu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			u.SetFocus(u.Layout.SubMenuList)
			return nil
		}
		return event
	})

	return menu
}
