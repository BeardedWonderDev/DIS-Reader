package debugUI

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	txtDebugSearch       *tview.InputField
	txtOutputPath        *tview.InputField
	formatDropdown       *tview.DropDown
	btnBatchSearch       *tview.Button
	selectedFormat       = types.DebugSearchOutputSQLite
	configuredOutputPath string
)

func (d *Debug) showDebugModal() {
	cfg := d.UI.GetConfig()
	selectedFormat = deriveConfiguredMode(cfg)
	configuredOutputPath = ""
	if cfg != nil {
		configuredOutputPath = strings.TrimSpace(cfg.DebugSearch.DefaultOutputPath)
	}

	txtDebugSearch = tview.NewInputField()
	txtDebugSearch.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	txtDebugSearch.SetFieldStyle(d.UI.GetTheme().Style.FieldStyle)
	txtDebugSearch.SetLabel("Search Term: ")

	formatDropdown = tview.NewDropDown()
	formatDropdown.SetLabel("Output Format: ")
	formatDropdown.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	formatDropdown.SetListStyles(d.UI.GetTheme().Style.ListMainTextStyle, d.UI.GetTheme().Style.ListSelectedStyle)
	formatDropdown.SetOptions([]string{"SQLite (.db)", "CSV (.csv)"}, func(text string, index int) {
		if index == 1 {
			selectedFormat = types.DebugSearchOutputCSV
		} else {
			selectedFormat = types.DebugSearchOutputSQLite
		}
		syncOutputPathWithFormat()
	})
	if selectedFormat == types.DebugSearchOutputCSV {
		formatDropdown.SetCurrentOption(1)
	} else {
		formatDropdown.SetCurrentOption(0)
	}

	txtOutputPath = tview.NewInputField()
	txtOutputPath.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	txtOutputPath.SetFieldStyle(d.UI.GetTheme().Style.FieldStyle)
	txtOutputPath.SetLabel("Output File: ")
	txtOutputPath.SetText(defaultOutputPath())

	txtDebugSearchDesc := tview.NewTextView()
	txtDebugSearchDesc.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	txtDebugSearchDesc.SetTextStyle(d.UI.GetTheme().Style.TextAreaStyle)
	txtDebugSearchDesc.SetDisabled(true)
	txtDebugSearchDesc.SetText("Enter a term to search all DIS tables for\n(ie.. Part #, Invoice #, Unit#, etc).\nResults are saved locally as SQLite or CSV based on your selection.")
	txtDebugSearchDesc.SetTextAlign(tview.AlignCenter)

	btnBatchSearch = tview.NewButton("Search")
	btnBatchSearch.SetStyle(d.UI.GetTheme().Style.ButtonStyle)

	layout := tview.NewGrid()
	layout.SetBorderPadding(1, 1, 1, 1)
	layout.SetBackgroundColor(d.UI.GetTheme().Colors.WindowColor)
	layout.AddItem(txtDebugSearch, 0, 0, 1, 1, 0, 0, true)
	layout.AddItem(formatDropdown, 1, 0, 1, 1, 0, 0, false)
	layout.AddItem(txtOutputPath, 2, 0, 1, 1, 0, 0, false)
	layout.AddItem(txtDebugSearchDesc, 3, 0, 1, 1, 0, 0, false)
	layout.AddItem(btnBatchSearch, 4, 0, 1, 1, 0, 0, false)

	wnd := d.UI.CreateModalDialog(types.CreateModalDialogParam{
		Title:         " DIS Batch Debug Search ",
		RootView:      layout,
		Draggable:     true,
		Size:          types.WinSize{X: 0, Y: 0, Width: 70, Height: 12},
		FallbackFocus: d.UI.GetLayout().SplitSidebar,
	})

	d.showDebugModal_SetInputCapture(wnd)
}

