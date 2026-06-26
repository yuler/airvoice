//go:build linux

package paste

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
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
		if p.Name() != "x11" {
			t.Errorf("expected x11, got %s", p.Name())
		}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "xclip", stdin: "hello world", args: []string{"-selection", "clipboard"}},
			{name: "xdotool", stdin: "", args: []string{"key", "ctrl+v"}},
		}
		for i := 0; i < 25; i++ {
			if len(calls) >= 2 {
				break
			}
			time.Sleep(10 * time.Millisecond)
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
		var mu sync.Mutex
		runCommand = func(name string, stdin string, args ...string) error {
			mu.Lock()
			calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
			mu.Unlock()
			return nil
		}
		calls = nil
		p := &waylandPaster{}
		if p.Name() != "wayland" {
			t.Errorf("expected wayland, got %s", p.Name())
		}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "wl-copy", stdin: "hello world", args: nil},
			{name: "ydotool", stdin: "", args: []string{"key", "CTRL+v"}},
		}
		for i := 0; i < 25; i++ {
			mu.Lock()
			length := len(calls)
			mu.Unlock()
			if length >= 2 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		mu.Lock()
		actualCalls := calls
		mu.Unlock()
		if !reflect.DeepEqual(actualCalls, expected) {
			t.Errorf("got calls %+v, expected %+v", actualCalls, expected)
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
		defer func() {
			goos = origGOOS
			os.Setenv("XDG_SESSION_TYPE", origSessionType)
			os.Setenv("WAYLAND_DISPLAY", origWaylandDisplay)
			os.Setenv("DISPLAY", origDisplay)
			lookPath = origLookPath
		}()

		goos = "linux"
		os.Unsetenv("XDG_SESSION_TYPE")
		os.Unsetenv("WAYLAND_DISPLAY")
		os.Unsetenv("DISPLAY")

		// Mock lookPath to succeed
		lookPath = func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}

		// Wayland
		os.Setenv("XDG_SESSION_TYPE", "wayland")
		p, err := New()
		if err != nil {
			t.Fatalf("New failed on wayland: %v", err)
		}
		if p.Name() != "wayland" {
			t.Errorf("expected wayland, got %s", p.Name())
		}

		// Wayland with missing ydotool
		lookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		_, err = New()
		if err == nil {
			t.Error("expected error on wayland when ydotool is missing, got nil")
		}

		// Restore mock
		lookPath = func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}

		// X11
		os.Setenv("XDG_SESSION_TYPE", "x11")
		p, err = New()
		if err != nil {
			t.Fatalf("New failed on x11: %v", err)
		}
		if p.Name() != "x11" {
			t.Errorf("expected x11, got %s", p.Name())
		}

		// Unknown
		os.Setenv("XDG_SESSION_TYPE", "unknown")
		_, err = New()
		if err == nil {
			t.Error("expected error on unknown session, got nil")
		}
	})
}
