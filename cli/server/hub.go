package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains the single active connection.
type Hub struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

// NewHub initializes and returns a Hub.
func NewHub() *Hub {
	return &Hub{}
}

// Set saves the connection and closes any previous connection.
func (h *Hub) Set(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn != nil {
		h.conn.Close()
	}
	h.conn = conn
}

// Clear removes the connection only if it matches the current active one
// (preventing clearing a new connection from an old connection teardown).
func (h *Hub) Clear(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn == conn {
		if h.conn != nil {
			h.conn.Close()
		}
		h.conn = nil
	}
}

// Get returns the current active connection.
func (h *Hub) Get() *websocket.Conn {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.conn
}
