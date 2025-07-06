package main

import (
	debugUI "github.com/BeardedWonderDev/DIS-Reader/cmd/debug"
	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/BeardedWonderDev/DIS-Reader/ui"
)

func main() {
	cfg := NewConfig()

	disReader, err := disreader.NewDISReaderService(cfg.DIS, nil)
	if err != nil {
		panic(err)
	}

	modules := []types.ViewModule{
		debugUI.NewDebugService(),
	}

	u := ui.NewUI(cfg, disReader, nil, modules)

	u.AttachToDISLogger()

	u.Run()

	disReader.Shutdown()
}
