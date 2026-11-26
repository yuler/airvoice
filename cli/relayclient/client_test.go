package relayclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/e2e"
	"github.com/yuler/airvoice/cli/protocol"
	"github.com/yuler/airvoice/relay"
)

type stubPaster struct {
	mu     sync.Mutex
	pasted []string
	err    error
	ch     chan string
}

func newStubPaster() *stubPaster {
	return &stubPaster{ch: make(chan string, 16)}
}

func (p *stubPaster) Paste(text string) error {
	p.mu.Lock()
	err := p.err
	if err == nil {
		p.pasted = append(p.pasted, text)
	}
	p.mu.Unlock()
	if err == nil {
		select {
		case p.ch <- text:
		default:
		}
	}
	return err
}

func (p *stubPaster) Name() string { return "stub" }

func (p *stubPaster) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pasted)
}

func TestValidateRelayURL(t *testing.T) {
	valid := []string{
		"wss://relay.airvoice.yuler.cc/ws",
		"wss://example.com:8443/ws",
		"ws://localhost:8787/ws",
		"ws://127.0.0.1:8787/ws",
		"ws://[::1]:8787/ws",
		"ws://192.168.1.50:8787/ws",
		"ws://10.0.0.5:8787/ws",
		"ws://100.64.1.2:8787/ws",
		"ws://100.127.255.1:8787/ws",
		"ws://myhost.local:8787/ws",
		"ws://[fd00::1]:8787/ws",
	}
	for _, u := range valid {
		if err := ValidateRelayURL(u); err != nil {
			t.Errorf("ValidateRelayURL(%q) unexpected err: %v", u, err)
		}
	}

	invalid := []string{
		"",
		"http://relay.airvoice.yuler.cc/ws",
		"https://relay.airvoice.yuler.cc/ws",
		"ws://relay.airvoice.yuler.cc/ws",
		"ws://8.8.8.8:8787/ws",
		"ws://100.63.0.1:8787/ws",
		"wss:///ws",
	}
	for _, u := range invalid {
		if err := ValidateRelayURL(u); err == nil {
			t.Errorf("ValidateRelayURL(%q) expected error, got nil", u)
		}
	}
}

func startTestRelay(t *testing.T, authKey string) (string, *relay.Server, *httptest.Server) {
	t.Helper()
	rs, err := relay.NewServer(relay.Config{
		AuthKey: authKey,
	})
	if err != nil {
		t.Fatalf("relay.NewServer: %v", err)
	}
	ts := httptest.NewServer(rs.Handler())
	t.Cleanup(func() {
		ts.Close()
		_ = rs.Close(context.Background())
	})
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	return wsURL, rs, ts
}

type simPhone struct {
	t      *testing.T
	conn   *websocket.Conn
	keys   e2e.DerivedKeys
	sender *e2e.Sender
	guard  *e2e.SessionGuard
}

func dialSimPhone(t *testing.T, wsURL, relayKey, authKey string) *simPhone {
	t.Helper()
	keys, err := e2e.DeriveKeysFromBase64(relayKey)
	if err != nil {
		t.Fatalf("DeriveKeys: %v", err)
	}
	headers := http.Header{}
	headers.Set("X-Airvoice-Room", keys.RoomID)
	headers.Set("X-Airvoice-Role", "client")
	headers.Set("X-Airvoice-Auth", keys.ClientAuth)
	if authKey != "" {
		headers.Set("Authorization", "Bearer "+authKey)
	}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		t.Fatalf("phone dial: %v", err)
	}
	sender, err := e2e.NewSender(keys.KeyC2H, e2e.DirC2H, keys.RoomID)
	if err != nil {
		t.Fatalf("phone NewSender: %v", err)
	}
	guard, err := e2e.NewSessionGuard(keys.KeyH2C, e2e.DirH2C, keys.RoomID)
	if err != nil {
		t.Fatalf("phone NewSessionGuard: %v", err)
	}
	return &simPhone{t: t, conn: conn, keys: keys, sender: sender, guard: guard}
}

