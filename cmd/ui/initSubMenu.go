package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// initSubMenu initializes the bookmark sidebar menu
func (u *UI) initSubMenu() *tview.List {
	subMenuList := tview.NewList().ShowSecondaryText(false)

	subMenuList.SetBorder(true)
	subMenuList.SetBorderPadding(1, 1, 1, 1)
	subMenuList.SetTitle(" Sub Menu ")

	u.initSubMenu_SetInputCapture(subMenuList)

	return subMenuList
}

func (u *UI) initSubMenu_SetInputCapture(subMenuList *tview.List) {
	subMenuList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			u.SetFocus(u.Layout.OutputPanel)
			return nil
		}
		return event
	})
}
