package pairing

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/yuler/airvoice/cli/e2e"
)

const (
	VersionLAN   = 1
	VersionRelay = 2
	ModeRelay    = "relay"
)

// Payload represents the v1 LAN pairing information sent to the peer.
type Payload struct {
	Version int    `json:"v"`
	WS      string `json:"ws"`
	Token   string `json:"token"`
}

// Marshal serializes the v1 LAN Payload to a JSON byte slice.
func (p *Payload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// PayloadV2 represents the v2 E2E Relay pairing payload embedded in QR codes.
type PayloadV2 struct {
	Version int    `json:"v"`
	Mode    string `json:"mode"`
	WS      string `json:"ws"`
	Key     string `json:"key"`
	Auth    string `json:"auth,omitempty"`
}

// Marshal serializes PayloadV2 to a JSON byte slice.
func (p *PayloadV2) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// PairingURI encodes this v2 relay payload into an airvoice://pair?... URI.
func (p *PayloadV2) PairingURI() (string, error) {
	return BuildPairingURI(p.WS, p.Key, p.Auth)
}

// BuildPairingURI constructs a copyable v2 relay pairing link:
// airvoice://pair?v=2&mode=relay&ws=...&key=...[&auth=...]
func BuildPairingURI(wsURL, relayKey, authKey string) (string, error) {
	wsURL = strings.TrimSpace(wsURL)
	relayKey = strings.TrimSpace(relayKey)
	authKey = strings.TrimSpace(authKey)

	if wsURL == "" {
		return "", errors.New("pairing: relay ws URL is required")
	}
	if _, err := e2e.ParseKey(relayKey); err != nil {
		return "", fmt.Errorf("pairing: invalid relay key: %w", err)
	}

	u := fmt.Sprintf(
		"airvoice://pair?v=%d&mode=%s&ws=%s&key=%s",
		VersionRelay,
		ModeRelay,
		escapeQueryParam(wsURL),
		escapeQueryParam(relayKey),
	)
	if authKey != "" {
		u += "&auth=" + escapeQueryParam(authKey)
	}
	return u, nil
}

// escapeQueryParam percent-encodes a query parameter value using %20 for spaces
// so Go, Swift URLComponents, and Android Uri decode it identically.
func escapeQueryParam(v string) string {
	return strings.ReplaceAll(url.QueryEscape(v), "+", "%20")
}
