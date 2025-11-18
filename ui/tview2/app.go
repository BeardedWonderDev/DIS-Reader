package tview2

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	tablePageSize     = 4
	tableVisibleRows  = 4
	approxColumnWidth = 18
	partsSearchLimit  = 200
	pageDebug         = "debug"
	pageParts         = "parts"
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
	pages      *tview.Pages
	navBar     *tview.TextView
	activePage string
	parts      *partsPanel

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
	viewer.parts = newPartsPanel(viewer)
	root := viewer.buildRoot()
	app.SetRoot(root, true)
	app.SetInputCapture(viewer.handleGlobalKeys)
	viewer.switchToPage(pageDebug)
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
		SetFixed(0, 0).
		SetSelectable(true, false)
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

func (v *viewerApp) buildDebugPage() tview.Primitive {
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

func (v *viewerApp) buildRoot() tview.Primitive {
	v.pages = tview.NewPages()
	v.pages.AddPage(pageDebug, v.buildDebugPage(), true, true)
	if v.parts != nil {
		v.pages.AddPage(pageParts, v.parts.Root(), true, false)
	}
	v.navBar = v.buildNavBar()
	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.navBar, 1, 0, false).
		AddItem(v.pages, 0, 1, true)

	return root
}

func (v *viewerApp) buildNavBar() *tview.TextView {
	text := tview.NewTextView().SetDynamicColors(true)
	text.SetBackgroundColor(v.theme.Colors.CommandBarColor)
	text.SetTextColor(v.theme.Colors.PrimaryText)
	text.SetBorder(true)
	text.SetTitle(" Navigation ")
	text.SetText(v.navBarText())
	return text
}

func (v *viewerApp) navBarText() string {
	var builder strings.Builder
	segments := []struct {
		page  string
		label string
	}{
		{pageDebug, "F1 Files Viewer"},
		{pageParts, "F2 Parts Lookup"},
	}
	for i, seg := range segments {
		if v.activePage == seg.page {
			builder.WriteString(fmt.Sprintf("[::b][green]%s[-::]", seg.label))
		} else {
			builder.WriteString(seg.label)
		}
		if i < len(segments)-1 {
			builder.WriteString("   •   ")
		}
	}
	builder.WriteString("   |   Tab = Focus  •  PgUp/PgDn = Page  •  Q = Quit")
	return builder.String()
}

func (v *viewerApp) updateNavBar() {
	if v.navBar == nil {
		return
	}
	v.navBar.SetText(v.navBarText())
}

