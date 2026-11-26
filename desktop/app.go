package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuler/airvoice/cli/config"
	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/relayclient"
	"github.com/yuler/airvoice/cli/server"
	qr "rsc.io/qr"
)

type ConnectionStatus struct {
	State      string `json:"state"`
	Mode       string `json:"mode"`
	DeviceName string `json:"deviceName"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Error      string `json:"error,omitempty"`
}

type Settings struct {
	Token          string `json:"token,omitempty"`
	RelayKey       string `json:"relayKey,omitempty"`
	Port           int    `json:"port"`
	AutoStart      bool   `json:"autoStart"`
	Language       string `json:"language"`
	ConnectionMode string `json:"connectionMode"`
	RelayURL       string `json:"relayUrl"`
	RelayAuthKey   string `json:"relayAuthKey"`
}

type App struct {
	ctx          context.Context
	server       *server.Server
	relayClient  *relayclient.Client
	history      *HistoryStore
	paster       paste.Paster
	token        string
	relayKey     string
	port         int
	settingsPath string
	settings     Settings
	mu           sync.RWMutex
	status       ConnectionStatus
	tray         *TrayManager
}

func NewApp() (*App, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	dbPath := filepath.Join(homeDir, ".airvoice", "history.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	history, err := NewHistoryStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize history store: %w", err)
	}

	settingsPath := config.Path()

	app := &App{
		history:      history,
		port:         7655,
		settingsPath: settingsPath,
		settings: Settings{
			Port:           7655,
			Language:       "zh-CN",
			ConnectionMode: config.ModeLAN,
		},
		status: ConnectionStatus{
			State: "disconnected",
			Mode:  config.ModeLAN,
		},
	}

	if err := app.loadSettings(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) getContext() context.Context {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ctx
}

func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	token := a.token
	relayKey := a.relayKey
	path := a.settingsPath
	a.mu.Unlock()

	if a.tray != nil {
		a.tray.Start()
	}
	server.SetLogHook(func(msg string) {
		if c := a.getContext(); c != nil {
			runtime.EventsEmit(c, "log_added", msg)
		}
	})
	if err := a.startActiveTransport(); err != nil {
		log.Printf("Failed to start transport: %v", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Server Error",
			Message: fmt.Sprintf("Failed to start Airvoice: %v\n\nPlease check your settings.", err),
		})
	}

	go config.WatchCredentials(ctx, path, time.Second, token, relayKey, func(updated config.Settings) {
		a.applyCredentials(updated.Token, updated.RelayKey, true)
	})
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.StopServer()
}

// GetPairingLink returns an airvoice://pair?... URI in Relay mode, or the v1 JSON payload in LAN mode.
func (a *App) GetPairingLink() (string, error) {
	a.mu.RLock()
	mode := config.NormalizeConnectionMode(a.settings.ConnectionMode)
	port := a.port
	token := a.token
	relayURL := config.ResolveRelayURL(a.settings.RelayURL, "")
	relayAuthKey := strings.TrimSpace(a.settings.RelayAuthKey)
	a.mu.RUnlock()

	if mode == config.ModeRelay {
		relayKey, err := a.ensureRelayKey()
		if err != nil {
			return "", err
		}
		return pairing.BuildPairingURI(relayURL, relayKey, relayAuthKey)
	}

	ip, err := pairing.LocalIPv4()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN IP: %w", err)
	}

	payload := pairing.Payload{
		Version: pairing.VersionLAN,
		WS:      fmt.Sprintf("ws://%s:%d/ws", ip, port),
		Token:   token,
	}
	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	return string(payloadBytes), nil
}

// RefreshPairing rotates both the LAN pairing token and the E2E relayKey,
// persists them, and disconnects/reconnects active sessions.
func (a *App) RefreshPairing() error {
	s, err := config.RotateToken(a.settingsPath)
	if err != nil {
		return err
	}
	a.applyCredentials(s.Token, s.RelayKey, true)
	return nil
}

func (a *App) GetQRCode() (string, error) {
	a.mu.RLock()
	mode := config.NormalizeConnectionMode(a.settings.ConnectionMode)
	port := a.port
	token := a.token
	relayURL := config.ResolveRelayURL(a.settings.RelayURL, "")
	relayAuthKey := strings.TrimSpace(a.settings.RelayAuthKey)
	a.mu.RUnlock()

	var payloadBytes []byte
	var err error

	if mode == config.ModeRelay {
		relayKey, keyErr := a.ensureRelayKey()
		if keyErr != nil {
			return "", keyErr
		}
		payload := pairing.PayloadV2{
			Version: pairing.VersionRelay,
			Mode:    pairing.ModeRelay,
			WS:      relayURL,
			Key:     relayKey,
			Auth:    relayAuthKey,
		}
		payloadBytes, err = payload.Marshal()
	} else {
		ip, ipErr := pairing.LocalIPv4()
		if ipErr != nil {
			return "", fmt.Errorf("failed to get LAN IP: %w", ipErr)
		}
		payload := pairing.Payload{
			Version: pairing.VersionLAN,
			WS:      fmt.Sprintf("ws://%s:%d/ws", ip, port),
			Token:   token,
		}
		payloadBytes, err = payload.Marshal()
	}
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	code, err := qr.Encode(string(payloadBytes), qr.M)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(code.PNG()), nil
}

func (a *App) GetConnectionStatus() ConnectionStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

func (a *App) loadSettings() error {
	s, err := config.EnsureToken(a.settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}
	mode := config.NormalizeConnectionMode(s.ConnectionMode)
	if mode == config.ModeRelay && strings.TrimSpace(s.RelayKey) == "" {
		s, err = config.EnsureRelayKey(a.settingsPath)
		if err != nil {
			return fmt.Errorf("failed to ensure relay key: %w", err)
		}
	}
	port := s.Port
	if port < 1024 || port > 65535 {
		port = 7655
	}
	lang := s.Language
	if lang == "" {
		lang = "zh-CN"
	}
	a.settings = Settings{
		Token:          s.Token,
		RelayKey:       s.RelayKey,
		Port:           port,
		AutoStart:      s.AutoStart,
		Language:       lang,
		ConnectionMode: mode,
		RelayURL:       strings.TrimSpace(s.RelayURL),
		RelayAuthKey:   strings.TrimSpace(s.RelayAuthKey),
	}
	a.token = s.Token
	a.relayKey = s.RelayKey
	a.port = port
	a.status.Mode = mode
	return nil
}

func (a *App) ensureRelayKey() (string, error) {
	a.mu.RLock()
	existing := strings.TrimSpace(a.relayKey)
	a.mu.RUnlock()
	if existing != "" {
		return existing, nil
	}
	s, err := config.EnsureRelayKey(a.settingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate relay key: %w", err)
	}
	a.mu.Lock()
	a.relayKey = s.RelayKey
	a.settings.RelayKey = s.RelayKey
	a.mu.Unlock()
	return s.RelayKey, nil
}

func (a *App) GetSettings() Settings {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.settings
}

func (a *App) SaveSettings(s Settings) error {
	mode := config.NormalizeConnectionMode(s.ConnectionMode)
	if s.Port < 1024 || s.Port > 65535 {
		return fmt.Errorf("invalid port: must be between 1024 and 65535")
	}

	relayURL := strings.TrimSpace(s.RelayURL)
	relayAuthKey := strings.TrimSpace(s.RelayAuthKey)
	if mode == config.ModeRelay {
		resolved := config.ResolveRelayURL(relayURL, "")
		if err := relayclient.ValidateRelayURL(resolved); err != nil {
			return err
		}
	}

	a.mu.RLock()
	oldMode := config.NormalizeConnectionMode(a.settings.ConnectionMode)
	modeChanged := oldMode != mode
	portChanged := a.port != s.Port
	relayChanged := a.settings.RelayURL != relayURL || a.settings.RelayAuthKey != relayAuthKey
	wasRunning := a.server != nil || a.relayClient != nil
	a.mu.RUnlock()

	if mode == config.ModeLAN && (portChanged || modeChanged) {
		if err := server.CheckPortAvailable(s.Port); err != nil {
			return fmt.Errorf("port %d is already in use: %w", s.Port, err)
		}
	}

	if err := config.Save(a.settingsPath, config.Settings{
		Port:           s.Port,
		AutoStart:      s.AutoStart,
		Language:       s.Language,
		ConnectionMode: mode,
		RelayURL:       relayURL,
		RelayAuthKey:   relayAuthKey,
	}); err != nil {
		return err
	}

	if mode == config.ModeRelay {
		if _, err := config.EnsureRelayKey(a.settingsPath); err != nil {
			return err
		}
	}

	disk, err := config.Load(a.settingsPath)
	if err != nil {
		return err
	}

	lang := s.Language
	if lang == "" {
		lang = a.settings.Language
	}

	a.mu.Lock()
	oldToken := a.token
	oldRelayKey := a.relayKey
	a.settings = Settings{
		Token:          disk.Token,
		RelayKey:       disk.RelayKey,
		Port:           s.Port,
		AutoStart:      s.AutoStart,
		Language:       lang,
		ConnectionMode: mode,
		RelayURL:       relayURL,
		RelayAuthKey:   relayAuthKey,
	}
	a.port = s.Port
	a.token = disk.Token
	a.relayKey = disk.RelayKey
	srv := a.server
	rc := a.relayClient
	ctx := a.ctx
	a.mu.Unlock()

	transportChanged := modeChanged || (mode == config.ModeLAN && portChanged) || (mode == config.ModeRelay && relayChanged)
	if transportChanged && (ctx != nil || wasRunning) {
		_ = a.StopServer()
		if err := a.startActiveTransport(); err != nil {
			return err
		}
		if ctx != nil {
			runtime.EventsEmit(ctx, "server_restarted")
			runtime.EventsEmit(ctx, "pairing_changed")
		}
		return nil
	}

	if srv != nil && disk.Token != "" && disk.Token != oldToken {
		srv.SetToken(disk.Token)
	}
	if rc != nil && disk.RelayKey != "" && disk.RelayKey != oldRelayKey {
		_ = rc.UpdateRelayKey(disk.RelayKey)
	}
	if ctx != nil && (modeChanged || relayChanged) {
		runtime.EventsEmit(ctx, "pairing_changed")
	}

	return nil
}

func (a *App) GetHistory(limit int) ([]HistoryEntry, error) {
	if a.history == nil {
		return nil, fmt.Errorf("history store not initialized")
	}
	return a.history.List(limit)
}

func (a *App) ClearHistory() error {
	if a.history == nil {
		return fmt.Errorf("history store not initialized")
	}
	return a.history.Clear()
}

func (a *App) SearchHistory(query string, limit int) ([]HistoryEntry, error) {
	if a.history == nil {
		return nil, fmt.Errorf("history store not initialized")
	}
	return a.history.Search(query, limit)
}

func getLocalHostname() string {
	name, _ := os.Hostname()
	return name
}
