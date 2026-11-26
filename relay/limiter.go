package relay

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrRateLimited   = errors.New("relay: rate limited")
	ErrQuotaExceeded = errors.New("relay: daily quota exceeded")
)

const (
	bucketIdleTTL     = 2 * time.Hour
	maxTrackedAuthIPs = 2048
)

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

type authFailEntry struct {
	count       int
	windowStart time.Time
}

// Limiter enforces per-room token-bucket rate limits, UTC daily message quotas
// per room and per IP, and failed-auth brute-force protection.
type Limiter struct {
	mu             sync.Mutex
	ratePerSec     float64
	rateBurst      float64
	dailyLimit     int
	authFailLimit  int
	authFailWindow time.Duration
	now            func() time.Time

	buckets   map[string]*tokenBucket
	dayNum    int64
	roomDaily map[string]int
	ipDaily   map[string]int
	authFails map[string]authFailEntry
}

func newLimiter(cfg Config) *Limiter {
	var ratePerSec float64
	if cfg.RatePerMinute > 0 {
		ratePerSec = float64(cfg.RatePerMinute) / 60.0
	}
	now := cfg.Now()
	return &Limiter{
		ratePerSec:     ratePerSec,
		rateBurst:      float64(cfg.RateBurst),
		dailyLimit:     cfg.DailyMessageLimit,
		authFailLimit:  cfg.AuthFailLimit,
		authFailWindow: cfg.AuthFailWindow,
		now:            cfg.Now,
		buckets:        make(map[string]*tokenBucket),
		dayNum:         utcDayNumber(now),
		roomDaily:      make(map[string]int),
		ipDaily:        make(map[string]int),
		authFails:      make(map[string]authFailEntry),
	}
}

func utcDayNumber(t time.Time) int64 {
	return t.UTC().Unix() / 86400
}

func (l *Limiter) rolloverLocked(now time.Time) {
	today := utcDayNumber(now)
	if today != l.dayNum {
		l.dayNum = today
		clear(l.roomDaily)
		clear(l.ipDaily)
	}
}

func (l *Limiter) bucketTokensLocked(roomID string, now time.Time) (*tokenBucket, float64) {
	b, ok := l.buckets[roomID]
	if !ok {
		b = &tokenBucket{tokens: l.rateBurst, lastRefill: now}
		l.buckets[roomID] = b
		return b, b.tokens
	}
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * l.ratePerSec
		if b.tokens > l.rateBurst {
			b.tokens = l.rateBurst
		}
		b.lastRefill = now
	}
	return b, b.tokens
}

// AllowRate checks and consumes one rate-limit token for roomID without
// touching daily message quotas (used for non-forwarded offline frames).
func (l *Limiter) AllowRate(roomID string) bool {
	if l.ratePerSec <= 0 {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, tokens := l.bucketTokensLocked(roomID, now)
	if tokens < 1.0 {
		return false
	}
	b.tokens -= 1.0
	return true
}

// AllowForward atomically enforces both UTC daily quotas (per room and per IP)
// and per-room token bucket rate limits. Counters are mutated only on success.
func (l *Limiter) AllowForward(roomID, ip string) error {
	if l.ratePerSec <= 0 && l.dailyLimit <= 0 {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.rolloverLocked(now)

	if l.dailyLimit > 0 {
		if l.roomDaily[roomID] >= l.dailyLimit || (ip != "" && l.ipDaily[ip] >= l.dailyLimit) {
			return ErrQuotaExceeded
		}
	}

	var b *tokenBucket
	if l.ratePerSec > 0 {
		var tokens float64
		b, tokens = l.bucketTokensLocked(roomID, now)
		if tokens < 1.0 {
			return ErrRateLimited
		}
	}

	if b != nil {
		b.tokens -= 1.0
	}
	if l.dailyLimit > 0 {
		l.roomDaily[roomID]++
		if ip != "" {
			l.ipDaily[ip]++
		}
	}
	return nil
}

// AllowAuthAttempt reports whether ip is allowed to attempt a handshake.
func (l *Limiter) AllowAuthAttempt(ip string) bool {
	if l.authFailLimit <= 0 || ip == "" {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.authFails[ip]
	if !ok {
		return true
	}
	if now.Sub(entry.windowStart) >= l.authFailWindow {
		delete(l.authFails, ip)
		return true
	}
	return entry.count < l.authFailLimit
}

// RecordAuthFailure records a failed authentication attempt from ip.
func (l *Limiter) RecordAuthFailure(ip string) {
	if l.authFailLimit <= 0 || ip == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.authFails[ip]
	if !ok || now.Sub(entry.windowStart) >= l.authFailWindow {
		if !ok && len(l.authFails) >= maxTrackedAuthIPs {
			l.sweepAuthFailsLocked(now)
		}
		l.authFails[ip] = authFailEntry{count: 1, windowStart: now}
		return
	}
	entry.count++
	l.authFails[ip] = entry
}

// Sweep removes expired daily maps, stale token buckets, and expired auth fail entries.
func (l *Limiter) Sweep() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.rolloverLocked(now)
	for id, b := range l.buckets {
		if now.Sub(b.lastRefill) >= bucketIdleTTL {
			delete(l.buckets, id)
		}
	}
	l.sweepAuthFailsLocked(now)
}

func (l *Limiter) sweepAuthFailsLocked(now time.Time) {
	for ip, entry := range l.authFails {
		if now.Sub(entry.windowStart) >= l.authFailWindow {
			delete(l.authFails, ip)
		}
	}
	if len(l.authFails) < maxTrackedAuthIPs {
		return
	}
	var oldestIP string
	var oldestTime time.Time
	for ip, entry := range l.authFails {
		if oldestIP == "" || entry.windowStart.Before(oldestTime) {
			oldestIP = ip
			oldestTime = entry.windowStart
		}
	}
	if oldestIP != "" {
		delete(l.authFails, oldestIP)
	}
}