func (v *viewerApp) switchToPage(page string) {
	if v.pages == nil {
		return
	}
	v.pages.SwitchToPage(page)
	v.activePage = page
	v.updateNavBar()
	switch page {
	case pageParts:
		if v.parts != nil {
			v.focusables = v.parts.focusables()
			v.focusIndex = 0
			v.focusPane(0)
		}
	default:
		v.focusables = []tview.Primitive{v.dbTable, v.tableTable, v.tableView, v.logView}
		v.focusIndex = 0
		v.focusPane(0)
	}
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

	if len(rows) == 0 {
		v.renderTableMessage("No rows on this page")
		return
	}

	groupSize := v.determineColumnGroupSize(len(columns))
	if groupSize <= 0 {
		groupSize = len(columns)
	}
	groupRanges := buildColumnRanges(len(columns), groupSize)

	rowCursor := 0
	visibleRows := tableVisibleRows
	if len(rows) < visibleRows {
		visibleRows = len(rows)
	}

	for idx, cr := range groupRanges {
		// header row for this group
		for colIdx, name := range columns[cr.start:cr.end] {
			v.tableView.SetCell(rowCursor, colIdx,
				tview.NewTableCell(name).
					SetSelectable(false).
					SetStyle(v.theme.Style.TableHeaderStyle))
		}
		rowCursor++

		// data rows
		for r := 0; r < visibleRows; r++ {
			data := rows[r]
			for colIdx := range columns[cr.start:cr.end] {
				val := ""
				sourceIdx := cr.start + colIdx
				if sourceIdx < len(data) {
					val = fmt.Sprint(data[sourceIdx])
				}
				v.tableView.SetCell(rowCursor, colIdx,
					tview.NewTableCell(val).
						SetStyle(v.theme.Style.TableCellStyle))
			}
			rowCursor++
		}

		// overflow hint
		if len(rows) > visibleRows {
			v.tableView.SetCell(rowCursor, 0,
				tview.NewTableCell("…").
					SetSelectable(false).
					SetAlign(tview.AlignCenter).
					SetStyle(v.theme.Style.TableCellStyle))
			rowCursor++
		}

		// spacer row
		if idx < len(groupRanges)-1 {
			v.tableView.SetCell(rowCursor, 0,
				tview.NewTableCell("").
					SetSelectable(false))
			rowCursor++
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
		SetExpansion(1).
		SetStyle(v.theme.Style.PickerCellStyle)
}

func (v *viewerApp) pickerMessageCell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetSelectable(false).
		SetAlign(tview.AlignCenter).
		SetExpansion(1).
		SetStyle(v.theme.Style.PickerCellStyle)
}

func (v *viewerApp) determineColumnGroupSize(totalColumns int) int {
	if totalColumns == 0 {
		return 0
	}
	_, _, width, _ := v.tableView.GetInnerRect()
	if width <= 0 {
		width = totalColumns * approxColumnWidth
	}
	target := approxColumnWidth + 2
	group := width / target
	if group < 1 {
		group = 1
	}
	if group > totalColumns {
		group = totalColumns
	}
	return group
}

type columnRange struct {
	start int
	end   int
}

func buildColumnRanges(total, size int) []columnRange {
	if size <= 0 {
		size = total
	}
	var ranges []columnRange
	for start := 0; start < total; start += size {
		end := start + size
		if end > total {
			end = total
		}
		ranges = append(ranges, columnRange{start: start, end: end})
	}
	return ranges
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
	case tcell.KeyF1:
		v.switchToPage(pageDebug)
		return nil
	case tcell.KeyF2:
		v.switchToPage(pageParts)
		return nil
	case tcell.KeyPgDn:
		if v.activePage == pageDebug {
			v.changePage(1)
			return nil
		}
	case tcell.KeyPgUp:
		if v.activePage == pageDebug {
			v.changePage(-1)
			return nil
		}
	case tcell.KeyCtrlL:
		if v.activePage == pageDebug {
			v.logView.SetText("")
			v.logLines = 0
			v.logUnread = false
			v.resumeLogFollow()
			return nil
		}
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
	if v.focusInTextEntry() {
		return event
	}
	switch event.Rune() {
	case 'q', 'Q':
		v.app.Stop()
		return nil
	case '1':
		v.switchToPage(pageDebug)
		return nil
	case '2':
		v.switchToPage(pageParts)
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
	legend := "F1 Files • F2 Parts • PgUp/PgDn Page • Tab Focus • Ctrl+L Clear Log • Q Quit"
	statusText := fmt.Sprintf("Host: %s  •  Page: %d  •  %s", host, v.page+1, legend)
	v.status.SetText(statusText)
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

func (v *viewerApp) focusInTextEntry() bool {
	if v.app == nil {
		return false
	}
	focus := v.app.GetFocus()
	switch focus.(type) {
	case *tview.InputField, *tview.TextArea:
		return true
	}
	return false
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

type partsPanel struct {
	viewer        *viewerApp
	root          *tview.Flex
	divisionInput *tview.InputField
	partInput     *tview.InputField
	queryInput    *tview.InputField
	vendorInput   *tview.InputField
	classInput    *tview.InputField
	fetchButton   *tview.Button
	searchButton  *tview.Button
	resultsTable  *tview.Table
	detailView    *tview.TextView
	status        *tview.TextView
	current       []*types.PartInventorySpec
}

func newPartsPanel(v *viewerApp) *partsPanel {
	p := &partsPanel{viewer: v}
	p.build()
	return p
}

func (p *partsPanel) Root() tview.Primitive {
	return p.root
}

func (p *partsPanel) build() {
	theme := p.viewer.theme
	p.divisionInput = tview.NewInputField()
	p.divisionInput.SetLabel("Division: ")
	p.divisionInput.SetFieldWidth(8)
	p.divisionInput.SetPlaceholder("001")
	p.divisionInput.SetBackgroundColor(theme.Colors.WindowColor)
	p.divisionInput.SetFieldStyle(theme.Style.FieldStyle)
	p.partInput = tview.NewInputField()
	p.partInput.SetLabel("Part #: ")
	p.partInput.SetPlaceholder("ABC123")
	p.partInput.SetBackgroundColor(theme.Colors.WindowColor)
	p.partInput.SetFieldStyle(theme.Style.FieldStyle)
	p.partInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			p.runLookup()
		}
	})
	p.queryInput = tview.NewInputField()
	p.queryInput.SetLabel("Search: ")
	p.queryInput.SetPlaceholder("description, vendor, etc")
	p.queryInput.SetBackgroundColor(theme.Colors.WindowColor)
	p.queryInput.SetFieldStyle(theme.Style.FieldStyle)
	p.queryInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			p.runSearch()
		}
	})
	p.vendorInput = tview.NewInputField()
	p.vendorInput.SetLabel("Vendor: ")
	p.vendorInput.SetFieldWidth(6)
	p.vendorInput.SetBackgroundColor(theme.Colors.WindowColor)
	p.vendorInput.SetFieldStyle(theme.Style.FieldStyle)
	p.classInput = tview.NewInputField()
	p.classInput.SetLabel("Class: ")
	p.classInput.SetFieldWidth(4)
	p.classInput.SetBackgroundColor(theme.Colors.WindowColor)
	p.classInput.SetFieldStyle(theme.Style.FieldStyle)
	p.fetchButton = tview.NewButton("Lookup Part")
	p.fetchButton.SetSelectedFunc(p.runLookup)
	p.fetchButton.SetStyle(theme.Style.ButtonStyle)
	p.searchButton = tview.NewButton("Search Parts")
	p.searchButton.SetSelectedFunc(p.runSearch)
	p.searchButton.SetStyle(theme.Style.ButtonStyle)
	p.resultsTable = tview.NewTable()
	p.resultsTable.SetFixed(1, 0)
	p.resultsTable.SetSelectable(true, false)
	p.resultsTable.SetBorder(true)
	p.resultsTable.SetTitle(" Results ")
	p.resultsTable.SetBorderColor(theme.Colors.BorderColor)
	p.resultsTable.SetTitleColor(theme.Colors.PrimaryText)
	p.resultsTable.SetSelectedStyle(theme.Style.TableSelectedStyle)
	p.resultsTable.SetSelectionChangedFunc(func(row, column int) {
		p.showResult(row - 1)
	})
	p.detailView = tview.NewTextView()
	p.detailView.SetDynamicColors(true)
	p.detailView.SetWrap(true)
	p.detailView.SetScrollable(true)
	p.detailView.SetBorder(true)
	p.detailView.SetTitle(" Part Details ")
	p.detailView.SetBorderColor(theme.Colors.BorderColor)
	p.detailView.SetTitleColor(theme.Colors.PrimaryText)
	p.detailView.SetBackgroundColor(theme.Colors.WindowColor)
	p.status = tview.NewTextView()
	p.status.SetDynamicColors(true)
	p.status.SetBorder(true)
	p.status.SetBorderColor(theme.Colors.BorderColor)
	p.status.SetBackgroundColor(theme.Colors.StatusBarBg)
	p.status.SetTextStyle(theme.Style.StatusBarStyle)
	title := tview.NewTextView()
	title.SetDynamicColors(true)
	title.SetBorder(true)
	title.SetTitle(" Parts Lookup ")
	title.SetBorderColor(theme.Colors.BorderColor)
	title.SetBackgroundColor(theme.Colors.CommandBarColor)
	title.SetTextColor(theme.Colors.PrimaryText)
	flexLookup := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(p.divisionInput, 0, 1, true).
		AddItem(p.partInput, 0, 2, false).
		AddItem(p.fetchButton, 20, 0, false)
	flexSearch := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(p.queryInput, 0, 2, false).
		AddItem(p.vendorInput, 0, 1, false).
		AddItem(p.classInput, 0, 1, false).
		AddItem(p.searchButton, 20, 0, false)
	controls := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flexLookup, 3, 0, true).
		AddItem(flexSearch, 3, 0, false)
	body := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(p.resultsTable, 0, 2, true).
		AddItem(p.detailView, 0, 3, false)
	p.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 3, 0, false).
		AddItem(controls, 6, 0, true).
		AddItem(body, 0, 1, true).
		AddItem(p.status, 3, 0, false)
	p.detailView.SetText("Use the lookup or search inputs to load part details.")
	p.setStatus("Enter a part number per division or run a search with filters. F1=Files, F2=Parts.")
}

