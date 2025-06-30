package ui

import (
	"github.com/rivo/tview"
)

// initSubMenu initializes the bookmark sidebar menu
func (u *UI) initSubMenu() *tview.List {
	subMenuList := tview.NewList().ShowSecondaryText(false)

	subMenuList.SetBorder(true)
	subMenuList.SetBorderPadding(1, 1, 1, 1)
	subMenuList.SetTitle(" Sub Menu ")

	return subMenuList
}
