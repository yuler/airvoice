package relay

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Role identifies the slot within a relay room.
type Role string

const (
	RoleHost   Role = "host"
	RoleClient Role = "client"

	// CloseCodeReplaced is sent when a new connection with matching auth replaces an existing peer.
	CloseCodeReplaced = 4001
	// CloseCodeIdleTimeout is sent when an orphan or inactive room is evicted.
	CloseCodeIdleTimeout = 4002

	peerSendBuffer = 32
)

var (
	ErrAuthMismatch    = errors.New("relay: room role auth mismatch")
	ErrIPRoomLimit     = errors.New("relay: per-ip room limit reached")
	ErrGlobalRoomLimit = errors.New("relay: global room limit reached")
	ErrPeerOffline     = errors.New("relay: peer is offline")
)

type closeRequest struct {
	code   int
	reason string
}

type peer struct {
	roomID   string
	role     Role
	authHash [32]byte
	ip       string
	conn     *websocket.Conn

	send    chan []byte
	closeCh chan closeRequest
	done    chan struct{}

	closeOnce sync.Once
}

func newPeer(roomID string, role Role, authHash [32]byte, ip string, conn *websocket.Conn) *peer {
	return &peer{
		roomID:   roomID,
		role:     role,
		authHash: authHash,
		ip:       ip,
		conn:     conn,
		send:     make(chan []byte, peerSendBuffer),
		closeCh:  make(chan closeRequest, 1),
		done:     make(chan struct{}),
	}
}

func (p *peer) enqueue(data []byte) bool {
	select {
	case <-p.done:
		return false
	default:
	}
	select {
	case p.send <- data:
		return true
	case <-p.done:
		return false
	default:
		// Slow consumer: close to protect hub and opposite peer.
		p.closeWith(websocket.ClosePolicyViolation, "slow consumer")
		return false
	}
}

func (p *peer) closeWith(code int, reason string) {
	p.closeOnce.Do(func() {
		select {
		case p.closeCh <- closeRequest{code: code, reason: reason}:
		default:
		}
		close(p.done)
	})
}

type room struct {
	id             string
	host           *peer
	client         *peer
	hostAuthHash   [32]byte
	hasHostAuth    bool
	clientAuthHash [32]byte
	hasClientAuth  bool
	lastActivity   time.Time
	orphanSince    time.Time
}

func (r *room) slot(role Role) (*peer, [32]byte, bool) {
	if role == RoleHost {
		return r.host, r.hostAuthHash, r.hasHostAuth
	}
	return r.client, r.clientAuthHash, r.hasClientAuth
}

func (r *room) opposite(role Role) *peer {
	if role == RoleHost {
		return r.client
	}
	return r.host
}

func (r *room) ips() (string, string) {
	var hostIP, clientIP string
	if r.host != nil {
		hostIP = r.host.ip
	}
	if r.client != nil {
		clientIP = r.client.ip
	}
	return hostIP, clientIP
}

type peerStatusFrame struct {
	Type   string `json:"type"`
	Online bool   `json:"online"`
}

type errorFrame struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func marshalPeerStatus(online bool) []byte {
	b, _ := json.Marshal(peerStatusFrame{Type: "peer_status", Online: online})
	return b
}

func marshalErrorFrame(code, message string) []byte {
	b, _ := json.Marshal(errorFrame{Type: "error", Code: code, Message: message})
	return b
}

// Hub manages active 1-host / 1-client rooms, slot ownership, IP/global room
// caps, verbatim E2E frame forwarding, and idle eviction.
type Hub struct {
	mu                 sync.Mutex
	rooms              map[string]*room
	ipRooms            map[string]int
	maxRoomsPerIP      int
	maxGlobalRooms     int
	orphanIdleTimeout  time.Duration
	sessionIdleTimeout time.Duration
	now                func() time.Time
}

func newHub(cfg Config) *Hub {
	return &Hub{
		rooms:              make(map[string]*room),
		ipRooms:            make(map[string]int),
		maxRoomsPerIP:      cfg.MaxRoomsPerIP,
		maxGlobalRooms:     cfg.MaxGlobalRooms,
		orphanIdleTimeout:  cfg.OrphanIdleTimeout,
		sessionIdleTimeout: cfg.SessionIdleTimeout,
		now:                cfg.Now,
	}
}

// PreCheckJoin validates room ownership and capacity before upgrading WebSocket.
func (h *Hub) PreCheckJoin(roomID string, role Role, authHash [32]byte, ip string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.checkJoinLocked(roomID, role, authHash, ip)
}

func (h *Hub) checkJoinLocked(roomID string, role Role, authHash [32]byte, ip string) error {
	r := h.rooms[roomID]
	if r != nil {
		_, storedHash, hasHash := r.slot(role)
		if hasHash && subtle.ConstantTimeCompare(authHash[:], storedHash[:]) != 1 {
			return ErrAuthMismatch
		}
	} else if h.maxGlobalRooms > 0 && len(h.rooms) >= h.maxGlobalRooms {
		return ErrGlobalRoomLimit
	}

	if h.maxRoomsPerIP > 0 && ip != "" {
		if h.wouldIncreaseIPRoomsLocked(r, role, ip) && h.ipRooms[ip] >= h.maxRoomsPerIP {
			return ErrIPRoomLimit
		}
	}
	return nil
}

func (h *Hub) wouldIncreaseIPRoomsLocked(r *room, role Role, ip string) bool {
	if r == nil {
		return true
	}
	hostIP, clientIP := r.ips()
	// Check whether ip is already counted for this room.
	if hostIP == ip || clientIP == ip {
		return false
	}
	return true
}

