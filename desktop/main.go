package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	trayManager := NewTrayManager(app)

	err := wails.Run(&options.App{
		Title:  "Airvoice",
		Width:  400,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Menu:      trayManager.GetApplicationMenu(),
		Bind: []interface{}{
			app,
		},
		Mac: trayManager.GetMacOptions(),
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
