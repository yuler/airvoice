package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains all active WebSocket connections.
type Hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

// NewHub initializes and returns a Hub.
func NewHub() *Hub {
	return &Hub{
		conns: make(map[*websocket.Conn]struct{}),
	}
}

// Add registers a connection.
func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = struct{}{}
}

// Remove unregisters a connection and closes it.
// Returns true if the connection was present.
func (h *Hub) Remove(conn *websocket.Conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.conns[conn]; ok {
		conn.Close()
		delete(h.conns, conn)
		return true
	}
	return false
}

// Count returns the number of active connections.
func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns)
}
