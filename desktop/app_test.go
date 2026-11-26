package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/config"
	"github.com/yuler/airvoice/cli/e2e"
	"github.com/yuler/airvoice/cli/protocol"
	"github.com/yuler/airvoice/relay"
)

func TestNewApp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if app.token == "" {
		t.Error("token should not be empty")
	}
	if app.port != 7655 {
		t.Errorf("port = %d, want 7655", app.port)
	}
}

func TestGetSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, _ := NewApp()
	settings := app.GetSettings()

	if settings.Port != 7655 {
		t.Errorf("settings.Port = %d, want 7655", settings.Port)
	}
	if settings.Language != "zh-CN" {
		t.Errorf("settings.Language = %s, want zh-CN", settings.Language)
	}
}

func TestGetQRCode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, _ := NewApp()
	qrCodeStr, err := app.GetQRCode()
	if err != nil {
		t.Fatalf("GetQRCode() error = %v", err)
	}

	if qrCodeStr == "" {
		t.Fatal("GetQRCode() returned empty string")
	}

	// Verify it's a data URL
	prefix := "data:image/png;base64,"
	if len(qrCodeStr) <= len(prefix) || qrCodeStr[:len(prefix)] != prefix {
		t.Fatalf("invalid QR code data URL prefix: %s", qrCodeStr)
	}

	rawB64 := qrCodeStr[len(prefix):]
	data, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode PNG image: %v", err)
	}

	bounds := img.Bounds()

	// Analyze unique colors and assert correct structure.
	var blackCount, whiteCount int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				blackCount++
			} else if r == 0xFFFF && g == 0xFFFF && b == 0xFFFF {
				whiteCount++
			}
		}
	}

	if bounds.Dx() < 100 || bounds.Dy() < 100 {
		t.Errorf("QR code dimensions are too small: %dx%d", bounds.Dx(), bounds.Dy())
	}
	if blackCount == 0 || whiteCount == 0 {
		t.Error("QR code should contain both black and white pixels")
	}
	// A correct QR code of scale=8 will have a substantial amount of both black and white pixels.
	// If it was bugged (1-pixel scale), blackCount would be extremely small (e.g. < 1000).
	if blackCount < 5000 {
		t.Errorf("QR code black pixels count is too low (%d), indicating scaling issue", blackCount)
	}
}

func TestRefreshPairing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	oldToken := app.token
	linkBefore, err := app.GetPairingLink()
	if err != nil {
		t.Fatalf("GetPairingLink() error = %v", err)
	}

	if err := app.RefreshPairing(); err != nil {
		t.Fatalf("RefreshPairing() error = %v", err)
	}
	if app.token == oldToken {
		t.Fatal("token should change after RefreshPairing")
	}

	linkAfter, err := app.GetPairingLink()
	if err != nil {
		t.Fatalf("GetPairingLink() after refresh error = %v", err)
	}
	if linkBefore == linkAfter {
		t.Fatal("pairing link should change after RefreshPairing")
	}

	disk, err := os.ReadFile(app.settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(disk, []byte(app.token)) {
		t.Fatal("refreshed token should be persisted")
	}
}

func TestRefreshPairingWhileConnected(t *testing.T) {
	app := newTestApp(t)
	app.status.State = "connected"
	if err := app.RefreshPairing(); err != nil {
		t.Fatalf("RefreshPairing while connected: %v", err)
	}
}

func TestNewAppReusesPersistedToken(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	app1, err := NewApp()
	if err != nil {
		t.Fatal(err)
	}
	token := app1.token

	app2, err := NewApp()
	if err != nil {
		t.Fatal(err)
	}
	if app2.token != token {
		t.Fatalf("token not reused: %q vs %q", app2.token, token)
	}
}

func TestSaveSettingsPreservesToken(t *testing.T) {
	app := newTestApp(t)
	token := app.token
	if err := app.SaveSettings(Settings{Port: 7655, Language: "en-US"}); err != nil {
		t.Fatal(err)
	}
	if app.token != token {
		t.Fatalf("token changed on SaveSettings: %q vs %q", app.token, token)
	}
}

