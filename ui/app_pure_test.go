package ui

import (
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestColorToHex(t *testing.T) {
	hex := colorToHex(tcell.ColorBlue)
	if hex != "#0000ff" {
		t.Fatalf("expected #0000ff, got %s", hex)
	}
}

func TestDetermineColumnGroupSizeUsesTableWidth(t *testing.T) {
	v := &viewerApp{tableView: tview.NewTable()}
	v.tableView.SetRect(0, 0, 100, 10)
	size := v.determineColumnGroupSize(10)
	if size != 5 { // 100 / (approxColumnWidth+2=20)
		t.Fatalf("expected group size 5, got %d", size)
	}
}

func TestDetermineColumnGroupSizeHandlesZeroWidth(t *testing.T) {
	v := &viewerApp{tableView: tview.NewTable()}
	size := v.determineColumnGroupSize(3)
	if size == 0 || size > 3 {
		t.Fatalf("expected group size between 1 and total columns, got %d", size)
	}
}

func TestBuildColumnRanges(t *testing.T) {
	ranges := buildColumnRanges(5, 2)
	expected := []columnRange{{0, 2}, {2, 4}, {4, 5}}
	if len(ranges) != len(expected) {
		t.Fatalf("unexpected length: %v", ranges)
	}
	for i := range expected {
		if ranges[i] != expected[i] {
			t.Fatalf("range %d mismatch: got %+v want %+v", i, ranges[i], expected[i])
		}
	}
}

func TestFormatIntHelpers(t *testing.T) {
	if s := formatRecentInts(nil, 3); s != "n/a" {
		t.Fatalf("expected n/a for empty slice, got %s", s)
	}
	if s := formatRecentInts([]int32{1, 2, 3, 4}, 2); s != "3, 4" {
		t.Fatalf("unexpected recent ints: %s", s)
	}
	if s := formatLeadingInts([]int32{5, 6, 7}, 2); s != "5, 6" {
		t.Fatalf("unexpected leading ints: %s", s)
	}
	if s := formatLeadingInt64([]int64{9, 8, 7}, 4); s != "9, 8, 7" {
		t.Fatalf("unexpected leading int64: %s", s)
	}
}

func TestIsLogAtBottom(t *testing.T) {
	v := &viewerApp{logView: tview.NewTextView(), logLines: 2}
	v.logView.SetRect(0, 0, 10, 5)
	if !v.isLogAtBottom() {
		t.Fatalf("expected at bottom when height >= lines")
	}
	v.logLines = 10
	v.logView.ScrollTo(3, 0)
	if v.isLogAtBottom() {
		t.Fatalf("expected not at bottom when scrolled up")
	}
}

func TestUpdateLogTitleUnreadFlag(t *testing.T) {
	v := &viewerApp{logView: tview.NewTextView()}
	v.updateLogTitle()
	if title := v.logView.GetTitle(); title != " Activity " {
		t.Fatalf("expected default title, got %s", title)
	}
	v.logUnread = true
	v.updateLogTitle()
	if title := v.logView.GetTitle(); title != " Activity (new) " {
		t.Fatalf("expected unread title, got %s", title)
	}
}

func TestFocusInTextEntry(t *testing.T) {
	app := tview.NewApplication()
	v := &viewerApp{app: app}
	input := tview.NewInputField()
	app.SetFocus(input)
	if !v.focusInTextEntry() {
		t.Fatalf("expected true when focus is input field")
	}
	app.SetFocus(tview.NewTextView())
	if v.focusInTextEntry() {
		t.Fatalf("expected false when focus is not text entry")
	}
}

func TestApplyPaneFocusColors(t *testing.T) {
	v := &viewerApp{theme: types.ResolveTheme("")}
	table := tview.NewTable()
	v.applyPaneFocus(table, true)
	if table.GetBorderColor() != v.theme.Colors.AccentColor {
		t.Fatalf("expected accent border color when focused")
	}
	v.applyPaneFocus(table, false)
	if table.GetBorderColor() != v.theme.Colors.BorderColor {
		t.Fatalf("expected border color when unfocused")
	}
}
