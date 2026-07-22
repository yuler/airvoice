//go:build linux

package paste

import (
	"fmt"
	"os"
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

func (w *waylandPaster) Paste(text string) error {
	if err := runCommand("wl-copy", text); err != nil {
		return err
	}
	time.Sleep(80 * time.Millisecond)

	if isHyprland() {
		return runCommand("hyprctl", "", "dispatch", "sendshortcut", "CTRL, V, activewindow")
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
		return "wayland-hyprland"
	}
	return "wayland"
}
