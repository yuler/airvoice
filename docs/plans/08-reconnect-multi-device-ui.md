# Reconnect, Multi-Device, and UI Polish Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persistent token (no rotation on disconnect), multi-device hub, iOS auto-reconnect on foreground, countdown bar at screen top, and header → status-bar style.

**Architecture:** CLI server keeps a single token for its entire process lifetime. Hub switches from single-connection to a connection map supporting multiple simultaneous devices. iOS `ConnectionManager` gains a `reconnect()` method that fires on `scenePhase → .active`. The countdown bar moves to the very top of the screen (above safe area, full screen width, zero spacing). Header elements split: connection status becomes a narrow top bar; theme toggle and pair button move into a toolbar at the bottom of the text editor area.

**Tech Stack:** Go (CLI server), Swift/SwiftUI (iOS)

---

## File Structure

### CLI (Go) — files to modify

| File | Change |
|------|--------|
| `cli/server/hub.go` | Rewrite: single `*websocket.Conn` → `map[*websocket.Conn]struct{}` |
| `cli/server/hub_test.go` | Create: unit tests for multi-conn hub |
| `cli/server/handler.go` | Remove `RotatePairing` call on disconnect; add paste mutex |
| `cli/server/server.go` | Remove `RotatePairing` method; add `SetToken` for tests; add `pasteMu` |
| `cli/server/server_test.go` | Update tests: remove token-rotation test, add multi-device + stable-token tests |
| `cli/pairing/session.go` | Add `PrintPairingWithToken` that accepts a pre-generated token |
| `cli/main.go` | Generate token once at startup, pass to server |

### iOS (Swift) — files to modify

| File | Change |
|------|--------|
| `ios/Airvoice/Services/ConnectionManager.swift` | Add `lastPayload` storage, `reconnect()` method |
| `ios/Airvoice/Views/HomeView.swift` | Move countdown bar to top; split header into status bar + relocated buttons; add `scenePhase` reconnect |
| `ios/Airvoice/Utilities/AppTheme.swift` | Add `statusBar*` color tokens |

---

## Task 1: Rewrite Hub to Support Multiple Connections

**Files:**
- Modify: `cli/server/hub.go`
- Create: `cli/server/hub_test.go`

- [ ] **Step 1: Write failing tests for multi-connection hub**

Create `cli/server/hub_test.go`:

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubMultipleConnections(t *testing.T) {
	hub := NewHub()
	if hub.Count() != 0 {
		t.Fatalf("expected 0 connections, got %d", hub.Count())
	}

	s := New(Config{Addr: ":0", Port: 7383, Hostname: "test", Version: "0.1.0"})
	setTestToken(s, "tok")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()
	time.Sleep(20 * time.Millisecond)

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	time.Sleep(20 * time.Millisecond)

	if s.hub.Count() != 2 {
		t.Fatalf("expected 2 connections, got %d", s.hub.Count())
	}

	// Both connections should still be readable (conn1 NOT kicked)
	ping := []byte(`{"type":"ping"}`)
	if err := conn1.WriteMessage(websocket.TextMessage, ping); err != nil {
		t.Fatal("conn1 should still be alive:", err)
	}
	var pong struct{ Type string `json:"type"` }
	if err := conn1.ReadJSON(&pong); err != nil {
		t.Fatal("conn1 read failed:", err)
	}
	if pong.Type != "pong" {
		t.Errorf("expected pong, got %q", pong.Type)
	}
}

func TestHubRemoveOnDisconnect(t *testing.T) {
	hub := NewHub()
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "test", Version: "0.1.0"})
	s.hub = hub
	setTestToken(s, "tok")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"

	conn1, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	time.Sleep(20 * time.Millisecond)

	conn2, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn2.Close()
	time.Sleep(20 * time.Millisecond)

	if hub.Count() != 2 {
		t.Fatalf("expected 2, got %d", hub.Count())
	}

	conn1.Close()
	time.Sleep(50 * time.Millisecond)

	if hub.Count() != 1 {
		t.Fatalf("expected 1 after disconnect, got %d", hub.Count())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd cli && go test ./server/ -run 'TestHubMultiple|TestHubRemove' -v`
Expected: Compilation errors — `Count()` does not exist, `Set` kicks old connection.

- [ ] **Step 3: Rewrite hub.go for multi-connection support**

Replace the entire contents of `cli/server/hub.go`:

```go
package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains all active WebSocket connections.
type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

// NewHub initializes and returns a Hub.
func NewHub() *Hub {
	return &Hub{
		conns: make(map[*websocket.Conn]struct{}),
	}
}

// Add registers a connection.
func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = struct{}{}
}

