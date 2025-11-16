package tview2

import (
	"fmt"
	"path/filepath"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	tablePageSize = 100
)

// Run starts the simplified tview-based SQLite viewer layout.
func Run(cfg *types.DISUIConfig, dis types.DISReaderService) error {
	viewer, err := newViewerApp(cfg, dis)
	if err != nil {
		return err
	}
	return viewer.app.Run()
}

type viewerApp struct {
	cfg *types.DISUIConfig
	dis types.DISReaderService
	app *tview.Application

	dbDir   string
	files   []string
	page    int
	perPage int

	currentFile  string
	currentTable string

	dbList    *tview.List
	tableList *tview.List
	tableView *tview.TextView
	logView   *tview.TextView
	status    *tview.TextView

	focusables []tview.Primitive
	focusIndex int
	logFollow  bool
	logUnread  bool
	logLines   int
}

func newViewerApp(cfg *types.DISUIConfig, dis types.DISReaderService) (*viewerApp, error) {
	dir, err := ensureDebugOutputDir(cfg)
	if err != nil {
		return nil, err
	}
	app := tview.NewApplication()
	viewer := &viewerApp{
		cfg:       cfg,
		dis:       dis,
		app:       app,
		dbDir:     dir,
		perPage:   tablePageSize,
		logFollow: true,
	}
	viewer.initWidgets()
	viewer.mountLayout()
	viewer.loadFiles()
	app.SetRoot(viewer.buildRoot(), true)
	app.SetInputCapture(viewer.handleGlobalKeys)
	viewer.focusables = []tview.Primitive{viewer.dbList, viewer.tableList, viewer.tableView, viewer.logView}
	viewer.focusIndex = 0
	app.SetFocus(viewer.dbList)
	return viewer, nil
}

func (v *viewerApp) initWidgets() {
	v.dbList = tview.NewList().ShowSecondaryText(false)
	v.dbList.SetBorder(true).SetTitle(" SQLite Files ")
	v.dbList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		v.selectFile(index)
	})

	v.tableList = tview.NewList().ShowSecondaryText(false)
	v.tableList.SetBorder(true).SetTitle(" Tables ")
	v.tableList.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		v.selectTable(index)
	})

	v.tableView = tview.NewTextView().SetDynamicColors(false)
	v.tableView.SetBorder(true).SetTitle(" Rows ")

	v.logView = tview.NewTextView().SetDynamicColors(false)
	v.logView.SetBorder(true).SetTitle(" Activity ")
	v.logView.SetScrollable(true)
	v.logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, col := v.logView.GetScrollOffset()
		switch event.Key() {
		case tcell.KeyUp:
			if row > 0 {
				v.logView.ScrollTo(row-1, col)
			}
			v.updateLogFollowAfterScroll()
			return nil
		case tcell.KeyDown:
			v.logView.ScrollTo(row+1, col)
			v.updateLogFollowAfterScroll()
			return nil
		case tcell.KeyPgUp:
			step := 10
			if row-step < 0 {
				step = row
			}
			v.logView.ScrollTo(row-step, col)
			v.updateLogFollowAfterScroll()
			return nil
		case tcell.KeyPgDn:
			v.logView.ScrollTo(row+10, col)
			v.updateLogFollowAfterScroll()
			return nil
		case tcell.KeyHome:
			v.logView.ScrollToBeginning()
			v.updateLogFollowAfterScroll()
			return nil
		case tcell.KeyEnd:
			v.logView.ScrollToEnd()
			v.resumeLogFollow()
			return nil
		}
		return event
	})
	v.logf("Bubble UI experimental mode enabled. Press q to quit.")
	v.logf("Tab arrows select panes • PgUp/PgDn change pages • Enter loads tables")

	v.status = tview.NewTextView().SetDynamicColors(true)
	v.status.SetBorder(true)
	v.updateStatus()
}

func (v *viewerApp) buildRoot() tview.Primitive {
	title := tview.NewTextView().SetDynamicColors(true)
	title.SetBorder(true).SetTitle(" DIS Reader ")
	fmt.Fprintf(title, "%s v%s — tview2 viewer", v.cfg.AppName, v.cfg.AppVersion)

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.dbList, 0, 2, true).
		AddItem(v.tableList, 0, 1, false)

	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.tableView, 0, 3, false).
		AddItem(v.logView, 0, 1, false)

	body := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 32, 0, true).
		AddItem(rightColumn, 0, 1, false)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 3, 0, false).
		AddItem(body, 0, 1, true).
		AddItem(v.status, 3, 0, false)

	return root
}

func (v *viewerApp) mountLayout() {}

func (v *viewerApp) loadFiles() {
	v.files = scanSQLiteOutputs(v.dbDir)
	v.dbList.Clear()
	if len(v.files) == 0 {
		v.dbList.AddItem("No SQLite outputs found", "", 0, nil)
		v.currentFile = ""
		v.tableList.Clear()
		v.tableView.SetText("Run a batch debug search to generate an SQLite file.")
		return
	}
	for _, path := range v.files {
		v.dbList.AddItem(filepath.Base(path), "", 0, nil)
	}
	v.dbList.SetCurrentItem(0)
	v.selectFile(0)
}

