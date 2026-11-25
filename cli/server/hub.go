package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains all active WebSocket connections.
type Hub struct {
	mu      sync.Mutex
	conns   map[*websocket.Conn]struct{}
	pending int
}

// NewHub initializes and returns a Hub.
func NewHub() *Hub {
	return &Hub{
		conns: make(map[*websocket.Conn]struct{}),
	}
}

// TryReserve reserves a connection slot if active + pending connections are below max.
func (h *Hub) TryReserve(max int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if max > 0 && len(h.conns)+h.pending >= max {
		return false
	}
	h.pending++
	return true
}

// CancelReserve releases a previously reserved connection slot after a failed upgrade.
func (h *Hub) CancelReserve() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pending > 0 {
		h.pending--
	}
}

// AddReserved promotes a reserved slot into an active connection.
func (h *Hub) AddReserved(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pending > 0 {
		h.pending--
	}
	h.conns[conn] = struct{}{}
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

// CloseAll closes all tracked connections and returns how many were closed.
func (h *Hub) CloseAll() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := len(h.conns)
	for conn := range h.conns {
		conn.Close()
	}
	h.conns = make(map[*websocket.Conn]struct{})
	return n
}
