package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yuler/airvoice/cli/pairing"
	"github.com/yuler/airvoice/cli/paste"
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
	paster       paste.Paster
	token        string
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

	settingsPath := filepath.Join(homeDir, ".airvoice", "settings.json")

	app := &App{
		history:      history,
		token:        uuid.New().String(),
		port:         7655,
		settingsPath: settingsPath,
		settings:     Settings{Port: 7655, Language: "zh-CN"},
		status:       ConnectionStatus{State: "disconnected"},
	}

	app.loadSettings()

	return app, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.tray != nil {
		a.tray.Start()
	}
	server.LogHook = func(msg string) {
		runtime.EventsEmit(a.ctx, "log_added", msg)
	}
	if err := a.StartServer(a.port); err != nil {
		log.Printf("Failed to start server: %v", err)
		runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Server Error",
			Message: fmt.Sprintf("Failed to start server: %v\n\nPlease check your settings and port occupancy.", err),
		})
	}
}

func (a *App) GetPairingLink() (string, error) {
	ip, err := pairing.LocalIPv4()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN IP: %w", err)
	}

	a.mu.RLock()
	port := a.port
	token := a.token
	a.mu.RUnlock()

	payload := pairing.Payload{
		Version: 1,
		WS:      fmt.Sprintf("ws://%s:%d/ws", ip, port),
		Token:   token,
	}

	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	return string(payloadBytes), nil
}

// RefreshPairing generates a new pairing token and invalidates pending connections.
func (a *App) RefreshPairing() error {
	a.mu.Lock()
	if a.status.State == "connected" {
		a.mu.Unlock()
		return fmt.Errorf("cannot refresh pairing while connected")
	}
	a.token = uuid.New().String()
	srv := a.server
	a.mu.Unlock()

	if srv != nil {
		srv.SetToken(a.token)
		srv.DisconnectClients()
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "log_added", " [airvoice] pairing token refreshed")
	}

	return nil
}

func (a *App) GetQRCode() (string, error) {
	ip, err := pairing.LocalIPv4()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN IP: %w", err)
	}

	a.mu.RLock()
	port := a.port
	a.mu.RUnlock()

	payload := pairing.Payload{
		Version: 1,
		WS:      fmt.Sprintf("ws://%s:%d/ws", ip, port),
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

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(code.PNG()), nil
}

func (a *App) GetConnectionStatus() ConnectionStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.status
}

func (a *App) StartServer(port int) error {
	if err := server.CheckPortAvailable(port); err != nil {
		return err
	}

	a.mu.Lock()
	if a.server != nil {
		a.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	a.port = port
	a.status = ConnectionStatus{
		State: "waiting",
		Port:  port,
	}
	if a.tray != nil {
		a.tray.UpdateStatus()
	}
	a.mu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "log_added", fmt.Sprintf(" [airvoice] listening on :%d (health: /health, ws: /ws)", port))
	}

	if a.paster == nil {
		paster, err := paste.New()
		if err != nil {
			return fmt.Errorf("failed to initialize paster: %w", err)
		}
		a.paster = paster
	}

	srv := server.New(server.Config{
		Addr:     fmt.Sprintf(":%d", port),
		Port:     port,
		Hostname: getLocalHostname(),
		Paster:   a.paster,
		OnTextReceived: func(content, device string) {
			if a.history != nil {
				a.history.Add(content, device)
			}
		},
		OnConnect: func(device string) {
			a.mu.Lock()
			host := ""
			if ip, err := pairing.LocalIPv4(); err == nil {
				host = ip
			}
			a.status = ConnectionStatus{
				State:      "connected",
				DeviceName: device,
				Host:       host,
				Port:       a.port,
			}
			status := a.status
			a.mu.Unlock()
			if a.tray != nil {
				a.tray.UpdateStatus()
			}
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "status_changed", status)
			}
		},
		OnDisconnect: func() {
			a.mu.Lock()
			a.status = ConnectionStatus{
				State: "waiting",
				Port:  a.port,
			}
			status := a.status
			a.mu.Unlock()
			if a.tray != nil {
				a.tray.UpdateStatus()
			}
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "status_changed", status)
			}
		},
	})

	srv.SetToken(a.token)

	a.mu.Lock()
	a.server = srv
	a.mu.Unlock()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.mu.Lock()
			a.status = ConnectionStatus{
				State: "disconnected",
				Port:  a.port,
			}
			status := a.status
			a.mu.Unlock()
			if a.ctx != nil {
				runtime.EventsEmit(a.ctx, "status_changed", status)
			}
		}
	}()

	return nil
}

func (a *App) StopServer() error {
	a.mu.Lock()
	srv := a.server
	a.server = nil
	a.status = ConnectionStatus{State: "disconnected"}
	status := a.status
	a.mu.Unlock()
	if a.tray != nil {
		a.tray.UpdateStatus()
	}

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "status_changed", status)
	}

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
	if s.Port < 1024 || s.Port > 65535 {
		s.Port = 7655
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
	if s.Port < 1024 || s.Port > 65535 {
		return fmt.Errorf("invalid port: must be between 1024 and 65535")
	}

	a.mu.RLock()
	portChanged := a.port != s.Port
	a.mu.RUnlock()

	if portChanged {
		if err := server.CheckPortAvailable(s.Port); err != nil {
			return fmt.Errorf("port %d is already in use: %w", s.Port, err)
		}
	}

	a.mu.Lock()
	a.settings = s
	a.port = s.Port
	a.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(a.settingsPath), 0700); err != nil {
		return fmt.Errorf("failed to create settings dir: %w", err)
	}
	if err := os.WriteFile(a.settingsPath, data, 0600); err != nil {
		return err
	}

	if portChanged && a.ctx != nil {
		_ = a.StopServer()
		_ = a.StartServer(s.Port)
		runtime.EventsEmit(a.ctx, "server_restarted")
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
