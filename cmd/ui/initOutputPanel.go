package ui

import (
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InitOutputPanel initializes the output panel on the main screen
func (u *UI) InitOutputPanel() typesUI.InitOutputPanelComponents {
	output := tview.NewTextArea()
	output.SetWrap(false)
	output.SetMaxLength(1)
	output.SetTextStyle(tcell.StyleDefault.
		Foreground(tcell.ColorGreen))

	u.initOutputPanel_handleTextArea(output)

	layout := tview.NewFlex()
	layout.SetDirection(tview.FlexRow)
	layout.SetBorder(true)
	layout.SetTitle(" Output ")
	layout.AddItem(output, 0, 1, true)

	return typesUI.InitOutputPanelComponents{
		Layout:   layout,
		TextArea: output,
		Buffer:   "",
	}
}

func (u *UI) initOutputPanel_handleTextArea(textarea *tview.TextArea) {
	textarea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Block backspace for editing the text area
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyBackspace2 {
			return nil
		}

		// Change focus to menu list when pressed TAB
		if event.Key() == tcell.KeyTAB {
			u.SetFocus(u.Layout.LogList)
			return nil
		}

		return event
	})
}
