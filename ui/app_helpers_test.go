package ui

import (
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/rivo/tview"
)

func newTestViewer() *viewerApp {
	theme := types.ResolveTheme(types.DefaultThemeName)
	return &viewerApp{
		theme:      theme,
		tableView:  tview.NewTable(),
		tableTable: tview.NewTable(),
		dbTable:    tview.NewTable(),
	}
}

func TestRenderTableMessageAndTitle(t *testing.T) {
	v := newTestViewer()
	v.renderTableMessage("hello")
	cell := v.tableView.GetCell(0, 0)
	if cell.Text != "hello" {
		t.Fatalf("expected message cell text 'hello', got %q", cell.Text)
	}

	v.updateTableTitle("parts")
	if title := v.tableView.GetTitle(); title != " parts " {
		t.Fatalf("expected title ' parts ', got %q", title)
	}
	v.updateTableTitle("   ")
	if title := v.tableView.GetTitle(); title != " Rows " {
		t.Fatalf("expected default title when blank table name, got %q", title)
	}
}

func TestPickerMessageCellIsNotSelectable(t *testing.T) {
	v := newTestViewer()
	cell := v.pickerMessageCell("none")
	if cell.Text != "none" {
		t.Fatalf("expected text to match")
	}
}

func TestDetermineColumnGroupSize(t *testing.T) {
	v := newTestViewer()
	// Set a known inner width
	v.tableView.SetRect(0, 0, 90, 10)
	group := v.determineColumnGroupSize(5)
	if group < 1 || group > 5 {
		t.Fatalf("expected group between 1 and 5, got %d", group)
	}

	// Width zero should fall back to approxColumnWidth logic and clamp to <= total columns
	v.tableView.SetRect(0, 0, 0, 0)
	if g := v.determineColumnGroupSize(3); g != 2 {
		t.Fatalf("expected derived group 2 when width zero, got %d", g)
	}
}
