# 14-desktop-theme-and-tray-fix-design

## Description
This design document addresses two desktop issues on the Linux platform:
1. **Theme Toggle failure**: Clicking the theme icon toggles the `.dark`/`.light` classes on the `html` element, but does not visually update the application. This happens because the colors in `tailwind.config.js` are hardcoded to dark mode hex values rather than using CSS variables.
2. **System Tray failure**: Wails v2 does not natively support system tray menus on Linux/Windows. The current implementation registers a window menu using Wails' `options.App.Menu`, which fails to render as a system tray icon.

## Proposed Solution

### Theme Toggle Fix
We will update `tailwind.config.js` to map all custom theme colors to CSS variables defined in `styles.css`. This enables the UI elements (using Tailwind classes like `bg-bg-primary` and `text-text-primary`) to instantly update when the `html` element's class switches between `.dark` and `.light`.

### System Tray Fix
We will introduce `github.com/getlantern/systray` for Linux/Windows while keeping Wails' native application menu for macOS to avoid threading conflicts.
1. **Platform Split**:
   - `tray.go`: Contains core `TrayManager` definition and shared utility methods.
   - `tray_systray.go` (`//go:build linux || windows`): Runs `systray` in a background goroutine, displays a minimalist white microphone tray icon, handles window show/hide/quit events, and updates the status label dynamically.
   - `tray_darwin.go` (`//go:build darwin`): Returns Wails' native menu bar for macOS.
2. **Dynamic Connection Status**:
   - Save a reference of `TrayManager` in the `App` struct.
   - Call `UpdateStatus()` on connection-related events in `app.go`.

---

## Architectural & Code Changes

### Files to Modify
- [tailwind.config.js](file:///home/yule/Sides/airvoice/desktop/frontend/tailwind.config.js): Update colors mapping to CSS variables.
- [desktop/tray.go](file:///home/yule/Sides/airvoice/desktop/tray.go): Refactor into core logic.
- [desktop/app.go](file:///home/yule/Sides/airvoice/desktop/app.go): Keep a reference to the tray manager and call `UpdateStatus()` on events.
- [desktop/main.go](file:///home/yule/Sides/airvoice/desktop/main.go): Wire the tray manager to the `App` instance.

### New Files
- [desktop/tray_systray.go](file:///home/yule/Sides/airvoice/desktop/tray_systray.go) (`//go:build linux || windows`): Implement system tray menu using `systray`.
- [desktop/tray_darwin.go](file:///home/yule/Sides/airvoice/desktop/tray_darwin.go) (`//go:build darwin`): Implement native application menu bar.
- [desktop/tray_icon.png](file:///home/yule/Sides/airvoice/desktop/tray_icon.png): Minimalist monochrome microphone icon (already generated).

---

## Verification Plan

### Automated Tests
- Run `go test ./desktop/...` to ensure it compiles and existing unit tests pass.

### Manual Verification
- Compile and launch the desktop app on Linux (`wails dev` or `wails build`).
- Verify that clicking the theme icon in the header toggles between dark (pure black) and light (white) mode.
- Verify that the tray icon is registered successfully in the system panel, and its menu options ("Show Window", "Hide Window", "Quit") work correctly.
- Verify that connecting/disconnecting the mobile app updates the tray menu text in real-time.