// Remove unregisters a connection and closes it.
// Returns true if the connection was present.
func (h *Hub) Remove(conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.conns[conn]; ok {
		conn.Close()
		delete(h.conns, conn)
		return true
	}
	return false
}

// Count returns the number of active connections.
func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd cli && go test ./server/ -run 'TestHubMultiple|TestHubRemove' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cli/server/hub.go cli/server/hub_test.go
git commit -m "feat(server): rewrite hub for multi-device connections"
```

---

## Task 2: Update Handler — No Token Rotation, Paste Serialization

**Files:**
- Modify: `cli/server/handler.go`
- Modify: `cli/server/server.go`

- [ ] **Step 1: Update handler.go to use new Hub API and remove RotatePairing**

Replace the entire contents of `cli/server/handler.go`:

```go
package server

import (
	"sync"

	"github.com/airvoice/airvoice/cli/protocol"
	"github.com/gorilla/websocket"
)

// handleConnection handles the WebSocket read/write loops.
func (s *Server) handleConnection(conn *websocket.Conn) {
	s.hub.Add(conn)
	defer func() {
		if s.hub.Remove(conn) {
			logStatus("client disconnected (active: %d)", s.hub.Count())
			if s.hub.Count() == 0 {
				logStatus("all clients disconnected — waiting for reconnection...")
			}
		}
	}()

	var writeMu sync.Mutex
	writeOutbound := func(outbound protocol.Outbound) error {
		data, err := outbound.Bytes()
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteMessage(websocket.TextMessage, data)
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logStatus("read error: %v", err)
			break
		}
		if messageType != websocket.TextMessage {
			logStatus("ignored frame type=%d", messageType)
			continue
		}

		inbound, err := protocol.ParseInbound(message)
		if err != nil {
			logStatus("invalid json: %v", err)
			continue
		}

		switch inbound.Type {
		case "hello":
			logStatus("hello from device=%q app=%q", inbound.Device, inbound.App)
			outbound := protocol.Outbound{
				Type:    "hello",
				Host:    s.cfg.Hostname,
				Version: s.cfg.Version,
			}
			if err := writeOutbound(outbound); err != nil {
				logStatus("write error: %v", err)
				return
			}
			logStatus("hello reply host=%s version=%s", outbound.Host, outbound.Version)

		case "text":
			logStatus("text id=%s len=%d preview=%q", inbound.ID, len(inbound.Content), previewText(inbound.Content, 40))
			go func(inbound protocol.Inbound) {
				// Serialize paste operations across all connections.
				s.pasteMu.Lock()
				err := s.cfg.Paster.Paste(inbound.Content)
				s.pasteMu.Unlock()

				var outbound protocol.Outbound
				if err != nil {
					logStatus("paste failed id=%s: %v", inbound.ID, err)
					outbound = protocol.Outbound{
						Type:    "ack",
						ID:      inbound.ID,
						OK:      false,
						Message: err.Error(),
					}
				} else {
					logStatus("paste ok id=%s backend=%s", inbound.ID, s.cfg.Paster.Name())
					outbound = protocol.Outbound{
						Type: "ack",
						ID:   inbound.ID,
						OK:   true,
					}
				}
				if err := writeOutbound(outbound); err != nil {
					logStatus("write error: %v", err)
				} else {
					if outbound.OK {
						logStatus("ack ok id=%s", outbound.ID)
					} else {
						logStatus("ack fail id=%s message=%s", outbound.ID, outbound.Message)
					}
				}
			}(inbound)

		case "ping":
			logStatus("ping")
			outbound := protocol.Outbound{
				Type: "pong",
			}
			if err := writeOutbound(outbound); err != nil {
				logStatus("write error: %v", err)
				return
			}
			logStatus("pong")

		default:
			logStatus("ignored message type=%q", inbound.Type)
		}
	}
}
```

- [ ] **Step 2: Update server.go — add pasteMu, remove RotatePairing, add SetToken**

Replace the entire contents of `cli/server/server.go`:

```go
package server

import (
	"net/http"
	"sync"

	"github.com/airvoice/airvoice/cli/paste"
	"github.com/gorilla/websocket"
)

// Config defines the configuration for the Server.
type Config struct {
	Addr, Hostname, Version string
	Port                     int
	Paster                   paste.Paster
}

// Server handles health checks and upgrades/coordinates websocket connections.
type Server struct {
	cfg      Config
	hub      *Hub
	upgrader websocket.Upgrader
	tokenMu  sync.RWMutex
	token    string
	pasteMu  sync.Mutex
}

