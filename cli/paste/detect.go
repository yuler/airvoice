package paste

import (
	"os"
	"runtime"
)

type SessionType string

const (
	SessionDarwin  SessionType = "darwin"
	SessionX11     SessionType = "x11"
	SessionWayland SessionType = "wayland"
	SessionUnknown SessionType = "unknown"
)

var goos = runtime.GOOS

func DetectSession() SessionType {
	if goos == "darwin" {
		return SessionDarwin
	}
	if goos == "linux" {
		if os.Getenv("XDG_SESSION_TYPE") == "wayland" || os.Getenv("WAYLAND_DISPLAY") != "" {
			return SessionWayland
		}
		if os.Getenv("DISPLAY") != "" || os.Getenv("XDG_SESSION_TYPE") == "x11" {
			return SessionX11
		}
	}
	return SessionUnknown
}
