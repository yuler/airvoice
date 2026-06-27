package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
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
	ctx     context.Context
	server  *server.Server
	history *HistoryStore
	token   string
	port    int
	mu      sync.RWMutex
	status  ConnectionStatus
}

func NewApp() *App {
	history, err := NewHistoryStore("airvoice.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize history store: %v\n", err)
	}
	return &App{
		token:   uuid.New().String(),
		port:    7383,
		history: history,
		status:  ConnectionStatus{State: "disconnected"},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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
	a.mu.Unlock()

	srv := server.New(server.Config{
		Addr:     fmt.Sprintf(":%d", port),
		Port:     port,
		Hostname: getLocalHostname(),
	})

	srv.SetToken(a.token)
	a.server = srv

	go srv.ListenAndServe()

	a.mu.Lock()
	a.status = ConnectionStatus{
		State: "waiting",
		Port:  a.port,
	}
	a.mu.Unlock()

	return nil
}

func (a *App) StopServer() error {
	a.server = nil
	a.mu.Lock()
	a.status = ConnectionStatus{State: "disconnected"}
	a.mu.Unlock()
	return nil
}

func (a *App) GetSettings() Settings {
	return Settings{
		Port:     a.port,
		Language: "zh-CN",
	}
}

func (a *App) SaveSettings(s Settings) error {
	a.port = s.Port
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

func getLocalHostname() string {
	name, _ := os.Hostname()
	return name
}
