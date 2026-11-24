package upgrade

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestRunUpgradeEndToEnd(t *testing.T) {
	newPayload := []byte("#!/bin/sh\necho airvoice 0.4.8\n")
	sum := sha256Hex(newPayload)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			w.Header().Set("Location", "https://github.com/yuler/airvoice/releases/tag/v0.4.8")
			w.WriteHeader(http.StatusFound)
		case "/releases/download/v0.4.8/checksums-cli.txt":
			fmt.Fprintf(w, "%s  airvoice-cli-linux-amd64\n", sum)
		case "/releases/download/v0.4.8/airvoice-cli-linux-amd64":
			_, _ = w.Write(newPayload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	binPath := filepath.Join(dir, "airvoice")
	if err := os.WriteFile(binPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(dir, "update-check.json")

	var out bytes.Buffer
	err := Run(context.Background(), Options{
		CurrentVersion: "0.4.7",
		ExecPath:       binPath,
		CachePath:      cachePath,
		GOOS:           "linux",
		GOARCH:         "amd64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Out:            &out,
	})
	if err != nil {
		t.Fatalf("Run err: %v", err)
	}

	gotBytes, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotBytes, newPayload) {
		t.Fatalf("binary content = %q, want %q", gotBytes, newPayload)
	}
	if !strings.Contains(out.String(), "Upgraded airvoice: v0.4.7 -> v0.4.8") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}

	cached, err := LoadCache(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if cached.LatestVersion != "0.4.8" {
		t.Fatalf("cached version = %q, want 0.4.8", cached.LatestVersion)
	}
}

func TestRunUpgradeChecksumMismatchLeavesOriginalUntouched(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/latest":
			w.Header().Set("Location", "/yuler/airvoice/releases/tag/v0.4.8")
			w.WriteHeader(http.StatusFound)
		case "/releases/download/v0.4.8/checksums-cli.txt":
			fmt.Fprintf(w, "%s  airvoice-cli-linux-amd64\n", strings.Repeat("a", 64))
		case "/releases/download/v0.4.8/airvoice-cli-linux-amd64":
			_, _ = w.Write([]byte("tampered-binary"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	binPath := filepath.Join(dir, "airvoice")
	if err := os.WriteFile(binPath, []byte("original-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := Run(context.Background(), Options{
		CurrentVersion: "0.4.7",
		ExecPath:       binPath,
		CachePath:      filepath.Join(dir, "update-check.json"),
		GOOS:           "linux",
		GOARCH:         "amd64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Out:            &out,
	})
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}

	gotBytes, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotBytes) != "original-binary" {
		t.Fatalf("original binary was modified: %q", gotBytes)
	}

	// Ensure no leftover temp files remain in dir.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".airvoice-upgrade-") {
			t.Fatalf("leftover temp file %q", e.Name())
		}
	}
}

func TestRunUpgradeCheckOnlyAndHomebrewGuard(t *testing.T) {
	var downloads atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/latest" {
			w.Header().Set("Location", "/yuler/airvoice/releases/tag/v0.4.8")
			w.WriteHeader(http.StatusFound)
			return
		}
		downloads.Add(1)
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	client := &Client{BaseURL: srv.URL, HTTPClient: srv.Client()}

	// --check with update available on Homebrew install.
	var out bytes.Buffer
	err := Run(context.Background(), Options{
		CurrentVersion: "0.4.7",
		CheckOnly:      true,
		ExecPath:       "/opt/homebrew/Cellar/airvoice/0.4.7/bin/airvoice",
		CachePath:      cachePath,
		Client:         client,
		Out:            &out,
	})
	if err != nil {
		t.Fatalf("check-only err: %v", err)
	}
	if !strings.Contains(out.String(), "Update available: v0.4.7 -> v0.4.8") ||
		!strings.Contains(out.String(), "brew upgrade airvoice") {
		t.Fatalf("unexpected check-only output:\n%s", out.String())
	}

	// Full upgrade on Homebrew install refuses direct overwrite.
	out.Reset()
	err = Run(context.Background(), Options{
		CurrentVersion: "0.4.7",
		ExecPath:       "/opt/homebrew/Cellar/airvoice/0.4.7/bin/airvoice",
		CachePath:      cachePath,
		Client:         client,
		Out:            &out,
	})
	if err == nil || !strings.Contains(err.Error(), "brew upgrade airvoice") {
		t.Fatalf("expected Homebrew guard error, got %v", err)
	}

	// Already up to date skips download.
	out.Reset()
	err = Run(context.Background(), Options{
		CurrentVersion: "0.4.8",
		ExecPath:       filepath.Join(t.TempDir(), "airvoice"),
		CachePath:      cachePath,
		Client:         client,
		Out:            &out,
	})
	if err != nil {
		t.Fatalf("up-to-date err: %v", err)
	}
	if !strings.Contains(out.String(), "already up to date") {
		t.Fatalf("unexpected up-to-date output:\n%s", out.String())
	}
	if downloads.Load() != 0 {
		t.Fatalf("expected 0 downloads, got %d", downloads.Load())
	}
}

func TestRunUpgradeExplicitVersionRollback(t *testing.T) {
	payload := []byte("v0.4.6-binary")
	sum := sha256Hex(payload)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/releases/download/v0.4.6/checksums-cli.txt":
			fmt.Fprintf(w, "%s *airvoice-cli-darwin-arm64\n", sum)
		case "/releases/download/v0.4.6/airvoice-cli-darwin-arm64":
			_, _ = w.Write(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	binPath := filepath.Join(dir, "airvoice")
	if err := os.WriteFile(binPath, []byte("v0.4.8-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := Run(context.Background(), Options{
		CurrentVersion: "0.4.8",
		TargetVersion:  "v0.4.6",
		ExecPath:       binPath,
		CachePath:      filepath.Join(dir, "update-check.json"),
		GOOS:           "darwin",
		GOARCH:         "arm64",
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Out:            &out,
	})
	if err != nil {
		t.Fatalf("rollback err: %v", err)
	}
	if !strings.Contains(out.String(), "Downgraded airvoice: v0.4.8 -> v0.4.6") {
		t.Fatalf("unexpected rollback output:\n%s", out.String())
	}
}

func TestReplaceExecutableWindowsFlowAndCleanup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "airvoice.exe")
	newBin := filepath.Join(dir, ".airvoice-upgrade-123.tmp")

	if err := os.WriteFile(target, []byte("old-exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBin, []byte("new-exe"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceExecutableForOS("windows", target, newBin); err != nil {
		t.Fatalf("replaceExecutableForOS(windows) err: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new-exe" {
		t.Fatalf("target = %q, want new-exe", got)
	}

	// Simulate leftover .old file and verify CleanupStaleBackup removes it.
	oldPath := target + ".old"
	if err := os.WriteFile(oldPath, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	CleanupStaleBackup(target)
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("expected %q to be removed, err=%v", oldPath, err)
	}
}
