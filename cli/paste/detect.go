package paste

import (
	"os"
	"os/exec"
	"runtime"
)

type SessionType string

const (
	SessionDarwin  SessionType = "darwin"
	SessionX11     SessionType = "x11"
	SessionWayland SessionType = "wayland"
	SessionWindows SessionType = "windows"
	SessionUnknown SessionType = "unknown"
)

var goos = runtime.GOOS

// detectLookPath is a package-level variable that can be overridden in tests.
var detectLookPath = exec.LookPath

func hasWlClipboardTools() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err1 := detectLookPath("wl-copy")
	_, err2 := detectLookPath("wl-paste")
	return err1 == nil && err2 == nil
}

func DetectSession() SessionType {
	if goos == "darwin" {
		return SessionDarwin
	}
	if goos == "windows" {
		return SessionWindows
	}
	if goos == "linux" {
		if os.Getenv("XDG_SESSION_TYPE") == "wayland" || os.Getenv("WAYLAND_DISPLAY") != "" {
			return SessionWayland
		}
		// Prefer wl-clipboard when both tools are installed (typical Wayland desktop).
		if hasWlClipboardTools() && os.Getenv("XDG_SESSION_TYPE") != "x11" {
			return SessionWayland
		}
		if os.Getenv("DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "x11" {
			return SessionX11
		}
		if hasWlClipboardTools() {
			return SessionWayland
		}
	}
	return SessionUnknown
}
