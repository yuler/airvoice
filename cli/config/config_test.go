package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yuler/airvoice/cli/e2e"
)

func TestPathUsesHomeAirvoice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	got := Path()
	want := filepath.Join(home, ".airvoice", "settings.json")
	if got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestEnsureTokenCreatesFileAndReusesToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	first, err := EnsureToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if first.Token == "" {
		t.Fatal("expected a token")
	}

	second, err := EnsureToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if second.Token != first.Token {
		t.Fatalf("token rotated on second load: %q vs %q", second.Token, first.Token)
	}
}

func TestEnsureTokenReplacesEmptyTokenAndPreservesFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	raw := []byte(`{"token":"   ","port":7655,"autoStart":true,"language":"en-US"}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := EnsureToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Token == "" {
		t.Fatal("expected minted token")
	}
	if got.Port != 7655 || !got.AutoStart || got.Language != "en-US" {
		t.Fatalf("lost settings fields: %+v", got)
	}
}

func TestRotateTokenPreservesOtherFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"token":"old-token","port":7655,"language":"zh-CN"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := RotateToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Token == "" || got.Token == "old-token" {
		t.Fatalf("token not rotated: %q", got.Token)
	}
	if got.Port != 7655 || got.Language != "zh-CN" {
		t.Fatalf("lost settings fields: %+v", got)
	}

	disk, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if disk.Token != got.Token {
		t.Fatalf("disk token %q != returned %q", disk.Token, got.Token)
	}
}

func TestSavePreservesTokenWhenIncomingEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if _, err := EnsureToken(path); err != nil {
		t.Fatal(err)
	}
	before, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := Save(path, Settings{Port: 9000, Language: "en-US", AutoStart: true}); err != nil {
		t.Fatal(err)
	}
	after, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if after.Token != before.Token {
		t.Fatalf("token dropped: %q vs %q", after.Token, before.Token)
	}
	if after.Port != 9000 || after.Language != "en-US" || !after.AutoStart {
		t.Fatalf("settings not saved: %+v", after)
	}
}

func TestWatchTokenFiresOnlyOnTokenChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	s, err := EnsureToken(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := make(chan string, 2)
	go WatchToken(ctx, path, 20*time.Millisecond, s.Token, func(token string) {
		got <- token
	})

	s.Language = "en-US"
	if err := writeSettings(path, s); err != nil {
		t.Fatal(err)
	}
	select {
	case tok := <-got:
		t.Fatalf("language-only write fired watch: %q", tok)
	case <-time.After(80 * time.Millisecond):
	}

	s.Token = "rotated-token"
	if err := writeSettings(path, s); err != nil {
		t.Fatal(err)
	}
	select {
	case tok := <-got:
		if tok != "rotated-token" {
			t.Fatalf("got %q", tok)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for token change")
	}
}

func writeSettings(path string, s Settings) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func TestEnsureTokenPropagatesErrorWithoutOverwritingCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	corrupt := []byte("{not valid json")
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureToken(path); err == nil {
		t.Fatal("expected EnsureToken to return error on corrupt JSON")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(corrupt) {
		t.Fatalf("corrupt file was overwritten: got %q, want %q", after, corrupt)
	}
}

func TestSavePropagatesErrorWithoutOverwritingCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	corrupt := []byte("{not valid json")
	if err := os.WriteFile(path, corrupt, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := Save(path, Settings{Port: 7655, Language: "en-US"}); err == nil {
		t.Fatal("expected Save to return error on corrupt JSON")
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(corrupt) {
		t.Fatalf("corrupt file was overwritten: got %q, want %q", after, corrupt)
	}
}

func TestWriteFileAtomicPermissionsAndConcurrentSaves(t *testing.T) {
	baseDir := t.TempDir()
	settingsDir := filepath.Join(baseDir, "nested-airvoice")
	path := filepath.Join(settingsDir, "settings.json")

	if _, err := EnsureToken(path); err != nil {
		t.Fatal(err)
	}

	if runtime.GOOS != "windows" {
		dirInfo, err := os.Stat(settingsDir)
		if err != nil {
			t.Fatal(err)
		}
		if perm := dirInfo.Mode().Perm(); perm != 0o700 {
			t.Errorf("dir perm = %o, want 0700", perm)
		}
		fileInfo, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if perm := fileInfo.Mode().Perm(); perm != 0o600 {
			t.Errorf("file perm = %o, want 0600", perm)
		}
	}

	// Concurrent atomic saves should not race on a fixed .tmp filename or leave temp files behind.
	// Skip parallel rename stress on Windows where renaming over a concurrently read file can fail with sharing violation.
	if runtime.GOOS != "windows" {
		var wg sync.WaitGroup
		errCh := make(chan error, 10)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if err := Save(path, Settings{Port: port, Language: "en-US"}); err != nil {
					errCh <- err
				}
			}(7655 + i)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			t.Fatalf("concurrent Save failed: %v", err)
		}
	} else {
		if err := Save(path, Settings{Port: 7656, Language: "en-US"}); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// Resulting file must still be valid JSON and no temp files left behind.
	if _, err := Load(path); err != nil {
		t.Fatalf("Load after concurrent saves failed: %v", err)
	}
	entries, err := os.ReadDir(settingsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".settings-") {
			t.Fatalf("leftover temp file found: %s", e.Name())
		}
	}
}

func TestEnsureRelayKeyLazilyCreatesWithoutTouchingToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	raw := []byte(`{"token":"existing-lan-token","port":7655,"futureFlag":true}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	first, err := EnsureRelayKey(path)
	if err != nil {
		t.Fatalf("EnsureRelayKey: %v", err)
	}
	if first.Token != "existing-lan-token" {
		t.Fatalf("token modified: got %q, want %q", first.Token, "existing-lan-token")
	}
	if _, err := e2e.ParseKey(first.RelayKey); err != nil {
		t.Fatalf("invalid relayKey %q: %v", first.RelayKey, err)
	}

	second, err := EnsureRelayKey(path)
	if err != nil {
		t.Fatalf("second EnsureRelayKey: %v", err)
	}
	if second.RelayKey != first.RelayKey {
		t.Fatalf("relayKey rotated unexpectedly: %q vs %q", second.RelayKey, first.RelayKey)
	}

	// Verify unknown fields on disk are preserved.
	diskBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rawMap map[string]any
	if err := json.Unmarshal(diskBytes, &rawMap); err != nil {
		t.Fatal(err)
	}
	if rawMap["futureFlag"] != true {
		t.Fatalf("unknown field futureFlag lost: %v", rawMap)
	}
}

