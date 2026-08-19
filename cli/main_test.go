package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuler/airvoice/cli/config"
)

func TestRunDefaultsToServe(t *testing.T) {
	orig := serveFn
	called := false
	var gotArgs []string
	serveFn = func(args []string) int {
		called = true
		gotArgs = append([]string(nil), args...)
		return 0
	}
	defer func() { serveFn = orig }()

	if code := run(nil); code != 0 || !called {
		t.Fatalf("run(nil) code=%d called=%v", code, called)
	}

	called = false
	if code := run([]string{"--port", "9000"}); code != 0 || !called {
		t.Fatalf("run(--port) code=%d called=%v", code, called)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "--port" {
		t.Fatalf("serve args = %v", gotArgs)
	}
}

func TestRunHelpListsCommands(t *testing.T) {
	orig := serveFn
	serveFn = func(args []string) int {
		t.Fatalf("serve should not start for help, args=%v", args)
		return 0
	}
	defer func() { serveFn = orig }()

	for _, args := range [][]string{{"help"}, {"--help"}, {"-h"}} {
		stdout := captureStdout(t, func() {
			if code := run(args); code != 0 {
				t.Fatalf("run(%v) code=%d, want 0", args, code)
			}
		})
		for _, want := range []string{
			"airvoice [--port 7654]",
			"airvoice settings",
			"airvoice token refresh",
			"airvoice doctor",
			"airvoice version",
			"airvoice help",
		} {
			if !strings.Contains(stdout, want) {
				t.Fatalf("run(%v) missing %q in:\n%s", args, want, stdout)
			}
		}
	}
}

func TestRunServeIsUnknown(t *testing.T) {
	stderr := captureStderr(t, func() {
		if code := run([]string{"serve"}); code != 1 {
			t.Fatalf("code = %d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "airvoice [--port 7654]") {
		t.Fatalf("usage missing default command:\n%s", stderr)
	}
	if strings.Contains(stderr, "airvoice serve") {
		t.Fatalf("usage still documents serve:\n%s", stderr)
	}
}

func TestRunSettingsPrintsJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	wantPath := filepath.Join(home, ".airvoice", "settings.json")
	stdout := captureStdout(t, func() {
		if code := run([]string{"settings"}); code != 0 {
			t.Fatalf("code = %d", code)
		}
	})
	if !strings.Contains(stdout, wantPath) {
		t.Fatalf("stdout missing settings path %q:\n%s", wantPath, stdout)
	}
	idx := strings.Index(stdout, "{")
	if idx < 0 {
		t.Fatalf("stdout missing JSON:\n%s", stdout)
	}
	var s config.Settings
	if err := json.Unmarshal([]byte(stdout[idx:]), &s); err != nil {
		t.Fatalf("stdout not JSON: %q err=%v", stdout, err)
	}
	if s.Token == "" {
		t.Fatal("expected token in settings output")
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatal(err)
	}
}

func TestRunTokenRefreshRotates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	first, err := config.EnsureToken(config.Path())
	if err != nil {
		t.Fatal(err)
	}

	stdout := captureStdout(t, func() {
		if code := run([]string{"token", "refresh"}); code != 0 {
			t.Fatalf("code = %d", code)
		}
	})
	newTok := strings.TrimSpace(stdout)
	if newTok == "" || newTok == first.Token {
		t.Fatalf("refresh printed %q, old %q", newTok, first.Token)
	}
	disk, err := config.Load(config.Path())
	if err != nil {
		t.Fatal(err)
	}
	if disk.Token != newTok {
		t.Fatalf("disk %q printed %q", disk.Token, newTok)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = w
	fn()
	_ = w.Close()
	os.Stderr = orig
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}
