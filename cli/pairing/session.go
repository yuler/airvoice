package pairing

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

// PrintPairing builds a fresh pairing payload, prints the QR code to stderr, and returns the token and WebSocket URL.
func PrintPairing(port int, banner string) (token, wsURL string, err error) {
	token = uuid.NewString()
	ip, err := LocalIPv4()
	if err != nil {
		return "", "", err
	}
	wsURL = fmt.Sprintf("ws://%s:%d/ws", ip, port)

	payload := &Payload{
		Version: 1,
		WS:      wsURL,
		Token:   token,
	}
	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", "", err
	}

	if banner != "" {
		fmt.Fprintf(os.Stderr, "\n [airvoice] %s\n\n", banner)
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}
	PrintQR(payloadBytes)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Token: %s\n", token)
	fmt.Fprintf(os.Stderr, "  WebSocket URL: %s\n\n", wsURL)
	fmt.Fprintf(os.Stderr, "  [airvoice] waiting for phone connection...\n\n")
	return token, wsURL, nil
}

// PrintPairingWithToken prints the QR code and pairing metadata for a given token.
func PrintPairingWithToken(port int, token string, banner string) (wsURL string, err error) {
	ip, err := LocalIPv4()
	if err != nil {
		return "", err
	}
	wsURL = fmt.Sprintf("ws://%s:%d/ws", ip, port)

	payload := &Payload{
		Version: 1,
		WS:      wsURL,
		Token:   token,
	}
	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", err
	}

	if banner != "" {
		fmt.Fprintf(os.Stderr, "\n [airvoice] %s\n\n", banner)
	} else {
		fmt.Fprintf(os.Stderr, "\n")
	}
	PrintQR(payloadBytes)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  Token: %s\n", token)
	fmt.Fprintf(os.Stderr, "  WebSocket URL: %s\n\n", wsURL)
	fmt.Fprintf(os.Stderr, "  [airvoice] waiting for phone connection...\n\n")
	return wsURL, nil
}
