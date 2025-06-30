package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/gdamore/tcell/v2"
)

func (u *UI) initShortcuts(modules []types.ViewModule) {
	// Build global shortcut map
	u.shortcuts = make(map[rune]func())
	u.shortcuts['c'] = u.Auth.ShowAuthModal
	u.shortcuts['q'] = u.QuitApplication
	// Module main-menu and sub-menu shortcuts
	for _, m := range modules {
		mod := m
		// Register main-menu shortcut, ensuring uniqueness
		r := mod.Shortcut()
		if _, exists := u.shortcuts[r]; exists {
			// find first available a-z, then 0-9
			for c := 'a'; c <= 'z'; c++ {
				if _, used := u.shortcuts[c]; !used {
					r = c
					break
				}
			}
			if _, exists2 := u.shortcuts[r]; exists2 {
				for c := '0'; c <= '9'; c++ {
					if _, used := u.shortcuts[c]; !used {
						r = c
						break
					}
				}
			}
		}
		u.shortcuts[r] = func() { mod.Activate() }
		// register submenu items
		for _, item := range mod.SubMenu().Items {
			r2 := item.Shortcut
			if _, exists := u.shortcuts[r2]; exists {
				// find fallback
				for c := 'a'; c <= 'z'; c++ {
					if _, used := u.shortcuts[c]; !used {
						r2 = c
						break
					}
				}
				if _, used2 := u.shortcuts[r2]; used2 {
					for c := '0'; c <= '9'; c++ {
						if _, used := u.shortcuts[c]; !used {
							r2 = c
							break
						}
					}
				}
			}
			if item.Selected != nil {
				u.shortcuts[r2] = item.Selected
			}
		}
	}
}

func (u *UI) handleShortcutEvent(event *tcell.EventKey) *tcell.EventKey {
	if fn, ok := u.shortcuts[event.Rune()]; ok {
		fn()
		return nil
	}

	switch event.Key() {
	case tcell.KeyTab:
		// keep existing behavior?
		u.SetFocus(u.Layout.OutputPanel)
		return nil
	case tcell.KeyPgUp, tcell.KeyUp:
		row, _ := u.Layout.LogList.GetScrollOffset()
		u.Layout.LogList.ScrollTo(row-1, 0)

		return nil
	case tcell.KeyPgDn, tcell.KeyDown:
		row, _ := u.Layout.LogList.GetScrollOffset()
		u.Layout.LogList.ScrollTo(row+1, 0)

		return nil
	}

	return event
}
