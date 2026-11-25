package server

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yuler/airvoice/cli/paste"
)

const (
	defaultMaxConnections   = 16
	defaultMaxMessageBytes  = 1 << 20 // 1 MiB
	defaultWriteTimeout     = 10 * time.Second
	defaultPingInterval     = 30 * time.Second
	defaultPongWait         = 90 * time.Second
	defaultMaxPendingPastes = 16
)

// Config defines the configuration for the Server.
type Config struct {
	Addr, Hostname, Version string
	Port                    int
	Paster                  paste.Paster
	OnTextReceived          func(content, device string)
	OnConnect               func(device string)
	OnDisconnect            func()

	MaxConnections   int
	MaxMessageBytes  int64
	WriteTimeout     time.Duration
	PingInterval     time.Duration
	PongWait         time.Duration
	MaxPendingPastes int
	AuthFailLimit    int
	AuthFailWindow   time.Duration
}

// Server handles health checks and upgrades/coordinates websocket connections.
type Server struct {
	cfg         Config
	hub         *Hub
	upgrader    websocket.Upgrader
	authLimiter *authRateLimiter
	tokenMu     sync.RWMutex
	token       string
	pasteMu     sync.Mutex
	httpServer  *http.Server
	mu          sync.Mutex
}

// New returns a newly configured Server instance.
func New(cfg Config) *Server {
	return &Server{
		cfg:         cfg,
		hub:         NewHub(),
		authLimiter: newAuthRateLimiter(cfg.AuthFailLimit, cfg.AuthFailWindow),
		upgrader: websocket.Upgrader{
			CheckOrigin: checkOrigin,
		},
	}
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

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) >= 7 && strings.EqualFold(auth[:7], "Bearer ") {
		if tok := strings.TrimSpace(auth[7:]); tok != "" {
			return tok
		}
	}
	return r.URL.Query().Get("token")
}

func (s *Server) maxConnections() int {
	if s.cfg.MaxConnections > 0 {
		return s.cfg.MaxConnections
	}
	return defaultMaxConnections
}

func (s *Server) maxMessageBytes() int64 {
	if s.cfg.MaxMessageBytes > 0 {
		return s.cfg.MaxMessageBytes
	}
	return defaultMaxMessageBytes
}

func (s *Server) writeTimeout() time.Duration {
	if s.cfg.WriteTimeout > 0 {
		return s.cfg.WriteTimeout
	}
	return defaultWriteTimeout
}

func (s *Server) pingInterval() time.Duration {
	if s.cfg.PingInterval > 0 {
		return s.cfg.PingInterval
	}
	return defaultPingInterval
}

func (s *Server) pongWait() time.Duration {
	if s.cfg.PongWait > 0 {
		return s.cfg.PongWait
	}
	return defaultPongWait
}

func (s *Server) maxPendingPastes() int {
	if s.cfg.MaxPendingPastes > 0 {
		return s.cfg.MaxPendingPastes
	}
	return defaultMaxPendingPastes
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
	srv := &http.Server{
		Addr:         s.cfg.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	s.mu.Lock()
	s.httpServer = srv
	s.mu.Unlock()
	return srv.ListenAndServe()
}

// DisconnectClients closes active WebSocket connections without stopping the server.
func (s *Server) DisconnectClients() {
	if s.hub.CloseAll() > 0 {
		logStatus("all clients disconnected — waiting for reconnection...")
		if s.cfg.OnDisconnect != nil {
			s.cfg.OnDisconnect()
		}
	}
}

// Close gracefully shuts down the server and closes all connections.
func (s *Server) Close() error {
	s.hub.CloseAll()

	s.mu.Lock()
	srv := s.httpServer
	s.mu.Unlock()
	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r.RemoteAddr)
	if !s.authLimiter.Allow(ip) {
		logStatus("ws rejected: rate limited %s", r.RemoteAddr)
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}

	token := extractToken(r)
	if !s.validToken(token) {
		s.authLimiter.RecordFailure(ip)
		logStatus("ws rejected: invalid token from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !s.hub.TryReserve(s.maxConnections()) {
		logStatus("ws rejected: max connections (%d) reached from %s", s.maxConnections(), r.RemoteAddr)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.hub.CancelReserve()
		logStatus("ws upgrade failed from %s: %v", r.RemoteAddr, err)
		return
	}

	s.hub.AddReserved(conn)
	logStatus("client connected from %s (active: %d)", r.RemoteAddr, s.hub.Count())
	if s.cfg.OnConnect != nil {
		s.cfg.OnConnect(r.RemoteAddr)
	}
	s.handleConnection(conn)
}

func (s *Server) validToken(token string) bool {
	s.tokenMu.RLock()
	expected := s.token
	s.tokenMu.RUnlock()
	if expected == "" || token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

// CheckPortAvailable checks if a TCP port is available to listen on all interfaces.
func CheckPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is already in use", port)
	}
	ln.Close()
	return nil
}