// New returns a newly configured Server instance.
func New(cfg Config) *Server {
	return &Server{
		cfg: cfg,
		hub: NewHub(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// SetToken sets the authentication token. Called once at startup.
func (s *Server) SetToken(token string) {
	s.tokenMu.Lock()
	s.token = token
	s.tokenMu.Unlock()
}

// ListenAndServe starts the HTTP/WS server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ws", s.handleWS)
	return http.ListenAndServe(s.cfg.Addr, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if !s.validToken(token) {
		logStatus("ws rejected: invalid token from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logStatus("ws upgrade failed from %s: %v", r.RemoteAddr, err)
		return
	}

	logStatus("client connected from %s (active: %d)", r.RemoteAddr, s.hub.Count()+1)
	s.handleConnection(conn)
}

func (s *Server) validToken(token string) bool {
	s.tokenMu.RLock()
	defer s.tokenMu.RUnlock()
	return token != "" && token == s.token
}
```

- [ ] **Step 3: Verify compilation**

Run: `cd cli && go build ./...`
Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add cli/server/handler.go cli/server/server.go
git commit -m "feat(server): remove token rotation, add paste serialization"
```

---

## Task 3: Update main.go and pairing — Generate Token Once at Startup

**Files:**
- Modify: `cli/main.go`
- Modify: `cli/pairing/session.go`

- [ ] **Step 1: Add PrintPairingWithToken to pairing/session.go**

Add this new function after the existing `PrintPairing` function in `cli/pairing/session.go`:

```go
// PrintPairingWithToken prints the QR code and pairing metadata for a given token.
func PrintPairingWithToken(port int, token string, banner string) (wsURL string, err error) {
	ip, err := LocalIPv4()
	if err != nil {
		return "", err
	}
	wsURL = fmt.Sprintf("ws://%s:%d/ws", ip, port)

	payload := &Payload{
		Version: 1,
		WS:      wsURL,
		Token:   token,
	}
	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", err
	}

	if banner != "" {
		fmt.Fprintf(os.Stderr, "\n [airvoice] %s\n\n", banner)
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}
	PrintQR(payloadBytes)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Token: %s\n", token)
	fmt.Fprintf(os.Stderr, "  WebSocket URL: %s\n\n", wsURL)
	fmt.Fprintf(os.Stderr, "  [airvoice] waiting for iPhone connection...\n\n")
	return wsURL, nil
}
```

- [ ] **Step 2: Update main.go to generate token once and use SetToken**

Replace the `case "serve":` block in `cli/main.go` (lines 26–65):

```go
	case "serve":
		port := 7383
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		portPtr := fs.Int("port", 7383, "port to listen on")
		if err := fs.Parse(os.Args[2:]); err == nil {
			port = *portPtr
		}

		paster, err := paste.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing paster: %v\n", err)
			os.Exit(1)
		}

		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "PC"
		}

		// Generate a stable token for the entire process lifetime.
		token := uuid.NewString()

		addr := fmt.Sprintf("0.0.0.0:%d", port)
		srv := server.New(server.Config{
			Addr:     addr,
			Port:     port,
			Hostname: hostname,
			Version:  version,
			Paster:   paster,
		})
		srv.SetToken(token)

		if _, err := pairing.PrintPairingWithToken(port, token, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating pairing session: %v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "  Paste backend: %s\n", paster.Name())
		fmt.Fprintf(os.Stderr, "  [airvoice] listening on %s (health: /health, ws: /ws)\n\n", addr)

		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			os.Exit(1)
		}
```

Add `"github.com/google/uuid"` to the imports in `main.go` alongside the existing imports.

- [ ] **Step 3: Verify build**

Run: `cd cli && go build ./...`
Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add cli/main.go cli/pairing/session.go
git commit -m "feat: generate stable token once at CLI startup"
```

---

## Task 4: Fix Existing Server Tests

**Files:**
- Modify: `cli/server/server_test.go`

- [ ] **Step 1: Rewrite server_test.go**

The `RotatePairing` method no longer exists. `Hub.Set/Clear` are now `Hub.Add/Remove`. The `TestRotatePairingOnDisconnect` test is replaced with `TestTokenStableAfterDisconnect`. The `TestHubConnectionLifecycle` test is updated for multi-connection (no kicking).

Replace the entire contents of `cli/server/server_test.go`:

```go
package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/airvoice/airvoice/cli/protocol"
	"github.com/gorilla/websocket"
)

type mockPaster struct {
	lastPaste string
	err       error
}

func (m *mockPaster) Paste(text string) error {
	m.lastPaste = text
	return m.err
}

func (m *mockPaster) Name() string {
	return "mock"
}

func TestServerHealth(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "test-host", Version: "0.1.0"})
	ts := httptest.NewServer(http.HandlerFunc(s.handleHealth))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerWSAuth(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "test-host", Version: "0.1.0"})
	setTestToken(s, "valid-token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	// Invalid token
	_, resp, err := websocket.DefaultDialer.Dial(wsURL+"?token=bad-token", nil)
	if err == nil {
		t.Error("expected dial to fail with invalid token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	// Valid token
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL+"?token=valid-token", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status 101, got %d", resp.StatusCode)
	}
}

func TestServerMessageHandling(t *testing.T) {
	paster := &mockPaster{}
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "host-pc", Version: "0.1.0", Paster: paster})
	setTestToken(s, "token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=token"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 1. Hello message
	helloMsg := protocol.Inbound{Type: "hello"}
	helloData, _ := json.Marshal(helloMsg)
	if err := conn.WriteMessage(websocket.TextMessage, helloData); err != nil {
		t.Fatal(err)
	}

	var helloResp protocol.Outbound
	if err := conn.ReadJSON(&helloResp); err != nil {
		t.Fatal(err)
	}
	if helloResp.Type != "hello" || helloResp.Host != "host-pc" || helloResp.Version != "0.1.0" {
		t.Errorf("unexpected hello response: %+v", helloResp)
	}

	// 2. Ping message
	pingMsg := protocol.Inbound{Type: "ping"}
	pingData, _ := json.Marshal(pingMsg)
	if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
		t.Fatal(err)
	}

	var pingResp protocol.Outbound
	if err := conn.ReadJSON(&pingResp); err != nil {
		t.Fatal(err)
	}
	if pingResp.Type != "pong" {
		t.Errorf("expected pong, got: %+v", pingResp)
	}

	// 3. Text message success
	textMsg := protocol.Inbound{Type: "text", ID: "msg-1", Content: "hello world"}
	textData, _ := json.Marshal(textMsg)
	if err := conn.WriteMessage(websocket.TextMessage, textData); err != nil {
		t.Fatal(err)
	}

	var ackResp protocol.Outbound
	if err := conn.ReadJSON(&ackResp); err != nil {
		t.Fatal(err)
	}
	if ackResp.Type != "ack" || ackResp.ID != "msg-1" || !ackResp.OK {
		t.Errorf("expected success ack, got: %+v", ackResp)
	}
	if paster.lastPaste != "hello world" {
		t.Errorf("expected paste content to be 'hello world', got '%s'", paster.lastPaste)
	}

	// 4. Text message paste failure
	paster.err = errors.New("paste failed")
	if err := conn.WriteMessage(websocket.TextMessage, textData); err != nil {
		t.Fatal(err)
	}

	var failAckResp protocol.Outbound
	if err := conn.ReadJSON(&failAckResp); err != nil {
		t.Fatal(err)
	}
	if failAckResp.Type != "ack" || failAckResp.OK || failAckResp.Message != "paste failed" {
		t.Errorf("expected failure ack, got: %+v", failAckResp)
	}
}

