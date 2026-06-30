package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

type TrayManager struct {
	app *App
	platformTray
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