func (p *partsPanel) focusables() []tview.Primitive {
	return []tview.Primitive{
		p.divisionInput,
		p.partInput,
		p.fetchButton,
		p.queryInput,
		p.vendorInput,
		p.classInput,
		p.searchButton,
		p.resultsTable,
		p.detailView,
	}
}

func (p *partsPanel) setStatus(msg string) {
	if msg == "" {
		msg = "Ready."
	}
	p.status.SetText(msg)
}

func (p *partsPanel) setBusy(busy bool, msg string) {
	p.fetchButton.SetDisabled(busy)
	p.searchButton.SetDisabled(busy)
	if busy {
		p.setStatus(fmt.Sprintf("[yellow]%s", msg))
	} else if msg != "" {
		p.setStatus(msg)
	}
}

func (p *partsPanel) runLookup() {
	division := strings.TrimSpace(p.divisionInput.GetText())
	part := strings.TrimSpace(p.partInput.GetText())
	if division == "" || part == "" {
		p.setStatus("[red]Division and Part # are required for lookup")
		return
	}
	svc := p.viewer.dis.PartService()
	if svc == nil {
		p.setStatus("[red]Part service unavailable")
		return
	}
	p.setBusy(true, fmt.Sprintf("Looking up %s/%s...", division, part))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	go func() {
		defer cancel()
		spec, err := svc.GetPart(ctx, division, part)
		p.viewer.app.QueueUpdateDraw(func() {
			p.setBusy(false, "")
			if err != nil {
				p.setStatus(fmt.Sprintf("[red]Lookup failed: %v", err))
				return
			}
			p.applyResults([]*types.PartInventorySpec{spec})
		})
	}()
}