func TestHubConnectionLifecycle(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "host-pc", Version: "0.1.0"})
	setTestToken(s, "token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=token"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()
	time.Sleep(20 * time.Millisecond)

	if s.hub.Count() != 1 {
		t.Fatalf("expected 1 connection, got %d", s.hub.Count())
	}

	// Second client does NOT kick the first
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	time.Sleep(20 * time.Millisecond)

	if s.hub.Count() != 2 {
		t.Fatalf("expected 2 connections, got %d", s.hub.Count())
	}

	// conn1 is still alive
	ping := []byte(`{"type":"ping"}`)
	if err := conn1.WriteMessage(websocket.TextMessage, ping); err != nil {
		t.Fatal("conn1 should still be alive:", err)
	}
	var pong struct{ Type string `json:"type"` }
	if err := conn1.ReadJSON(&pong); err != nil {
		t.Fatal(err)
	}
	if pong.Type != "pong" {
		t.Errorf("expected pong from conn1, got %q", pong.Type)
	}

	// Close conn1, only conn2 remains
	conn1.Close()
	time.Sleep(50 * time.Millisecond)

	if s.hub.Count() != 1 {
		t.Fatalf("expected 1 after disconnect, got %d", s.hub.Count())
	}
}

func TestTokenStableAfterDisconnect(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "host-pc", Version: "0.1.0"})
	setTestToken(s, "stable-token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=stable-token"

	// Connect and disconnect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	// Token should NOT have changed
	s.tokenMu.RLock()
	currentToken := s.token
	s.tokenMu.RUnlock()
	if currentToken != "stable-token" {
		t.Fatalf("token rotated unexpectedly: got %q, want %q", currentToken, "stable-token")
	}

	// Reconnect with same token should work
	conn2, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected 101, got %d", resp.StatusCode)
	}
}

