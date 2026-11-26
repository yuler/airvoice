package relay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/e2e"
)

func newTestServer(t *testing.T, cfg Config) (*Server, string) {
	t.Helper()
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		ts.Close()
		_ = srv.Close(context.Background())
	})
	return srv, ts.URL
}

func wsURL(base string) string {
	return "ws" + strings.TrimPrefix(base, "http") + "/ws"
}

func dialPeer(t *testing.T, baseURL, roomID string, role Role, auth, bearer string, extraHeaders http.Header) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	h := http.Header{}
	if roomID != "" {
		h.Set(HeaderRoom, roomID)
	}
	if role != "" {
		h.Set(HeaderRole, string(role))
	}
	if auth != "" {
		h.Set(HeaderAuth, auth)
	}
	if bearer != "" {
		h.Set("Authorization", "Bearer "+bearer)
	}
	for k, vals := range extraHeaders {
		for _, v := range vals {
			h.Add(k, v)
		}
	}
	return websocket.DefaultDialer.Dial(wsURL(baseURL), h)
}

func mustDialPeer(t *testing.T, baseURL, roomID string, role Role, auth, bearer string, extraHeaders http.Header) *websocket.Conn {
	t.Helper()
	conn, resp, err := dialPeer(t, baseURL, roomID, role, auth, bearer, extraHeaders)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("dial %s failed (status=%d): %v", role, status, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func readJSON(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal(%q): %v", string(data), err)
	}
	return m
}

func newDerivedPair(t *testing.T) e2e.DerivedKeys {
	t.Helper()
	k, err := e2e.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	d, err := e2e.DeriveKeysFromBase64(k)
	if err != nil {
		t.Fatalf("DeriveKeysFromBase64: %v", err)
	}
	return d
}

func TestHealthEndpoint(t *testing.T) {
	_, base := newTestServer(t, DefaultConfig())
	resp, err := http.Get(base + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "ok" {
		t.Fatalf("unexpected /health: status=%d body=%q", resp.StatusCode, string(body))
	}
}

func TestRelayE2ERoundTripAndPeerStatus(t *testing.T) {
	_, base := newTestServer(t, DefaultConfig())
	keys := newDerivedPair(t)

	// Host joins alone -> receives peer_status(online: false).
	host := mustDialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
	msg := readJSON(t, host)
	if msg["type"] != "peer_status" || msg["online"] != false {
		t.Fatalf("expected host peer_status false, got %v", msg)
	}

	// Host sends an E2E frame while client is offline -> receives peer_offline error.
	sidH, _ := e2e.NewSessionID()
	offlineEnv, _ := e2e.Seal(keys.KeyH2C, e2e.DirH2C, keys.RoomID, sidH, 1, []byte(`{"type":"hello","ts":1700000000}`))
	rawOffline, _ := json.Marshal(offlineEnv)
	if err := host.WriteMessage(websocket.TextMessage, rawOffline); err != nil {
		t.Fatalf("host write: %v", err)
	}
	errMsg := readJSON(t, host)
	if errMsg["type"] != "error" || errMsg["code"] != "peer_offline" {
		t.Fatalf("expected peer_offline, got %v", errMsg)
	}

	// Client joins -> both receive peer_status(online: true).
	client := mustDialPeer(t, base, keys.RoomID, RoleClient, keys.ClientAuth, "", nil)
	for name, c := range map[string]*websocket.Conn{"host": host, "client": client} {
		m := readJSON(t, c)
		if m["type"] != "peer_status" || m["online"] != true {
			t.Fatalf("%s expected peer_status true, got %v", name, m)
		}
	}

	// Client sends encrypted hello + text -> Host opens with SessionGuard.
	now := time.Now()
	clientSender, _ := e2e.NewSender(keys.KeyC2H, e2e.DirC2H, keys.RoomID)
	hostGuard, _ := e2e.NewSessionGuard(keys.KeyC2H, e2e.DirC2H, keys.RoomID, e2e.WithClock(func() time.Time { return now }))

	helloPlain, _ := json.Marshal(map[string]any{"type": "hello", "device": "iPhone", "ts": now.Unix()})
	helloEnv, _ := clientSender.Seal(helloPlain)
	helloWire, _ := json.Marshal(helloEnv)
	if err := client.WriteMessage(websocket.TextMessage, helloWire); err != nil {
		t.Fatalf("client write hello: %v", err)
	}

	_ = host.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, gotHelloWire, err := host.ReadMessage()
	if err != nil {
		t.Fatalf("host read hello: %v", err)
	}
	var gotHelloEnv e2e.Envelope
	if err := json.Unmarshal(gotHelloWire, &gotHelloEnv); err != nil {
		t.Fatalf("unmarshal hello env: %v", err)
	}
	decryptedHello, err := hostGuard.Open(gotHelloEnv)
	if err != nil || string(decryptedHello) != string(helloPlain) {
		t.Fatalf("host open hello: err=%v got=%q", err, decryptedHello)
	}

	// Host disconnects -> Client receives peer_status(online: false).
	_ = host.Close()
	offlineNotice := readJSON(t, client)
	if offlineNotice["type"] != "peer_status" || offlineNotice["online"] != false {
		t.Fatalf("expected client peer_status false after host close, got %v", offlineNotice)
	}
}

func TestSlotOwnershipAndReplacement(t *testing.T) {
	_, base := newTestServer(t, DefaultConfig())
	keys := newDerivedPair(t)
	otherKeys := newDerivedPair(t)

	host1 := mustDialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
	_ = readJSON(t, host1) // peer_status false

	// Attacker with different hostAuth gets HTTP 403.
	_, resp, err := dialPeer(t, base, keys.RoomID, RoleHost, otherKeys.HostAuth, "", nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for mismatched hostAuth, got err=%v resp=%v", err, resp)
	}

	// Legitimate host with matching hostAuth replaces host1 (4001 Replaced).
	host2 := mustDialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
	_ = readJSON(t, host2) // peer_status false

	_ = host1.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err = host1.ReadMessage()
	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != CloseCodeReplaced {
		t.Fatalf("expected close code %d on replaced host1, got %v", CloseCodeReplaced, err)
	}
}

func TestHeaderOnlyAuthAndBearerKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AuthKey = "super-secret-relay-key"
	_, base := newTestServer(t, cfg)
	keys := newDerivedPair(t)

	// Missing Bearer -> 401
	_, resp, err := dialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without bearer, got err=%v resp=%v", err, resp)
	}

	// Wrong Bearer -> 401
	_, resp, err = dialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, "wrong-key", nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong bearer, got err=%v resp=%v", err, resp)
	}

	// Query-string credentials must NOT work.
	qURL := wsURL(base) + "?room=" + keys.RoomID + "&role=host&auth=" + keys.HostAuth + "&token=" + cfg.AuthKey
	_, resp, err = websocket.DefaultDialer.Dial(qURL, nil)
	if err == nil || resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for query-string auth, got err=%v resp=%v", err, resp)
	}

	// Valid Bearer + headers -> 101 Switching Protocols
	host := mustDialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, cfg.AuthKey, nil)
	msg := readJSON(t, host)
	if msg["type"] != "peer_status" {
		t.Fatalf("expected peer_status, got %v", msg)
	}
}