func (p *partsPanel) runSearch() {
	division := strings.TrimSpace(p.divisionInput.GetText())
	query := strings.TrimSpace(p.queryInput.GetText())
	vendor := strings.TrimSpace(p.vendorInput.GetText())
	classCode := strings.TrimSpace(p.classInput.GetText())
	if division == "" && query == "" {
		p.setStatus("[red]Provide a division or search text to narrow results")
		return
	}
	svc := p.viewer.dis.PartService()
	if svc == nil {
		p.setStatus("[red]Part service unavailable")
		return
	}
	filters := make([]types.Filter, 0)
	if division != "" {
		filters = append(filters, types.Filter{Column: "PIMDIV", Operator: "=", Value: division})
	}
	if vendor != "" {
		filters = append(filters, types.Filter{Column: "PIMVEN", Operator: "=", Value: vendor})
	}
	if classCode != "" {
		filters = append(filters, types.Filter{Column: "PIMCLS", Operator: "=", Value: classCode})
	}
	lp := types.ListParams{
		Limit:     partsSearchLimit,
		SortBy:    "PIMPRT",
		SortOrder: types.SortOrderAsc,
		Query:     query,
		Filters:   filters,
	}
	p.setBusy(true, "Searching parts...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	go func() {
		defer cancel()
		rows, err := svc.ListParts(ctx, lp)
		p.viewer.app.QueueUpdateDraw(func() {
			p.setBusy(false, "")
			if err != nil {
				p.setStatus(fmt.Sprintf("[red]Search failed: %v", err))
				return
			}
			p.applyResults(rows)
		})
	}()
}

func (p *partsPanel) applyResults(items []*types.PartInventorySpec) {
	p.current = items
	p.resultsTable.Clear()
	headers := []string{"Div", "Part", "Vendor", "Description", "Avail", "OnHand"}
	for col, label := range headers {
		cell := tview.NewTableCell(label).
			SetSelectable(false).
			SetStyle(p.viewer.theme.Style.TableHeaderStyle)
		p.resultsTable.SetCell(0, col, cell)
	}
	if len(items) == 0 {
		msg := tview.NewTableCell("No parts found").
			SetSelectable(false).
			SetStyle(p.viewer.theme.Style.TableCellStyle)
		p.resultsTable.SetCell(1, 0, msg)
		p.setStatus("No matching parts")
		p.detailView.SetText("Use the filters above to search for parts.")
		return
	}
	for i, item := range items {
		row := i + 1
		p.resultsTable.SetCell(row, 0, tview.NewTableCell(strings.TrimSpace(item.Division)).SetStyle(p.viewer.theme.Style.TableCellStyle))
		p.resultsTable.SetCell(row, 1, tview.NewTableCell(strings.TrimSpace(item.PartNumber)).SetStyle(p.viewer.theme.Style.TableCellStyle))
		p.resultsTable.SetCell(row, 2, tview.NewTableCell(strings.TrimSpace(item.VendorCode)).SetStyle(p.viewer.theme.Style.TableCellStyle))
		desc := strings.TrimSpace(item.Description)
		if len(desc) > 32 {
			desc = desc[:32] + "…"
		}
		p.resultsTable.SetCell(row, 3, tview.NewTableCell(desc).SetStyle(p.viewer.theme.Style.TableCellStyle))
		p.resultsTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%d", item.AvailableQty)).SetStyle(p.viewer.theme.Style.TableCellStyle))
		p.resultsTable.SetCell(row, 5, tview.NewTableCell(fmt.Sprintf("%d", item.OnHandQty)).SetStyle(p.viewer.theme.Style.TableCellStyle))
	}
	p.resultsTable.Select(1, 0)
	p.showResult(0)
	p.setStatus(fmt.Sprintf("Loaded %d part(s).", len(items)))
}

