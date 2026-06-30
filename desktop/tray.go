package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TrayManager struct {
	app *App
}

func NewTrayManager(app *App) *TrayManager {
	return &TrayManager{app: app}
}

func (t *TrayManager) GetMacOptions() *mac.Options {
	return &mac.Options{
		About: &mac.AboutInfo{
			Title:   "Airvoice",
			Message: "Speak on your phone, type on your desktop",
		},
	}
}

func (t *TrayManager) GetTrayMenu() *menu.Menu {
	m := menu.NewMenu()

	m.AddText("Show Window", keys.CmdOrCtrl("0"), func(_ *menu.CallbackData) {
		runtime.WindowShow(t.app.ctx)
	})
	m.AddText("Hide Window", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		runtime.WindowHide(t.app.ctx)
	})
	m.AddSeparator()
	m.AddText(t.statusLabel(), nil, nil)
	m.AddSeparator()
	m.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(t.app.ctx)
	})

	return m
}

func (t *TrayManager) statusLabel() string {
	status := t.app.GetConnectionStatus()
	switch status.State {
	case "connected":
		return fmt.Sprintf("Connected: %s", status.DeviceName)
	case "waiting":
		return "Waiting for device"
	default:
		return "Disconnected"
	}
}
