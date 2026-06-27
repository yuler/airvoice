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
	s.SetToken("valid-token")
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
	s.SetToken("token")
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
	s.SetToken("token")
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
	s.SetToken("stable-token")
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