func setTestToken(s *Server, token string) {
	s.tokenMu.Lock()
	s.token = token
	s.tokenMu.Unlock()
}
```

- [ ] **Step 2: Run all server tests**

Run: `cd cli && go test ./server/ -v`
Expected: All tests PASS.

- [ ] **Step 3: Run all CLI tests**

Run: `cd cli && go test ./... -v`
Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
git add cli/server/server_test.go
git commit -m "test(server): update tests for multi-device hub and stable token"
```

---

## Task 5: iOS — ConnectionManager Reconnect Support

**Files:**
- Modify: `ios/Airvoice/Services/ConnectionManager.swift`

- [ ] **Step 1: Add lastPayload and reconnect() to ConnectionManager**

Replace the entire contents of `ios/Airvoice/Services/ConnectionManager.swift`:

```swift
import Foundation
import UIKit
import Combine

@MainActor
class ConnectionManager: ObservableObject {
    enum ConnectionState: Equatable {
        case disconnected
        case connecting
        case connected
        case error(String)
    }

    @Published var state: ConnectionState = .disconnected
    @Published var hostName: String? = nil

    var onAck: ((String, Bool, String?) -> Void)?
    var onTransportError: ((String) -> Void)?

    /// The most recent pairing payload, kept in memory for reconnection.
    private(set) var lastPayload: PairingPayload?

    private var webSocketTask: URLSessionWebSocketTask?
    private var session: URLSession?
    private var isDisconnecting = false

    /// Whether a reconnect attempt has already been scheduled/in-progress.
    private var isReconnecting = false

    func connect(payload: PairingPayload) {
        disconnect()

        lastPayload = payload
        isDisconnecting = false
        isReconnecting = false
        state = .connecting
        hostName = nil

        connectWithPayload(payload)
    }

    /// Reconnect using the stored payload. No-op if no payload or already connected.
    func reconnect() {
        guard let payload = lastPayload else { return }
        guard state != .connected, state != .connecting else { return }
        guard !isReconnecting else { return }

        isReconnecting = true
        isDisconnecting = false
        state = .connecting
        hostName = nil

        connectWithPayload(payload)
    }

    func disconnect() {
        isDisconnecting = true
        isReconnecting = false
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session?.invalidateAndCancel()
        session = nil
        state = .disconnected
        hostName = nil
    }

    /// Clear stored payload (e.g. on 401 when CLI has restarted).
    func clearStoredPayload() {
        lastPayload = nil
    }

    /// Whether we have credentials to attempt reconnection.
    var canReconnect: Bool {
        lastPayload != nil
    }

    @discardableResult
    func sendText(id: String, content: String) -> Bool {
        guard state == .connected, let task = webSocketTask else {
            return false
        }

        let outbound = OutboundText(
            id: id,
            content: content,
            ts: Int(Date().timeIntervalSince1970)
        )

        do {
            let data = try JSONEncoder().encode(outbound)
            guard let jsonString = String(data: data, encoding: .utf8) else {
                return false
            }
            task.send(.string(jsonString)) { [weak self] error in
                if let error {
                    Task { @MainActor [weak self] in
                        guard let self, !self.isDisconnecting else { return }
                        let message = "Send failed: \(error.localizedDescription)"
                        self.state = .error(message)
                        self.onTransportError?(message)
                    }
                }
            }
            return true
        } catch {
            state = .error("Encoding failed: \(error.localizedDescription)")
            onTransportError?("Encoding failed: \(error.localizedDescription)")
            return false
        }
    }

    // MARK: - Private

    private func connectWithPayload(_ payload: PairingPayload) {
        // Clean up any existing connection first.
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        session?.invalidateAndCancel()
        session = nil

        guard var components = URLComponents(string: payload.ws) else {
            state = .error("Invalid WebSocket URL")
            isReconnecting = false
            return
        }

        var queryItems = components.queryItems ?? []
        queryItems.append(URLQueryItem(name: "token", value: payload.token))
        components.queryItems = queryItems

        guard let url = components.url else {
            state = .error("Invalid URL generated")
            isReconnecting = false
            return
        }

        let session = URLSession(configuration: .default)
        self.session = session
        let task = session.webSocketTask(with: url)
        webSocketTask = task

        task.resume()
        receiveMessage(task: task)
        sendHello(task: task)
    }

    private func sendHello(task: URLSessionWebSocketTask) {
        let device = UIDevice.current.name
        let hello = OutboundHello(device: device, app: "0.1.0")

        do {
            let data = try JSONEncoder().encode(hello)
            guard let jsonString = String(data: data, encoding: .utf8) else {
                return
            }
            task.send(.string(jsonString)) { [weak self] error in
                if let error {
                    Task { @MainActor [weak self] in
                        guard let self, !self.isDisconnecting else { return }
                        self.state = .error("Hello failed: \(error.localizedDescription)")
                        self.isReconnecting = false
                    }
                }
            }
        } catch {
            state = .error("Hello encoding failed: \(error.localizedDescription)")
            isReconnecting = false
        }
    }

    private func receiveMessage(task: URLSessionWebSocketTask) {
        task.receive { [weak self] result in
            Task { @MainActor [weak self] in
                guard let self, self.webSocketTask === task else { return }

                switch result {
                case .success(let message):
                    switch message {
                    case .string(let text):
                        self.processIncomingText(text)
                    case .data(let data):
                        if let text = String(data: data, encoding: .utf8) {
                            self.processIncomingText(text)
                        }
                    @unknown default:
                        break
                    }
                    self.receiveMessage(task: task)

                case .failure(let error):
                    if !self.isDisconnecting {
                        self.state = .error("Connection lost: \(error.localizedDescription)")
                        self.hostName = nil
                        self.isReconnecting = false
                        self.onTransportError?("Connection lost: \(error.localizedDescription)")
                    }
                }
            }
        }
    }

    private func processIncomingText(_ text: String) {
        guard let data = text.data(using: .utf8) else { return }

        do {
            let msg = try JSONDecoder().decode(InboundMessage.self, from: data)
            switch msg.type {
            case "hello":
                state = .connected
                hostName = msg.host ?? "Unknown Server"
                isReconnecting = false
            case "ack":
                if let id = msg.id {
                    onAck?(id, msg.ok ?? false, msg.message)
                }
            case "pong":
                break
            default:
                break
            }
        } catch {
            print("[airvoice] decode error: \(error) raw=\(text)")
        }
    }
}
```

