//go:build gui

package main

import (
	"context"
	"embed"
	"log"

	"github.com/BeardedWonderDev/DIS-Reader/internal/installer"
	"github.com/BeardedWonderDev/DIS-Reader/internal/servicectl"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/*
var assets embed.FS

type App struct {
	ctrl servicectl.Controller
}

func NewApp() *App {
	return &App{ctrl: servicectl.New()}
}

func (a *App) InstallService(ctx context.Context) (string, error) {
	// assumes binary already present
	if err := a.ctrl.Install(ctx); err != nil {
		return "", err
	}
	return "install ok", nil
}

func (a *App) InstallBinary(ctx context.Context) (string, error) {
	path, err := installer.InstallLatest(ctx, "")
	if err != nil {
		return "", err
	}
	return "binary installed at " + path, nil
}

func (a *App) InstallBinaryAndService(ctx context.Context) (string, error) {
	path, err := installer.InstallLatest(ctx, "")
	if err != nil {
		return "", err
	}
	if err := a.ctrl.Install(ctx); err != nil {
		return "", err
	}
	return "binary installed at " + path + "; service registered", nil
}

func (a *App) StartService(ctx context.Context) (string, error) {
	if err := a.ctrl.Start(ctx); err != nil {
		return "", err
	}
	return "start ok", nil
}

func (a *App) StopService(ctx context.Context) (string, error) {
	if err := a.ctrl.Stop(ctx); err != nil {
		return "", err
	}
	return "stop ok", nil
}

func (a *App) RestartService(ctx context.Context) (string, error) {
	if err := a.ctrl.Restart(ctx); err != nil {
		return "", err
	}
	return "restart ok", nil
}

func (a *App) StatusService(ctx context.Context) (string, error) {
	out, err := a.ctrl.Status(ctx)
	return out, err
}

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:         "DIS Agent Control",
		Width:         960,
		Height:        720,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 12, G: 18, B: 32, A: 255},
		Bind:             []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