func (p *partsPanel) showResult(index int) {
	if index < 0 || index >= len(p.current) {
		p.detailView.SetText("Select a part to view its details.")
		return
	}
	p.detailView.SetText(p.formatPartDetail(p.current[index]))
}

func (p *partsPanel) formatPartDetail(spec *types.PartInventorySpec) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "[::b]%s[-::-] — %s\n", strings.TrimSpace(spec.PartNumber), strings.TrimSpace(spec.Description))
	fmt.Fprintf(&builder, "Division: %s   Vendor: %s   Class: %s   Quick: %s\n",
		strings.TrimSpace(spec.Division), strings.TrimSpace(spec.VendorCode), strings.TrimSpace(spec.ItemClass), strings.TrimSpace(spec.QuickCode))
	fmt.Fprintf(&builder, "Bin: %s   Tax: %s   Active: %s   Stock: %s\n",
		strings.TrimSpace(spec.BinLocation), strings.TrimSpace(spec.TaxCode), strings.TrimSpace(spec.ActiveFlag), strings.TrimSpace(spec.StockFlag))
	fmt.Fprintf(&builder, "On Hand: %d   Available: %d   Reserved SA/WO: %d/%d   Returnable: %s\n",
		spec.OnHandQty, spec.AvailableQty, spec.ReservedSA, spec.ReservedWO, strings.TrimSpace(spec.ReturnableFlag))
	fmt.Fprintf(&builder, "Avg Cost: %.2f   Price1(%s): %.2f   Price2(%s): %.2f\n",
		spec.AvgCost, strings.TrimSpace(spec.Price1Basis), spec.Price1Value, strings.TrimSpace(spec.Price2Basis), spec.Price2Value)
	fmt.Fprintf(&builder, "Price3(%s): %.2f   Price4(%s): %.2f\n",
		strings.TrimSpace(spec.Price3Basis), spec.Price3Value, strings.TrimSpace(spec.Price4Basis), spec.Price4Value)
	fmt.Fprintf(&builder, "Order Min/Max: %d/%d   Mult: %d   Std Order: %d\n",
		spec.MinQty, spec.MaxQty, spec.OrderMultiplier, spec.OrderQty)
	fmt.Fprintf(&builder, "Last Receipt: %d (%d qty)   Last Price Update: %d\n",
		spec.LastReceiptDate, spec.LastReceiptQty, spec.DateLastPriceUpdate)
	fmt.Fprintf(&builder, "Core: %s (qty %d price %.2f)   Supersession: %s/%s/%s\n",
		strings.TrimSpace(spec.CorePartNumber), spec.CoreQty, spec.CorePrice,
		strings.TrimSpace(spec.Sup1Part), strings.TrimSpace(spec.Sup2Part), strings.TrimSpace(spec.Sup3Part))
	usage := formatRecentInts(spec.MonthlyUsage, 6)
	purchases := formatRecentInts(spec.MonthlyPurchases, 6)
	fmt.Fprintf(&builder, "Monthly usage (latest 6): %s\n", usage)
	fmt.Fprintf(&builder, "Monthly purchases (latest 6): %s\n", purchases)
	fmt.Fprintf(&builder, "Year usage qty (Y0-Y2): %s\n", formatLeadingInts(spec.YearUsageQty, 3))
	fmt.Fprintf(&builder, "Year usage values (Y0-Y2): %s\n", formatLeadingInt64(spec.YearUsageValue, 3))
	comment := strings.TrimSpace(spec.Comment)
	if comment != "" {
		fmt.Fprintf(&builder, "Comment: %s\n", comment)
	}
	return builder.String()
}

func formatRecentInts(vals []int32, count int) string {
	if len(vals) == 0 {
		return "n/a"
	}
	start := len(vals) - count
	if start < 0 {
		start = 0
	}
	vals = vals[start:]
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ", ")
}

func formatLeadingInts(vals []int32, count int) string {
	if len(vals) == 0 {
		return "n/a"
	}
	if len(vals) < count {
		count = len(vals)
	}
	parts := make([]string, count)
	for i := 0; i < count; i++ {
		parts[i] = fmt.Sprintf("%d", vals[i])
	}
	return strings.Join(parts, ", ")
}

func formatLeadingInt64(vals []int64, count int) string {
	if len(vals) == 0 {
		return "n/a"
	}
	if len(vals) < count {
		count = len(vals)
	}
	parts := make([]string, count)
	for i := 0; i < count; i++ {
		parts[i] = fmt.Sprintf("%d", vals[i])
	}
	return strings.Join(parts, ", ")
}
