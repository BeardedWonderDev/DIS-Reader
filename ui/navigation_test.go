package ui

import (
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/rivo/tview"
)

func TestSwitchToPageUpdatesFocusables(t *testing.T) {
	app := tview.NewApplication()
	v := &viewerApp{
		app:        app,
		pages:      tview.NewPages(),
		navBar:     tview.NewTextView(),
		theme:      types.ResolveTheme(""),
		dbTable:    tview.NewTable(),
		tableTable: tview.NewTable(),
		tableView:  tview.NewTable(),
		logView:    tview.NewTextView(),
	}
	v.parts = newPartsPanel(v)

	// default path
	v.switchToPage(pageDebug)
	if len(v.focusables) == 0 {
		t.Fatalf("expected default focusables set")
	}

	// parts path
	v.parts = newPartsPanel(v)
	v.switchToPage(pageParts)
	if len(v.focusables) == 0 {
		t.Fatalf("expected parts focusables set")
	}
	if v.activePage != pageParts {
		t.Fatalf("expected activePage to be parts")
	}
}
