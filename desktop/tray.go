package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/yuler/airvoice/cli/config"
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
		if status.Mode == config.ModeRelay {
			if status.DeviceName != "" {
				return fmt.Sprintf("Connected (Relay · E2E): %s", status.DeviceName)
			}
			return "Connected (Relay · E2E)"
		}
		if status.DeviceName != "" {
			return fmt.Sprintf("Connected (LAN): %s", status.DeviceName)
		}
		return "Connected (LAN)"
	case "waiting":
		return "Waiting for phone"
	case "connecting":
		if status.Mode == config.ModeRelay {
			return "Connecting to relay..."
		}
		return "Connecting..."
	case "error":
		return "Relay error"
	default:
		if status.Mode == config.ModeRelay {
			return "Relay disconnected (reconnecting...)"
		}
		return "Disconnected"
	}
}