func TestMessageTooLargeRateAndDailyQuota(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxMessageBytes = 512
	cfg.RatePerMinute = 60
	cfg.RateBurst = 2
	cfg.DailyMessageLimit = 3
	_, base := newTestServer(t, cfg)
	keys := newDerivedPair(t)

	host := mustDialPeer(t, base, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
	_ = readJSON(t, host)
	client := mustDialPeer(t, base, keys.RoomID, RoleClient, keys.ClientAuth, "", nil)
	_ = readJSON(t, host)
	_ = readJSON(t, client)

	// 1. Oversized frame -> message_too_large error, connection remains alive.
	bigFrame := `{"type":"e2e","v":1,"ct":"` + strings.Repeat("a", 600) + `"}`
	if err := client.WriteMessage(websocket.TextMessage, []byte(bigFrame)); err != nil {
		t.Fatalf("write oversized: %v", err)
	}
	errMsg := readJSON(t, client)
	if errMsg["type"] != "error" || errMsg["code"] != "message_too_large" {
		t.Fatalf("expected message_too_large, got %v", errMsg)
	}

	// 2. Burst of 2 succeeds, 3rd hits rate_limited.
	validFrame := []byte(`{"type":"e2e","v":1,"sid":"AAAAAAAAAAA","seq":1,"iv":"AAAAAAAAAAAAAAAA","ct":"AAAA"}`)
	for i := 0; i < 2; i++ {
		if err := client.WriteMessage(websocket.TextMessage, validFrame); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		_ = readJSON(t, host)
	}
	if err := client.WriteMessage(websocket.TextMessage, validFrame); err != nil {
		t.Fatalf("write 3rd: %v", err)
	}
	rateMsg := readJSON(t, client)
	if rateMsg["type"] != "error" || rateMsg["code"] != "rate_limited" {
		t.Fatalf("expected rate_limited, got %v", rateMsg)
	}

	// 3. Test daily quota with burst > daily limit on a fresh server.
	cfg2 := DefaultConfig()
	cfg2.RatePerMinute = 0 // disable rate limit
	cfg2.DailyMessageLimit = 2
	_, base2 := newTestServer(t, cfg2)
	keys2 := newDerivedPair(t)
	h2 := mustDialPeer(t, base2, keys2.RoomID, RoleHost, keys2.HostAuth, "", nil)
	_ = readJSON(t, h2)
	c2 := mustDialPeer(t, base2, keys2.RoomID, RoleClient, keys2.ClientAuth, "", nil)
	_ = readJSON(t, h2)
	_ = readJSON(t, c2)

	for i := 0; i < 2; i++ {
		_ = c2.WriteMessage(websocket.TextMessage, validFrame)
		_ = readJSON(t, h2)
	}
	_ = c2.WriteMessage(websocket.TextMessage, validFrame)
	quotaMsg := readJSON(t, c2)
	if quotaMsg["type"] != "error" || quotaMsg["code"] != "quota_exceeded" {
		t.Fatalf("expected quota_exceeded, got %v", quotaMsg)
	}
}

func TestPerIPAndGlobalRoomLimitsAndTrustProxy(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRoomsPerIP = 2
	cfg.MaxGlobalRooms = 3
	cfg.TrustProxy = true
	_, base := newTestServer(t, cfg)

	ip1Header := http.Header{"X-Forwarded-For": []string{"9.9.9.9, 203.0.113.10"}}
	ip2Header := http.Header{"X-Forwarded-For": []string{"203.0.113.20"}}

	// Same IP opens 2 rooms -> OK. Host + Client in room 1 from same IP still counts as 1 room.
	k1 := newDerivedPair(t)
	k2 := newDerivedPair(t)
	k3 := newDerivedPair(t)
	k4 := newDerivedPair(t)

	_ = mustDialPeer(t, base, k1.RoomID, RoleHost, k1.HostAuth, "", ip1Header)
	_ = mustDialPeer(t, base, k1.RoomID, RoleClient, k1.ClientAuth, "", ip1Header)
	_ = mustDialPeer(t, base, k2.RoomID, RoleHost, k2.HostAuth, "", ip1Header)

	// 3rd room from same real IP (even with different spoofed left XFF!) -> 429 room_limit_reached.
	spoofedHeader := http.Header{"X-Forwarded-For": []string{"1.1.1.1, 203.0.113.10"}}
	_, resp, err := dialPeer(t, base, k3.RoomID, RoleHost, k3.HostAuth, "", spoofedHeader)
	if err == nil || resp == nil || resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for per-IP room cap, got err=%v resp=%v", err, resp)
	}

	// Different IP opens 3rd global room -> OK.
	_ = mustDialPeer(t, base, k3.RoomID, RoleHost, k3.HostAuth, "", ip2Header)

	// 4th global room -> 503 room_limit_reached.
	_, resp, err = dialPeer(t, base, k4.RoomID, RoleHost, k4.HostAuth, "", ip2Header)
	if err == nil || resp == nil || resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for global room cap, got err=%v resp=%v", err, resp)
	}
}

