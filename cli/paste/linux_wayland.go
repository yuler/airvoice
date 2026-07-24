//go:build linux

package paste

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type waylandPaster struct{}

// isHyprland reports whether the current session is Hyprland.
var isHyprland = detectHyprland

func detectHyprland() bool {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		return true
	}
	return os.Getenv("XDG_CURRENT_DESKTOP") == "Hyprland"
}

// terminalClasses is a set of window class names that are terminals.
var terminalClasses = map[string]bool{
	"Alacritty":         true,
	"kitty":             true,
	"foot":              true,
	"wezterm":           true,
	"gnome-terminal":    true,
	"konsole":           true,
	"xfce4-terminal":    true,
	"tilix":             true,
	"terminator":        true,
	"Terminal":          true,
	"io.github.cellbots.Terminal": true,
}

// isTerminalWindow reports whether the given window class looks like a terminal.
func isTerminalWindow(class string) bool {
	if terminalClasses[class] {
		return true
	}
	for c := range terminalClasses {
		if len(class) > len(c) && class[:len(c)+1] == c+"." {
			return true
		}
	}
	return false
}

// hyprctlWindow is the JSON shape returned by hyprctl activewindow -j.
type hyprctlWindow struct {
	Class string `json:"class"`
}

// activeWindowClass returns the window class of the currently focused window.
// Returns empty string on any failure.
var activeWindowClass = detectActiveWindowClass

func detectActiveWindowClass() string {
	out, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil {
		return ""
	}
	var w hyprctlWindow
	if err := json.Unmarshal(out, &w); err != nil {
		return ""
	}
	return w.Class
}

func (w *waylandPaster) Paste(text string) error {
	if err := runCommand("wl-copy", text); err != nil {
		return err
	}
	time.Sleep(80 * time.Millisecond)

	if isHyprland() {
		modifiers := "CTRL, V"
		if isTerminalWindow(activeWindowClass()) {
			modifiers = "CTRL SHIFT, V"
		}
		return runCommand("hyprctl", "", "dispatch", "sendshortcut", modifiers+", activewindow")
	}

	ensureYdotoolSocket()
	return runCommand("ydotool", "", "key", "CTRL+v")
}

func ensureYdotoolSocket() {
	if os.Getenv("YDOTOOL_SOCKET") != "" {
		return
	}
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if _, err := os.Stat(runtimeDir + "/.ydotool_socket"); err == nil {
			os.Setenv("YDOTOOL_SOCKET", runtimeDir+"/.ydotool_socket")
			return
		}
	}
	if _, err := os.Stat("/tmp/.ydotool_socket"); err == nil {
		os.Setenv("YDOTOOL_SOCKET", "/tmp/.ydotool_socket")
		return
	}
	uidSocket := fmt.Sprintf("/run/user/%d/ydotool/socket", os.Getuid())
	if _, err := os.Stat(uidSocket); err == nil {
		os.Setenv("YDOTOOL_SOCKET", uidSocket)
	}
}

func (w *waylandPaster) Name() string {
	if isHyprland() {
		return "wayland-hyprland (wl-copy + hyprctl sendshortcut)"
	}
	return "wayland (wl-copy + ydotool)"
}
