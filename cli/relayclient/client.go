package relayclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/e2e"
	"github.com/yuler/airvoice/cli/paste"
	"github.com/yuler/airvoice/cli/protocol"
	"github.com/yuler/airvoice/cli/server"
)

const (
	headerRoom = "X-Airvoice-Room"
	headerRole = "X-Airvoice-Role"
	headerAuth = "X-Airvoice-Auth"
	roleHost   = "host"

	defaultInitialBackoff = 2 * time.Second
	defaultMaxBackoff     = 30 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultReadTimeout    = 90 * time.Second
	defaultMaxReadBytes   = 128 * 1024
)

// RelayState describes the host relay client lifecycle state.
type RelayState string

const (
	StateDisconnected RelayState = "disconnected"
	StateConnecting   RelayState = "connecting"
	StateWaiting      RelayState = "waiting"
	StateConnected    RelayState = "connected"
	StateError        RelayState = "error"
)

// AuthError indicates the relay rejected host authentication (HTTP 401/403).
type AuthError struct {
	StatusCode int
	Message    string
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("relay auth failed (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("relay auth failed (HTTP %d)", e.StatusCode)
}

// Config configures the outbound Relay Host WebSocket client.
type Config struct {
	RelayURL         string
	RelayKey         string
	AuthKey          string
	Hostname         string
	Version          string
	Paster           paste.Paster
	MaxPendingPastes int

	Dialer         *websocket.Dialer
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	Now            func() time.Time

	OnStateChange  func(state RelayState, detail string)
	OnPeerStatus   func(online bool, device string)
	OnTextReceived func(content, device string)
	OnRelayError   func(code, message string)
	OnAuthError    func(statusCode int, message string)
}

// Client maintains an outbound E2E-encrypted WebSocket connection to an
// airvoice-relay server in the "host" role.
type Client struct {
	cfg     Config
	pasteMu sync.Mutex

	mu           sync.Mutex
	relayKey     string
	keys         e2e.DerivedKeys
	activeConn   *websocket.Conn
	reconnectNow bool
	closed       bool

	ctx       context.Context
	cancel    context.CancelFunc
	wakeupCh  chan struct{}
	startOnce sync.Once
	started   bool
	done      chan struct{}
}

// New validates cfg, derives E2E keys from RelayKey, and returns a Client.
func New(cfg Config) (*Client, error) {
	cfg.RelayURL = strings.TrimSpace(cfg.RelayURL)
	cfg.RelayKey = strings.TrimSpace(cfg.RelayKey)
	cfg.AuthKey = strings.TrimSpace(cfg.AuthKey)

	if err := ValidateRelayURL(cfg.RelayURL); err != nil {
		return nil, err
	}
	keys, err := e2e.DeriveKeysFromBase64(cfg.RelayKey)
	if err != nil {
		return nil, fmt.Errorf("relayclient: %w", err)
	}

	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = defaultInitialBackoff
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = defaultMaxBackoff
	}
	if cfg.MaxBackoff < cfg.InitialBackoff {
		cfg.MaxBackoff = cfg.InitialBackoff
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		cfg:      cfg,
		relayKey: cfg.RelayKey,
		keys:     keys,
		ctx:      ctx,
		cancel:   cancel,
		wakeupCh: make(chan struct{}, 1),
		done:     make(chan struct{}),
	}, nil
}

// Start runs the relay reconnect loop in a background goroutine.
// Stop it with Close.
func (c *Client) Start() {
	c.startOnce.Do(func() {
		c.mu.Lock()
		c.started = true
		c.mu.Unlock()
		go func() {
			defer close(c.done)
			_ = c.Run(c.ctx)
		}()
	})
}

// Close shuts down the relay client, closes any active connection, and waits
// for the background goroutine (if started via Start) to exit.
func (c *Client) Close() error {
	c.mu.Lock()
	c.closed = true
	conn := c.activeConn
	started := c.started
	c.mu.Unlock()

	c.cancel()
	if conn != nil {
		_ = conn.Close()
	}
	c.signalWakeup()

	if started {
		<-c.done
	}
	return nil
}

// UpdateRelayKey rotates the active relayKey and forces an immediate reconnect
// into the newly derived room.
func (c *Client) UpdateRelayKey(newKey string) error {
	newKey = strings.TrimSpace(newKey)
	keys, err := e2e.DeriveKeysFromBase64(newKey)
	if err != nil {
		return fmt.Errorf("relayclient: %w", err)
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.relayKey = newKey
	c.keys = keys
	c.reconnectNow = true
	conn := c.activeConn
	c.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
	c.signalWakeup()
	return nil
}

// Reconnect drops the active relay socket (retiring the current E2E session)
// and reconnects immediately.
func (c *Client) Reconnect() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.reconnectNow = true
	conn := c.activeConn
	c.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
	c.signalWakeup()
}

