package e2e

import (
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

const (
	// KeySize is the length in bytes of the master relay pre-shared key.
	KeySize = 32
	// RoomIDSize is the raw byte length of the derived room identifier.
	RoomIDSize = 16
	// AuthSize is the raw byte length of derived host/client room auth tokens.
	AuthSize = 32

	infoRoomID     = "airvoice-relay-room-v1"
	infoHostAuth   = "airvoice-relay-host-auth-v1"
	infoClientAuth = "airvoice-relay-client-v1"
	infoKeyC2H     = "airvoice-e2e-c2h-v1"
	infoKeyH2C     = "airvoice-e2e-h2c-v1"
)

var (
	ErrInvalidKey = errors.New("e2e: relay key must be 32 bytes base64url-encoded without padding")
)

// DerivedKeys holds all identifiers, relay auth tokens, and AES-256-GCM keys
// derived via HKDF-SHA256 from a 32-byte master relayKey.
type DerivedKeys struct {
	RoomID     string // 22-char unpadded base64url (16 bytes)
	HostAuth   string // 43-char unpadded base64url (32 bytes)
	ClientAuth string // 43-char unpadded base64url (32 bytes)
	KeyC2H     []byte // 32-byte AES-256 key (Phone -> Host)
	KeyH2C     []byte // 32-byte AES-256 key (Host -> Phone)
}

// GenerateKey generates a cryptographically random 32-byte relayKey encoded
// as unpadded base64url (43 characters).
func GenerateKey() (string, error) {
	raw := make([]byte, KeySize)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("e2e: generate relay key: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// ParseKey decodes an unpadded base64url-encoded 32-byte relayKey.
func ParseKey(encoded string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(raw) != KeySize {
		return nil, ErrInvalidKey
	}
	// Enforce canonical unpadded base64url representation.
	if base64.RawURLEncoding.EncodeToString(raw) != encoded {
		return nil, ErrInvalidKey
	}
	return raw, nil
}

// DeriveKeys derives all relay identifiers and per-direction E2E encryption keys
// from a raw 32-byte master relayKey using RFC 5869 HKDF-SHA256 with nil salt.
func DeriveKeys(rawKey []byte) (DerivedKeys, error) {
	if len(rawKey) != KeySize {
		return DerivedKeys{}, ErrInvalidKey
	}

	roomRaw, err := hkdf.Key(sha256.New, rawKey, nil, infoRoomID, RoomIDSize)
	if err != nil {
		return DerivedKeys{}, fmt.Errorf("e2e: derive roomId: %w", err)
	}
	hostAuthRaw, err := hkdf.Key(sha256.New, rawKey, nil, infoHostAuth, AuthSize)
	if err != nil {
		return DerivedKeys{}, fmt.Errorf("e2e: derive hostAuth: %w", err)
	}
	clientAuthRaw, err := hkdf.Key(sha256.New, rawKey, nil, infoClientAuth, AuthSize)
	if err != nil {
		return DerivedKeys{}, fmt.Errorf("e2e: derive clientAuth: %w", err)
	}
	keyC2H, err := hkdf.Key(sha256.New, rawKey, nil, infoKeyC2H, KeySize)
	if err != nil {
		return DerivedKeys{}, fmt.Errorf("e2e: derive keyC2H: %w", err)
	}
	keyH2C, err := hkdf.Key(sha256.New, rawKey, nil, infoKeyH2C, KeySize)
	if err != nil {
		return DerivedKeys{}, fmt.Errorf("e2e: derive keyH2C: %w", err)
	}

	return DerivedKeys{
		RoomID:     base64.RawURLEncoding.EncodeToString(roomRaw),
		HostAuth:   base64.RawURLEncoding.EncodeToString(hostAuthRaw),
		ClientAuth: base64.RawURLEncoding.EncodeToString(clientAuthRaw),
		KeyC2H:     keyC2H,
		KeyH2C:     keyH2C,
	}, nil
}

// DeriveKeysFromBase64 decodes an unpadded base64url relayKey and derives all keys.
func DeriveKeysFromBase64(encoded string) (DerivedKeys, error) {
	raw, err := ParseKey(encoded)
	if err != nil {
		return DerivedKeys{}, err
	}
	return DeriveKeys(raw)
}
