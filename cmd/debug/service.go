package debugUI

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	utilsUI "github.com/BeardedWonderDev/DIS-Reader/ui/utils"
)

type Debug struct {
	UI           types.UI
	ProgressChan chan types.ProgressStatus
	EventChan    chan types.TableEvent
}

func NewDebugService() *Debug {
	return &Debug{
		ProgressChan: make(chan types.ProgressStatus),
		EventChan:    make(chan types.TableEvent),
	}
}

// Name returns the label for this module in the main menu.
func (d *Debug) Name() string {
	return "Debug Search"
}

// Shortcut returns a rune for quick activation from the main menu.
func (d *Debug) Shortcut() rune {
	return 'd'
}

// Init is called once after UI and Logger are ready. It stores the UI reference
// and initializes the submenu for this module.
func (d *Debug) Init(ui types.UI) {
	d.UI = ui
}

// Activate is called when the user selects this module from the main menu.
// It makes the submenu visible and sets focus to it.
func (d *Debug) Activate() {
	// Populate and show the submenu
	utilsUI.BuildMenu(d.SubMenu(), d.UI.GetLayout().SubMenuList)
	d.UI.SetOutputPanel(d.OutputPanel())
	d.UI.SetFocus(d.UI.GetLayout().SplitSidebar)
}

// SubMenu returns this module's submenu definition.
func (d *Debug) SubMenu() types.MenuDefinition {
	return types.MenuDefinition{
		Title: " Debug Menu ",
		Items: []types.MenuItem{
			{
				MainText:      "Start Batch Search",
				SecondaryText: "",
				Shortcut:      's',
				Selected: func() {
					d.showDebugModal()
				},
			},
		},
	}
}

// OutputPanel returns the panel definition to display in the main output area.
func (d *Debug) OutputPanel() types.PanelDefinition {
	// For now, returning an empty panel or you can implement a real view.
	return types.PanelDefinition{
		Name: "Debug Output",
		View: nil, // Replace with actual tview.Primitive if available
	}
}
