package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/yuler/airvoice/cli/paste"
	"github.com/gorilla/websocket"
)

// Config defines the configuration for the Server.
type Config struct {
	Addr, Hostname, Version string
	Port                     int
	Paster                   paste.Paster
	OnTextReceived           func(content, device string)
	OnConnect               func(device string)
	OnDisconnect            func()
}

// Server handles health checks and upgrades/coordinates websocket connections.
type Server struct {
	cfg        Config
	hub        *Hub
	upgrader   websocket.Upgrader
	tokenMu    sync.RWMutex
	token      string
	pasteMu    sync.Mutex
	httpServer *http.Server
	mu         sync.Mutex
}

// New returns a newly configured Server instance.
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
	s.hub.CloseAll()
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
	token := r.URL.Query().Get("token")
	if !s.validToken(token) {
		logStatus("ws rejected: invalid token from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logStatus("ws upgrade failed from %s: %v", r.RemoteAddr, err)
		return
	}

	logStatus("client connected from %s (active: %d)", r.RemoteAddr, s.hub.Count()+1)
	if s.cfg.OnConnect != nil {
		s.cfg.OnConnect(r.RemoteAddr)
	}
	s.handleConnection(conn)
}

func (s *Server) validToken(token string) bool {
	s.tokenMu.RLock()
	defer s.tokenMu.RUnlock()
	return token != "" && token == s.token
}
