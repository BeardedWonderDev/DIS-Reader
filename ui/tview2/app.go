package tview2

import (
	"fmt"
	"path/filepath"
	"strings"

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
	cfg   *types.DISUIConfig
	dis   types.DISReaderService
	app   *tview.Application
	theme *types.Theme

	dbDir      string
	files      []string
	tableNames []string
	page       int
	perPage    int

	currentFile  string
	currentTable string

	dbTable    *tview.Table
	tableTable *tview.Table
	tableView  *tview.Table
	logView    *tview.TextView
	status     *tview.TextView

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
	theme := types.ResolveTheme(cfg.Theme)
	viewer := &viewerApp{
		cfg:       cfg,
		dis:       dis,
		app:       app,
		theme:     theme,
		dbDir:     dir,
		perPage:   tablePageSize,
		logFollow: true,
	}
	viewer.initWidgets()
	viewer.mountLayout()
	viewer.loadFiles()
	app.SetRoot(viewer.buildRoot(), true)
	app.SetInputCapture(viewer.handleGlobalKeys)
	viewer.focusables = []tview.Primitive{viewer.dbTable, viewer.tableTable, viewer.tableView, viewer.logView}
	viewer.focusPane(0)
	return viewer, nil
}

func (v *viewerApp) initWidgets() {
	v.dbTable = v.newPickerTable(" SQLite Files ")
	v.dbTable.SetSelectionChangedFunc(func(row, column int) {
		v.selectFile(row)
	})

	v.tableTable = v.newPickerTable(" Tables ")
	v.tableTable.SetSelectionChangedFunc(func(row, column int) {
		v.selectTable(row)
	})

	v.tableView = tview.NewTable().
		SetFixed(1, 0).
		SetSelectable(true, true)
	v.tableView.SetBorder(true).SetTitle(" Rows ")
	v.tableView.SetBorders(true)
	v.tableView.SetBorderColor(v.theme.Colors.BorderColor)
	v.tableView.SetTitleColor(v.theme.Colors.PrimaryText)
	v.tableView.SetBackgroundColor(v.theme.Colors.WindowColor)
	v.tableView.SetSelectedStyle(v.theme.Style.TableSelectedStyle)

	v.logView = tview.NewTextView().SetDynamicColors(false)
	v.logView.SetBorder(true).SetTitle(" Activity ")
	v.logView.SetBorderColor(v.theme.Colors.BorderColor)
	v.logView.SetTitleColor(v.theme.Colors.PrimaryText)
	v.logView.SetBackgroundColor(v.theme.Colors.WindowColor)
	v.logView.SetTextColor(v.theme.Colors.PrimaryText)
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
	v.status.SetBorderColor(v.theme.Colors.BorderColor)
	v.status.SetBackgroundColor(v.theme.Colors.StatusBarBg)
	v.status.SetTextStyle(v.theme.Style.StatusBarStyle)
	v.status.SetTextAlign(tview.AlignCenter)
	v.updateStatus()
}

func (v *viewerApp) buildRoot() tview.Primitive {
	title := tview.NewTextView().SetDynamicColors(true)
	title.SetBorder(true).SetTitle(" DIS Reader ")
	title.SetBorderColor(v.theme.Colors.BorderColor)
	title.SetBackgroundColor(v.theme.Colors.CommandBarColor)
	title.SetTextColor(v.theme.Colors.PrimaryText)
	fmt.Fprintf(title, "%s v%s — tview2 viewer", v.cfg.AppName, v.cfg.AppVersion)

	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.dbTable, 0, 2, true).
		AddItem(v.tableTable, 0, 1, false)

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
	v.renderDBPicker()
	if len(v.files) == 0 {
		v.currentFile = ""
		v.tableNames = nil
		v.renderTablePicker()
		v.renderTableMessage("Run a batch debug search to generate an SQLite file.")
		return
	}
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
		v.tableNames = nil
		v.renderTablePicker()
		v.renderTableMessage(err.Error())
		return
	}
	v.tableNames = tables
	v.renderTablePicker()
	if len(tables) == 0 {
		v.renderTableMessage("No tables found")
		return
	}
	v.selectTable(0)
	v.logf("Opened %s", filepath.Base(path))
	v.updateStatus()
}

func (v *viewerApp) selectTable(index int) {
	tableCount := len(v.tableNames)
	if tableCount == 0 {
		return
	}
	if index < 0 || index >= tableCount {
		return
	}
	tableName := v.tableNames[index]
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
		v.renderTableMessage("Select a file and table to view rows")
		v.updateTableTitle("")
		return
	}
	columns, rows, err := readSQLiteRows(v.currentFile, v.currentTable, v.page, v.perPage)
	if err != nil {
		v.renderTableMessage(err.Error())
		v.updateTableTitle(v.currentTable)
		v.logf("Failed to read rows: %v", err)
		return
	}
	v.populateTable(columns, rows)
	v.updateTableTitle(v.currentTable)
}

