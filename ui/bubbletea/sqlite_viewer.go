package bubbletea

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteViewerModel struct {
	files        []string
	selectedFile int

	tables        []string
	selectedTable int

	columns []string
	rows    [][]string

	page     int
	pageSize int

	status     string
	errMsg     string
	loading    bool
	defaultDir string
	width      int
}

func newSQLiteViewerModel() sqliteViewerModel {
	return sqliteViewerModel{
		pageSize: 100,
		status:   "No debug-search output loaded yet",
	}
}

func (s sqliteViewerModel) withDefaultDir(dir string) sqliteViewerModel {
	s.defaultDir = dir
	if dir != "" {
		if files := scanSQLiteOutputs(dir); len(files) > 0 {
			s.files = files
			s.selectedFile = 0
		}
	}
	return s
}

func (s sqliteViewerModel) Update(msg tea.Msg, active bool) (sqliteViewerModel, tea.Cmd, bool) {
	var cmds []tea.Cmd
	handled := false

	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = m.Width
	case tea.KeyMsg:
		if !active {
			break
		}
		switch m.String() {
		case "left", "h":
			if s.moveFile(-1) {
				handled = true
				var cmd tea.Cmd
				s, cmd = s.requestTablesLoad()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		case "right", "l":
			if s.moveFile(1) {
				handled = true
				var cmd tea.Cmd
				s, cmd = s.requestTablesLoad()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		case "up", "k":
			if s.moveTable(-1) {
				handled = true
				var cmd tea.Cmd
				s, cmd = s.requestRowsLoad()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		case "down", "j":
			if s.moveTable(1) {
				handled = true
				var cmd tea.Cmd
				s, cmd = s.requestRowsLoad()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		case "pgdn", ".", "f":
			if s.advancePage(1) {
				handled = true
				var cmd tea.Cmd
				s, cmd = s.requestRowsLoad()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		case "pgup", ",", "b":
			if s.advancePage(-1) {
				handled = true
				var cmd tea.Cmd
				s, cmd = s.requestRowsLoad()
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		case "r":
			handled = true
			var cmd tea.Cmd
			s, cmd = s.requestTablesLoad()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case sqliteTablesLoadedMsg:
		if len(s.files) == 0 || s.currentFile() != m.File {
			break
		}
		if m.Err != nil {
			s.errMsg = m.Err.Error()
			s.loading = false
			s.tables = nil
			s.rows = nil
			break
		}
		s.errMsg = ""
		s.tables = m.Tables
		s.selectedTable = 0
		s.page = 0
		s.loading = false
		s.status = fmt.Sprintf("Loaded %d tables from %s", len(m.Tables), filepath.Base(m.File))
		var cmd tea.Cmd
		s, cmd = s.requestRowsLoad()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case sqliteRowsLoadedMsg:
		if len(s.files) == 0 || s.currentFile() != m.File || s.currentTable() != m.Table || s.page != m.Page {
			break
		}
		if m.Err != nil {
			s.errMsg = m.Err.Error()
			s.loading = false
			break
		}
		s.errMsg = ""
		s.columns = m.Columns
		s.rows = m.Rows
		s.loading = false
		s.status = fmt.Sprintf("%s — %s (page %d)", filepath.Base(m.File), m.Table, m.Page+1)
	case sqliteScanMsg:
		if len(m.Files) > 0 {
			s.files = m.Files
			s.selectedFile = 0
			var cmd tea.Cmd
			s, cmd = s.requestTablesLoad()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) == 0 {
		return s, nil, handled
	}
	return s, tea.Batch(cmds...), handled
}

func (s sqliteViewerModel) View() string {
	if len(s.files) == 0 {
		return placeholderStyle.Render("Run a batch debug search to generate an SQLite file, then inspect it here.")
	}

	fileLine := fieldLabelStyle.Render("File:") + " " + infoStyle.Render(filepath.Base(s.currentFile()))
	if len(s.files) > 1 {
		fileLine += helpStyle.Render("  ←/→ switch")
	}

	tableLine := fieldLabelStyle.Render("Table:")
	if len(s.tables) == 0 {
		tableLine += " (no tables)"
	} else {
		tableLine += " " + infoStyle.Render(s.currentTable())
		if len(s.tables) > 1 {
			tableLine += helpStyle.Render("  ↑/↓ switch tables")
		}
	}

	var rowsView string
	if len(s.rows) == 0 {
		if s.loading {
			rowsView = noticeStyle.Render("Loading rows…")
		} else if s.errMsg != "" {
			rowsView = errorStyle.Render(s.errMsg)
		} else {
			rowsView = noticeStyle.Render("No rows on this page")
		}
	} else {
		rowsView = s.renderTable()
	}

	shortcutLines := []string{
		"←/→ files", "↑/↓ tables", "PgUp/PgDn pages", "r reload tables", "Ctrl+C quit",
	}
	pager := helpStyle.Render(strings.Join(shortcutLines, "  "))

	status := s.status
	if s.errMsg != "" {
		status = errorStyle.Render(s.errMsg)
	} else if status != "" {
		status = noticeStyle.Render(status)
	}

	sections := []string{sectionTitleStyle.Render("SQLite Viewer"), fileLine, tableLine, rowsView, status, pager}
	return strings.Join(sections, "\n\n")
}

func (s sqliteViewerModel) StatusLine() string {
	if len(s.files) == 0 {
		return "SQLite Viewer: waiting for output"
	}
	info := fmt.Sprintf("SQLite Viewer: %s", filepath.Base(s.currentFile()))
	if tbl := s.currentTable(); tbl != "" {
		info += fmt.Sprintf(" • %s (page %d)", tbl, s.page+1)
	}
	return info
}

func (s sqliteViewerModel) currentFile() string {
	if len(s.files) == 0 || s.selectedFile >= len(s.files) {
		return ""
	}
	return s.files[s.selectedFile]
}

func (s sqliteViewerModel) currentTable() string {
	if len(s.tables) == 0 || s.selectedTable >= len(s.tables) {
		return ""
	}
	return s.tables[s.selectedTable]
}

func (s sqliteViewerModel) moveFile(delta int) bool {
	if len(s.files) == 0 {
		return false
	}
	newIdx := s.selectedFile + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(s.files) {
		newIdx = len(s.files) - 1
	}
	if newIdx == s.selectedFile {
		return false
	}
	s.selectedFile = newIdx
	s.tables = nil
	s.rows = nil
	s.page = 0
	return true
}

func (s sqliteViewerModel) moveTable(delta int) bool {
	if len(s.tables) == 0 {
		return false
	}
	newIdx := s.selectedTable + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(s.tables) {
		newIdx = len(s.tables) - 1
	}
	if newIdx == s.selectedTable {
		return false
	}
	s.selectedTable = newIdx
	s.page = 0
	return true
}

func (s sqliteViewerModel) advancePage(delta int) bool {
	if len(s.tables) == 0 {
		return false
	}
	newPage := s.page + delta
	if newPage < 0 {
		newPage = 0
	}
	if newPage == s.page {
		return false
	}
	s.page = newPage
	return true
}

func (s sqliteViewerModel) AddOutputFile(path string) (sqliteViewerModel, tea.Cmd) {
	abs := filepath.Clean(path)
	for _, existing := range s.files {
		if existing == abs {
			return s, nil
		}
	}
	s.files = append([]string{abs}, s.files...)
	s.selectedFile = 0
	s.tables = nil
	s.rows = nil
	s.page = 0
	return s, loadTablesCmd(abs)
}

func (s sqliteViewerModel) requestTablesLoad() (sqliteViewerModel, tea.Cmd) {
	file := s.currentFile()
	if file == "" {
		return s, nil
	}
	s.loading = true
	return s, loadTablesCmd(file)
}

func (s sqliteViewerModel) requestRowsLoad() (sqliteViewerModel, tea.Cmd) {
	file := s.currentFile()
	table := s.currentTable()
	if file == "" || table == "" {
		return s, nil
	}
	s.loading = true
	return s, loadRowsCmd(file, table, s.page, s.pageSize)
}

func (s sqliteViewerModel) renderTable() string {
	if len(s.columns) == 0 {
		return noticeStyle.Render("No columns to display")
	}
	colWidths := make([]int, len(s.columns))
	maxWidth := s.width
	if maxWidth <= 0 {
		maxWidth = 80
	}
	for i, col := range s.columns {
		colWidths[i] = len(col)
	}
	for _, row := range s.rows {
		for i, val := range row {
			if l := len(val); l > colWidths[i] {
				colWidths[i] = l
			}
		}
	}
	padding := 3 * (len(colWidths) - 1)
	total := padding
	for _, w := range colWidths {
		if w > 40 {
			w = 40
		}
		total += w
	}
	if total > maxWidth {
		scale := float64(maxWidth-padding) / float64(total-padding)
		for i := range colWidths {
			colWidths[i] = int(float64(colWidths[i]) * scale)
			if colWidths[i] < 4 {
				colWidths[i] = 4
			}
		}
	}
	formatRow := func(cells []string, style lipgloss.Style) string {
		parts := make([]string, len(cells))
		for i, cell := range cells {
			parts[i] = style.Width(colWidths[i]).Render(truncate(cell, colWidths[i]))
		}
		return strings.Join(parts, " │ ")
	}
	header := formatRow(s.columns, sectionTitleStyle)
	separator := strings.Repeat("─", len(header))
	var builder strings.Builder
	builder.WriteString(header)
	builder.WriteString("\n" + separator)
	for _, row := range s.rows {
		builder.WriteString("\n" + formatRow(row, lipgloss.NewStyle()))
	}
	return builder.String()
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 1 {
		return text[:max]
	}
	return text[:max-1] + "…"
}

func loadTablesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		tables, err := readSQLiteTables(path)
		return sqliteTablesLoadedMsg{File: path, Tables: tables, Err: err}
	}
}

func loadRowsCmd(path, table string, page, pageSize int) tea.Cmd {
	return func() tea.Msg {
		columns, rows, err := readSQLiteRows(path, table, page, pageSize)
		return sqliteRowsLoadedMsg{File: path, Table: table, Columns: columns, Rows: rows, Page: page, Err: err}
	}
}

func readSQLiteTables(path string) ([]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func readSQLiteRows(path, table string, page, pageSize int) ([]string, [][]string, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", table, pageSize, page*pageSize)
	rows, err := db.Query(query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var result [][]string
	for rows.Next() {
		raw := make([]interface{}, len(columns))
		rawPtrs := make([]interface{}, len(columns))
		for i := range raw {
			rawPtrs[i] = &raw[i]
		}
		if err := rows.Scan(rawPtrs...); err != nil {
			return nil, nil, err
		}
		formatted := make([]string, len(columns))
		for i, val := range raw {
			if b, ok := val.([]byte); ok {
				formatted[i] = string(b)
			} else if val == nil {
				formatted[i] = "<nil>"
			} else {
				formatted[i] = fmt.Sprint(val)
			}
		}
		result = append(result, formatted)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return columns, result, nil
}

type sqliteTablesLoadedMsg struct {
	File   string
	Tables []string
	Err    error
}

type sqliteRowsLoadedMsg struct {
	File    string
	Table   string
	Columns []string
	Rows    [][]string
	Page    int
	Err     error
}

type sqliteScanMsg struct {
	Files []string
}