func TestOrphanAndSessionIdleEviction(t *testing.T) {
	var mu sync.Mutex
	fakeNow := time.Unix(1700000000, 0)
	nowFn := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return fakeNow
	}
	advance := func(d time.Duration) {
		mu.Lock()
		fakeNow = fakeNow.Add(d)
		mu.Unlock()
	}

	cfg := DefaultConfig()
	cfg.OrphanIdleTimeout = 5 * time.Minute
	cfg.SessionIdleTimeout = 10 * time.Minute
	cfg.EvictInterval = 20 * time.Millisecond
	cfg.Now = nowFn
	srv, base := newTestServer(t, cfg)

	// 1. Orphan eviction after 5 minutes.
	k1 := newDerivedPair(t)
	orphan := mustDialPeer(t, base, k1.RoomID, RoleHost, k1.HostAuth, "", nil)
	_ = readJSON(t, orphan)

	advance(6 * time.Minute)
	_ = orphan.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err := orphan.ReadMessage()
	var closeErr *websocket.CloseError
	if !errors.As(err, &closeErr) || closeErr.Code != CloseCodeIdleTimeout {
		t.Fatalf("expected orphan close code %d, got %v", CloseCodeIdleTimeout, err)
	}

	// 2. Paired session eviction after 10 minutes of inactivity.
	k2 := newDerivedPair(t)
	h := mustDialPeer(t, base, k2.RoomID, RoleHost, k2.HostAuth, "", nil)
	_ = readJSON(t, h)
	c := mustDialPeer(t, base, k2.RoomID, RoleClient, k2.ClientAuth, "", nil)
	_ = readJSON(t, h)
	_ = readJSON(t, c)

	advance(11 * time.Minute)
	for name, conn := range map[string]*websocket.Conn{"host": h, "client": c} {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := conn.ReadMessage()
		var ce *websocket.CloseError
		if !errors.As(err, &ce) || ce.Code != CloseCodeIdleTimeout {
			t.Fatalf("expected session close code %d on %s, got %v", CloseCodeIdleTimeout, name, err)
		}
	}

	if srv.Hub().RoomCount() != 0 {
		t.Fatalf("expected 0 rooms after eviction, got %d", srv.Hub().RoomCount())
	}
}

