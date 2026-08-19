package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Settings is the on-disk ~/.airvoice/settings.json document.
type Settings struct {
	Token     string `json:"token,omitempty"`
	Port      int    `json:"port,omitempty"`
	AutoStart bool   `json:"autoStart,omitempty"`
	Language  string `json:"language,omitempty"`
}

// Path returns $HOME/.airvoice/settings.json.
func Path() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	return filepath.Join(home, ".airvoice", "settings.json")
}

// Load reads settings. A missing file returns a zero Settings and nil error.
func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Settings{}, nil
	}
	if err != nil {
		return Settings{}, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("parse settings: %w", err)
	}
	return s, nil
}

// EnsureToken returns persisted settings, minting and writing a token if needed.
func EnsureToken(path string) (Settings, error) {
	s, err := Load(path)
	if err != nil {
		s = Settings{}
	}
	if strings.TrimSpace(s.Token) != "" {
		s.Token = strings.TrimSpace(s.Token)
		return s, nil
	}
	s.Token = uuid.NewString()
	if err := writeFile(path, s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// RotateToken replaces the pairing token and preserves other fields.
func RotateToken(path string) (Settings, error) {
	s, err := EnsureToken(path)
	if err != nil {
		return Settings{}, err
	}
	s.Token = uuid.NewString()
	if err := writeFile(path, s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// Save writes settings. An empty incoming token keeps the existing token (or mints one).
func Save(path string, incoming Settings) error {
	existing, err := Load(path)
	if err != nil {
		existing = Settings{}
	}
	if strings.TrimSpace(incoming.Token) == "" {
		incoming.Token = existing.Token
	}
	if strings.TrimSpace(incoming.Token) == "" {
		incoming.Token = uuid.NewString()
	}
	return writeFile(path, incoming)
}

// WatchToken polls path and calls onChange when the token string changes.
func WatchToken(ctx context.Context, path string, interval time.Duration, current string, onChange func(token string)) {
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s, err := Load(path)
			if err != nil {
				continue
			}
			tok := strings.TrimSpace(s.Token)
			if tok == "" || tok == current {
				continue
			}
			current = tok
			onChange(tok)
		}
	}
}

func writeFile(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