func TestSaveSettingsPreservesExternallyRotatedToken(t *testing.T) {
	app := newTestApp(t)
	oldToken := app.token

	rotated, err := config.RotateToken(app.settingsPath)
	if err != nil {
		t.Fatalf("RotateToken: %v", err)
	}
	if rotated.Token == oldToken {
		t.Fatal("expected rotated token to differ from initial token")
	}

	if err := app.SaveSettings(Settings{Port: app.port, Language: "en-US", AutoStart: true}); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	if app.token != rotated.Token {
		t.Fatalf("app.token = %q, want externally rotated token %q", app.token, rotated.Token)
	}
	disk, err := config.Load(app.settingsPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if disk.Token != rotated.Token {
		t.Fatalf("disk token = %q, want externally rotated token %q", disk.Token, rotated.Token)
	}
}

type stubPaster struct{}

func (stubPaster) Paste(string) error { return nil }
func (stubPaster) Name() string       { return "stub" }

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestDisconnectClientsKeepsServerRunning(t *testing.T) {
	app := newTestApp(t)
	app.paster = stubPaster{}

	port := freePort(t)
	if err := app.StartServer(port); err != nil {
		t.Fatalf("StartServer: %v", err)
	}
	defer func() { _ = app.StopServer() }()

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d/ws?token=%s", port, app.token)

	// Wait briefly for ListenAndServe goroutine to bind.
	var conn *websocket.Conn
	var err error
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
		if err == nil {
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("failed to connect websocket client: %v", err)
	}
	defer conn.Close()

	// Wait for OnConnect to update status to connected.
	deadline = time.Now().Add(time.Second)
	for app.GetConnectionStatus().State != "connected" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := app.GetConnectionStatus().State; got != "connected" {
		t.Fatalf("status = %q, want connected", got)
	}

	// DisconnectClients should drop client, move state to waiting, and keep server running.
	app.DisconnectClients()

	if got := app.GetConnectionStatus().State; got != "waiting" {
		t.Fatalf("status after DisconnectClients = %q, want waiting", got)
	}

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected disconnected client read to fail")
	}

	// Server must still be running and accept a new connection.
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected server to remain listening after DisconnectClients, got err=%v", err)
	}
	conn2.Close()
}

func TestStartServerPasterFailureRollsBackStatus(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux session-detection failure test")
	}
	app := newTestApp(t)
	t.Setenv("XDG_SESSION_TYPE", "unknown")
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", "")
	t.Setenv("PATH", "")

	port := freePort(t)
	if err := app.StartServer(port); err == nil {
		_ = app.StopServer()
		t.Fatal("expected StartServer to fail when paster initialization fails")
	}
	if got := app.GetConnectionStatus().State; got != "disconnected" {
		t.Fatalf("status after paster failure = %q, want disconnected", got)
	}
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
	app, err := NewApp()
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestRelayPairingLinkQRAndRefresh(t *testing.T) {
	app := newTestApp(t)

	if err := app.SaveSettings(Settings{
		Port:           7655,
		Language:       "en-US",
		ConnectionMode: config.ModeRelay,
		RelayURL:       "wss://relay.airvoice.yuler.cc/ws",
		RelayAuthKey:   "my-secret",
	}); err != nil {
		t.Fatalf("SaveSettings(relay): %v", err)
	}

	link1, err := app.GetPairingLink()
	if err != nil {
		t.Fatalf("GetPairingLink(relay): %v", err)
	}
	if !strings.HasPrefix(link1, "airvoice://pair?v=2&mode=relay&ws=wss%3A%2F%2Frelay.airvoice.yuler.cc%2Fws&key=") {
		t.Fatalf("unexpected relay pairing link: %q", link1)
	}
	if !strings.Contains(link1, "&auth=my-secret") {
		t.Fatalf("expected auth query param in relay pairing link: %q", link1)
	}

	qrData, err := app.GetQRCode()
	if err != nil {
		t.Fatalf("GetQRCode(relay): %v", err)
	}
	if !strings.HasPrefix(qrData, "data:image/png;base64,") {
		t.Fatalf("unexpected QR data URL: %q", qrData[:min(30, len(qrData))])
	}

	oldRelayKey := app.relayKey
	if err := app.RefreshPairing(); err != nil {
		t.Fatalf("RefreshPairing: %v", err)
	}
	if app.relayKey == "" || app.relayKey == oldRelayKey {
		t.Fatalf("relayKey did not rotate: %q vs %q", app.relayKey, oldRelayKey)
	}

	link2, err := app.GetPairingLink()
	if err != nil {
		t.Fatalf("GetPairingLink after refresh: %v", err)
	}
	if link2 == link1 {
		t.Fatal("expected relay pairing link to change after RefreshPairing")
	}
}

func TestSaveSettingsRejectsInsecureRelayURL(t *testing.T) {
	app := newTestApp(t)
	err := app.SaveSettings(Settings{
		Port:           7655,
		Language:       "en-US",
		ConnectionMode: config.ModeRelay,
		RelayURL:       "ws://relay.airvoice.yuler.cc/ws",
	})
	if err == nil {
		t.Fatal("expected SaveSettings to reject public ws:// relay URL")
	}
}