func TestConfigFromEnvAndFlags(t *testing.T) {
	t.Setenv("AIRVOICE_RELAY_ADDR", "127.0.0.1:9999")
	t.Setenv("AIRVOICE_RELAY_AUTH_KEY", "secret")
	t.Setenv("AIRVOICE_RELAY_MAX_MESSAGE_BYTES", "32768")
	t.Setenv("AIRVOICE_RELAY_RATE_PER_MINUTE", "0")
	t.Setenv("AIRVOICE_RELAY_DAILY_MESSAGE_LIMIT", "0")
	t.Setenv("AIRVOICE_RELAY_ORPHAN_IDLE_TIMEOUT", "5m")
	t.Setenv("AIRVOICE_RELAY_TRUST_PROXY", "true")

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}
	if cfg.Addr != "127.0.0.1:9999" || cfg.AuthKey != "secret" || cfg.MaxMessageBytes != 32768 ||
		cfg.RatePerMinute != 0 || cfg.DailyMessageLimit != 0 || cfg.OrphanIdleTimeout != 5*time.Minute || !cfg.TrustProxy {
		t.Fatalf("unexpected env config: %+v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestLimiterUTCRolloverAndPerIPDailyCap(t *testing.T) {
	now := time.Date(2026, 11, 26, 23, 59, 0, 0, time.UTC)
	cfg := DefaultConfig()
	cfg.RatePerMinute = 0
	cfg.DailyMessageLimit = 2
	cfg.Now = func() time.Time { return now }

	l := newLimiter(cfg.normalized())

	// Same IP across two different rooms shares the per-IP daily cap (2).
	if err := l.AllowForward("room-1", "203.0.113.5"); err != nil {
		t.Fatalf("msg 1 failed: %v", err)
	}
	if err := l.AllowForward("room-2", "203.0.113.5"); err != nil {
		t.Fatalf("msg 2 failed: %v", err)
	}
	if err := l.AllowForward("room-3", "203.0.113.5"); !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("msg 3 expected ErrQuotaExceeded, got %v", err)
	}

	// Cross UTC midnight -> daily counters roll over.
	now = now.Add(2 * time.Minute)
	if err := l.AllowForward("room-3", "203.0.113.5"); err != nil {
		t.Fatalf("after UTC rollover expected allow, got %v", err)
	}
}

