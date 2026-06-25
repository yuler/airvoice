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

func setTestToken(s *Server, token string) {
	s.tokenMu.Lock()
	s.token = token
	s.tokenMu.Unlock()
}

func TestRotatePairingOnDisconnect(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7383, Hostname: "host-pc", Version: "0.1.0"})
	setTestToken(s, "token-old")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=token-old"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	s.tokenMu.RLock()
	newToken := s.token
	s.tokenMu.RUnlock()
	if newToken == "" || newToken == "token-old" {
		t.Fatalf("expected rotated token, got %q", newToken)
	}

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected dial to fail with stale token")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}

	freshURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=" + newToken
	conn2, resp, err := websocket.DefaultDialer.Dial(freshURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("expected status 101, got %d", resp.StatusCode)
	}
}
