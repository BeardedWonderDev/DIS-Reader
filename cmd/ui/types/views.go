package typesUI

import "github.com/rivo/tview"

type ComponentLayout struct {
	MainMenu    MainMenuDefinition
	SubMenuList *tview.List
	LogList     *tview.TextView
	OutputPanel InitOutputPanelComponents
}
