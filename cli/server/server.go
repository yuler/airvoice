package server

import (
	"net/http"
	"sync"

	"github.com/airvoice/airvoice/cli/pairing"
	"github.com/airvoice/airvoice/cli/paste"
	"github.com/gorilla/websocket"
)

// Config defines the configuration for the Server.
type Config struct {
	Addr, Hostname, Version string
	Port                     int
	Paster                   paste.Paster
}

// Server handles health checks and upgrades/coordinates websocket connections.
type Server struct {
	cfg      Config
	hub      *Hub
	upgrader websocket.Upgrader
	tokenMu  sync.RWMutex
	token    string
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

// ListenAndServe starts the HTTP/WS server.
func (s *Server) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ws", s.handleWS)
	return http.ListenAndServe(s.cfg.Addr, mux)
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

	logStatus("client connected from %s", r.RemoteAddr)
	s.handleConnection(conn)
}

// RotatePairing issues a new token and prints a fresh QR code to stderr.
func (s *Server) RotatePairing(banner string) error {
	token, _, err := pairing.PrintPairing(s.cfg.Port, banner)
	if err != nil {
		return err
	}
	s.tokenMu.Lock()
	s.token = token
	s.tokenMu.Unlock()
	return nil
}

func (s *Server) validToken(token string) bool {
	s.tokenMu.RLock()
	defer s.tokenMu.RUnlock()
	return token != "" && token == s.token
}
