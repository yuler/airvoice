package e2e

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Direction identifies the traffic direction bound into AES-GCM AAD.
type Direction string

const (
	// EnvelopeType is the wire JSON type for encrypted relay frames.
	EnvelopeType = "e2e"
	// EnvelopeVersion is the current E2E envelope schema version.
	EnvelopeVersion = 1

	// DirC2H represents Phone -> Host frames (encrypted with KeyC2H).
	DirC2H Direction = "c2h"
	// DirH2C represents Host -> Phone frames (encrypted with KeyH2C).
	DirH2C Direction = "h2c"

	// NonceSize is the AES-256-GCM nonce length in bytes (16-char base64url).
	NonceSize = 12
	// SessionIDSize is the per-connection session ID length in bytes (11-char base64url).
	SessionIDSize = 8

	// DefaultHelloMaxSkew is the maximum allowed clock skew on seq==1 hello timestamps.
	DefaultHelloMaxSkew = 5 * time.Minute
	// DefaultRetiredSIDLimit is the bounded ring size for retired session IDs.
	DefaultRetiredSIDLimit = 256
)

var (
	ErrInvalidEnvelope  = errors.New("e2e: invalid envelope")
	ErrInvalidDirection = errors.New("e2e: invalid direction")
	ErrInvalidSessionID = errors.New("e2e: invalid session id")
	ErrInvalidNonce     = errors.New("e2e: invalid nonce")
	ErrAuthFailed       = errors.New("e2e: authentication failed")
	ErrReplayedSession  = errors.New("e2e: replayed retired session id")
	ErrReplayedSequence = errors.New("e2e: replayed or out-of-order sequence")
	ErrHelloRequired    = errors.New("e2e: new session must start with hello at seq 1")
	ErrStaleHello       = errors.New("e2e: hello timestamp outside allowed skew")
)

// Envelope is the wire JSON representation of an end-to-end encrypted frame.
type Envelope struct {
	Type string `json:"type"`
	V    int    `json:"v"`
	SID  string `json:"sid"`
	Seq  uint64 `json:"seq"`
	IV   string `json:"iv"`
	CT   string `json:"ct"`
}

