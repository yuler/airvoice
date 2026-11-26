package relay

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/e2e"
)

const (
	HeaderRoom = "X-Airvoice-Room"
	HeaderRole = "X-Airvoice-Role"
	HeaderAuth = "X-Airvoice-Auth"

	// hardReadSlack lets frames slightly over MaxMessageBytes be read so the
	// relay can return a structured message_too_large error without disconnecting,
	// while still aborting abusive multi-megabyte frames early.
	hardReadSlack = 64 * 1024
)

// Server is the zero-knowledge Airvoice WebSocket relay HTTP server.
type Server struct {
	cfg        Config
	hub        *Hub
	limiter    *Limiter
	upgrader   websocket.Upgrader
	mux        *http.ServeMux
	httpServer *http.Server

	mu        sync.Mutex
	stopEvict chan struct{}
	evictDone chan struct{}
	closeOnce sync.Once
}

// NewServer constructs a Relay Server and starts its background eviction loop.
// Call Close when finished to release resources.
func NewServer(cfg Config) (*Server, error) {
	cfg = cfg.normalized()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		cfg:     cfg,
		hub:     newHub(cfg),
		limiter: newLimiter(cfg),
		upgrader: websocket.Upgrader{
			CheckOrigin: checkOrigin,
		},
		stopEvict: make(chan struct{}),
		evictDone: make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ws", s.handleWS)
	s.mux = mux

	go s.evictLoop()
	return s, nil
}

// Handler returns the HTTP handler serving /health and /ws.
func (s *Server) Handler() http.Handler {
	return s.mux
}

// Hub returns the underlying room Hub (useful for diagnostics and tests).
func (s *Server) Hub() *Hub {
	return s.hub
}

// ListenAndServe binds cfg.Addr and serves HTTP/WebSocket traffic.
func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	s.mu.Lock()
	s.httpServer = srv
	s.mu.Unlock()
	return srv.ListenAndServe()
}

// Close stops background eviction, disconnects all peers, and shuts down HTTP.
func (s *Server) Close(ctx context.Context) error {
	var err error
	s.closeOnce.Do(func() {
		close(s.stopEvict)
		<-s.evictDone

		s.hub.CloseAll()

		s.mu.Lock()
		srv := s.httpServer
		s.mu.Unlock()
		if srv != nil {
			err = srv.Shutdown(ctx)
		}
	})
	return err
}

