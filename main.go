package main

import (
	"os"
	"strings"

	debugUI "github.com/BeardedWonderDev/DIS-Reader/cmd/debug"
	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/BeardedWonderDev/DIS-Reader/ui"
	bubbleui "github.com/BeardedWonderDev/DIS-Reader/ui/bubbletea"
)

func main() {
	cfg := NewConfig()

	disReader, err := disreader.NewDISReaderService(cfg.DIS, nil)
	if err != nil {
		panic(err)
	}
	defer disReader.Shutdown()

	if shouldUseBubbleUI() {
		if err := bubbleui.Run(cfg, disReader); err != nil {
			panic(err)
		}
		return
	}

	modules := []types.ViewModule{
		debugUI.NewDebugService(),
	}

	u := ui.NewUI(cfg, disReader, nil, modules)

	u.AttachToDISLogger()

	if err := u.Run(); err != nil {
		panic(err)
	}
}

func shouldUseBubbleUI() bool {
	mode := strings.TrimSpace(os.Getenv("DISREADER_UI"))
	switch strings.ToLower(mode) {
	case "bubble", "bubbletea", "charm", "bubble-ui":
		return true
	default:
		return false
	}
}
