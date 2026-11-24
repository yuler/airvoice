package upgrade

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry is persisted at ~/.airvoice/update-check.json.
type CacheEntry struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
}

// CachePath returns $HOME/.airvoice/update-check.json.
func CachePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".airvoice", "update-check.json")
}

// LoadCache reads the update-check cache. A missing file returns a zero entry and nil error.
func LoadCache(path string) (CacheEntry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return CacheEntry{}, nil
	}
	if err != nil {
		return CacheEntry{}, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CacheEntry{}, fmt.Errorf("parse update cache: %w", err)
	}
	entry.LatestVersion = Normalize(entry.LatestVersion)
	return entry, nil
}

// SaveCache atomically writes the update-check cache file.
func SaveCache(path string, entry CacheEntry) error {
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	entry.LatestVersion = Normalize(entry.LatestVersion)
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, ".update-check-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	_ = os.Chmod(tmpPath, 0o600)
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}
	return nil
}

// IsFresh reports whether entry was checked within ttl relative to now.
func IsFresh(entry CacheEntry, now time.Time, ttl time.Duration) bool {
	if entry.CheckedAt.IsZero() || entry.LatestVersion == "" || ttl <= 0 {
		return false
	}
	age := now.Sub(entry.CheckedAt)
	return age >= 0 && age < ttl
}
