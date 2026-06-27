package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/server"
	qr "rsc.io/qr"
)

type ConnectionStatus struct {
	State      string `json:"state"`
	DeviceName string `json:"deviceName"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
}

type Settings struct {
	Port      int    `json:"port"`
	AutoStart bool   `json:"autoStart"`
	Language  string `json:"language"`
}

type App struct {
	ctx          context.Context
	server       *server.Server
	history      *HistoryStore
	token        string
	port         int
	settingsPath string
	settings     Settings
	mu           sync.RWMutex
	status       ConnectionStatus
}

func NewApp() (*App, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	dbPath := filepath.Join(homeDir, ".airvoice", "history.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	history, err := NewHistoryStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize history store: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".airvoice", "settings.json")

	app := &App{
		history:      history,
		token:        uuid.New().String(),
		port:         7383,
		settingsPath: settingsPath,
		settings:     Settings{Port: 7383, Language: "zh-CN"},
		status:       ConnectionStatus{State: "disconnected"},
	}

	app.loadSettings()

	return app, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetPairingLink() (string, error) {
	ip, err := pairing.LocalIPv4()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN IP: %w", err)
	}

	payload := pairing.Payload{
		Version: 1,
		WS:      fmt.Sprintf("ws://%s:%d/ws", ip, a.port),
		Token:   a.token,
	}

	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	return string(payloadBytes), nil
}

func (a *App) GetQRCode() (string, error) {
	ip, err := pairing.LocalIPv4()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN IP: %w", err)
	}

	payload := pairing.Payload{
		Version: 1,
		WS:      fmt.Sprintf("ws://%s:%d/ws", ip, a.port),
		Token:   a.token,
	}

	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	code, err := qr.Encode(string(payloadBytes), qr.M)
	if err != nil {
		return "", fmt.Errorf("failed to encode QR: %w", err)
	}

	img := code.Image()

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (a *App) GetConnectionStatus() ConnectionStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

func (a *App) StartServer(port int) error {
	a.mu.Lock()
	a.port = port
	a.status = ConnectionStatus{
		State: "waiting",
		Port:  port,
	}
	a.mu.Unlock()

	srv := server.New(server.Config{
		Addr:     fmt.Sprintf(":%d", port),
		Port:     port,
		Hostname: getLocalHostname(),
		OnTextReceived: func(content, device string) {
			if a.history != nil {
				a.history.Add(content, device)
			}
		},
		OnConnect: func(device string) {
			a.mu.Lock()
			a.status = ConnectionStatus{
				State:      "connected",
				DeviceName: device,
				Port:       a.port,
			}
			a.mu.Unlock()
		},
		OnDisconnect: func() {
			a.mu.Lock()
			a.status = ConnectionStatus{
				State: "waiting",
				Port:  a.port,
			}
			a.mu.Unlock()
		},
	})

	srv.SetToken(a.token)

	a.mu.Lock()
	a.server = srv
	a.mu.Unlock()

	go srv.ListenAndServe()

	return nil
}

func (a *App) StopServer() error {
	a.mu.Lock()
	srv := a.server
	a.server = nil
	a.status = ConnectionStatus{State: "disconnected"}
	a.mu.Unlock()

	if srv != nil {
		return srv.Close()
	}
	return nil
}

func (a *App) loadSettings() {
	data, err := os.ReadFile(a.settingsPath)
	if err != nil {
		return
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return
	}
	a.settings = s
	a.port = s.Port
}

func (a *App) GetSettings() Settings {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.settings
}

func (a *App) SaveSettings(s Settings) error {
	a.mu.Lock()
	a.settings = s
	a.port = s.Port
	a.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(a.settingsPath), 0755); err != nil {
		return fmt.Errorf("failed to create settings dir: %w", err)
	}
	return os.WriteFile(a.settingsPath, data, 0644)
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
