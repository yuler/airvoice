//go:build darwin

package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type platformTray struct {
	menu *menu.Menu
}

func (t *TrayManager) initPlatform() {
	// No background initialization needed for macOS Menu Bar
}

func (t *TrayManager) UpdateStatus() {
	// No-op for macOS native menu bar status
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

	t.menu = m
	return m
}
