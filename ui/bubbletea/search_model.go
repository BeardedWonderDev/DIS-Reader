package bubbletea

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

const (
	focusSearchField = iota
	focusFormatField
	focusOutputField
	focusRunButton
	focusCount
)

type searchModel struct {
	cfg *types.DISUIConfig
	dis types.DISReaderService

	termInput   textinput.Model
	outputInput textinput.Model
	formats     []types.DebugSearchOutputMode
	formatIndex int

	focused int
	running bool

	progressBar   progress.Model
	progressWidth int

	statusMsg string
	errMsg    string
	startedAt time.Time

	lastProgress types.ProgressStatus
	lastEvent    *types.TableEvent

	progressCh chan types.ProgressStatus
	eventCh    chan types.TableEvent

	lastOutputPath string
	authenticated  bool
	defaultDir     string
}

func newSearchModel(cfg *types.DISUIConfig, dis types.DISReaderService) searchModel {
	defaultDir := ""
	if cfg != nil {
		if dir, err := ensureDebugOutputDir(cfg); err == nil {
			defaultDir = dir
		}
	}
	term := textinput.New()
	term.Placeholder = "Search term (Part #, Invoice #, Unit #, etc.)"
	term.CharLimit = 256

	output := textinput.New()
	output.CharLimit = 512
	outputPlaceholder := defaultOutputPath(cfg, types.DebugSearchOutputSQLite)
	output.Placeholder = outputPlaceholder
	if cfg != nil {
		output.SetValue(strings.TrimSpace(cfg.DebugSearch.DefaultOutputPath))
	}

	pb := progress.New(
		progress.WithDefaultGradient(),
		progress.WithScaledGradient("#7B61FF", "#4ADE80"),
	)
	pb.Width = 50

	formats := []types.DebugSearchOutputMode{types.DebugSearchOutputSQLite, types.DebugSearchOutputCSV}
	formatIdx := 0
	if cfg != nil && strings.EqualFold(cfg.DebugSearch.DefaultOutputMode, string(types.DebugSearchOutputCSV)) {
		formatIdx = 1
		output.Placeholder = defaultOutputPath(cfg, types.DebugSearchOutputCSV)
	}

	return searchModel{
		cfg:           cfg,
		dis:           dis,
		termInput:     term,
		outputInput:   output,
		formats:       formats,
		formatIndex:   formatIdx,
		focused:       focusSearchField,
		progressBar:   pb,
		progressWidth: 50,
		statusMsg:     "Ready for batch debug search",
		authenticated: false,
		defaultDir:    defaultDir,
	}
}

func (s searchModel) WithAuthState(ok bool) searchModel {
	s.authenticated = ok
	if ok && strings.Contains(strings.ToLower(s.errMsg), "auth") {
		s.errMsg = ""
		s.statusMsg = "Ready for batch debug search"
	}
	return s
}

