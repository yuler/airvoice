package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal("Failed to initialize app: ", err)
	}
	defer app.history.Close()

	trayManager := NewTrayManager(app)

	err = wails.Run(&options.App{
		Title:  "Airvoice",
		Width:  320,
		Height: 480,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Menu:      trayManager.GetTrayMenu(),
		Bind: []interface{}{
			app,
		},
		Mac: trayManager.GetMacOptions(),
	})

	if err != nil {
		log.Fatal("Error: ", err)
	}
}
