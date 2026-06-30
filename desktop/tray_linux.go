//go:build linux

package main

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
)

type platformTray struct{}

func (t *TrayManager) initPlatform() {
}

func (t *TrayManager) UpdateStatus() {
}

func (t *TrayManager) GetTrayMenu() *menu.Menu {
	return nil
}