// Run blocks while maintaining the relay connection with exponential backoff
// reconnects (InitialBackoff -> MaxBackoff). It returns nil on context
// cancellation or Close, and returns *AuthError immediately on HTTP 401/403.
func (c *Client) Run(parentCtx context.Context) error {
	runCtx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	// Bridge c.Close() cancellation into runCtx.
	go func() {
		select {
		case <-c.ctx.Done():
			cancel()
		case <-runCtx.Done():
		}
	}()

	backoff := c.cfg.InitialBackoff
	for {
		if runCtx.Err() != nil {
			return nil
		}

		c.mu.Lock()
		keys := c.keys
		c.reconnectNow = false
		c.mu.Unlock()

		c.emitState(StateConnecting, "")
		conn, resp, err := c.dial(runCtx, keys)
		if err != nil {
			if runCtx.Err() != nil {
				return nil
			}
			if authErr := checkAuthError(resp); authErr != nil {
				server.LogStatus("relay auth rejected: %v", authErr)
				c.emitState(StateError, authErr.Error())
				if c.cfg.OnAuthError != nil {
					c.cfg.OnAuthError(authErr.StatusCode, authErr.Message)
				}
				return authErr
			}

			server.LogStatus("relay connect failed: %v (retrying in %s)", err, backoff)
			c.emitState(StateDisconnected, err.Error())
			if !c.waitBackoff(runCtx, backoff) {
				return nil
			}
			backoff = nextBackoff(backoff, c.cfg.MaxBackoff)
			continue
		}

		backoff = c.cfg.InitialBackoff
		server.LogStatus("relay connected (%s)", RelayHost(c.cfg.RelayURL))
		c.emitState(StateWaiting, "")

		c.serveConn(runCtx, conn, keys)

		if runCtx.Err() != nil {
			return nil
		}
		if c.consumeReconnectNow() {
			backoff = c.cfg.InitialBackoff
			continue
		}

		server.LogStatus("relay disconnected — reconnecting in %s...", backoff)
		c.emitState(StateDisconnected, "")
		if !c.waitBackoff(runCtx, backoff) {
			return nil
		}
		backoff = nextBackoff(backoff, c.cfg.MaxBackoff)
	}
}

func (c *Client) dial(ctx context.Context, keys e2e.DerivedKeys) (*websocket.Conn, *http.Response, error) {
	dialer := c.cfg.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}

	headers := http.Header{}
	headers.Set(headerRoom, keys.RoomID)
	headers.Set(headerRole, roleHost)
	headers.Set(headerAuth, keys.HostAuth)
	if c.cfg.AuthKey != "" {
		headers.Set("Authorization", "Bearer "+c.cfg.AuthKey)
	}

	conn, resp, err := dialer.DialContext(ctx, c.cfg.RelayURL, headers)
	if err != nil && resp != nil && resp.Body != nil {
		// Read a small snippet for diagnostics and ensure body is closed.
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(strings.NewReader(strings.TrimSpace(string(snippet))))
	}
	return conn, resp, err
}

func checkAuthError(resp *http.Response) *AuthError {
	if resp == nil {
		return nil
	}
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		return nil
	}
	msg := http.StatusText(resp.StatusCode)
	if resp.Body != nil {
		if body, err := io.ReadAll(io.LimitReader(resp.Body, 256)); err == nil {
			if trimmed := strings.TrimSpace(string(body)); trimmed != "" {
				msg = trimmed
			}
		}
		_ = resp.Body.Close()
	}
	return &AuthError{StatusCode: resp.StatusCode, Message: msg}
}

