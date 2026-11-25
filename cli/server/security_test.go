package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/protocol"
)

func TestWSOriginCheck(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7654, Hostname: "test-host", Version: "0.1.0"})
	s.SetToken("valid-token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=valid-token"

	t.Run("empty origin allowed", func(t *testing.T) {
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("expected empty Origin to succeed, got err=%v", err)
		}
		defer conn.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("expected 101, got %d", resp.StatusCode)
		}
	})

	t.Run("same host origin allowed", func(t *testing.T) {
		headers := http.Header{"Origin": {ts.URL}}
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatalf("expected same-host Origin %q to succeed, got err=%v", ts.URL, err)
		}
		defer conn.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("expected 101, got %d", resp.StatusCode)
		}
	})

	t.Run("cross-site origin rejected with 403", func(t *testing.T) {
		headers := http.Header{"Origin": {"https://evil.example"}}
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err == nil {
			t.Fatal("expected cross-site Origin to be rejected")
		}
		if resp == nil || resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got resp=%v", resp)
		}
	})

	t.Run("different port origin rejected with 403", func(t *testing.T) {
		headers := http.Header{"Origin": {"http://127.0.0.1:1"}}
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err == nil {
			t.Fatal("expected mismatched port Origin to be rejected")
		}
		if resp == nil || resp.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got resp=%v", resp)
		}
	})
}

func TestWSHeaderAndQueryAuth(t *testing.T) {
	s := New(Config{Addr: ":0", Port: 7654, Hostname: "test-host", Version: "0.1.0"})
	s.SetToken("secret-token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	t.Run("bearer header auth succeeds without query token", func(t *testing.T) {
		headers := http.Header{"Authorization": {"Bearer secret-token"}}
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		if err != nil {
			t.Fatalf("expected Bearer auth to succeed, got err=%v", err)
		}
		defer conn.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("expected 101, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid bearer header takes precedence over valid query token", func(t *testing.T) {
		headers := http.Header{"Authorization": {"Bearer wrong-token"}}
		_, resp, err := websocket.DefaultDialer.Dial(wsURL+"?token=secret-token", headers)
		if err == nil {
			t.Fatal("expected invalid Bearer token to fail")
		}
		if resp == nil || resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got resp=%v", resp)
		}
	})

	t.Run("query token fallback succeeds", func(t *testing.T) {
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL+"?token=secret-token", nil)
		if err != nil {
			t.Fatalf("expected query token auth to succeed, got err=%v", err)
		}
		defer conn.Close()
		if resp.StatusCode != http.StatusSwitchingProtocols {
			t.Fatalf("expected 101, got %d", resp.StatusCode)
		}
	})

	t.Run("empty configured token rejects all", func(t *testing.T) {
		sEmpty := New(Config{Addr: ":0", Port: 7654})
		if sEmpty.validToken("") {
			t.Fatal("expected empty token to be invalid when server token is empty")
		}
	})
}

func TestWSAuthRateLimit(t *testing.T) {
	s := New(Config{
		Addr:           ":0",
		Port:           7654,
		AuthFailLimit:  3,
		AuthFailWindow: time.Minute,
	})
	s.SetToken("valid-token")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	for i := 0; i < 3; i++ {
		_, resp, err := websocket.DefaultDialer.Dial(wsURL+"?token=bad", nil)
		if err == nil {
			t.Fatalf("attempt %d: expected auth failure", i+1)
		}
		if resp == nil || resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got resp=%v", i+1, resp)
		}
	}

	// 4th attempt should be rate limited with 429 even with valid token.
	_, resp, err := websocket.DefaultDialer.Dial(wsURL+"?token=valid-token", nil)
	if err == nil {
		t.Fatal("expected 4th attempt to be rate limited")
	}
	if resp == nil || resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got resp=%v", resp)
	}
}

