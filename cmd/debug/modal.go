package debugUI

import (
	"fmt"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	txtDebugSearch *tview.InputField
	btnBatchSearch *tview.Button
)

func (d *Debug) showDebugModal() {
	txtDebugSearch = tview.NewInputField()
	txtDebugSearch.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	txtDebugSearch.SetFieldStyle(d.UI.GetTheme().Style.FieldStyle)
	txtDebugSearch.SetLabel("Search Term: ")

	txtDebugSearchDesc := tview.NewTextView()
	txtDebugSearchDesc.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	txtDebugSearchDesc.SetTextStyle(d.UI.GetTheme().Style.TextAreaStyle)
	txtDebugSearchDesc.SetDisabled(true)
	txtDebugSearchDesc.SetText("Enter a term to search all DIS tables for\n(ie.. Part #, Invoice #, Unit#, etc).\nResults will be stored locally and visable in DIS Reader.")
	txtDebugSearchDesc.SetTextAlign(tview.AlignCenter)

	btnBatchSearch = tview.NewButton("Search")
	btnBatchSearch.SetStyle(d.UI.GetTheme().Style.ButtonStyle)

	layout := tview.NewGrid()
	layout.SetBorderPadding(1, 1, 1, 1)
	layout.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	layout.AddItem(txtDebugSearch, 0, 0, 1, 1, 0, 0, true)
	layout.AddItem(txtDebugSearchDesc, 1, 0, 2, 1, 0, 0, false)
	layout.AddItem(btnBatchSearch, 3, 0, 1, 1, 0, 0, false)

	wnd := d.UI.CreateModalDialog(types.CreateModalDialogParam{
		Title:         " DIS Batch Debug Search ",
		RootView:      layout,
		Draggable:     true,
		Size:          types.WinSize{X: 0, Y: 0, Width: 70, Height: 10},
		FallbackFocus: d.UI.GetLayout().MainMenu,
	})

	d.showDebugModal_SetInputCapture(wnd)
}

func (d *Debug) showDebugModal_SetInputCapture(wnd *winman.WindowBase) {

	txtDebugSearch.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			d.UI.GetWinMan().RemoveWindow(wnd)
			d.UI.SetFocus(d.UI.GetLayout().MainMenu)
			return nil

		case tcell.KeyTAB:
			d.UI.SetFocus(btnBatchSearch)

		}

		return event
	})

	btnBatchSearch.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			d.UI.SetFocus(txtDebugSearch)
			return nil
		case tcell.KeyEnter:
			d.runBatchSearch(wnd)
			return nil
		}

		return event
	})

	btnBatchSearch.SetSelectedFunc(func() {
		d.runBatchSearch(wnd)
	})
}

func (d *Debug) runBatchSearch(wnd *winman.WindowBase) {
	if !d.UI.GetAuth().IsAuthenticated() {
		d.UI.GetAuth().ShowAuthModal()
		return
	}

	go func() {
		searchTerm := txtDebugSearch.GetText()
		sqliteFile := "test.db"

		d.UI.GetLogger().Info("Starting Search for search term", "term", searchTerm)

		d.UI.GetDIS().RunDebugSearch(
			searchTerm,
			sqliteFile,
			d.ProgressChan,
			d.EventChan,
		)
	}()

	go func() {
		for progress := range d.ProgressChan {
			d.UI.GetLogger().Info("Search progress update",
				"runID", progress.RunID,
				"completed", progress.CompletedQueries,
				"total", progress.TotalQueries,
				"percent", fmt.Sprintf("%.2f", progress.PercentComplete),
			)
		}
	}()

	go func() {
		for event := range d.EventChan {
			d.UI.GetLogger().Info("Debug search event",
				"table", event.TableName,
				"type", event.EventType,
				"rows", event.RowCount,
				"cols", event.ColumnCount,
				"sample", event.SampleRow,
			)
		}
	}()

	// Remove the window and restore focus to menu list
	d.UI.CloseModalDialog(wnd, d.UI.GetLayout().MainMenu)
}
