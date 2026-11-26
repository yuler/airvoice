package relayclient

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateRelayURL verifies that rawURL is a valid WebSocket relay URL.
// Public hosts must use wss://; ws:// is permitted only for local/private hosts
// (localhost, .local, loopback, RFC 1918, link-local, Tailscale 100.64.0.0/10, ULA fc00::/7).
func ValidateRelayURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("relay URL is empty")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid relay URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "wss" && scheme != "ws" {
		return fmt.Errorf("relay URL must use wss:// or ws:// scheme, got %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("relay URL is missing host")
	}
	if scheme == "ws" && !IsLocalHost(host) {
		return fmt.Errorf("relay URL %q must use wss:// for non-local hosts", rawURL)
	}
	return nil
}

// RelayHost extracts the display host (e.g. "relay.airvoice.yuler.cc") from a relay URL.
func RelayHost(rawURL string) string {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// IsLocalHost reports whether host is localhost, .local, loopback, private
// (RFC 1918 / ULA fc00::/7), link-local, or Tailscale CGNAT (100.64.0.0/10).
func IsLocalHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	if h == "" {
		return false
	}
	if h == "localhost" || strings.HasSuffix(h, ".local") {
		return true
	}
	ip := net.ParseIP(h)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		// Tailscale / CGNAT 100.64.0.0/10
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
	}
	return false
}