func TestDesktopRelayLifecycleAndModeSwitch(t *testing.T) {
	rs, err := relay.NewServer(relay.Config{})
	if err != nil {
		t.Fatalf("relay.NewServer: %v", err)
	}
	ts := httptest.NewServer(rs.Handler())
	defer func() {
		ts.Close()
		_ = rs.Close(context.Background())
	}()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	app := newTestApp(t)
	app.paster = stubPaster{}

	// Start in LAN mode, then switch to Relay mode via SaveSettings.
	lanPort := freePort(t)
	if err := app.StartServer(lanPort); err != nil {
		t.Fatalf("StartServer: %v", err)
	}
	defer func() { _ = app.StopServer() }()

	if err := app.SaveSettings(Settings{
		Port:           lanPort,
		Language:       "en-US",
		ConnectionMode: config.ModeRelay,
		RelayURL:       wsURL,
	}); err != nil {
		t.Fatalf("SaveSettings(relay): %v", err)
	}

	// Wait until Desktop relay client connects and enters waiting state.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		st := app.GetConnectionStatus()
		if st.Mode == config.ModeRelay && st.State == "waiting" {
			break
		}
		time.Sleep(15 * time.Millisecond)
	}
	if st := app.GetConnectionStatus(); st.Mode != config.ModeRelay || st.State != "waiting" {
		t.Fatalf("status = %+v, want mode=relay state=waiting", st)
	}

	// Simulate an encrypted phone connecting through the relay.
	keys, err := e2e.DeriveKeysFromBase64(app.relayKey)
	if err != nil {
		t.Fatal(err)
	}
	headers := http.Header{}
	headers.Set("X-Airvoice-Room", keys.RoomID)
	headers.Set("X-Airvoice-Role", "client")
	headers.Set("X-Airvoice-Auth", keys.ClientAuth)
	phoneConn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("phone dial: %v", err)
	}
	defer phoneConn.Close()

	sender, _ := e2e.NewSender(keys.KeyC2H, e2e.DirC2H, keys.RoomID)
	guard, _ := e2e.NewSessionGuard(keys.KeyH2C, e2e.DirH2C, keys.RoomID)

	sendPhone := func(in protocol.Inbound) {
		raw, _ := json.Marshal(in)
		env, err := sender.Seal(raw)
		if err != nil {
			t.Fatal(err)
		}
		wire, _ := json.Marshal(env)
		if err := phoneConn.WriteMessage(websocket.TextMessage, wire); err != nil {
			t.Fatal(err)
		}
	}

	sendPhone(protocol.Inbound{Type: "hello", Device: "iPhone Test", TS: time.Now().Unix()})
	sendPhone(protocol.Inbound{Type: "text", ID: "d1", Content: "desktop relay text", Device: "iPhone Test"})

	// Read encrypted hello reply + ack.
	_ = phoneConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	gotAck := false
	for !gotAck {
		_, data, err := phoneConn.ReadMessage()
		if err != nil {
			t.Fatalf("phone read: %v", err)
		}
		var env e2e.Envelope
		if err := json.Unmarshal(data, &env); err != nil || env.Type != e2e.EnvelopeType {
			continue
		}
		pt, err := guard.Open(env)
		if err != nil {
			t.Fatalf("phone guard.Open: %v", err)
		}
		var out protocol.Outbound
		_ = json.Unmarshal(pt, &out)
		if out.Type == "ack" && out.ID == "d1" && out.OK {
			gotAck = true
		}
	}

	st := app.GetConnectionStatus()
	if st.State != "connected" || st.Mode != config.ModeRelay || st.DeviceName != "iPhone Test" {
		t.Fatalf("status after relay hello = %+v", st)
	}

	entries, err := app.GetHistory(10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(entries) == 0 || entries[0].Content != "desktop relay text" {
		t.Fatalf("history missing relay text: %+v", entries)
	}

	// Switch back to LAN mode -> stops relay client and restarts LAN server.
	nextPort := freePort(t)
	if err := app.SaveSettings(Settings{
		Port:           nextPort,
		Language:       "en-US",
		ConnectionMode: config.ModeLAN,
	}); err != nil {
		t.Fatalf("SaveSettings(lan): %v", err)
	}
	if app.relayClient != nil {
		t.Fatal("expected relayClient to be nil after switching back to LAN")
	}
	if app.server == nil {
		t.Fatal("expected LAN server to be running after switching back to LAN")
	}
}
