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
			"airvoice upgrade [--check] [--version X.Y.Z]",
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

func TestRunDispatchesUpgrade(t *testing.T) {
	orig := upgradeFn
	var gotArgs []string
	called := false
	upgradeFn = func(args []string) int {
		called = true
		gotArgs = append([]string(nil), args...)
		return 0
	}
	defer func() { upgradeFn = orig }()

	if code := run([]string{"upgrade", "--check", "--version", "0.4.8"}); code != 0 || !called {
		t.Fatalf("run(upgrade) code=%d called=%v", code, called)
	}
	want := []string{"--check", "--version", "0.4.8"}
	if len(gotArgs) != len(want) {
		t.Fatalf("upgrade args = %v, want %v", gotArgs, want)
	}
	for i := range want {
		if gotArgs[i] != want[i] {
			t.Fatalf("upgrade args[%d] = %q, want %q", i, gotArgs[i], want[i])
		}
	}
}

func TestRunUpgradeRejectsPositionalArgs(t *testing.T) {
	stderr := captureStderr(t, func() {
		if code := runUpgrade([]string{"unexpected"}); code != 1 {
			t.Fatalf("code = %d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "airvoice upgrade [--check] [--version X.Y.Z]") {
		t.Fatalf("stderr missing upgrade usage:\n%s", stderr)
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

func TestParseServeFlagsRelaySemantics(t *testing.T) {
	// Bare invocation stays LAN mode.
	opts, err := parseServeFlags(nil, io.Discard)
	if err != nil {
		t.Fatalf("parseServeFlags(nil): %v", err)
	}
	if opts.useRelay() || opts.port != 7654 {
		t.Fatalf("bare opts = %+v, want LAN port 7654", opts)
	}

	// --relay enables relay mode without URL override.
	opts, err = parseServeFlags([]string{"--relay"}, io.Discard)
	if err != nil {
		t.Fatalf("parseServeFlags(--relay): %v", err)
	}
	if !opts.useRelay() || opts.relayURL != "" {
		t.Fatalf("--relay opts = %+v", opts)
	}

	// --relay-url implicitly enables relay mode.
	opts, err = parseServeFlags([]string{"--relay-url", "wss://custom.example/ws", "--relay-auth-key", "k123"}, io.Discard)
	if err != nil {
		t.Fatalf("parseServeFlags(--relay-url): %v", err)
	}
	if !opts.useRelay() || opts.relayURL != "wss://custom.example/ws" || opts.relayAuthKey != "k123" {
		t.Fatalf("--relay-url opts = %+v", opts)
	}
}

func TestRunSettingsMasksRelayAuthKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if err := config.Save(config.Path(), config.Settings{
		Port:           7655,
		ConnectionMode: config.ModeRelay,
		RelayURL:       "wss://relay.example.com/ws",
		RelayAuthKey:   "top-secret-key",
	}); err != nil {
		t.Fatal(err)
	}

	stdout := captureStdout(t, func() {
		if code := run([]string{"settings"}); code != 0 {
			t.Fatalf("code = %d", code)
		}
	})
	if strings.Contains(stdout, "top-secret-key") {
		t.Fatalf("settings output leaked raw relayAuthKey:\n%s", stdout)
	}
	idx := strings.Index(stdout, "{")
	if idx < 0 {
		t.Fatalf("stdout missing JSON:\n%s", stdout)
	}
	var s config.Settings
	if err := json.Unmarshal([]byte(stdout[idx:]), &s); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if s.RelayAuthKey != "********" {
		t.Fatalf("RelayAuthKey = %q, want masked", s.RelayAuthKey)
	}
	if s.RelayURL != "wss://relay.example.com/ws" || s.ConnectionMode != config.ModeRelay {
		t.Fatalf("unexpected relay settings: %+v", s)
	}
}

func TestRunTokenRefreshRotatesRelayKeyToo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	s1, err := config.EnsureRelayKey(config.Path())
	if err != nil {
		t.Fatal(err)
	}
	stdout := captureStdout(t, func() {
		if code := run([]string{"token", "refresh"}); code != 0 {
			t.Fatalf("code = %d", code)
		}
	})
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected refreshed token on stdout")
	}
	s2, err := config.Load(config.Path())
	if err != nil {
		t.Fatal(err)
	}
	if s2.RelayKey == "" || s2.RelayKey == s1.RelayKey {
		t.Fatalf("relayKey not rotated: %q vs %q", s2.RelayKey, s1.RelayKey)
	}
}

func TestRunRelayRejectsInsecurePublicWS(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	stderr := captureStderr(t, func() {
		if code := runServer([]string{"--relay-url", "ws://relay.airvoice.yuler.cc/ws"}); code != 1 {
			t.Fatalf("code = %d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "wss://") {
		t.Fatalf("expected wss:// error in stderr, got:\n%s", stderr)
	}
}
