//go:build linux

package paste

import (
	"errors"
	"os"
	"path/filepath"
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

		origResolve := resolveYdotoolSocket
		resolveYdotoolSocket = func() string { return "" }
		defer func() { resolveYdotoolSocket = origResolve }()

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

	t.Run("waylandPaster passes YDOTOOL_SOCKET via child env without mutating process env", func(t *testing.T) {
		t.Setenv("YDOTOOL_SOCKET", "")

		origIsHyprland := isHyprland
		isHyprland = func() bool { return false }
		defer func() { isHyprland = origIsHyprland }()

		origResolve := resolveYdotoolSocket
		resolveYdotoolSocket = func() string { return "/run/user/1000/.ydotool_socket" }
		defer func() { resolveYdotoolSocket = origResolve }()

		origRunWithEnv := runCommandWithEnv
		defer func() { runCommandWithEnv = origRunWithEnv }()

		var gotEnv []string
		var envCalls []commandCall
		runCommand = func(name string, stdin string, args ...string) error {
			calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
			return nil
		}
		runCommandWithEnv = func(extraEnv []string, name string, stdin string, args ...string) error {
			gotEnv = append([]string(nil), extraEnv...)
			envCalls = append(envCalls, commandCall{name: name, stdin: stdin, args: args})
			return nil
		}
		calls = nil

		p := &waylandPaster{}
		if err := p.Paste("hello env"); err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		if os.Getenv("YDOTOOL_SOCKET") != "" {
			t.Fatalf("process YDOTOOL_SOCKET was mutated to %q", os.Getenv("YDOTOOL_SOCKET"))
		}
		wantEnv := []string{"YDOTOOL_SOCKET=/run/user/1000/.ydotool_socket"}
		if !reflect.DeepEqual(gotEnv, wantEnv) {
			t.Errorf("extraEnv = %v, want %v", gotEnv, wantEnv)
		}
		wantEnvCalls := []commandCall{
			{name: "ydotool", stdin: "", args: []string{"key", "CTRL+v"}},
		}
		if !reflect.DeepEqual(envCalls, wantEnvCalls) {
			t.Errorf("envCalls = %+v, want %+v", envCalls, wantEnvCalls)
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
			{name: "hyprctl", stdin: "", args: []string{"dispatch", `hl.dsp.send_shortcut({ mods = "CTRL", key = "V", window = "activewindow" })`}},
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
			{name: "hyprctl", stdin: "", args: []string{"dispatch", `hl.dsp.send_shortcut({ mods = "CTRL,SHIFT", key = "V", window = "activewindow" })`}},
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

func TestIsTerminalWindow(t *testing.T) {
	terminals := []string{
		"Alacritty", "kitty", "foot", "wezterm", "konsole",
		"herdr", "Herdr", "HERDR", "ghostty", "Ghostty",
		"org.wezfurlong.wezterm", "footclient",
	}
	for _, class := range terminals {
		if !isTerminalWindow(class) {
			t.Errorf("expected %q to be detected as terminal", class)
		}
	}
	nonTerminals := []string{"firefox", "Code", "chromium", "", "nautilus"}
	for _, class := range nonTerminals {
		if isTerminalWindow(class) {
			t.Errorf("expected %q NOT to be detected as terminal", class)
		}
	}
}

func TestDetectYdotoolSocketAndOwnership(t *testing.T) {
	origTmpPath := tmpYdotoolSocketPath
	defer func() { tmpYdotoolSocketPath = origTmpPath }()

	t.Run("respects existing YDOTOOL_SOCKET", func(t *testing.T) {
		t.Setenv("YDOTOOL_SOCKET", "/custom/socket")
		if got := detectYdotoolSocket(); got != "" {
			t.Fatalf("expected empty fallback when YDOTOOL_SOCKET is set, got %q", got)
		}
	})

	t.Run("prefers XDG_RUNTIME_DIR socket without mutating env", func(t *testing.T) {
		runtimeDir := t.TempDir()
		sock := filepath.Join(runtimeDir, ".ydotool_socket")
		if err := os.WriteFile(sock, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("YDOTOOL_SOCKET", "")
		t.Setenv("XDG_RUNTIME_DIR", runtimeDir)
		tmpYdotoolSocketPath = filepath.Join(t.TempDir(), "missing")

		if got := detectYdotoolSocket(); got != sock {
			t.Fatalf("got %q, want %q", got, sock)
		}
		if os.Getenv("YDOTOOL_SOCKET") != "" {
			t.Fatalf("expected YDOTOOL_SOCKET env to remain unset, got %q", os.Getenv("YDOTOOL_SOCKET"))
		}
	})

	t.Run("uses tmp fallback only when owned by current user and not a symlink", func(t *testing.T) {
		dir := t.TempDir()
		owned := filepath.Join(dir, ".ydotool_socket")
		if err := os.WriteFile(owned, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("YDOTOOL_SOCKET", "")
		t.Setenv("XDG_RUNTIME_DIR", filepath.Join(dir, "empty-runtime"))
		tmpYdotoolSocketPath = owned

		if got := detectYdotoolSocket(); got != owned {
			t.Fatalf("expected owned tmp fallback %q, got %q", owned, got)
		}

		// Symlink must be rejected.
		link := filepath.Join(dir, "symlink_socket")
		if err := os.Symlink(owned, link); err != nil {
			t.Fatal(err)
		}
		if isOwnedByCurrentUser(link) {
			t.Fatal("expected symlink to be rejected by isOwnedByCurrentUser")
		}

		// File owned by another user (root) must be rejected when running as non-root.
		if os.Getuid() != 0 {
			if isOwnedByCurrentUser("/etc/passwd") {
				t.Fatal("expected root-owned file to be rejected by isOwnedByCurrentUser")
			}
		}
	})
}

func TestRunCommandCaptureEnvDoesNotMutateProcessEnv(t *testing.T) {
	t.Setenv("AIRVOICE_CHILD_ONLY", "")
	_, err := runCommandCaptureEnv(
		[]string{"AIRVOICE_CHILD_ONLY=expected-val"},
		"sh", "", "-c", `test "$AIRVOICE_CHILD_ONLY" = "expected-val"`,
	)
	if err != nil {
		t.Fatalf("expected child process to observe extraEnv, got err=%v", err)
	}
	if got := os.Getenv("AIRVOICE_CHILD_ONLY"); got != "" {
		t.Fatalf("expected parent env to remain empty, got %q", got)
	}
}