func TestAuthRateLimiterWindowAndBoundedCleanup(t *testing.T) {
	now := time.Now()
	l := newAuthRateLimiter(2, time.Minute)
	l.now = func() time.Time { return now }

	l.RecordFailure("10.0.0.1")
	l.RecordFailure("10.0.0.1")
	if l.Allow("10.0.0.1") {
		t.Fatal("expected IP to be blocked after reaching limit")
	}

	// Advance past window: should reset.
	now = now.Add(time.Minute + time.Second)
	if !l.Allow("10.0.0.1") {
		t.Fatal("expected IP to be allowed after window expires")
	}

	// Insert more than maxTrackedAuthIPs entries and verify map stays bounded.
	for i := 0; i < maxTrackedAuthIPs+50; i++ {
		l.RecordFailure(fmt.Sprintf("192.0.%d.%d", i/256, i%256))
	}
	if len(l.entries) > maxTrackedAuthIPs {
		t.Fatalf("expected entries <= %d, got %d", maxTrackedAuthIPs, len(l.entries))
	}
}

func TestWSMaxConnections(t *testing.T) {
	s := New(Config{
		Addr:           ":0",
		Port:           7654,
		MaxConnections: 2,
	})
	s.SetToken("tok")
	ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("conn1 failed: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("conn2 failed: %v", err)
	}
	defer conn2.Close()

	// 3rd connection exceeds cap -> 503.
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected 3rd connection to be rejected")
	}
	if resp == nil || resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got resp=%v", resp)
	}

	// Closing conn1 frees a slot.
	conn1.Close()
	deadline := time.Now().Add(time.Second)
	for s.hub.Count() > 1 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	conn3, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("expected conn3 to succeed after disconnect, got err=%v", err)
	}
	defer conn3.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101, got %d", resp.StatusCode)
	}
}

func TestWSReadLimitAndKeepalive(t *testing.T) {
	t.Run("read limit rejects oversized message", func(t *testing.T) {
		s := New(Config{
			Addr:            ":0",
			Port:            7654,
			MaxMessageBytes: 64,
		})
		s.SetToken("tok")
		ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		oversized := `{"type":"text","id":"1","content":"` + strings.Repeat("a", 128) + `"}`
		_ = conn.WriteMessage(websocket.TextMessage, []byte(oversized))

		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		if _, _, err := conn.ReadMessage(); err == nil {
			t.Fatal("expected connection to be closed after oversized message")
		}
	})

	t.Run("keepalive pings keep responsive client alive and drop silent client", func(t *testing.T) {
		s := New(Config{
			Addr:         ":0",
			Port:         7654,
			PingInterval: 40 * time.Millisecond,
			PongWait:     180 * time.Millisecond,
		})
		s.SetToken("tok")
		ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"

		// Responsive client (goroutine reading so gorilla default ping handler replies with pong).
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		pingCount := atomic.Int32{}
		origPingHandler := conn.PingHandler()
		conn.SetPingHandler(func(appData string) error {
			pingCount.Add(1)
			return origPingHandler(appData)
		})

		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()

		// Wait longer than PongWait: connection should still be alive because pongs refresh deadline.
		time.Sleep(240 * time.Millisecond)
		if s.hub.Count() != 1 {
			t.Fatalf("expected responsive client to remain connected, count=%d", s.hub.Count())
		}
		if pingCount.Load() == 0 {
			t.Fatal("expected client to receive at least one ping frame")
		}
		conn.Close()
		<-readDone

		// Silent client that never reads (so never sends pongs) should be dropped by PongWait.
		silentConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer silentConn.Close()

		deadline := time.Now().Add(time.Second)
		for s.hub.Count() > 0 && time.Now().Before(deadline) {
			time.Sleep(15 * time.Millisecond)
		}
		if s.hub.Count() != 0 {
			t.Fatalf("expected silent client to be dropped after PongWait, count=%d", s.hub.Count())
		}
	})
}

type blockingPaster struct {
	started  chan string
	release  chan struct{}
	inFlight atomic.Int32
	maxSeen  atomic.Int32
}

func (p *blockingPaster) Paste(text string) error {
	cur := p.inFlight.Add(1)
	defer p.inFlight.Add(-1)
	for {
		prev := p.maxSeen.Load()
		if cur <= prev || p.maxSeen.CompareAndSwap(prev, cur) {
			break
		}
	}
	if p.started != nil {
		p.started <- text
	}
	if p.release != nil {
		<-p.release
	}
	return nil
}

func (p *blockingPaster) Name() string { return "blocking-mock" }