- [ ] **Step 2: Build the iOS project to verify compilation**

Run: `cd ios && xcodebuild -project Airvoice.xcodeproj -scheme Airvoice -destination 'platform=iOS Simulator,name=iPhone 16' build 2>&1 | tail -5`
Expected: `** BUILD SUCCEEDED **`

- [ ] **Step 3: Commit**

```bash
git add ios/Airvoice/Services/ConnectionManager.swift
git commit -m "feat(ios): add reconnect support with in-memory payload"
```

---

## Task 6: iOS — HomeView UI Overhaul (Countdown Bar + Status Bar + Reconnect)

**Files:**
- Modify: `ios/Airvoice/Views/HomeView.swift`
- Modify: `ios/Airvoice/Utilities/AppTheme.swift`

- [ ] **Step 1: Add status bar color tokens to AppTheme**

Add these computed properties to `AppTheme` in `ios/Airvoice/Utilities/AppTheme.swift`, after the existing `toastBackground` property (around line 85):

```swift
    var statusBarConnected: Color { Color(hex: "00ac3a") }
    var statusBarConnecting: Color { Color(hex: "ffae00") }
    var statusBarError: Color { Color(hex: "e2162a") }
    var statusBarDisconnected: Color { Color(hex: "8f8f8f") }
```

- [ ] **Step 2: Rewrite HomeView.swift**

Replace the entire contents of `ios/Airvoice/Views/HomeView.swift`:

