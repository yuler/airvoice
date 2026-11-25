package server

import (
	"net"
	"sync"
	"time"
)

const (
	defaultAuthFailLimit  = 20
	defaultAuthFailWindow = time.Minute
	maxTrackedAuthIPs     = 1024
)

type authFailEntry struct {
	count       int
	windowStart time.Time
}

type authRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]authFailEntry
	now     func() time.Time
}

func newAuthRateLimiter(limit int, window time.Duration) *authRateLimiter {
	if limit <= 0 {
		limit = defaultAuthFailLimit
	}
	if window <= 0 {
		window = defaultAuthFailWindow
	}
	return &authRateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string]authFailEntry),
		now:     time.Now,
	}
}

// Allow reports whether the given IP is under the failed-auth rate limit.
func (l *authRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.entries[ip]
	if !ok {
		return true
	}
	if now.Sub(entry.windowStart) >= l.window {
		delete(l.entries, ip)
		return true
	}
	return entry.count < l.limit
}

// RecordFailure records a failed authentication attempt for the given IP.
func (l *authRateLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	entry, ok := l.entries[ip]
	if !ok || now.Sub(entry.windowStart) >= l.window {
		if !ok && len(l.entries) >= maxTrackedAuthIPs {
			l.cleanupLocked(now)
		}
		l.entries[ip] = authFailEntry{count: 1, windowStart: now}
		return
	}
	entry.count++
	l.entries[ip] = entry
}

func (l *authRateLimiter) cleanupLocked(now time.Time) {
	for ip, entry := range l.entries {
		if now.Sub(entry.windowStart) >= l.window {
			delete(l.entries, ip)
		}
	}
	if len(l.entries) < maxTrackedAuthIPs {
		return
	}
	var oldestIP string
	var oldestTime time.Time
	for ip, entry := range l.entries {
		if oldestIP == "" || entry.windowStart.Before(oldestTime) {
			oldestIP = ip
			oldestTime = entry.windowStart
		}
	}
	if oldestIP != "" {
		delete(l.entries, oldestIP)
	}
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