func (d *Debug) showDebugModal_SetInputCapture(wnd *winman.WindowBase) {

	txtDebugSearch.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			d.UI.GetWinMan().RemoveWindow(wnd)
			d.UI.SetFocus(d.UI.GetLayout().SplitSidebar)
			return nil

		case tcell.KeyTAB:
			d.UI.SetFocus(formatDropdown)

		}

		return event
	})

	formatDropdown.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			d.UI.GetWinMan().RemoveWindow(wnd)
			d.UI.SetFocus(d.UI.GetLayout().SplitSidebar)
			return nil
		case tcell.KeyTAB:
			d.UI.SetFocus(txtOutputPath)
			return nil
		}
		return event
	})

	txtOutputPath.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			d.UI.GetWinMan().RemoveWindow(wnd)
			d.UI.SetFocus(d.UI.GetLayout().SplitSidebar)
			return nil
		case tcell.KeyTAB:
			d.UI.SetFocus(btnBatchSearch)
			return nil
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
	searchTerm := strings.TrimSpace(txtDebugSearch.GetText())
	if searchTerm == "" {
		d.UI.GetLogger().Error("Search term is required")
		return
	}

	outputPath := strings.TrimSpace(txtOutputPath.GetText())
	if outputPath == "" {
		outputPath = defaultOutputPath()
		txtOutputPath.SetText(outputPath)
	}
	opts := types.DebugSearchOptions{
		OutputMode: selectedFormat,
		OutputPath: outputPath,
	}

	go func(term string, options types.DebugSearchOptions) {
		d.UI.GetDIS().RunDebugSearch(
			term,
			options,
			d.ProgressChan,
			d.EventChan,
		)
	}(searchTerm, opts)

	go func() {
		for progress := range d.ProgressChan {
			d.UI.GetLogger().Info("Search progress update",
				slog.String("run_id", progress.RunID),
				slog.Int("completed", progress.CompletedQueries),
				slog.Int("total", progress.TotalQueries),
				slog.Float64("percent", progress.PercentComplete),
			)
		}
	}()

	go func() {
		for event := range d.EventChan {
			d.UI.GetLogger().Info("Debug search event",
				slog.String("table", event.TableName),
				slog.String("type", event.EventType),
				slog.Int("rows", event.RowCount),
				slog.Int("cols", event.ColumnCount),
				slog.Any("sample", event.SampleRow),
			)
			if strings.EqualFold(event.EventType, "run_failed") {
				msg := fmt.Sprintf("Debug search failed: %v", event.SampleRow["error"])
				d.showRunFailedModal(msg)
			}
		}
	}()

	// Remove the window and restore focus to menu list
	d.UI.CloseModalDialog(wnd, d.UI.GetLayout().SplitSidebar)
}

func defaultOutputPath() string {
	base := strings.TrimSpace(configuredOutputPath)
	if base == "" {
		if selectedFormat == types.DebugSearchOutputCSV {
			return "debug-search.csv"
		}
		return "debug-search.db"
	}
	return ensureExtension(base, selectedFormat)
}

func (d *Debug) showRunFailedModal(msg string) {
	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"OK"})

	var wnd *winman.WindowBase
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if wnd != nil {
			d.UI.CloseModalDialog(wnd, d.UI.GetLayout().SplitSidebar)
		}
	})

	wnd = d.UI.CreateModalDialog(types.CreateModalDialogParam{
		Title:         " Debug Search Failed ",
		RootView:      modal,
		Draggable:     true,
		Resizeable:    false,
		Size:          types.WinSize{Width: 60, Height: 7},
		FallbackFocus: d.UI.GetLayout().SplitSidebar,
	})
}

func syncOutputPathWithFormat() {
	if txtOutputPath == nil {
		return
	}
	current := strings.TrimSpace(txtOutputPath.GetText())
	if current == "" {
		txtOutputPath.SetText(defaultOutputPath())
		return
	}
	lower := strings.ToLower(current)
	otherExt := extensionForMode(flipMode(selectedFormat))
	if strings.HasSuffix(lower, otherExt) {
		base := current[:len(current)-len(otherExt)]
		txtOutputPath.SetText(base + extensionForMode(selectedFormat))
	}
}

func ensureExtension(path string, mode types.DebugSearchOutputMode) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "debug-search" + extensionForMode(mode)
	}
	ext := extensionForMode(mode)
	if strings.HasSuffix(strings.ToLower(path), ext) {
		return path
	}
	other := extensionForMode(flipMode(mode))
	if strings.HasSuffix(strings.ToLower(path), other) {
		path = path[:len(path)-len(other)]
	}
	return path + ext
}

func extensionForMode(mode types.DebugSearchOutputMode) string {
	if mode == types.DebugSearchOutputCSV {
		return ".csv"
	}
	return ".db"
}

func flipMode(mode types.DebugSearchOutputMode) types.DebugSearchOutputMode {
	if mode == types.DebugSearchOutputCSV {
		return types.DebugSearchOutputSQLite
	}
	return types.DebugSearchOutputCSV
}

func deriveConfiguredMode(cfg *types.DISUIConfig) types.DebugSearchOutputMode {
	if cfg == nil {
		return types.DebugSearchOutputSQLite
	}
	return normalizeOutputMode(cfg.DebugSearch.DefaultOutputMode)
}

func normalizeOutputMode(val string) types.DebugSearchOutputMode {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case string(types.DebugSearchOutputCSV):
		return types.DebugSearchOutputCSV
	default:
		return types.DebugSearchOutputSQLite
	}
}
