//go:build linux

package paste

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

type commandCall struct {
	name  string
	stdin string
	args  []string
}

func TestLinuxPasters(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	var calls []commandCall
	runCommand = func(name string, stdin string, args ...string) error {
		calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
		return nil
	}

	t.Run("x11Paster success", func(t *testing.T) {
		calls = nil
		p := &x11Paster{}
		if p.Name() != "x11 (xclip + xdotool)" {
			t.Errorf("expected x11 (xclip + xdotool), got %s", p.Name())
		}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "xclip", stdin: "hello world", args: []string{"-selection", "clipboard"}},
			{name: "xdotool", stdin: "", args: []string{"key", "ctrl+v"}},
		}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("got calls %+v, expected %+v", calls, expected)
		}
	})

	t.Run("x11Paster first command failure", func(t *testing.T) {
		runCommand = func(name string, stdin string, args ...string) error {
			if name == "xclip" {
				return errors.New("xclip failed")
			}
			return nil
		}
		p := &x11Paster{}
		err := p.Paste("hello")
		if err == nil || err.Error() != "xclip failed" {
			t.Errorf("expected xclip failed error, got: %v", err)
		}
	})

	t.Run("waylandPaster success", func(t *testing.T) {
		origIsHyprland := isHyprland
		isHyprland = func() bool { return false }
		defer func() { isHyprland = origIsHyprland }()

		runCommand = func(name string, stdin string, args ...string) error {
			calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
			return nil
		}
		calls = nil
		p := &waylandPaster{}
		if p.Name() != "wayland (wl-copy + ydotool)" {
			t.Errorf("expected wayland (wl-copy + ydotool), got %s", p.Name())
		}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "wl-copy", stdin: "hello world", args: nil},
			{name: "ydotool", stdin: "", args: []string{"key", "CTRL+v"}},
		}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("got calls %+v, expected %+v", calls, expected)
		}
	})

	t.Run("waylandPaster hyprland success", func(t *testing.T) {
		origIsHyprland := isHyprland
		isHyprland = func() bool { return true }
		defer func() { isHyprland = origIsHyprland }()

		origActiveWindowClass := activeWindowClass
		activeWindowClass = func() string { return "firefox" }
		defer func() { activeWindowClass = origActiveWindowClass }()

		runCommand = func(name string, stdin string, args ...string) error {
			calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
			return nil
		}
		calls = nil
		p := &waylandPaster{}
		if p.Name() != "wayland-hyprland (wl-copy + hyprctl sendshortcut)" {
			t.Errorf("expected wayland-hyprland (wl-copy + hyprctl sendshortcut), got %s", p.Name())
		}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "wl-copy", stdin: "hello world", args: nil},
			{name: "hyprctl", stdin: "", args: []string{"dispatch", "sendshortcut", "CTRL, V, activewindow"}},
		}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("got calls %+v, expected %+v", calls, expected)
		}
	})

	t.Run("waylandPaster hyprland terminal uses Ctrl+Shift+V", func(t *testing.T) {
		origIsHyprland := isHyprland
		isHyprland = func() bool { return true }
		defer func() { isHyprland = origIsHyprland }()

		origActiveWindowClass := activeWindowClass
		activeWindowClass = func() string { return "Alacritty" }
		defer func() { activeWindowClass = origActiveWindowClass }()

		runCommand = func(name string, stdin string, args ...string) error {
			calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
			return nil
		}
		calls = nil
		p := &waylandPaster{}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "wl-copy", stdin: "hello world", args: nil},
			{name: "hyprctl", stdin: "", args: []string{"dispatch", "sendshortcut", "CTRL SHIFT, V, activewindow"}},
		}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("got calls %+v, expected %+v", calls, expected)
		}
	})

	t.Run("waylandPaster first command failure", func(t *testing.T) {
		runCommand = func(name string, stdin string, args ...string) error {
			if name == "wl-copy" {
				return errors.New("wl-copy failed")
			}
			return nil
		}
		p := &waylandPaster{}
		err := p.Paste("hello")
		if err == nil || err.Error() != "wl-copy failed" {
			t.Errorf("expected wl-copy failed error, got: %v", err)
		}
	})

	t.Run("New function switches", func(t *testing.T) {
		origGOOS := goos
		origSessionType := os.Getenv("XDG_SESSION_TYPE")
		origWaylandDisplay := os.Getenv("WAYLAND_DISPLAY")
		origDisplay := os.Getenv("DISPLAY")
		origLookPath := lookPath
		origDetectLookPath := detectLookPath
		origIsHyprland := isHyprland
		defer func() {
			goos = origGOOS
			os.Setenv("XDG_SESSION_TYPE", origSessionType)
			os.Setenv("WAYLAND_DISPLAY", origWaylandDisplay)
			os.Setenv("DISPLAY", origDisplay)
			lookPath = origLookPath
			detectLookPath = origDetectLookPath
			isHyprland = origIsHyprland
		}()

		goos = "linux"
		os.Unsetenv("XDG_SESSION_TYPE")
		os.Unsetenv("WAYLAND_DISPLAY")
		os.Unsetenv("DISPLAY")
		isHyprland = func() bool { return false }

		mockLookPath := func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}
		lookPath = mockLookPath
		detectLookPath = mockLookPath

		// Wayland
		os.Setenv("XDG_SESSION_TYPE", "wayland")
		p, err := New()
		if err != nil {
			t.Fatalf("New failed on wayland: %v", err)
		}
		if p.Name() != "wayland (wl-copy + ydotool)" {
			t.Errorf("expected wayland (wl-copy + ydotool), got %s", p.Name())
		}

		// Wayland with missing wl-clipboard tools
		mockMissing := func(file string) (string, error) {
			return "", errors.New("not found")
		}
		lookPath = mockMissing
		detectLookPath = mockMissing
		_, err = New()
		if err == nil {
			t.Error("expected error on wayland when wl-clipboard tools are missing, got nil")
		}

		// Restore mock
		lookPath = mockLookPath
		detectLookPath = mockLookPath

		// Hyprland Wayland
		isHyprland = func() bool { return true }
		p, err = New()
		if err != nil {
			t.Fatalf("New failed on hyprland wayland: %v", err)
		}
		if p.Name() != "wayland-hyprland (wl-copy + hyprctl sendshortcut)" {
			t.Errorf("expected wayland-hyprland (wl-copy + hyprctl sendshortcut), got %s", p.Name())
		}

		// Hyprland Wayland with missing hyprctl
		lookPath = func(file string) (string, error) {
			if file == "hyprctl" {
				return "", errors.New("not found")
			}
			return "/usr/bin/" + file, nil
		}
		_, err = New()
		if err == nil {
			t.Error("expected error on hyprland wayland when hyprctl is missing, got nil")
		}

		// Restore mock
		lookPath = func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}
		isHyprland = func() bool { return false }

		// X11
		os.Setenv("XDG_SESSION_TYPE", "x11")
		p, err = New()
		if err != nil {
			t.Fatalf("New failed on x11: %v", err)
		}
		if p.Name() != "x11 (xclip + xdotool)" {
			t.Errorf("expected x11 (xclip + xdotool), got %s", p.Name())
		}

		// Unknown
		os.Setenv("XDG_SESSION_TYPE", "unknown")
		detectLookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		_, err = New()
		if err == nil {
			t.Error("expected error on unknown session, got nil")
		}
	})
}