func TestRotateTokenRotatesBothTokenAndRelayKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	s1, err := EnsureToken(path)
	if err != nil {
		t.Fatal(err)
	}
	s1, err = EnsureRelayKey(path)
	if err != nil {
		t.Fatal(err)
	}

	s2, err := RotateToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if s2.Token == "" || s2.Token == s1.Token {
		t.Fatalf("token not rotated: %q vs %q", s2.Token, s1.Token)
	}
	if s2.RelayKey == "" || s2.RelayKey == s1.RelayKey {
		t.Fatalf("relayKey not rotated: %q vs %q", s2.RelayKey, s1.RelayKey)
	}
	if _, err := e2e.ParseKey(s2.RelayKey); err != nil {
		t.Fatalf("rotated relayKey invalid: %v", err)
	}
}

func TestSavePreservesRelayKeyAndUnknownFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"token":"tok1","unknownKey":"keep-me"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := EnsureRelayKey(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := Save(path, Settings{
		Port:           7656,
		Language:       "en-US",
		ConnectionMode: ModeRelay,
		RelayURL:       "wss://relay.example.com/ws",
		RelayAuthKey:   "secret-auth",
	}); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Token != "tok1" {
		t.Fatalf("token = %q, want tok1", loaded.Token)
	}
	if loaded.RelayKey != s.RelayKey {
		t.Fatalf("relayKey lost: %q vs %q", loaded.RelayKey, s.RelayKey)
	}
	if loaded.ConnectionMode != ModeRelay || loaded.RelayURL != "wss://relay.example.com/ws" || loaded.RelayAuthKey != "secret-auth" {
		t.Fatalf("relay settings not saved: %+v", loaded)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["unknownKey"] != "keep-me" {
		t.Fatalf("unknownKey lost: %v", m)
	}
}

func TestWatchCredentialsDetectsRelayKeyChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	s, err := EnsureToken(path)
	if err != nil {
		t.Fatal(err)
	}
	s, err = EnsureRelayKey(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := make(chan Settings, 2)
	go WatchCredentials(ctx, path, 20*time.Millisecond, s.Token, s.RelayKey, func(updated Settings) {
		got <- updated
	})

	rotated, err := RotateToken(path)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case u := <-got:
		if u.Token != rotated.Token || u.RelayKey != rotated.RelayKey {
			t.Fatalf("got %+v, want token=%q relayKey=%q", u, rotated.Token, rotated.RelayKey)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for credential rotation")
	}
}

func TestHelpersNormalizeAndResolve(t *testing.T) {
	if got := NormalizeConnectionMode("RELAY"); got != ModeRelay {
		t.Fatalf("NormalizeConnectionMode(RELAY) = %q", got)
	}
	if got := NormalizeConnectionMode(""); got != ModeLAN {
		t.Fatalf("NormalizeConnectionMode('') = %q", got)
	}
	if got := ResolveRelayURL("", ""); got != DefaultRelayURL {
		t.Fatalf("ResolveRelayURL default = %q", got)
	}
	if got := ResolveRelayURL("", "wss://saved.example/ws"); got != "wss://saved.example/ws" {
		t.Fatalf("ResolveRelayURL saved = %q", got)
	}
	if got := ResolveRelayURL("wss://flag.example/ws", "wss://saved.example/ws"); got != "wss://flag.example/ws" {
		t.Fatalf("ResolveRelayURL override = %q", got)
	}
}