```swift
import SwiftUI
import Combine

struct HomeView: View {
    @AppStorage("appTheme") private var appThemeRaw = AppTheme.light.rawValue
    @Environment(\.scenePhase) private var scenePhase

    @StateObject private var connection = ConnectionManager()
    @StateObject private var autoSend = AutoSendController()
    @StateObject private var viewModel = HomeViewModel()

    @FocusState private var isEditorFocused: Bool

    @State private var showScanner = false

    private var theme: AppTheme {
        AppTheme(rawValue: appThemeRaw) ?? .light
    }

    var body: some View {
        ZStack(alignment: .top) {
            theme.background.ignoresSafeArea()

            VStack(spacing: 0) {
                // — Status bar (narrow, full width, at the very top)
                statusBar

                // — Countdown bar (full width, no spacing, directly below status bar)
                AutoSendCountdownBar(
                    active: autoSend.countdownActive,
                    token: autoSend.countdownToken,
                    duration: autoSend.autoSendDelay
                )

                // — Main content
                VStack(spacing: 20) {
                    editorSection
                    bottomControls
                }
                .padding(.top, 12)
            }
        }
        .toast(message: $viewModel.toastMessage, isError: $viewModel.isToastError, theme: theme)
        .sheet(isPresented: $showScanner) {
            scannerSheetView
        }
        .onAppear {
            viewModel.wire(connection: connection, autoSend: autoSend)
            focusEditorForKeyboard()
        }
        .onChange(of: showScanner) { _, isShowing in
            if isShowing {
                isEditorFocused = false
            } else {
                focusEditorForKeyboard()
            }
        }
        .onChange(of: connection.state) { _, newValue in
            UIApplication.shared.isIdleTimerDisabled = (newValue == .connected)
            if newValue == .connected {
                viewModel.flushPendingAutoSend(connection: connection, autoSend: autoSend)
            }
        }
        .onChange(of: scenePhase) { _, newPhase in
            if newPhase == .active {
                if !showScanner {
                    focusEditorForKeyboard()
                }
                // Auto-reconnect when returning to foreground
                if connection.state != .connected,
                   connection.state != .connecting,
                   connection.canReconnect {
                    connection.reconnect()
                }
            }
        }
    }

    // MARK: - Status Bar

    private var statusBar: some View {
        HStack(spacing: 8) {
            Circle()
                .fill(.white.opacity(0.9))
                .frame(width: 6, height: 6)

            Text(statusText)
                .font(.caption2)
                .fontWeight(.medium)
                .foregroundColor(.white.opacity(0.95))

            Spacer()

            Button {
                var next = theme
                next.toggle()
                appThemeRaw = next.rawValue
            } label: {
                Image(systemName: theme == .light ? "moon.fill" : "sun.max.fill")
                    .font(.caption2.weight(.semibold))
                    .foregroundColor(.white.opacity(0.9))
            }

            Button {
                showScanner = true
            } label: {
                Image(systemName: "qrcode.viewfinder")
                    .font(.caption2.weight(.semibold))
                    .foregroundColor(.white.opacity(0.9))
            }
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 6)
        .background(statusBarColor)
    }

    private var statusBarColor: Color {
        switch connection.state {
        case .disconnected: return theme.statusBarDisconnected
        case .connecting: return theme.statusBarConnecting
        case .connected: return theme.statusBarConnected
        case .error: return theme.statusBarError
        }
    }

    private var statusText: String {
        switch connection.state {
        case .disconnected: return "未连接"
        case .connecting: return "连接中..."
        case .connected: return "已连接: \(connection.hostName ?? "电脑")"
        case .error(let msg): return msg
        }
    }

    // MARK: - Editor

    private var editorSection: some View {
        ZStack(alignment: .topLeading) {
            if viewModel.text.isEmpty {
                Text("在此输入，或使用键盘麦克风语音输入...")
                    .foregroundColor(theme.placeholderText)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                    .allowsHitTesting(false)
            }

            TextEditor(text: $viewModel.text)
                .focused($isEditorFocused)
                .scrollContentBackground(.hidden)
                .foregroundColor(theme.primaryText)
                .font(.body)
                .padding(8)
                .onChange(of: viewModel.text) { _, newValue in
                    autoSend.textDidChange(newValue)
                    if newValue.isEmpty, !showScanner {
                        isEditorFocused = true
                    }
                }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(theme.secondaryBackground)
        .clipShape(RoundedRectangle(cornerRadius: 16))
        .overlay(
            RoundedRectangle(cornerRadius: 16)
                .stroke(theme.border, lineWidth: 1)
        )
        .padding(.horizontal, 20)
    }

    // MARK: - Bottom Controls

    private var bottomControls: some View {
        VStack(spacing: 12) {
            if autoSend.inFlight {
                HStack(spacing: 6) {
                    ProgressView()
                        .controlSize(.small)
                        .tint(theme.secondaryText)
                    Text("发送中")
                        .font(.caption)
                        .foregroundColor(theme.secondaryText)
                    Button {
                        viewModel.cancelSend(autoSend: autoSend)
                    } label: {
                        Text("取消")
                            .font(.caption.weight(.semibold))
                            .foregroundColor(theme.accent)
                    }
                }
                .padding(.horizontal, 12)
                .padding(.vertical, 6)
                .background(theme.chipBackground)
                .cornerRadius(14)
            } else if connection.state != .connected {
                Text(connection.canReconnect ? "连接中断，正在重连..." : "请先扫码连接电脑")
                    .font(.caption)
                    .foregroundColor(theme.secondaryText)
            }

            Button(action: {
                viewModel.manualSend(connection: connection, autoSend: autoSend)
            }) {
                HStack(spacing: 8) {
                    Image(systemName: "paperplane.fill")
                    Text("发送到电脑")
                }
                .font(.subheadline.weight(.semibold))
                .foregroundColor(theme.primaryText)
                .frame(maxWidth: .infinity)
                .frame(height: 44)
                .background(
                    theme.sendButtonBackground.opacity(
                        connection.state == .connected ? 1 : 0.5
                    )
                )
                .cornerRadius(22)
            }
            .disabled(connection.state != .connected || autoSend.inFlight)

            InputMethodTipsView(theme: theme)
        }
        .padding(.horizontal, 20)
        .padding(.bottom, 20)
    }

    // MARK: - Helpers

    private func focusEditorForKeyboard() {
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) {
            isEditorFocused = true
        }
    }

    private var scannerSheetView: some View {
        NavigationStack {
            QRScannerView(
                onScan: { rawPayload in
                    do {
                        let payload = try PairingPayload.decode(from: rawPayload)
                        connection.connect(payload: payload)
                        showScanner = false
                    } catch {
                        viewModel.showToast("无效的二维码", isError: true)
                    }
                },
                onError: { error in
                    viewModel.showToast("相机错误: \(error.localizedDescription)", isError: true)
                }
            )
            .navigationTitle("扫描电脑端二维码")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("取消") {
                        showScanner = false
                    }
                }
            }
            .ignoresSafeArea(edges: .bottom)
        }
    }
}

/// Thin bar that drains from full width to empty over `duration`, signalling
/// the idle countdown before an automatic send. Restarts whenever `token` changes.
private struct AutoSendCountdownBar: View {
    let active: Bool
    let token: Int
    let duration: Double

    @State private var progress: CGFloat = 0

    var body: some View {
        Rectangle()
            .fill(Color.red)
            .frame(height: 3)
            .scaleEffect(x: min(max(progress, 0), 1), y: 1, anchor: .leading)
            .frame(maxWidth: .infinity, alignment: .leading)
            .opacity(active ? 1 : 0)
            .animation(.easeOut(duration: 0.15), value: active)
            .onChange(of: token) { _, _ in
                progress = 1
                withAnimation(.linear(duration: duration)) {
                    progress = 0
                }
            }
    }
}
```