func (p *simPhone) close() { _ = p.conn.Close() }

func (p *simPhone) waitPeerOnline() {
	p.t.Helper()
	_ = p.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			p.t.Fatalf("phone waitPeerOnline read: %v", err)
		}
		var hdr relayFrameHeader
		if err := json.Unmarshal(data, &hdr); err == nil && hdr.Type == "peer_status" && hdr.Online {
			return
		}
	}
}

func (p *simPhone) sendEncrypted(msg protocol.Inbound) e2e.Envelope {
	p.t.Helper()
	raw, err := json.Marshal(msg)
	if err != nil {
		p.t.Fatalf("marshal inbound: %v", err)
	}
	env, err := p.sender.Seal(raw)
	if err != nil {
		p.t.Fatalf("phone seal: %v", err)
	}
	wire, _ := json.Marshal(env)
	if err := p.conn.WriteMessage(websocket.TextMessage, wire); err != nil {
		p.t.Fatalf("phone write: %v", err)
	}
	return env
}

func (p *simPhone) readEncrypted() protocol.Outbound {
	p.t.Helper()
	_ = p.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			p.t.Fatalf("phone readEncrypted: %v", err)
		}
		var hdr relayFrameHeader
		_ = json.Unmarshal(data, &hdr)
		if hdr.Type != e2e.EnvelopeType {
			continue
		}
		var env e2e.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			p.t.Fatalf("phone unmarshal env: %v", err)
		}
		pt, err := p.guard.Open(env)
		if err != nil {
			p.t.Fatalf("phone guard.Open failed: %v", err)
		}
		var out protocol.Outbound
		if err := json.Unmarshal(pt, &out); err != nil {
			p.t.Fatalf("phone unmarshal outbound: %v", err)
		}
		return out
	}
}

