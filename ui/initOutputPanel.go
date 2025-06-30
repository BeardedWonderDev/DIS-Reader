package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
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

// ClearOutputPanel resets the output panel to its default empty state.
func (u *UI) ClearOutputPanel() {
	// Recreate the output panel flex using the existing initializer
	newPanel := u.InitOutputPanel()
	u.Layout.OutputPanel = newPanel
}

// SetOutputPanel applies the given PanelDefinition: sets the title and view.
func (u *UI) SetOutputPanel(pd types.PanelDefinition) {
	// clear previous content
	panel := u.Layout.OutputPanel
	panel.Clear()
	// Set the panel title
	panel.SetTitle(" " + pd.Name + " ")
	// Insert the provided view into the flex (full size)
	panel.AddItem(pd.View, 0, 1, false)
}
