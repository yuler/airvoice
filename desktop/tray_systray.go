//go:build linux || windows || darwin

package main

import (
	_ "embed"
	"runtime"
	"sync"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/menu"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray_icon.png
var trayIcon []byte

type platformTray struct {
	once       sync.Once
	statusItem *systray.MenuItem
	showItem   *systray.MenuItem
	hideItem   *systray.MenuItem
	quitItem   *systray.MenuItem
}

// Start registers the system tray after Wails has initialized the platform UI loop.
func (t *TrayManager) Start() {
	t.once.Do(func() {
		systray.Register(t.onReady, t.onExit)
	})
}

func (t *TrayManager) onReady() {
	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(trayIcon, trayIcon)
	} else {
		systray.SetIcon(trayIcon)
	}
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
			case _, ok := <-t.showItem.ClickedCh:
				if !ok {
					return
				}
				if t.app.ctx != nil {
					wailsruntime.WindowShow(t.app.ctx)
				}
			case _, ok := <-t.hideItem.ClickedCh:
				if !ok {
					return
				}
				if t.app.ctx != nil {
					wailsruntime.WindowHide(t.app.ctx)
				}
			case _, ok := <-t.quitItem.ClickedCh:
				if !ok {
					return
				}
				if t.app.ctx != nil {
					wailsruntime.Quit(t.app.ctx)
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
	// Returning nil disables Wails application menu; tray menu comes from systray.
	return nil
}
