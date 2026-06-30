# Desktop Theme and System Tray Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the theme toggle by using CSS variables in Tailwind and implement a native system tray for Linux/Windows using `getlantern/systray` while preserving macOS native menus.

**Architecture:** Use platform-specific Go files with build tags (`//go:build linux || windows` and `//go:build darwin`) to isolate the `systray` dependency from macOS. Update the Tailwind configuration to map theme colors to the CSS variables in `styles.css`.

**Tech Stack:** Go (Wails v2, `github.com/getlantern/systray`), Vue 3, Tailwind CSS v3

---

### Task 1: Update Tailwind Configuration to Use CSS Variables

**Files:**
- Modify: [tailwind.config.js](file:///home/yule/Sides/airvoice/desktop/frontend/tailwind.config.js)

- [ ] **Step 1: Replace hardcoded colors with CSS variables**
  Update the `colors` block to map variables.
  
  Code to write:
  ```javascript
  colors: {
    'bg-primary': 'var(--color-bg-primary)',
    'bg-secondary': 'var(--color-bg-secondary)',
    'border-default': 'var(--color-border-default)',
    'accent-blue': 'var(--color-accent-blue)',
    
    // Map both standard text keys used in components
    'text-primary': 'var(--color-primary-text)',
    'text-secondary': 'var(--color-secondary-text)',
    'text-muted': 'var(--color-muted-text)',
    
    'primary-text': 'var(--color-primary-text)',
    'secondary-text': 'var(--color-secondary-text)',
    'muted-text': 'var(--color-muted-text)',
    
    'status-success': 'var(--color-status-success)',
    'status-warning': 'var(--color-status-warning)',
    'status-error': 'var(--color-status-error)',
    'status-neutral': 'var(--color-status-neutral)',
  },
  ```

- [ ] **Step 2: Commit changes**
  ```bash
  git add desktop/frontend/tailwind.config.js
  git commit -m "feat: map tailwind config colors to CSS variables for theme toggle"
  ```

---

### Task 2: Create Linux/Windows System Tray Implementation

**Files:**
- Create: [desktop/tray_systray.go](file:///home/yule/Sides/airvoice/desktop/tray_systray.go)

- [ ] **Step 1: Create `desktop/tray_systray.go`**
  Implement the system tray lifecycle using `getlantern/systray`.

  Code to write:
  ```go
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
  ```

- [ ] **Step 2: Commit the new file**
  ```bash
  git add desktop/tray_systray.go
  git commit -m "feat: implement linux/windows system tray using getlantern/systray"
  ```

---

### Task 3: Create macOS Native Menu Fallback

**Files:**
- Create: [desktop/tray_darwin.go](file:///home/yule/Sides/airvoice/desktop/tray_darwin.go)

- [ ] **Step 1: Create `desktop/tray_darwin.go`**
  Implement Wails menu fallback for macOS (which does not use `systray` to prevent thread/linker conflicts).

  Code to write:
  ```go
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
  ```

- [ ] **Step 2: Commit the new file**
  ```bash
  git add desktop/tray_darwin.go
  git commit -m "feat: implement native macOS app menu fallback"
  ```

---

### Task 4: Refactor Common Tray Manager Definition

**Files:**
- Modify: [desktop/tray.go](file:///home/yule/Sides/airvoice/desktop/tray.go)

- [ ] **Step 1: Simplify `desktop/tray.go` to shared declarations**
  Remove platform-specific fields and method overrides, keeping only `TrayManager` and its constructor.

  Code to write:
  ```go
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
  	t := &TrayManager{app: app}
  	t.initPlatform()
  	return t
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
  ```

- [ ] **Step 2: Commit simplify changes**
  ```bash
  git add desktop/tray.go
  git commit -m "refactor: simplify desktop/tray.go to hold shared types"
  ```

---

### Task 5: Wires Tray Events to the Main Application Lifecycle

**Files:**
- Modify: [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go)
- Modify: [desktop/main.go](file:///home/yule/Sides/airvoice/desktop/main.go)

- [ ] **Step 1: Add tray reference to `App` struct and trigger status updates**
  Update [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go):
  1. Add `tray *TrayManager` to `App` struct field (near line 44).
  2. Call `a.tray.UpdateStatus()` on events.

  Code diff to write:
  ```diff
  type App struct {
  	ctx          context.Context
  	server       *server.Server
  	history      *HistoryStore
  	token        string
  	port         int
  	settingsPath string
  	settings     Settings
  	mu           sync.RWMutex
  	status       ConnectionStatus
+ 	tray         *TrayManager
  }
  ```
  And in `StartServer`, `StopServer`, `OnConnect` callbacks:
  ```go
  // In OnConnect callback:
  a.mu.Unlock()
  if a.tray != nil {
      a.tray.UpdateStatus()
  }
  if a.ctx != nil {
  
  // In OnDisconnect callback:
  a.mu.Unlock()
  if a.tray != nil {
      a.tray.UpdateStatus()
  }
  if a.ctx != nil {
  
  // In StartServer connection update:
  a.port = port
  a.status = ConnectionStatus{
      State: "waiting",
      Port:  port,
  }
  if a.tray != nil {
      a.tray.UpdateStatus()
  }
  a.mu.Unlock()
  
  // In StopServer:
  a.status = ConnectionStatus{State: "disconnected"}
  if a.tray != nil {
      a.tray.UpdateStatus()
  }
  status := a.status
  ```

- [ ] **Step 2: Assign `app.tray` in `desktop/main.go`**
  Update [desktop/main.go](file:///home/yule/Sides/airvoice/desktop/main.go) to link the `trayManager` reference.

  Code diff to write:
  ```diff
  	trayManager := NewTrayManager(app)
+ 	app.tray = trayManager
  ```

- [ ] **Step 3: Run `go mod tidy` in the workspace**
  Run to add `github.com/getlantern/systray` dependency:
  `mise exec -- go mod tidy`
  Expected output: downloads the systray dependency.

- [ ] **Step 4: Commit integration changes**
  ```bash
  git add desktop/app.go desktop/main.go go.mod go.sum
  git commit -m "feat: wire tray status updates into app connection lifecycle and add dependencies"
  ```

---

### Task 6: Verification and Compilation

- [ ] **Step 1: Run unit tests**
  Run: `mise run desktop:test`
  Expected output: Tests pass successfully.

- [ ] **Step 2: Perform manual build check**
  Run: `go build -o tmp/test_build ./desktop`
  Expected: Compiled file successfully (Note: if package webkit2gtk-4.0 is not installed locally, it may error on library dependencies, but should verify compilation dependencies and syntax errors).
