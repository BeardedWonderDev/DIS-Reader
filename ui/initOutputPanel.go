package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InitOutputPanel initializes the output panel on the main screen
func (u *UI) InitOutputPanel() *tview.Flex {
	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.SetBorder(true)
	layout.SetTitle(" Output ")

	layout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Change focus to menu list when pressed TAB
		if event.Key() == tcell.KeyTAB {
			u.SetFocus(u.Layout.SplitSidebar)
			return nil
		}

		return event
	})

	return layout
}
