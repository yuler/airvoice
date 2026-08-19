package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
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
