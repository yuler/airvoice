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
	"github.com/yuler/airvoice/cli/e2e"
)

const (
	ModeLAN         = "lan"
	ModeRelay       = "relay"
	DefaultRelayURL = "wss://relay.airvoice.yuler.cc/ws"
)

// Settings is the on-disk ~/.airvoice/settings.json document.
type Settings struct {
	Token          string `json:"token,omitempty"`
	RelayKey       string `json:"relayKey,omitempty"`
	Port           int    `json:"port,omitempty"`
	AutoStart      bool   `json:"autoStart,omitempty"`
	Language       string `json:"language,omitempty"`
	ConnectionMode string `json:"connectionMode,omitempty"`
	RelayURL       string `json:"relayUrl,omitempty"`
	RelayAuthKey   string `json:"relayAuthKey,omitempty"`
}

var knownSettingKeys = []string{
	"token",
	"relayKey",
	"port",
	"autoStart",
	"language",
	"connectionMode",
	"relayUrl",
	"relayAuthKey",
}

// NormalizeConnectionMode returns ModeRelay if mode is "relay" (case-insensitive),
// otherwise ModeLAN.
func NormalizeConnectionMode(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), ModeRelay) {
		return ModeRelay
	}
	return ModeLAN
}

// ResolveRelayURL applies precedence: override -> saved -> DefaultRelayURL.
func ResolveRelayURL(override, saved string) string {
	if u := strings.TrimSpace(override); u != "" {
		return u
	}
	if u := strings.TrimSpace(saved); u != "" {
		return u
	}
	return DefaultRelayURL
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
		return Settings{}, err
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

// EnsureRelayKey returns persisted settings, lazily minting and persisting a
// 32-byte base64url relayKey without modifying Token.
func EnsureRelayKey(path string) (Settings, error) {
	s, err := Load(path)
	if err != nil {
		return Settings{}, err
	}
	trimmed := strings.TrimSpace(s.RelayKey)
	if trimmed != "" {
		if _, err := e2e.ParseKey(trimmed); err == nil {
			s.RelayKey = trimmed
			return s, nil
		}
	}
	key, err := e2e.GenerateKey()
	if err != nil {
		return Settings{}, err
	}
	s.RelayKey = key
	if err := writeFile(path, s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// RotateToken replaces both the LAN pairing token and the E2E relayKey,
// preserving all other settings fields.
func RotateToken(path string) (Settings, error) {
	s, err := Load(path)
	if err != nil {
		return Settings{}, err
	}
	key, err := e2e.GenerateKey()
	if err != nil {
		return Settings{}, err
	}
	s.Token = uuid.NewString()
	s.RelayKey = key
	if err := writeFile(path, s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

// Save writes settings. Empty incoming Token or RelayKey preserves existing
// persisted credentials (minting a Token if none exists yet).
func Save(path string, incoming Settings) error {
	existing, err := Load(path)
	if err != nil {
		return err
	}
	if strings.TrimSpace(incoming.Token) == "" {
		incoming.Token = existing.Token
	}
	if strings.TrimSpace(incoming.Token) == "" {
		incoming.Token = uuid.NewString()
	}
	if strings.TrimSpace(incoming.RelayKey) == "" {
		incoming.RelayKey = existing.RelayKey
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

// WatchCredentials polls path and calls onChange whenever either Token or
// RelayKey changes externally (e.g. via `airvoice token refresh`).
func WatchCredentials(ctx context.Context, path string, interval time.Duration, currentToken, currentRelayKey string, onChange func(Settings)) {
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
			rk := strings.TrimSpace(s.RelayKey)

			tokChanged := tok != "" && tok != currentToken
			rkChanged := rk != "" && rk != currentRelayKey
			if !tokChanged && !rkChanged {
				continue
			}
			if tokChanged {
				currentToken = tok
				s.Token = tok
			}
			if rkChanged {
				currentRelayKey = rk
				s.RelayKey = rk
			}
			onChange(s)
		}
	}
}

func writeFile(path string, s Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	merged, err := mergeWithExisting(path, s)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmpFile, err := os.CreateTemp(dir, ".settings-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

// mergeWithExisting overlays Settings onto any existing JSON object on disk,
// preserving unknown fields from future versions cleanly.
func mergeWithExisting(path string, s Settings) (map[string]any, error) {
	merged := make(map[string]any)
	if raw, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(raw, &merged); err != nil {
			return nil, fmt.Errorf("parse settings: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	encoded, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var known map[string]any
	if err := json.Unmarshal(encoded, &known); err != nil {
		return nil, err
	}

	for _, k := range knownSettingKeys {
		delete(merged, k)
	}
	for k, v := range known {
		merged[k] = v
	}
	return merged, nil
}