func (v *viewerApp) selectFile(index int) {
	if index < 0 || index >= len(v.files) {
		return
	}
	path := v.files[index]
	if path == v.currentFile {
		return
	}
	v.currentFile = path
	v.page = 0
	tables, err := readSQLiteTables(path)
	if err != nil {
		v.logf("Failed to read tables: %v", err)
		v.tableList.Clear()
		v.tableView.SetText(err.Error())
		return
	}
	v.tableList.Clear()
	if len(tables) == 0 {
		v.tableList.AddItem("(empty)", "", 0, nil)
		v.tableView.SetText("No tables found")
		return
	}
	for _, tbl := range tables {
		v.tableList.AddItem(tbl, "", 0, nil)
	}
	v.tableList.SetCurrentItem(0)
	v.selectTable(0)
	v.logf("Opened %s", filepath.Base(path))
	v.updateStatus()
}

func (v *viewerApp) selectTable(index int) {
	tableCount := v.tableList.GetItemCount()
	if tableCount == 0 {
		return
	}
	if index < 0 || index >= tableCount {
		return
	}
	tableName, _ := v.tableList.GetItemText(index)
	if tableName == "(empty)" {
		return
	}
	if tableName == v.currentTable {
		return
	}
	v.currentTable = tableName
	v.page = 0
	v.loadRows()
	v.logf("Viewing table %s", tableName)
	v.updateStatus()
}

func (v *viewerApp) loadRows() {
	if v.currentFile == "" || v.currentTable == "" {
		v.tableView.SetText("Select a file and table to view rows")
		return
	}
	columns, rows, err := readSQLiteRows(v.currentFile, v.currentTable, v.page, v.perPage)
	if err != nil {
		v.tableView.SetText(err.Error())
		v.logf("Failed to read rows: %v", err)
		return
	}
	tableText := formatTable(columns, rows, 120)
	header := fmt.Sprintf("%s — page %d", v.currentTable, v.page+1)
	v.tableView.SetText(fmt.Sprintf("%s\n\n%s", header, tableText))
}

func (v *viewerApp) changePage(delta int) {
	if v.currentFile == "" || v.currentTable == "" {
		return
	}
	nextPage := v.page + delta
	if nextPage < 0 {
		nextPage = 0
	}
	if nextPage == v.page {
		return
	}
	v.page = nextPage
	v.loadRows()
	v.logf("Moved to page %d", v.page+1)
	v.updateStatus()
}

func (v *viewerApp) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyPgDn:
		v.changePage(1)
		return nil
	case tcell.KeyPgUp:
		v.changePage(-1)
		return nil
	case tcell.KeyCtrlL:
		v.logView.SetText("")
		v.logLines = 0
		v.logUnread = false
		v.resumeLogFollow()
		return nil
	case tcell.KeyCtrlC:
		v.app.Stop()
		return nil
	case tcell.KeyTAB:
		v.cycleFocus(1)
		return nil
	case tcell.KeyBacktab:
		v.cycleFocus(-1)
		return nil
	}
	switch event.Rune() {
	case 'q', 'Q':
		v.app.Stop()
		return nil
	}
	return event
}

func (v *viewerApp) logf(format string, args ...interface{}) {
	fmt.Fprintf(v.logView, "%s\n", fmt.Sprintf(format, args...))
	v.logLines++
	if v.logFollow {
		v.logView.ScrollToEnd()
		v.logUnread = false
		v.updateLogTitle()
	} else {
		v.logUnread = true
		v.updateLogTitle()
	}
}

func (v *viewerApp) updateStatus() {
	host := "<unset>"
	if v.cfg != nil && v.cfg.DIS != nil && v.cfg.DIS.Host != "" {
		host = v.cfg.DIS.Host
	}
	file := "<none>"
	if v.currentFile != "" {
		file = filepath.Base(v.currentFile)
	}
	table := "<none>"
	if v.currentTable != "" {
		table = v.currentTable
	}
	fmt.Fprintf(v.status, "Host: %s | File: %s | Table: %s | Page: %d", host, file, table, v.page+1)
}

func (v *viewerApp) cycleFocus(delta int) {
	if len(v.focusables) == 0 {
		return
	}
	v.focusIndex = (v.focusIndex + delta + len(v.focusables)) % len(v.focusables)
	v.app.SetFocus(v.focusables[v.focusIndex])
}

func (v *viewerApp) pauseLogFollow() {
	if v.logFollow {
		v.logFollow = false
		v.updateLogTitle()
	}
}

func (v *viewerApp) resumeLogFollow() {
	if v.logFollow {
		if v.logUnread {
			v.logUnread = false
			v.updateLogTitle()
		}
		return
	}
	v.logFollow = true
	v.logUnread = false
	v.updateLogTitle()
	v.logView.ScrollToEnd()
}

func (v *viewerApp) updateLogTitle() {
	title := " Activity "
	if v.logUnread {
		title = " Activity (new) "
	}
	v.logView.SetTitle(title)
}

func (v *viewerApp) updateLogFollowAfterScroll() {
	if v.isLogAtBottom() {
		v.resumeLogFollow()
	} else {
		v.pauseLogFollow()
	}
}

func (v *viewerApp) isLogAtBottom() bool {
	_, _, _, height := v.logView.GetInnerRect()
	if height <= 0 {
		return true
	}
	if v.logLines <= height {
		return true
	}
	row, _ := v.logView.GetScrollOffset()
	return row+height >= v.logLines
}
