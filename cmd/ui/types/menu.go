package typesUI

import "github.com/rivo/tview"

type MainMenuDefinition struct {
	MenuList *tview.List

	RootMenu  MenuDefinition
	DebugMenu MenuDefinition
}

type MenuDefinition struct {
	Title string
	Items []MenuItem
}

type MenuItem struct {
	MainText      string // The main text of the list item.
	SecondaryText string // A secondary text to be shown underneath the main text.
	Shortcut      rune   // The key to select the list item directly, 0 if there is no shortcut.
	Selected      func() // The optional function which is called when the item is selected.
}
