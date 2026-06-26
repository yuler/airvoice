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
	s.SetToken("tok")
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
	s.SetToken("tok")
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