func (s *Server) evictLoop() {
	defer close(s.evictDone)
	ticker := time.NewTicker(s.cfg.EvictInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopEvict:
			return
		case <-ticker.C:
			s.hub.EvictIdle()
			s.limiter.Sweep()
		}
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write([]byte("ok"))
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	ip := extractClientIP(r, s.cfg.TrustProxy)
	if !s.limiter.AllowAuthAttempt(ip) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	// Optional private relay access control (header-only, never query string).
	if s.cfg.AuthKey != "" {
		if !validBearer(r.Header.Get("Authorization"), s.cfg.AuthKey) {
			s.limiter.RecordAuthFailure(ip)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	roomID := strings.TrimSpace(r.Header.Get(HeaderRoom))
	roleStr := strings.TrimSpace(r.Header.Get(HeaderRole))
	authToken := strings.TrimSpace(r.Header.Get(HeaderAuth))

	if !validBase64URLBytes(roomID, e2e.RoomIDSize) {
		http.Error(w, "Bad Request: invalid room header", http.StatusBadRequest)
		return
	}
	role := Role(roleStr)
	if role != RoleHost && role != RoleClient {
		http.Error(w, "Bad Request: invalid role header", http.StatusBadRequest)
		return
	}
	if !validBase64URLBytes(authToken, e2e.AuthSize) {
		s.limiter.RecordAuthFailure(ip)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authHash := sha256.Sum256([]byte(authToken))

	if err := s.hub.PreCheckJoin(roomID, role, authHash, ip); err != nil {
		s.writeJoinError(w, ip, err)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[relay] ws upgrade failed ip=%s: %v", ip, err)
		return
	}

	p := newPeer(roomID, role, authHash, ip, conn)
	if err := s.hub.Register(p); err != nil {
		// Rare concurrent race between PreCheckJoin and Upgrade.
		code := websocket.CloseTryAgainLater
		if errors.Is(err, ErrAuthMismatch) {
			s.limiter.RecordAuthFailure(ip)
			code = websocket.ClosePolicyViolation
		}
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, err.Error()),
			time.Now().Add(s.cfg.WriteTimeout),
		)
		_ = conn.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.writeLoop(p)
	}()

	s.readLoop(p)
	s.hub.Unregister(p)
	p.closeWith(websocket.CloseNormalClosure, "")
	wg.Wait()
}

func (s *Server) writeJoinError(w http.ResponseWriter, ip string, err error) {
	switch {
	case errors.Is(err, ErrAuthMismatch):
		s.limiter.RecordAuthFailure(ip)
		http.Error(w, "Forbidden", http.StatusForbidden)
	case errors.Is(err, ErrIPRoomLimit):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write(marshalErrorFrame("room_limit_reached", "too many active rooms for this IP"))
	case errors.Is(err, ErrGlobalRoomLimit):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write(marshalErrorFrame("room_limit_reached", "relay server room capacity reached"))
	default:
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

type minimalEnvelope struct {
	Type string `json:"type"`
}

func (s *Server) readLoop(p *peer) {
	hardLimit := s.cfg.MaxMessageBytes + hardReadSlack
	p.conn.SetReadLimit(hardLimit)
	_ = p.conn.SetReadDeadline(time.Now().Add(s.cfg.PongWait))
	p.conn.SetPongHandler(func(string) error {
		return p.conn.SetReadDeadline(time.Now().Add(s.cfg.PongWait))
	})

	for {
		msgType, data, err := p.conn.ReadMessage()
		if err != nil {
			return
		}
		_ = p.conn.SetReadDeadline(time.Now().Add(s.cfg.PongWait))

		if msgType != websocket.TextMessage {
			continue
		}
		if int64(len(data)) > s.cfg.MaxMessageBytes {
			p.enqueue(marshalErrorFrame("message_too_large", "frame exceeds maximum allowed size"))
			continue
		}

		var env minimalEnvelope
		if err := json.Unmarshal(data, &env); err != nil || env.Type != e2e.EnvelopeType {
			continue
		}

		opp := s.hub.ForwardTarget(p)
		if opp == nil {
			if !s.limiter.AllowRate(p.roomID) {
				p.enqueue(marshalErrorFrame("rate_limited", "too many messages; try again shortly"))
				continue
			}
			p.enqueue(marshalErrorFrame("peer_offline", "peer is offline"))
			continue
		}

		if err := s.limiter.AllowForward(p.roomID, p.ip); err != nil {
			switch {
			case errors.Is(err, ErrQuotaExceeded):
				p.enqueue(marshalErrorFrame("quota_exceeded", "daily message quota exceeded"))
			case errors.Is(err, ErrRateLimited):
				p.enqueue(marshalErrorFrame("rate_limited", "too many messages; try again shortly"))
			}
			continue
		}

		opp.enqueue(data)
	}
}

func (s *Server) writeLoop(p *peer) {
	ticker := time.NewTicker(s.cfg.PingInterval)
	defer func() {
		ticker.Stop()
		_ = p.conn.Close()
	}()

	writeClose := func(req closeRequest) {
		_ = p.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(req.code, req.reason),
			time.Now().Add(s.cfg.WriteTimeout),
		)
	}

	for {
		// Prioritize shutdown/replacement close signals over queued data frames.
		select {
		case req := <-p.closeCh:
			writeClose(req)
			return
		default:
		}

		select {
		case req := <-p.closeCh:
			writeClose(req)
			return
		case <-p.done:
			// Drain any pending close request so explicit close codes are never
			// dropped when closeCh and done become ready simultaneously.
			select {
			case req := <-p.closeCh:
				writeClose(req)
			default:
			}
			return
		case msg := <-p.send:
			_ = p.conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
			if err := p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = p.conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func validBearer(header, expected string) bool {
	if len(header) < 7 || !strings.EqualFold(header[:7], "Bearer ") {
		return false
	}
	got := strings.TrimSpace(header[7:])
	if got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}

func validBase64URLBytes(encoded string, expectedBytes int) bool {
	if encoded == "" {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(raw) != expectedBytes {
		return false
	}
	return base64.RawURLEncoding.EncodeToString(raw) == encoded
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}

// extractClientIP extracts the client IP address. When trustProxy is true,
// it trusts the right-most valid IP in X-Forwarded-For (appended by a single
// reverse proxy such as Caddy or Nginx, preventing client header spoofing),
// falling back to X-Real-IP and then RemoteAddr.
func extractClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				candidate := strings.TrimSpace(parts[i])
				if ip := net.ParseIP(candidate); ip != nil {
					return ip.String()
				}
			}
		}
		if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
			if ip := net.ParseIP(xri); ip != nil {
				return ip.String()
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