// NewSessionID generates a random 8-byte session ID encoded as unpadded base64url.
func NewSessionID() (string, error) {
	raw := make([]byte, SessionIDSize)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("e2e: generate session id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// ValidSessionID reports whether sid is a canonical 8-byte unpadded base64url string.
func ValidSessionID(sid string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(sid)
	if err != nil || len(raw) != SessionIDSize {
		return false
	}
	return base64.RawURLEncoding.EncodeToString(raw) == sid
}

// ValidDirection reports whether dir is DirC2H or DirH2C.
func ValidDirection(dir Direction) bool {
	return dir == DirC2H || dir == DirH2C
}

// AAD constructs the Additional Authenticated Data bytes bound to every frame:
// airvoice-e2e-v1|<dir>|<roomId>|<sid>|<seq>
func AAD(dir Direction, roomID, sid string, seq uint64) []byte {
	return []byte(fmt.Sprintf("airvoice-e2e-v1|%s|%s|%s|%d", dir, roomID, sid, seq))
}

// Seal encrypts plaintext with AES-256-GCM using a fresh random 12-byte nonce.
func Seal(key []byte, dir Direction, roomID, sid string, seq uint64, plaintext []byte) (Envelope, error) {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return Envelope{}, fmt.Errorf("e2e: generate nonce: %w", err)
	}
	return SealWithNonce(key, dir, roomID, sid, seq, nonce, plaintext)
}

// SealWithNonce encrypts plaintext with AES-256-GCM using an explicit 12-byte nonce.
// Intended for deterministic cross-platform test vector generation and verification.
func SealWithNonce(key []byte, dir Direction, roomID, sid string, seq uint64, nonce, plaintext []byte) (Envelope, error) {
	if len(key) != KeySize {
		return Envelope{}, ErrInvalidKey
	}
	if !ValidDirection(dir) {
		return Envelope{}, ErrInvalidDirection
	}
	if roomID == "" || seq == 0 {
		return Envelope{}, ErrInvalidEnvelope
	}
	if !ValidSessionID(sid) {
		return Envelope{}, ErrInvalidSessionID
	}
	if len(nonce) != NonceSize {
		return Envelope{}, ErrInvalidNonce
	}

	gcm, err := newGCM(key)
	if err != nil {
		return Envelope{}, err
	}
	aad := AAD(dir, roomID, sid, seq)
	ciphertext := gcm.Seal(nil, nonce, plaintext, aad)

	return Envelope{
		Type: EnvelopeType,
		V:    EnvelopeVersion,
		SID:  sid,
		Seq:  seq,
		IV:   base64.RawURLEncoding.EncodeToString(nonce),
		CT:   base64.RawURLEncoding.EncodeToString(ciphertext),
	}, nil
}

// Open decrypts and authenticates an Envelope without stateful replay tracking.
func Open(key []byte, dir Direction, roomID string, env Envelope) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	if !ValidDirection(dir) {
		return nil, ErrInvalidDirection
	}
	if env.Type != EnvelopeType || env.V != EnvelopeVersion || roomID == "" || env.Seq == 0 {
		return nil, ErrInvalidEnvelope
	}
	if !ValidSessionID(env.SID) {
		return nil, ErrInvalidSessionID
	}

	nonce, err := base64.RawURLEncoding.DecodeString(env.IV)
	if err != nil || len(nonce) != NonceSize || base64.RawURLEncoding.EncodeToString(nonce) != env.IV {
		return nil, ErrInvalidNonce
	}
	ct, err := base64.RawURLEncoding.DecodeString(env.CT)
	if err != nil || base64.RawURLEncoding.EncodeToString(ct) != env.CT {
		return nil, ErrInvalidEnvelope
	}

	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	if len(ct) < gcm.Overhead() {
		return nil, ErrAuthFailed
	}

	aad := AAD(dir, roomID, env.SID, env.Seq)
	plaintext, err := gcm.Open(nil, nonce, ct, aad)
	if err != nil {
		return nil, ErrAuthFailed
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("e2e: aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("e2e: aes gcm: %w", err)
	}
	return gcm, nil
}

// Sender manages monotonically increasing sequence numbers for an outgoing
// per-connection E2E session.
type Sender struct {
	mu     sync.Mutex
	key    []byte
	dir    Direction
	roomID string
	sid    string
	seq    uint64
}

// NewSender creates an outgoing E2E Sender with a fresh random session ID.
func NewSender(key []byte, dir Direction, roomID string) (*Sender, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	if !ValidDirection(dir) {
		return nil, ErrInvalidDirection
	}
	if roomID == "" {
		return nil, ErrInvalidEnvelope
	}
	sid, err := NewSessionID()
	if err != nil {
		return nil, err
	}
	keyCopy := append([]byte(nil), key...)
	return &Sender{
		key:    keyCopy,
		dir:    dir,
		roomID: roomID,
		sid:    sid,
	}, nil
}

// Reset generates a fresh session ID and resets the outgoing sequence counter to 0.
func (s *Sender) Reset() error {
	sid, err := NewSessionID()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.sid = sid
	s.seq = 0
	s.mu.Unlock()
	return nil
}

// Seal encrypts plaintext with the next sequence number in the current session.
func (s *Sender) Seal(plaintext []byte) (Envelope, error) {
	s.mu.Lock()
	s.seq++
	sid := s.sid
	seq := s.seq
	s.mu.Unlock()
	return Seal(s.key, s.dir, s.roomID, sid, seq, plaintext)
}

// SessionGuard enforces AEAD verification, seq==1 hello gating with timestamp
// freshness, strictly increasing sequence numbers, and retired-session rejection.
type SessionGuard struct {
	mu           sync.Mutex
	key          []byte
	dir          Direction
	roomID       string
	maxSkew      time.Duration
	retiredLimit int
	now          func() time.Time

	activeSID    string
	lastSeq      uint64
	retired      map[string]struct{}
	retiredOrder []string
}

// SessionGuardOption customizes a SessionGuard.
type SessionGuardOption func(*SessionGuard)

// WithClock overrides the clock used for hello timestamp freshness validation.
func WithClock(fn func() time.Time) SessionGuardOption {
	return func(g *SessionGuard) {
		if fn != nil {
			g.now = fn
		}
	}
}

// WithMaxSkew overrides the allowed clock skew for hello timestamps.
func WithMaxSkew(d time.Duration) SessionGuardOption {
	return func(g *SessionGuard) {
		if d > 0 {
			g.maxSkew = d
		}
	}
}

// WithRetiredLimit overrides the maximum number of retired session IDs remembered.
func WithRetiredLimit(n int) SessionGuardOption {
	return func(g *SessionGuard) {
		if n > 0 {
			g.retiredLimit = n
		}
	}
}

// NewSessionGuard creates a stateful inbound E2E verifier for a single direction.
func NewSessionGuard(key []byte, dir Direction, roomID string, opts ...SessionGuardOption) (*SessionGuard, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	if !ValidDirection(dir) {
		return nil, ErrInvalidDirection
	}
	if roomID == "" {
		return nil, ErrInvalidEnvelope
	}
	g := &SessionGuard{
		key:          append([]byte(nil), key...),
		dir:          dir,
		roomID:       roomID,
		maxSkew:      DefaultHelloMaxSkew,
		retiredLimit: DefaultRetiredSIDLimit,
		now:          time.Now,
		retired:      make(map[string]struct{}),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(g)
		}
	}
	return g, nil
}