func TestClientEndToEndEncryptedRoundTripAndReconnect(t *testing.T) {
	wsURL, _, _ := startTestRelay(t, "private-relay-secret")
	relayKey, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	paster := newStubPaster()
	peerEvents := make(chan bool, 8)
	textEvents := make(chan string, 8)

	client, err := New(Config{
		RelayURL:       wsURL,
		RelayKey:       relayKey,
		AuthKey:        "private-relay-secret",
		Hostname:       "dev-box",
		Version:        "0.6.0",
		Paster:         paster,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     200 * time.Millisecond,
		OnPeerStatus: func(online bool, device string) {
			peerEvents <- online
		},
		OnTextReceived: func(content, device string) {
			textEvents <- content
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	client.Start()
	defer func() { _ = client.Close() }()

	// Session 1: Phone joins, handshakes, sends text & ping.
	phone1 := dialSimPhone(t, wsURL, relayKey, "private-relay-secret")
	phone1.waitPeerOnline()

	phone1.sendEncrypted(protocol.Inbound{
		Type:   "hello",
		Device: "iPhone 16",
		App:    "Airvoice iOS",
		TS:     time.Now().Unix(),
	})
	helloReply := phone1.readEncrypted()
	if helloReply.Type != "hello" || helloReply.Host != "dev-box" || helloReply.Version != "0.6.0" || helloReply.TS == 0 {
		t.Fatalf("unexpected hello reply: %+v", helloReply)
	}

	select {
	case online := <-peerEvents:
		if !online {
			t.Fatal("expected peer online=true")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for peer online callback")
	}

	phone1.sendEncrypted(protocol.Inbound{
		Type:    "text",
		ID:      "msg-1",
		Content: "你好 relay E2E",
		Device:  "iPhone 16",
	})
	ack := phone1.readEncrypted()
	if ack.Type != "ack" || ack.ID != "msg-1" || !ack.OK {
		t.Fatalf("unexpected ack: %+v", ack)
	}

	select {
	case got := <-paster.ch:
		if got != "你好 relay E2E" {
			t.Fatalf("pasted %q, want %q", got, "你好 relay E2E")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for paste")
	}

	phone1.sendEncrypted(protocol.Inbound{Type: "ping"})
	pong := phone1.readEncrypted()
	if pong.Type != "pong" {
		t.Fatalf("expected pong, got %+v", pong)
	}

	// Disconnect phone1 -> host should observe offline.
	phone1.close()
	select {
	case online := <-peerEvents:
		if online {
			t.Fatal("expected peer online=false on disconnect")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for peer offline callback")
	}

	// Session 2: Fresh phone connects; host must reset outgoing Sender to seq==1.
	phone2 := dialSimPhone(t, wsURL, relayKey, "private-relay-secret")
	defer phone2.close()
	phone2.waitPeerOnline()

	phone2.sendEncrypted(protocol.Inbound{
		Type:   "hello",
		Device: "Pixel 9",
		TS:     time.Now().Unix(),
	})
	hello2 := phone2.readEncrypted()
	if hello2.Type != "hello" || hello2.TS == 0 {
		t.Fatalf("session 2 hello failed: %+v", hello2)
	}
	phone2.sendEncrypted(protocol.Inbound{Type: "text", ID: "msg-2", Content: "second session"})
	ack2 := phone2.readEncrypted()
	if ack2.Type != "ack" || ack2.ID != "msg-2" || !ack2.OK {
		t.Fatalf("session 2 ack failed: %+v", ack2)
	}
}

func TestClientRejectsReplayAndTamperedFrames(t *testing.T) {
	wsURL, _, _ := startTestRelay(t, "")
	relayKey, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	paster := newStubPaster()
	client, err := New(Config{
		RelayURL:       wsURL,
		RelayKey:       relayKey,
		Hostname:       "host",
		Version:        "0.6.0",
		Paster:         paster,
		InitialBackoff: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Start()
	defer func() { _ = client.Close() }()

	phone := dialSimPhone(t, wsURL, relayKey, "")
	defer phone.close()
	phone.waitPeerOnline()

	// Handshake + valid text.
	phone.sendEncrypted(protocol.Inbound{Type: "hello", Device: "TestPhone", TS: time.Now().Unix()})
	_ = phone.readEncrypted()
	env := phone.sendEncrypted(protocol.Inbound{Type: "text", ID: "m1", Content: "legit"})
	ack := phone.readEncrypted()
	if !ack.OK {
		t.Fatalf("expected ok ack, got %+v", ack)
	}

	// Replay the exact same envelope -> must be dropped and session retired.
	wire, _ := json.Marshal(env)
	if err := phone.conn.WriteMessage(websocket.TextMessage, wire); err != nil {
		t.Fatal(err)
	}
	// Follow up with a new seq on the same (now retired) sid -> must also be dropped.
	phone.sendEncrypted(protocol.Inbound{Type: "text", ID: "m2", Content: "should-drop"})

	time.Sleep(150 * time.Millisecond)
	if paster.count() != 1 {
		t.Fatalf("expected only 1 paste, got %d", paster.count())
	}

	// A fresh session ID with valid hello should recover cleanly.
	if err := phone.sender.Reset(); err != nil {
		t.Fatal(err)
	}
	phone.sendEncrypted(protocol.Inbound{Type: "hello", Device: "TestPhone", TS: time.Now().Unix()})
	helloAgain := phone.readEncrypted()
	if helloAgain.Type != "hello" {
		t.Fatalf("expected recovery hello, got %+v", helloAgain)
	}
}

func TestClientStopsImmediatelyOnAuthError(t *testing.T) {
	wsURL, _, _ := startTestRelay(t, "correct-secret")
	relayKey, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	var gotStatus int
	client, err := New(Config{
		RelayURL:       wsURL,
		RelayKey:       relayKey,
		AuthKey:        "wrong-secret",
		InitialBackoff: 50 * time.Millisecond,
		OnAuthError: func(statusCode int, message string) {
			gotStatus = statusCode
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	runErr := client.Run(ctx)
	var authErr *AuthError
	if !errors.As(runErr, &authErr) {
		t.Fatalf("expected *AuthError, got %v", runErr)
	}
	if authErr.StatusCode != http.StatusUnauthorized || gotStatus != http.StatusUnauthorized {
		t.Fatalf("status=%d callback=%d, want 401", authErr.StatusCode, gotStatus)
	}
}

func TestClientUpdateRelayKeySwitchesRoom(t *testing.T) {
	wsURL, _, _ := startTestRelay(t, "")
	key1, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	key2, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	paster := newStubPaster()
	client, err := New(Config{
		RelayURL:       wsURL,
		RelayKey:       key1,
		Hostname:       "host",
		Paster:         paster,
		InitialBackoff: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Start()
	defer func() { _ = client.Close() }()

	// Rotate to key2; phone using key2 must pair and paste.
	time.Sleep(80 * time.Millisecond)
	if err := client.UpdateRelayKey(key2); err != nil {
		t.Fatalf("UpdateRelayKey: %v", err)
	}

	phone := dialSimPhone(t, wsURL, key2, "")
	defer phone.close()
	phone.waitPeerOnline()

	phone.sendEncrypted(protocol.Inbound{Type: "hello", Device: "NewPhone", TS: time.Now().Unix()})
	_ = phone.readEncrypted()
	phone.sendEncrypted(protocol.Inbound{Type: "text", ID: "rot-1", Content: "after rotation"})
	ack := phone.readEncrypted()
	if !ack.OK {
		t.Fatalf("expected ok ack after rotation, got %+v", ack)
	}
}

func TestClientSurfacesRelayErrorFrames(t *testing.T) {
	upgrader := websocket.Upgrader{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"error","code":"rate_limited","message":"slow down"}`))
		time.Sleep(300 * time.Millisecond)
	}))
	defer ts.Close()

	key, _ := e2e.GenerateKey()
	errCh := make(chan string, 2)
	client, err := New(Config{
		RelayURL:       "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws",
		RelayKey:       key,
		InitialBackoff: 50 * time.Millisecond,
		OnRelayError: func(code, message string) {
			errCh <- code + ":" + message
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Start()
	defer func() { _ = client.Close() }()

	select {
	case got := <-errCh:
		if got != "rate_limited:slow down" {
			t.Fatalf("got %q, want rate_limited:slow down", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for relay error callback")
	}
}

type gatePaster struct {
	mu      sync.Mutex
	gateFor string
	started chan struct{}
	release chan struct{}
	done    chan struct{}
}

func (p *gatePaster) Paste(text string) error {
	p.mu.Lock()
	gated := text == p.gateFor
	p.mu.Unlock()
	if gated {
		close(p.started)
		<-p.release
		close(p.done)
	}
	return nil
}

func (p *gatePaster) Name() string { return "gate" }

func TestClientDropsStaleAckAfterSessionReset(t *testing.T) {
	wsURL, _, _ := startTestRelay(t, "")
	relayKey, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	paster := &gatePaster{
		gateFor: "stale-paste",
		started: make(chan struct{}),
		release: make(chan struct{}),
		done:    make(chan struct{}),
	}
	peerEvents := make(chan bool, 8)
	client, err := New(Config{
		RelayURL:       wsURL,
		RelayKey:       relayKey,
		Hostname:       "dev-box",
		Version:        "0.6.0",
		Paster:         paster,
		InitialBackoff: 50 * time.Millisecond,
		OnPeerStatus: func(online bool, _ string) {
			peerEvents <- online
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Start()
	defer func() { _ = client.Close() }()

	phone1 := dialSimPhone(t, wsURL, relayKey, "")
	phone1.waitPeerOnline()
	phone1.sendEncrypted(protocol.Inbound{Type: "hello", Device: "Phone1", TS: time.Now().Unix()})
	if reply := phone1.readEncrypted(); reply.Type != "hello" {
		t.Fatalf("phone1 expected hello, got %+v", reply)
	}
	<-peerEvents

	// Start a paste on phone1 and hold it in flight across peer disconnect/reconnect.
	phone1.sendEncrypted(protocol.Inbound{Type: "text", ID: "stale-1", Content: "stale-paste"})
	select {
	case <-paster.started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for stale paste to start")
	}

	phone1.close()
	select {
	case online := <-peerEvents:
		if online {
			t.Fatal("expected peer offline after phone1 closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for peer offline")
	}

	phone2 := dialSimPhone(t, wsURL, relayKey, "")
	defer phone2.close()
	phone2.waitPeerOnline()

	// Release phone1's in-flight paste; its stale ack must be dropped so it
	// does not consume seq=1 of phone2's fresh session.
	close(paster.release)
	select {
	case <-paster.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for stale paste to finish")
	}

	phone2.sendEncrypted(protocol.Inbound{Type: "hello", Device: "Phone2", TS: time.Now().Unix()})
	if reply := phone2.readEncrypted(); reply.Type != "hello" {
		t.Fatalf("phone2 expected hello on seq=1, got %+v", reply)
	}

	phone2.sendEncrypted(protocol.Inbound{Type: "text", ID: "fresh-1", Content: "fresh-paste"})
	ack := phone2.readEncrypted()
	if ack.Type != "ack" || ack.ID != "fresh-1" || !ack.OK {
		t.Fatalf("phone2 expected fresh ack, got %+v", ack)
	}
}

func TestClientConcurrentWritesPreserveMonotonicSequence(t *testing.T) {
	wsURL, _, _ := startTestRelay(t, "")
	relayKey, err := e2e.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	// Keep each in-flight wave (waveSize * 2 = 16 frames) well below the
	// relay's per-peer send buffer (32) so CI under -race never triggers
	// slow-consumer eviction while still exercising concurrent ack+pong writes.
	const (
		waves    = 3
		waveSize = 8
	)
	paster := newStubPaster()
	client, err := New(Config{
		RelayURL:         wsURL,
		RelayKey:         relayKey,
		Hostname:         "dev-box",
		Version:          "0.6.0",
		Paster:           paster,
		MaxPendingPastes: waveSize,
		InitialBackoff:   50 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	client.Start()
	defer func() { _ = client.Close() }()

	phone := dialSimPhone(t, wsURL, relayKey, "")
	defer phone.close()
	phone.waitPeerOnline()

	phone.sendEncrypted(protocol.Inbound{Type: "hello", Device: "Phone", TS: time.Now().Unix()})
	if reply := phone.readEncrypted(); reply.Type != "hello" {
		t.Fatalf("expected hello, got %+v", reply)
	}

	acks := 0
	pongs := 0
	for w := 0; w < waves; w++ {
		for i := 0; i < waveSize; i++ {
			idx := w*waveSize + i
			phone.sendEncrypted(protocol.Inbound{
				Type:    "text",
				ID:      fmt.Sprintf("msg-%d", idx),
				Content: fmt.Sprintf("hello %d", idx),
			})
			phone.sendEncrypted(protocol.Inbound{Type: "ping"})
		}
		for i := 0; i < waveSize*2; i++ {
			out := phone.readEncrypted()
			switch out.Type {
			case "ack":
				if !out.OK {
					t.Fatalf("unexpected failed ack: %+v", out)
				}
				acks++
			case "pong":
				pongs++
			default:
				t.Fatalf("unexpected outbound message: %+v", out)
			}
		}
	}
	want := waves * waveSize
	if acks != want || pongs != want {
		t.Fatalf("got acks=%d pongs=%d, want %d/%d", acks, pongs, want, want)
	}
}