func (v *viewerApp) populateTable(columns []string, rows [][]string) {
	v.tableView.Clear()
	v.tableView.SetFixed(1, 0)
	if len(columns) == 0 {
		v.renderTableMessage("No columns available")
		return
	}

	headerStyle := v.theme.Style.TableHeaderStyle
	for colIdx, name := range columns {
		cell := tview.NewTableCell(name).
			SetSelectable(false).
			SetStyle(headerStyle)
		v.tableView.SetCell(0, colIdx, cell)
	}

	if len(rows) == 0 {
		msg := tview.NewTableCell("No rows on this page").
			SetSelectable(false).
			SetStyle(v.theme.Style.TableCellStyle)
		v.tableView.SetCell(1, 0, msg)
		return
	}

	for rowIdx, row := range rows {
		for colIdx, value := range row {
			cell := tview.NewTableCell(value).
				SetStyle(v.theme.Style.TableCellStyle)
			v.tableView.SetCell(rowIdx+1, colIdx, cell)
		}
	}
}

func (v *viewerApp) renderTableMessage(message string) {
	v.tableView.Clear()
	v.tableView.SetFixed(0, 0)
	cell := tview.NewTableCell(message).
		SetSelectable(false).
		SetStyle(v.theme.Style.TableCellStyle)
	v.tableView.SetCell(0, 0, cell)
}

func (v *viewerApp) updateTableTitle(tableName string) {
	title := " Rows "
	if strings.TrimSpace(tableName) != "" {
		title = fmt.Sprintf(" %s ", tableName)
	}
	v.tableView.SetTitle(title)
}

func (v *viewerApp) newPickerTable(title string) *tview.Table {
	tbl := tview.NewTable().
		SetSelectable(true, false).
		SetFixed(0, 0)
	tbl.SetBorder(true).
		SetTitle(title).
		SetBorderColor(v.theme.Colors.BorderColor).
		SetBackgroundColor(v.theme.Colors.WindowColor).
		SetTitleColor(v.theme.Colors.PrimaryText)
	tbl.SetSelectedStyle(v.theme.Style.PickerSelectedStyle)
	return tbl
}

func (v *viewerApp) renderDBPicker() {
	v.dbTable.Clear()
	if len(v.files) == 0 {
		v.dbTable.SetSelectable(false, false)
		v.dbTable.SetCell(0, 0, v.pickerMessageCell("No SQLite outputs found"))
		return
	}
	v.dbTable.SetSelectable(true, false)
	for row, path := range v.files {
		name := filepath.Base(path)
		v.dbTable.SetCell(row, 0, v.pickerDataCell(name))
	}
	v.dbTable.Select(0, 0)
}

func (v *viewerApp) renderTablePicker() {
	v.tableTable.Clear()
	if len(v.tableNames) == 0 {
		v.tableTable.SetSelectable(false, false)
		v.tableTable.SetCell(0, 0, v.pickerMessageCell("(no tables)"))
		return
	}
	v.tableTable.SetSelectable(true, false)
	for row, name := range v.tableNames {
		v.tableTable.SetCell(row, 0, v.pickerDataCell(name))
	}
	v.tableTable.Select(0, 0)
}

func (v *viewerApp) pickerDataCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetAlign(tview.AlignCenter).
		SetStyle(v.theme.Style.PickerCellStyle)
}

func (v *viewerApp) pickerMessageCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetAlign(tview.AlignCenter).
		SetStyle(v.theme.Style.PickerCellStyle)
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
	statusLine := fmt.Sprintf("Host: %s    Page: %d", host, v.page+1)
	legendLine := "PgUp/PgDn Page  •  Tab Focus  •  Ctrl+L Clear Log  •  Q Quit"
	v.status.SetText(fmt.Sprintf("%s\n%s", statusLine, legendLine))
}

func (v *viewerApp) cycleFocus(delta int) {
	if len(v.focusables) == 0 {
		return
	}
	v.focusPane(v.focusIndex + delta)
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

func (v *viewerApp) focusPane(index int) {
	if len(v.focusables) == 0 {
		return
	}
	v.focusIndex = (index + len(v.focusables)) % len(v.focusables)
	for i, p := range v.focusables {
		v.applyPaneFocus(p, i == v.focusIndex)
	}
	v.app.SetFocus(v.focusables[v.focusIndex])
}

func (v *viewerApp) applyPaneFocus(p tview.Primitive, focused bool) {
	borderColor := v.theme.Colors.BorderColor
	titleColor := v.theme.Colors.PrimaryText
	attr := tcell.AttrNone
	if focused {
		borderColor = v.theme.Colors.AccentColor
		titleColor = v.theme.Colors.AccentColor
		attr = tcell.AttrBold
	}

	switch pane := p.(type) {
	case *tview.Table:
		pane.SetBorderColor(borderColor)
		pane.SetTitleColor(titleColor)
		pane.SetBorderAttributes(attr)
	case *tview.TextView:
		pane.SetBorderColor(borderColor)
		pane.SetTitleColor(titleColor)
		pane.SetBorderAttributes(attr)
	}
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
