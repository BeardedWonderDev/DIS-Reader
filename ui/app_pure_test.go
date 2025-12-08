package ui

import (
	"testing"

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
