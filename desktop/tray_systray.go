//go:build linux || windows || darwin

package main

import (
	_ "embed"
	"runtime"
	"sync"

	"fyne.io/systray"
	"github.com/wailsapp/wails/v2/pkg/menu"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray_icon.png
var trayIcon []byte

type platformTray struct {
	once       sync.Once
	mu         sync.Mutex
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

	showItem := systray.AddMenuItem("Show Window", "Show Airvoice Window")
	hideItem := systray.AddMenuItem("Hide Window", "Hide Airvoice Window")
	systray.AddSeparator()
	statusItem := systray.AddMenuItem(t.statusLabel(), "")
	statusItem.Disable()
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Quit Airvoice")

	t.mu.Lock()
	t.showItem = showItem
	t.hideItem = hideItem
	t.statusItem = statusItem
	t.quitItem = quitItem
	t.mu.Unlock()

	// Handle clicks in a background loop
	go func() {
		for {
			select {
			case _, ok := <-showItem.ClickedCh:
				if !ok {
					return
				}
				if ctx := t.app.getContext(); ctx != nil {
					wailsruntime.WindowShow(ctx)
				}
			case _, ok := <-hideItem.ClickedCh:
				if !ok {
					return
				}
				if ctx := t.app.getContext(); ctx != nil {
					wailsruntime.WindowHide(ctx)
				}
			case _, ok := <-quitItem.ClickedCh:
				if !ok {
					return
				}
				if ctx := t.app.getContext(); ctx != nil {
					wailsruntime.Quit(ctx)
				}
				systray.Quit()
			}
		}
	}()
}

func (t *TrayManager) onExit() {
}

func (t *TrayManager) UpdateStatus() {
	t.mu.Lock()
	item := t.statusItem
	t.mu.Unlock()
	if item != nil {
		item.SetTitle(t.statusLabel())
	}
}

func (t *TrayManager) GetTrayMenu() *menu.Menu {
	// Returning nil disables Wails application menu; tray menu comes from systray.
	return nil
}
