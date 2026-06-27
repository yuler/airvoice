# A5 cli/server Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the `cli/server` package containing the websocket server, hub connection manager, request handlers, and write comprehensive tests.

**Architecture:** 
- `Hub` manages the single active websocket connection concurrently using `sync.Mutex`.
- `handler.go` handles protocol message routing (`hello`, `text`, `ping`).
- `server.go` hosts `/health` and `/ws` endpoints with token authorization query validation.

**Tech Stack:** Go (1.22), gorilla/websocket

---

### Task 1: Create `cli/server/hub.go`

**Files:**
- Create: `cli/server/hub.go`

- [ ] **Step 1: Create `hub.go` with thread-safe Set, Clear, and Get methods.**

```go
package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func NewHub() *Hub {
	return &Hub{}
}

func (h *Hub) Set(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn != nil {
		h.conn.Close()
	}
	h.conn = conn
}

func (h *Hub) Clear(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn == conn {
		if h.conn != nil {
			h.conn.Close()
		}
		h.conn = nil
	}
}

func (h *Hub) Get() *websocket.Conn {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.conn
}
```

---

### Task 2: Create `cli/server/handler.go`

**Files:**
- Create: `cli/server/handler.go`

- [ ] **Step 1: Implement handler logic to parse websocket messages and dispatch based on protocol.**

```go
package server

import (
	"github.com/yuler/airvoice/cli/protocol"
	"github.com/gorilla/websocket"
)

func (s *Server) handleConnection(conn *websocket.Conn) {
	s.hub.Set(conn)
	defer s.hub.Clear(conn)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType != websocket.TextMessage {
			continue
		}

		inbound, err := protocol.ParseInbound(message)
		if err != nil {
			continue
		}

		var outbound protocol.Outbound
		switch inbound.Type {
		case "hello":
			outbound = protocol.Outbound{
				Type:    "hello",
				Host:    s.cfg.Hostname,
				Version: s.cfg.Version,
			}
		case "text":
			err := s.cfg.Paster.Paste(inbound.Content)
			if err != nil {
				outbound = protocol.Outbound{
					Type:    "ack",
					ID:      inbound.ID,
					OK:      false,
					Message: err.Error(),
				}
			} else {
				outbound = protocol.Outbound{
					Type: "ack",
					ID:   inbound.ID,
					OK:   true,
				}
			}
		case "ping":
			outbound = protocol.Outbound{
				Type: "pong",
			}
		default:
			continue
		}

		data, err := outbound.Bytes()
		if err != nil {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}
}
```

---

### Task 3: Create `cli/server/server.go`

**Files:**
- Create: `cli/server/server.go`

- [ ] **Step 1: Define `Config` and `Server` struct, and HTTP handlers for `/health` and `/ws`.**

```go
package server

import (
	"net/http"

	"github.com/yuler/airvoice/cli/paste"
	"github.com/gorilla/websocket"
)

type Config struct {
	Addr, Token, Hostname, Version string
	Paster                         paste.Paster
}

type Server struct {
	cfg      Config
	hub      *Hub
	upgrader websocket.Upgrader
}

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
	if token != s.cfg.Token {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.handleConnection(conn)
}
```

---

### Task 4: Create Tests in `cli/server/server_test.go`

**Files:**
- Create: `cli/server/server_test.go`

- [ ] **Step 1: Write websocket server tests verifying routing, auth, message parsing, and hub behaviors.**

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

	"github.com/yuler/airvoice/cli/protocol"
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
	s := New(Config{Addr: ":0", Token: "test-token", Hostname: "test-host", Version: "0.1.0"})
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
	s := New(Config{Addr: ":0", Token: "valid-token", Hostname: "test-host", Version: "0.1.0"})
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
	s := New(Config{Addr: ":0", Token: "token", Hostname: "host-pc", Version: "0.1.0", Paster: paster})
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

	if err := conn.ReadJSON(&ackResp); err != nil {
		t.Fatal(err)
	}
	if ackResp.Type != "ack" || ackResp.OK || ackResp.Message != "paste failed" {
		t.Errorf("expected failure ack, got: %+v", ackResp)
	}
}

func TestHubConnectionLifecycle(t *testing.T) {
	s := New(Config{Addr: ":0", Token: "token", Hostname: "host-pc", Version: "0.1.0"})
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=token"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn1.Close()

	// Wait briefly for connection 1 to register in hub
	time.Sleep(10 * time.Millisecond)
	active1 := s.hub.Get()
	if active1 == nil {
		t.Fatal("expected hub to have connection 1")
	}

	// Connect second client, should kick/close connection 1
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()

	time.Sleep(10 * time.Millisecond)
	active2 := s.hub.Get()
	if active2 == nil {
		t.Fatal("expected hub to have connection 2")
	}

	// Verify that reading from connection 1 now fails (was closed)
	_, _, err = conn1.ReadMessage()
	if err == nil {
		t.Error("expected connection 1 to be closed by hub override")
	}

	// Clear old connection, shouldn't clear active one
	s.hub.Clear(active1)
	if s.hub.Get() == nil {
		t.Error("hub cleared active connection 2 when cleaning up connection 1")
	}

	// Clear active connection, should remove it
	s.hub.Clear(active2)
	if s.hub.Get() != nil {
		t.Error("hub failed to clear active connection")
	}
}
```

---

### Task 5: Run tests and verify

- [ ] **Step 1: Execute `go test ./cli/server/...`**

Run: `go test -v ./cli/server/...`
Expected: ALL PASS

---

### Task 6: Git commit

- [ ] **Step 1: Stage and commit the server package.**

Run: `git add cli/server/ && git commit -m "feat: A5 cli/server package"`