func (h *Hub) adjustIPCountsLocked(beforeHost, beforeClient, afterHost, afterClient string) {
	if beforeHost != "" {
		h.decIPLocked(beforeHost)
	}
	if beforeClient != "" && beforeClient != beforeHost {
		h.decIPLocked(beforeClient)
	}
	if afterHost != "" {
		h.ipRooms[afterHost]++
	}
	if afterClient != "" && afterClient != afterHost {
		h.ipRooms[afterClient]++
	}
}

func (h *Hub) decIPLocked(ip string) {
	if n := h.ipRooms[ip] - 1; n > 0 {
		h.ipRooms[ip] = n
	} else {
		delete(h.ipRooms, ip)
	}
}

// Register attaches an upgraded peer to its room, replacing any existing peer
// in the same role slot (closing it with 4001) and emitting peer_status frames.
func (h *Hub) Register(p *peer) error {
	h.mu.Lock()
	if err := h.checkJoinLocked(p.roomID, p.role, p.authHash, p.ip); err != nil {
		h.mu.Unlock()
		return err
	}

	now := h.now()
	r := h.rooms[p.roomID]
	if r == nil {
		r = &room{
			id:           p.roomID,
			lastActivity: now,
		}
		h.rooms[p.roomID] = r
	}

	beforeHost, beforeClient := r.ips()
	var replaced *peer
	if p.role == RoleHost {
		replaced = r.host
		r.host = p
		if !r.hasHostAuth {
			r.hostAuthHash = p.authHash
			r.hasHostAuth = true
		}
	} else {
		replaced = r.client
		r.client = p
		if !r.hasClientAuth {
			r.clientAuthHash = p.authHash
			r.hasClientAuth = true
		}
	}
	afterHost, afterClient := r.ips()
	h.adjustIPCountsLocked(beforeHost, beforeClient, afterHost, afterClient)

	r.lastActivity = now
	opp := r.opposite(p.role)
	if opp != nil {
		r.orphanSince = time.Time{}
	} else {
		r.orphanSince = now
	}
	h.mu.Unlock()

	if replaced != nil {
		replaced.closeWith(CloseCodeReplaced, "Replaced")
	}

	if opp != nil {
		onlineMsg := marshalPeerStatus(true)
		p.enqueue(onlineMsg)
		opp.enqueue(onlineMsg)
	} else {
		p.enqueue(marshalPeerStatus(false))
	}
	return nil
}

// Unregister detaches p if it still occupies its room slot, notifying the
// opposite peer or deleting the room when both slots are empty.
func (h *Hub) Unregister(p *peer) {
	h.mu.Lock()
	r := h.rooms[p.roomID]
	if r == nil {
		h.mu.Unlock()
		return
	}

	cur, _, _ := r.slot(p.role)
	if cur != p {
		// Already replaced or evicted.
		h.mu.Unlock()
		return
	}

	beforeHost, beforeClient := r.ips()
	if p.role == RoleHost {
		r.host = nil
	} else {
		r.client = nil
	}
	afterHost, afterClient := r.ips()
	h.adjustIPCountsLocked(beforeHost, beforeClient, afterHost, afterClient)

	opp := r.opposite(p.role)
	if opp == nil {
		delete(h.rooms, p.roomID)
	} else {
		r.orphanSince = h.now()
	}
	h.mu.Unlock()

	if opp != nil {
		opp.enqueue(marshalPeerStatus(false))
	}
}

// Opposite returns the currently attached opposite peer for p, if online,
// and updates room lastActivity when forwarding.
func (h *Hub) ForwardTarget(p *peer) *peer {
	h.mu.Lock()
	defer h.mu.Unlock()

	r := h.rooms[p.roomID]
	if r == nil {
		return nil
	}
	cur, _, _ := r.slot(p.role)
	if cur != p {
		return nil
	}
	opp := r.opposite(p.role)
	if opp != nil {
		r.lastActivity = h.now()
	}
	return opp
}

// EvictIdle closes and removes rooms that exceeded orphan or session idle timeouts.
func (h *Hub) EvictIdle() int {
	h.mu.Lock()
	now := h.now()
	var evicted []*peer

	for id, r := range h.rooms {
		bothOnline := r.host != nil && r.client != nil
		shouldEvict := false
		if !bothOnline && !r.orphanSince.IsZero() && h.orphanIdleTimeout > 0 && now.Sub(r.orphanSince) >= h.orphanIdleTimeout {
			shouldEvict = true
		} else if bothOnline && h.sessionIdleTimeout > 0 && now.Sub(r.lastActivity) >= h.sessionIdleTimeout {
			shouldEvict = true
		}
		if !shouldEvict {
			continue
		}

		beforeHost, beforeClient := r.ips()
		if r.host != nil {
			evicted = append(evicted, r.host)
			r.host = nil
		}
		if r.client != nil {
			evicted = append(evicted, r.client)
			r.client = nil
		}
		h.adjustIPCountsLocked(beforeHost, beforeClient, "", "")
		delete(h.rooms, id)
	}
	h.mu.Unlock()

	for _, p := range evicted {
		p.closeWith(CloseCodeIdleTimeout, "Idle Timeout")
	}
	return len(evicted)
}

// CloseAll closes all active peers and clears all rooms (used on graceful shutdown).
func (h *Hub) CloseAll() {
	h.mu.Lock()
	var all []*peer
	for id, r := range h.rooms {
		if r.host != nil {
			all = append(all, r.host)
			r.host = nil
		}
		if r.client != nil {
			all = append(all, r.client)
			r.client = nil
		}
		delete(h.rooms, id)
	}
	clear(h.ipRooms)
	h.mu.Unlock()

	for _, p := range all {
		p.closeWith(websocket.CloseGoingAway, "Server shutting down")
	}
}

// RoomCount returns the current number of active rooms.
func (h *Hub) RoomCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.rooms)
}