type helloHeader struct {
	Type string `json:"type"`
	TS   int64  `json:"ts"`
}

// Open verifies replay rules, authenticates and decrypts env, and advances
// session state only when all checks succeed.
func (g *SessionGuard) Open(env Envelope) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if env.Type != EnvelopeType || env.V != EnvelopeVersion || env.Seq == 0 {
		return nil, ErrInvalidEnvelope
	}
	if !ValidSessionID(env.SID) {
		return nil, ErrInvalidSessionID
	}
	if _, wasRetired := g.retired[env.SID]; wasRetired {
		return nil, ErrReplayedSession
	}

	if g.activeSID != "" && env.SID == g.activeSID {
		if env.Seq <= g.lastSeq {
			return nil, ErrReplayedSequence
		}
		plaintext, err := Open(g.key, g.dir, g.roomID, env)
		if err != nil {
			return nil, err
		}
		g.lastSeq = env.Seq
		return plaintext, nil
	}

	// Brand new session ID: must start with seq == 1 carrying a fresh hello frame.
	if env.Seq != 1 {
		return nil, ErrHelloRequired
	}
	plaintext, err := Open(g.key, g.dir, g.roomID, env)
	if err != nil {
		return nil, err
	}

	var hdr helloHeader
	if err := json.Unmarshal(plaintext, &hdr); err != nil || hdr.Type != "hello" {
		return nil, ErrHelloRequired
	}
	if !g.validHelloTS(hdr.TS) {
		return nil, ErrStaleHello
	}

	if g.activeSID != "" {
		g.retireLocked(g.activeSID)
	}
	g.activeSID = env.SID
	g.lastSeq = 1
	return plaintext, nil
}

// RetireCurrent retires the active session ID (if any) so it can never be replayed
// after a disconnect or peer reset.
func (g *SessionGuard) RetireCurrent() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.activeSID != "" {
		g.retireLocked(g.activeSID)
		g.activeSID = ""
		g.lastSeq = 0
	}
}

func (g *SessionGuard) validHelloTS(ts int64) bool {
	if ts <= 0 {
		return false
	}
	diff := g.now().Unix() - ts
	if diff < 0 {
		diff = -diff
	}
	return diff <= int64(g.maxSkew/time.Second)
}

func (g *SessionGuard) retireLocked(sid string) {
	if _, exists := g.retired[sid]; exists {
		return
	}
	if len(g.retiredOrder) >= g.retiredLimit {
		evicted := g.retiredOrder[0]
		copy(g.retiredOrder, g.retiredOrder[1:])
		g.retiredOrder = g.retiredOrder[:len(g.retiredOrder)-1]
		delete(g.retired, evicted)
	}
	g.retired[sid] = struct{}{}
	g.retiredOrder = append(g.retiredOrder, sid)
}
