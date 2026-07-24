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
	if goos != "linux" {
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
		// DISPLAY alone means X11 (or XWayland with missing Wayland env). Do not
		// prefer Wayland just because wl-clipboard is installed — that breaks
		// pure X11 sessions that also have wl-clipboard for other reasons.
		if os.Getenv("XDG_SESSION_TYPE") == "x11" || os.Getenv("DISPLAY") != "" {
			return SessionX11
		}
		// Last resort: wl-clipboard present and no DISPLAY (some Wayland setups).
		if hasWlClipboardTools() {
			return SessionWayland
		}
	}
	return SessionUnknown
}