func (s searchModel) Init() (searchModel, tea.Cmd) {
	s.focused = focusSearchField
	var cmds []tea.Cmd
	if cmd := s.termInput.Focus(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	s.outputInput.Blur()
	if len(cmds) == 0 {
		return s, nil
	}
	return s, tea.Batch(cmds...)
}

func (s searchModel) SetWidth(width int) searchModel {
	if width < 20 {
		width = 20
	}
	s.progressWidth = width - 6
	if s.progressWidth < 10 {
		s.progressWidth = 10
	}
	s.progressBar.Width = s.progressWidth
	return s
}

func (s searchModel) Update(msg tea.Msg, focused bool) (searchModel, tea.Cmd, bool) {
	var cmds []tea.Cmd
	handled := false

	// Keep inputs responsive for cursor blink.
	var cmd tea.Cmd
	s.termInput, cmd = s.termInput.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	s.outputInput, cmd = s.outputInput.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !focused && msg.String() != "ctrl+r" {
			break
		}
		if s.running {
			switch msg.String() {
			case "tab", "shift+tab", "ctrl+n", "ctrl+p", "ctrl+r":
				handled = true
			}
			break
		}
		switch msg.String() {
		case "tab":
			handled = true
			s.focused = (s.focused + 1) % focusCount
			if cmd := s.applyFocus(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "shift+tab":
			handled = true
			s.focused = (s.focused - 1 + focusCount) % focusCount
			if cmd := s.applyFocus(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "ctrl+n":
			handled = true
			s.focused = (s.focused + 1) % focusCount
			if cmd := s.applyFocus(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "ctrl+p":
			handled = true
			s.focused = (s.focused - 1 + focusCount) % focusCount
			if cmd := s.applyFocus(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		case "left", "right", "space":
			if s.focused == focusFormatField {
				handled = true
				s.toggleFormat()
			}
		case "enter":
			if s.focused == focusRunButton {
				handled = true
				var c tea.Cmd
				s, c = s.startSearch()
				if c != nil {
					cmds = append(cmds, c)
				}
			}
		case "ctrl+r":
			handled = true
			var c tea.Cmd
			s, c = s.startSearch()
			if c != nil {
				cmds = append(cmds, c)
			}
		}
	case searchProgressMsg:
		s.lastProgress = msg.Status
		percent := msg.Status.PercentComplete / 100.0
		if percent < 0 {
			percent = 0
		}
		if percent > 1 {
			percent = 1
		}
		progressCmd := s.progressBar.SetPercent(percent)
		cmds = append(cmds, progressCmd)
		s.statusMsg = fmt.Sprintf("Running %s — %d/%d queries (%.1f%%)", msg.Status.RunID, msg.Status.CompletedQueries, msg.Status.TotalQueries, msg.Status.PercentComplete)
		if s.progressCh != nil {
			cmds = append(cmds, waitProgressCmd(s.progressCh))
		}
	case searchProgressStreamClosedMsg:
		s.progressCh = nil
	case searchEventMsg:
		evt := msg.Event
		s.lastEvent = &evt
		s.statusMsg = fmt.Sprintf("Table %s — last event %s (%d rows)", evt.TableName, evt.EventType, evt.RowCount)
		if s.eventCh != nil {
			cmds = append(cmds, waitEventCmd(s.eventCh))
		}
	case searchEventStreamClosedMsg:
		s.eventCh = nil
	case searchFinishedMsg:
		s.running = false
		s.statusMsg = fmt.Sprintf("Search complete → %s", msg.OutputPath)
		s.lastOutputPath = msg.OutputPath
		s.progressCh = nil
		s.eventCh = nil
		cmds = append(cmds, newLogCmd("Search complete: %s", msg.OutputPath))
	case tea.WindowSizeMsg:
		s = s.SetWidth(msg.Width)
	}

	if len(cmds) == 0 {
		return s, nil, handled
	}
	return s, tea.Batch(cmds...), handled
}

func (s searchModel) View() string {
	sections := []string{
		sectionTitleStyle.Render("Batch Debug Search"),
		fieldLabelStyle.Render("Search Term") + "\n" + s.termInput.View(),
		s.renderFormatRow(),
		fieldLabelStyle.Render("Output File") + "\n" + s.outputInput.View(),
		s.renderRunButton(),
		s.renderStatus(),
	}

	if s.running || s.lastProgress.TotalQueries > 0 {
		sections = append(sections, s.progressBar.View())
	}

	if s.lastEvent != nil {
		sections = append(sections, s.renderLastEvent())
	}

	sections = append(sections, helpStyle.Render("Tab/Shift+Tab navigate • Space toggles format • Enter/Ctrl+R runs search • c opens DIS Config"))

	return strings.Join(sections, "\n\n")
}

func (s searchModel) renderFormatRow() string {
	current := s.formats[s.formatIndex]
	label := fmt.Sprintf("Format: %s", formatName(current))
	if s.focused == focusFormatField {
		return selectedFormatStyle.Render(label)
	}
	return fieldLabelStyle.Render(label)
}

func (s searchModel) renderRunButton() string {
	style := buttonStyle.Copy()
	label := "▶ Run Debug Search"
	if s.running {
		style = style.Copy().Foreground(lipgloss.Color("#666"))
		label = "Running…"
	} else if s.focused == focusRunButton {
		style = buttonFocusedStyle
	}
	return style.Render(label)
}

func (s searchModel) renderStatus() string {
	if s.errMsg != "" {
		return errorStyle.Render(s.errMsg)
	}
	if s.statusMsg == "" {
		return noticeStyle.Render("Idle")
	}
	return noticeStyle.Render(s.statusMsg)
}

func (s searchModel) renderLastEvent() string {
	evt := s.lastEvent
	if evt == nil {
		return ""
	}
	summary := fmt.Sprintf("Last Table: %s • Event: %s • Rows: %d • Columns: %d", evt.TableName, evt.EventType, evt.RowCount, evt.ColumnCount)
	line := summary
	if evt.SampleRow != nil && len(evt.SampleRow) > 0 {
		pairs := 0
		parts := make([]string, 0, len(evt.SampleRow))
		for k, v := range evt.SampleRow {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			pairs++
			if pairs >= 4 {
				break
			}
		}
		line = line + "\n" + sampleStyle.Render(strings.Join(parts, ", "))
	}
	return infoStyle.Render(line)
}

func (s searchModel) StatusLine() string {
	switch {
	case s.running:
		return fmt.Sprintf("Debug Search: running (%s)", s.statusMsg)
	case s.lastOutputPath != "":
		return fmt.Sprintf("Debug Search: last output %s", filepath.Base(s.lastOutputPath))
	default:
		return "Debug Search: idle"
	}
}

func (s searchModel) toggleFormat() {
	s.formatIndex = (s.formatIndex + 1) % len(s.formats)
	if s.formats[s.formatIndex] == types.DebugSearchOutputCSV {
		s.outputInput.Placeholder = defaultOutputPath(s.cfg, types.DebugSearchOutputCSV)
	} else {
		s.outputInput.Placeholder = defaultOutputPath(s.cfg, types.DebugSearchOutputSQLite)
	}
}

func (s searchModel) applyFocus() tea.Cmd {
	var cmds []tea.Cmd
	if s.focused == focusSearchField {
		if cmd := s.termInput.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		s.termInput.Blur()
	}
	if s.focused == focusOutputField {
		if cmd := s.outputInput.Focus(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else {
		s.outputInput.Blur()
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (s searchModel) startSearch() (searchModel, tea.Cmd) {
	if s.running {
		return s, nil
	}
	if !s.authenticated {
		s.errMsg = "DIS authentication required"
		s.statusMsg = "Press c to configure DIS connection"
		cmds := []tea.Cmd{
			newLogCmd("Debug search blocked: authenticate to DIS first"),
			requireAuthCmd(),
		}
		return s, tea.Batch(cmds...)
	}
	term := strings.TrimSpace(s.termInput.Value())
	if term == "" {
		s.errMsg = "Search term is required"
		s.statusMsg = "Enter a search term"
		return s, nil
	}

	mode := s.formats[s.formatIndex]
	path := strings.TrimSpace(s.outputInput.Value())
	if path == "" {
		path = defaultOutputPath(s.cfg, mode)
		s.outputInput.SetValue(path)
	}
	path = ensureExtension(path, mode)

	progressCh := make(chan types.ProgressStatus)
	eventCh := make(chan types.TableEvent)

	s.progressCh = progressCh
	s.eventCh = eventCh
	s.running = true
	s.errMsg = ""
	s.startedAt = time.Now()
	s.lastProgress = types.ProgressStatus{}
	s.lastEvent = nil
	s.statusMsg = "Starting debug search"
	s.lastOutputPath = path

	cmds := []tea.Cmd{
		runDebugSearchCmd(s.dis, term, types.DebugSearchOptions{OutputMode: mode, OutputPath: path}, progressCh, eventCh),
		waitProgressCmd(progressCh),
		waitEventCmd(eventCh),
		newLogCmd("Run debug search: term=%s, format=%s, output=%s", term, mode, path),
	}
	cmds = append(cmds, s.progressBar.SetPercent(0))

	return s, tea.Batch(cmds...)
}

func defaultOutputPath(cfg *types.DISUIConfig, mode types.DebugSearchOutputMode) string {
	if cfg == nil {
		base := "debug-output/debug-search"
		if mode == types.DebugSearchOutputCSV {
			return base + ".csv"
		}
		return base + ".db"
	}
	base := strings.TrimSpace(cfg.DebugSearch.DefaultOutputPath)
	if base == "" {
		base = filepath.Join("debug-output", "debug-search")
		if mode == types.DebugSearchOutputCSV {
			return base + ".csv"
		}
		return base + ".db"
	}
	return ensureExtension(base, mode)
}

func ensureExtension(path string, mode types.DebugSearchOutputMode) string {
	targetExt := extensionForMode(mode)
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, targetExt) {
		return path
	}
	// strip other debug extensions
	for _, m := range []types.DebugSearchOutputMode{types.DebugSearchOutputSQLite, types.DebugSearchOutputCSV} {
		ext := extensionForMode(m)
		if strings.HasSuffix(lower, ext) {
			path = path[:len(path)-len(ext)]
			break
		}
	}
	return path + targetExt
}

func extensionForMode(mode types.DebugSearchOutputMode) string {
	if mode == types.DebugSearchOutputCSV {
		return ".csv"
	}
	return ".db"
}

func formatName(mode types.DebugSearchOutputMode) string {
	switch mode {
	case types.DebugSearchOutputCSV:
		return "CSV (.csv)"
	default:
		return "SQLite (.db)"
	}
}

func runDebugSearchCmd(dis types.DISReaderService, term string, opts types.DebugSearchOptions, progressCh chan types.ProgressStatus, eventCh chan types.TableEvent) tea.Cmd {
	return func() tea.Msg {
		dis.RunDebugSearch(term, opts, progressCh, eventCh)
		close(progressCh)
		close(eventCh)
		return searchFinishedMsg{OutputPath: opts.OutputPath}
	}
}

func requireAuthCmd() tea.Cmd {
	return func() tea.Msg {
		return requireAuthMsg{}
	}
}

func waitProgressCmd(ch <-chan types.ProgressStatus) tea.Cmd {
	return func() tea.Msg {
		status, ok := <-ch
		if !ok {
			return searchProgressStreamClosedMsg{}
		}
		return searchProgressMsg{Status: status}
	}
}

func waitEventCmd(ch <-chan types.TableEvent) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return searchEventStreamClosedMsg{}
		}
		return searchEventMsg{Event: evt}
	}
}

type searchProgressMsg struct {
	Status types.ProgressStatus
}

type searchProgressStreamClosedMsg struct{}

type searchEventMsg struct {
	Event types.TableEvent
}

type searchEventStreamClosedMsg struct{}

type searchFinishedMsg struct {
	OutputPath string
}

type requireAuthMsg struct{}
