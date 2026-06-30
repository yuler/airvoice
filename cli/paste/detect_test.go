package paste

import (
	"os"
	"testing"
)

func TestDetectSession(t *testing.T) {
	// Save environment and goos, restore at end
	origGOOS := goos
	origSessionType := os.Getenv("XDG_SESSION_TYPE")
	origWaylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	origDisplay := os.Getenv("DISPLAY")
	defer func() {
		goos = origGOOS
		os.Setenv("XDG_SESSION_TYPE", origSessionType)
		os.Setenv("WAYLAND_DISPLAY", origWaylandDisplay)
		os.Setenv("DISPLAY", origDisplay)
	}()

	tests := []struct {
		name     string
		targetOS string
		env      map[string]string
		want     SessionType
	}{
		{
			name:     "darwin",
			targetOS: "darwin",
			want:     SessionDarwin,
		},
		{
			name:     "linux wayland by XDG",
			targetOS: "linux",
			env:      map[string]string{"XDG_SESSION_TYPE": "wayland"},
			want:     SessionWayland,
		},
		{
			name:     "linux wayland by WAYLAND_DISPLAY",
			targetOS: "linux",
			env:      map[string]string{"WAYLAND_DISPLAY": "wayland-0"},
			want:     SessionWayland,
		},
		{
			name:     "linux X11 by XDG",
			targetOS: "linux",
			env:      map[string]string{"XDG_SESSION_TYPE": "x11"},
			want:     SessionX11,
		},
		{
			name:     "linux X11 by DISPLAY",
			targetOS: "linux",
			env:      map[string]string{"DISPLAY": ":0"},
			want:     SessionX11,
		},
		{
			name:     "linux unknown",
			targetOS: "linux",
			want:     SessionUnknown,
		},
		{
			name:     "windows session",
			targetOS: "windows",
			want:     SessionWindows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goos = tt.targetOS
			os.Unsetenv("XDG_SESSION_TYPE")
			os.Unsetenv("WAYLAND_DISPLAY")
			os.Unsetenv("DISPLAY")
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			got := DetectSession()
			if got != tt.want {
				t.Errorf("DetectSession() = %v, want %v", got, tt.want)
			}
		})
	}
}
