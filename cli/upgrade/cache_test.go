package upgrade

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheLoadSaveAndFreshness(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".airvoice", "update-check.json")

	// Missing file returns zero entry and nil error.
	entry, err := LoadCache(path)
	if err != nil {
		t.Fatalf("LoadCache missing err: %v", err)
	}
	if !entry.CheckedAt.IsZero() || entry.LatestVersion != "" {
		t.Fatalf("expected zero entry, got %+v", entry)
	}

	now := time.Date(2026, 11, 24, 17, 0, 0, 0, time.UTC)
	if err := SaveCache(path, CacheEntry{CheckedAt: now, LatestVersion: "v0.4.8"}); err != nil {
		t.Fatalf("SaveCache err: %v", err)
	}

	loaded, err := LoadCache(path)
	if err != nil {
		t.Fatalf("LoadCache err: %v", err)
	}
	if loaded.LatestVersion != "0.4.8" {
		t.Fatalf("LatestVersion = %q, want 0.4.8", loaded.LatestVersion)
	}
	if !loaded.CheckedAt.Equal(now) {
		t.Fatalf("CheckedAt = %v, want %v", loaded.CheckedAt, now)
	}

	// Freshness checks.
	if !IsFresh(loaded, now.Add(23*time.Hour), 24*time.Hour) {
		t.Fatal("expected cache to be fresh at 23h")
	}
	if IsFresh(loaded, now.Add(24*time.Hour), 24*time.Hour) {
		t.Fatal("expected cache to be stale at 24h")
	}
	if IsFresh(loaded, now.Add(-time.Minute), 24*time.Hour) {
		t.Fatal("expected negative clock skew to be treated as stale")
	}

	// Corrupt JSON returns an error.
	if err := os.WriteFile(path, []byte("{bad json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCache(path); err == nil {
		t.Fatal("expected parse error for corrupt cache")
	}
}