Key changes summary:
- `AutoSendCountdownBar` moved from below the text editor to directly below the status bar. Changed from `Capsule()` to `Rectangle()` for full-width edge-to-edge rendering.
- Old header (statusBadge + themeToggleButton + pair button) replaced with a narrow colored status bar containing status text + icon-only theme toggle + icon-only QR button.
- `scenePhase → .active` now triggers `connection.reconnect()`.
- Disconnected hint shows "连接中断，正在重连..." when `canReconnect` is true.

- [ ] **Step 3: Commit AppTheme changes**

```bash
git add ios/Airvoice/Utilities/AppTheme.swift
git commit -m "feat(ios): add status bar color tokens to AppTheme"
```

- [ ] **Step 4: Build iOS project**

Run: `cd ios && xcodebuild -project Airvoice.xcodeproj -scheme Airvoice -destination 'platform=iOS Simulator,name=iPhone 16' build 2>&1 | tail -5`
Expected: `** BUILD SUCCEEDED **`

- [ ] **Step 5: Commit HomeView**

```bash
git add ios/Airvoice/Views/HomeView.swift
git commit -m "feat(ios): status bar header, top countdown bar, auto-reconnect on foreground"
```

---

## Task 7: Verify Full Build — CLI + iOS

- [ ] **Step 1: Run all CLI tests**

Run: `cd cli && go test ./... -v`
Expected: All tests PASS.

- [ ] **Step 2: Build iOS**

Run: `cd ios && xcodebuild -project Airvoice.xcodeproj -scheme Airvoice -destination 'platform=iOS Simulator,name=iPhone 16' build 2>&1 | tail -5`
Expected: `** BUILD SUCCEEDED **`

- [ ] **Step 3: Final commit with all changes**

If any uncommitted files remain:
```bash
git add -A
git commit -m "chore: final cleanup for reconnect + multi-device + UI overhaul"
```
