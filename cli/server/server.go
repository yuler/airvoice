package server

import (
	"net/http"

	"github.com/airvoice/airvoice/cli/paste"
	"github.com/gorilla/websocket"
)

// Config defines the configuration for the Server.
type Config struct {
	Addr, Token, Hostname, Version string
	Paster                         paste.Paster
}

// Server handles health checks and upgrades/coordinates websocket connections.
type Server struct {
	cfg      Config
	hub      *Hub
	upgrader websocket.Upgrader
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
	if token != s.cfg.Token {
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
