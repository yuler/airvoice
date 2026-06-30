//go:build linux || windows

package main

import (
	_ "embed"
	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray_icon.png
var trayIcon []byte

type platformTray struct {
	statusItem *systray.MenuItem
	showItem   *systray.MenuItem
	hideItem   *systray.MenuItem
	quitItem   *systray.MenuItem
}

func (t *TrayManager) initPlatform() {
	go func() {
		systray.Run(t.onReady, t.onExit)
	}()
}

func (t *TrayManager) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTooltip("Airvoice")

	t.showItem = systray.AddMenuItem("Show Window", "Show Airvoice Window")
	t.hideItem = systray.AddMenuItem("Hide Window", "Hide Airvoice Window")
	systray.AddSeparator()
	t.statusItem = systray.AddMenuItem(t.statusLabel(), "")
	t.statusItem.Disable()
	systray.AddSeparator()
	t.quitItem = systray.AddMenuItem("Quit", "Quit Airvoice")

	// Handle clicks in a background loop
	go func() {
		for {
			select {
			case <-t.showItem.ClickedCh:
				if t.app.ctx != nil {
					runtime.WindowShow(t.app.ctx)
				}
			case <-t.hideItem.ClickedCh:
				if t.app.ctx != nil {
					runtime.WindowHide(t.app.ctx)
				}
			case <-t.quitItem.ClickedCh:
				if t.app.ctx != nil {
					runtime.Quit(t.app.ctx)
				}
				systray.Quit()
			}
		}
	}()
}

func (t *TrayManager) onExit() {
}

func (t *TrayManager) UpdateStatus() {
	if t.statusItem != nil {
		t.statusItem.SetTitle(t.statusLabel())
	}
}

func (t *TrayManager) GetTrayMenu() *menu.Menu {
	// Returning nil disables application menu on Linux/Windows
	return nil
}