func TestPasteSerializationAndQueueOverflow(t *testing.T) {
	t.Run("pasteMu serializes concurrent pastes across connections", func(t *testing.T) {
		paster := &blockingPaster{
			started: make(chan string, 10),
			release: make(chan struct{}),
		}
		s := New(Config{Addr: ":0", Port: 7654, Paster: paster})
		s.SetToken("tok")
		ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn1.Close()

		conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn2.Close()

		msg1, _ := json.Marshal(protocol.Inbound{Type: "text", ID: "c1", Content: "from-1"})
		msg2, _ := json.Marshal(protocol.Inbound{Type: "text", ID: "c2", Content: "from-2"})

		if err := conn1.WriteMessage(websocket.TextMessage, msg1); err != nil {
			t.Fatal(err)
		}
		if err := conn2.WriteMessage(websocket.TextMessage, msg2); err != nil {
			t.Fatal(err)
		}

		// Wait for first paste to enter Paste(), give second a chance to collide if unlocked.
		<-paster.started
		time.Sleep(40 * time.Millisecond)
		if got := paster.maxSeen.Load(); got != 1 {
			t.Fatalf("expected at most 1 concurrent paste, saw %d", got)
		}

		// Release first paste, then second should enter.
		paster.release <- struct{}{}
		<-paster.started
		paster.release <- struct{}{}

		var ack1, ack2 protocol.Outbound
		if err := conn1.ReadJSON(&ack1); err != nil {
			t.Fatal(err)
		}
		if err := conn2.ReadJSON(&ack2); err != nil {
			t.Fatal(err)
		}
		if !ack1.OK || !ack2.OK {
			t.Fatalf("expected both acks OK, got ack1=%+v ack2=%+v", ack1, ack2)
		}
		if got := paster.maxSeen.Load(); got != 1 {
			t.Fatalf("expected serialized pastes (maxSeen=1), got %d", got)
		}
	})

	t.Run("overflow returns clean ack error", func(t *testing.T) {
		paster := &blockingPaster{
			started: make(chan string, 10),
			release: make(chan struct{}),
		}
		s := New(Config{
			Addr:             ":0",
			Port:             7654,
			Paster:           paster,
			MaxPendingPastes: 2,
		})
		s.SetToken("tok")
		ts := httptest.NewServer(http.HandlerFunc(s.handleWS))
		defer ts.Close()

		wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "?token=tok"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		// Send 1st frame and wait until it is actively blocked in Paste().
		sendText := func(id, content string) {
			data, _ := json.Marshal(protocol.Inbound{Type: "text", ID: id, Content: content})
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				t.Fatal(err)
			}
		}

		sendText("1", "one")
		<-paster.started

		// 2nd frame fills the remaining pending slot (cap=2).
		sendText("2", "two")
		// 3rd frame exceeds capacity and must receive an overflow ack error immediately.
		sendText("3", "three")

		var overflowAck protocol.Outbound
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		if err := conn.ReadJSON(&overflowAck); err != nil {
			t.Fatalf("expected overflow ack, got err=%v", err)
		}
		if overflowAck.ID != "3" || overflowAck.OK || overflowAck.Message != "too many pending paste requests" {
			t.Fatalf("unexpected overflow ack: %+v", overflowAck)
		}

		// Unblock frames 1 and 2 and verify they both succeed in order.
		paster.release <- struct{}{}
		<-paster.started
		paster.release <- struct{}{}

		var ack1, ack2 protocol.Outbound
		if err := conn.ReadJSON(&ack1); err != nil {
			t.Fatal(err)
		}
		if err := conn.ReadJSON(&ack2); err != nil {
			t.Fatal(err)
		}
		if ack1.ID != "1" || !ack1.OK || ack2.ID != "2" || !ack2.OK {
			t.Fatalf("expected acks 1 and 2 OK, got %+v and %+v", ack1, ack2)
		}
	})
}

func TestSetLogHookConcurrent(t *testing.T) {
	defer SetLogHook(nil)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			SetLogHook(func(string) {})
			SetLogHook(nil)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			logStatus("concurrent test %d", i)
		}
	}()
	wg.Wait()
}