type relayFrameHeader struct {
	Type    string `json:"type"`
	Online  bool   `json:"online"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) serveConn(ctx context.Context, conn *websocket.Conn, keys e2e.DerivedKeys) {
	c.mu.Lock()
	if c.closed || ctx.Err() != nil {
		c.mu.Unlock()
		_ = conn.Close()
		return
	}
	c.activeConn = conn
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		if c.activeConn == conn {
			c.activeConn = nil
		}
		c.mu.Unlock()
		_ = conn.Close()
	}()

	// Close conn immediately when context is canceled.
	connDone := make(chan struct{})
	defer close(connDone)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-connDone:
		}
	}()

	guard, err := e2e.NewSessionGuard(keys.KeyC2H, e2e.DirC2H, keys.RoomID, e2e.WithClock(c.cfg.Now))
	if err != nil {
		server.LogStatus("relay e2e guard init failed: %v", err)
		return
	}
	sender, err := e2e.NewSender(keys.KeyH2C, e2e.DirH2C, keys.RoomID)
	if err != nil {
		server.LogStatus("relay e2e sender init failed: %v", err)
		return
	}

	var writeMu sync.Mutex
	conn.SetReadLimit(defaultMaxReadBytes)
	_ = conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
	conn.SetPingHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
		writeMu.Lock()
		defer writeMu.Unlock()
		err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(c.cfg.WriteTimeout))
		if errors.Is(err, websocket.ErrCloseSent) {
			return nil
		}
		return err
	})
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
	})

	writeEncrypted := func(outbound protocol.Outbound) error {
		plaintext, err := outbound.Bytes()
		if err != nil {
			return err
		}
		env, err := sender.Seal(plaintext)
		if err != nil {
			return err
		}
		wire, err := json.Marshal(env)
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
		return conn.WriteMessage(websocket.TextMessage, wire)
	}

	var (
		proc          *server.MessageProcessor
		peerConnected bool
	)

	resetSession := func(emitWaiting bool) {
		if proc != nil {
			proc.Close()
			proc = nil
		}
		guard.RetireCurrent()
		if err := sender.Reset(); err != nil {
			server.LogStatus("relay e2e sender reset failed: %v", err)
		}
		if peerConnected {
			peerConnected = false
			if c.cfg.OnPeerStatus != nil {
				c.cfg.OnPeerStatus(false, "")
			}
		}
		if emitWaiting {
			c.emitState(StateWaiting, "")
		}
	}

	newProc := func() *server.MessageProcessor {
		return server.NewMessageProcessor(server.ProcessorConfig{
			Hostname:         c.cfg.Hostname,
			Version:          c.cfg.Version,
			Paster:           c.cfg.Paster,
			PasteMu:          &c.pasteMu,
			MaxPendingPastes: c.cfg.MaxPendingPastes,
			Now:              c.cfg.Now,
			OnHello: func(device, _ string) {
				peerConnected = true
				if c.cfg.OnPeerStatus != nil {
					c.cfg.OnPeerStatus(true, device)
				}
				c.emitState(StateConnected, device)
			},
			OnTextReceived: c.cfg.OnTextReceived,
			WriteOutbound:  writeEncrypted,
		})
	}

	defer resetSession(false)

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() == nil && !c.isReconnectNow() {
				server.LogStatus("relay read error: %v", err)
			}
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))

		if msgType != websocket.TextMessage {
			continue
		}

		var hdr relayFrameHeader
		if err := json.Unmarshal(data, &hdr); err != nil {
			server.LogStatus("relay invalid frame json")
			continue
		}

		switch hdr.Type {
		case "peer_status":
			if hdr.Online {
				server.LogStatus("peer online")
				resetSession(false)
				proc = newProc()
			} else {
				server.LogStatus("peer offline")
				resetSession(true)
			}

		case "error":
			server.LogStatus("relay error [%s]: %s", hdr.Code, hdr.Message)
			if c.cfg.OnRelayError != nil {
				c.cfg.OnRelayError(hdr.Code, hdr.Message)
			}

		case e2e.EnvelopeType:
			var env e2e.Envelope
			if err := json.Unmarshal(data, &env); err != nil {
				server.LogStatus("relay e2e rejected: malformed envelope")
				resetSession(true)
				continue
			}
			plaintext, err := guard.Open(env)
			if err != nil {
				server.LogStatus("relay e2e rejected: %v", err)
				resetSession(true)
				continue
			}
			if proc == nil {
				proc = newProc()
			}
			if err := proc.HandleRaw(plaintext); err != nil {
				return
			}

		default:
			server.LogStatus("relay ignored frame type=%q", hdr.Type)
		}
	}
}

func (c *Client) emitState(state RelayState, detail string) {
	if c.cfg.OnStateChange != nil {
		c.cfg.OnStateChange(state, detail)
	}
}

func (c *Client) signalWakeup() {
	select {
	case c.wakeupCh <- struct{}{}:
	default:
	}
}

func (c *Client) consumeReconnectNow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	v := c.reconnectNow
	c.reconnectNow = false
	return v
}

func (c *Client) isReconnectNow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reconnectNow
}

func (c *Client) waitBackoff(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-c.wakeupCh:
		return ctx.Err() == nil
	case <-timer.C:
		return ctx.Err() == nil
	}
}

func nextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}
	return next
}
