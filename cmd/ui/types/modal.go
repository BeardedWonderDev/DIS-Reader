package typesUI

import "github.com/rivo/tview"

type WinSize struct {
	X      int
	Y      int
	Width  int
	Height int
}

type CreateModalDialogParam struct {
	Title         string
	RootView      tview.Primitive
	Draggable     bool
	Resizeable    bool
	Size          WinSize
	FallbackFocus tview.Primitive
}