// TestSmokeLiveRelay verifies a deployed relay when AIRVOICE_RELAY_SMOKE_URL is set
// (e.g. wss://relay.airvoice.yuler.cc/ws or ws://127.0.0.1:18787/ws).
func TestSmokeLiveRelay(t *testing.T) {
	wsTarget := strings.TrimSpace(os.Getenv("AIRVOICE_RELAY_SMOKE_URL"))
	if wsTarget == "" {
		t.Skip("AIRVOICE_RELAY_SMOKE_URL not set")
	}
	bearer := strings.TrimSpace(os.Getenv("AIRVOICE_RELAY_SMOKE_AUTH_KEY"))

	keys := newDerivedPair(t)
	otherKeys := newDerivedPair(t)

	dial := func(role Role, auth string) (*websocket.Conn, *http.Response, error) {
		h := http.Header{}
		h.Set(HeaderRoom, keys.RoomID)
		h.Set(HeaderRole, string(role))
		h.Set(HeaderAuth, auth)
		if bearer != "" {
			h.Set("Authorization", "Bearer "+bearer)
		}
		return websocket.DefaultDialer.Dial(wsTarget, h)
	}

	host, resp, err := dial(RoleHost, keys.HostAuth)
	if err != nil {
		t.Fatalf("dial host (%v): %v", resp, err)
	}
	defer host.Close()
	if m := readJSON(t, host); m["type"] != "peer_status" || m["online"] != false {
		t.Fatalf("expected host peer_status false, got %v", m)
	}

	// Verify slot hijack rejection (403)
	if _, hResp, hErr := dial(RoleHost, otherKeys.HostAuth); hErr == nil || hResp == nil || hResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 on host hijack, got err=%v resp=%v", hErr, hResp)
	}

	client, resp, err := dial(RoleClient, keys.ClientAuth)
	if err != nil {
		t.Fatalf("dial client (%v): %v", resp, err)
	}
	defer client.Close()

	for name, c := range map[string]*websocket.Conn{"host": host, "client": client} {
		if m := readJSON(t, c); m["type"] != "peer_status" || m["online"] != true {
			t.Fatalf("%s expected peer_status true, got %v", name, m)
		}
	}

	now := time.Now().Unix()
	cSender, _ := e2e.NewSender(keys.KeyC2H, e2e.DirC2H, keys.RoomID)
	hSender, _ := e2e.NewSender(keys.KeyH2C, e2e.DirH2C, keys.RoomID)
	hGuard, _ := e2e.NewSessionGuard(keys.KeyC2H, e2e.DirC2H, keys.RoomID)
	cGuard, _ := e2e.NewSessionGuard(keys.KeyH2C, e2e.DirH2C, keys.RoomID)

	send := func(conn *websocket.Conn, s *e2e.Sender, payload []byte) {
		env, err := s.Seal(payload)
		if err != nil {
			t.Fatalf("seal: %v", err)
		}
		raw, _ := json.Marshal(env)
		if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	recv := func(conn *websocket.Conn, g *e2e.SessionGuard) string {
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var env e2e.Envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		pt, err := g.Open(env)
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		return string(pt)
	}

	cHello, _ := json.Marshal(map[string]any{"type": "hello", "device": "smoke-phone", "ts": now})
	send(client, cSender, cHello)
	if got := recv(host, hGuard); got != string(cHello) {
		t.Fatalf("hello mismatch: %q", got)
	}

	wantText := `{"type":"text","id":"smoke-1","content":"跨网中继 E2E 冒烟测试 🎙️"}`
	send(client, cSender, []byte(wantText))
	if got := recv(host, hGuard); got != wantText {
		t.Fatalf("text mismatch: %q", got)
	}

	hHello, _ := json.Marshal(map[string]any{"type": "hello", "host": "smoke-pc", "version": "0.5.0", "ts": now})
	send(host, hSender, hHello)
	if got := recv(client, cGuard); got != string(hHello) {
		t.Fatalf("host hello mismatch: %q", got)
	}

	wantAck := `{"type":"ack","id":"smoke-1","ok":true}`
	send(host, hSender, []byte(wantAck))
	if got := recv(client, cGuard); got != wantAck {
		t.Fatalf("ack mismatch: %q", got)
	}
}
