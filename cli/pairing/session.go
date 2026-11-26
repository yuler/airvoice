package pairing

import (
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/mdp/qrterminal/v3"
)

// PrintPairing builds a fresh pairing payload, prints the QR code to stderr, and returns the token and WebSocket URL.
func PrintPairing(port int, banner string) (token, wsURL string, err error) {
	return printPairing(port, uuid.NewString(), banner, "")
}

func printPairing(port int, token, banner, version string) (string, string, error) {
	ip, err := LocalIPv4()
	if err != nil {
		return "", "", err
	}
	wsURL := fmt.Sprintf("ws://%s:%d/ws", ip, port)

	payload := &Payload{
		Version: VersionLAN,
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
	if version != "" {
		fmt.Fprintf(os.Stderr, "  Airvoice %s\n", version)
	}
	fmt.Fprintf(os.Stderr, "  Token: %s\n", token)
	fmt.Fprintf(os.Stderr, "  WebSocket URL: %s\n\n", wsURL)
	fmt.Fprintf(os.Stderr, "  [airvoice] waiting for phone connection...\n\n")
	return token, wsURL, nil
}

// PrintPairingWithToken prints the QR code and pairing metadata for a given token.
func PrintPairingWithToken(port int, token, banner, version string) (wsURL string, err error) {
	_, wsURL, err = printPairing(port, token, banner, version)
	return wsURL, err
}

// PrintRelayPairing renders the v2 relay QR code, relay URL, and copyable
// airvoice://pair?... URI to os.Stderr and returns the URI.
func PrintRelayPairing(wsURL, relayKey, authKey, banner, version string) (string, error) {
	return fprintRelayPairing(os.Stderr, wsURL, relayKey, authKey, banner, version)
}

func fprintRelayPairing(w io.Writer, wsURL, relayKey, authKey, banner, version string) (string, error) {
	uri, err := BuildPairingURI(wsURL, relayKey, authKey)
	if err != nil {
		return "", err
	}
	payload := &PayloadV2{
		Version: VersionRelay,
		Mode:    ModeRelay,
		WS:      wsURL,
		Key:     relayKey,
		Auth:    authKey,
	}
	payloadBytes, err := payload.Marshal()
	if err != nil {
		return "", err
	}

	if banner != "" {
		fmt.Fprintf(w, "\n [airvoice] %s\n\n", banner)
	} else {
		fmt.Fprintf(w, "\n")
	}
	printQRTo(w, payloadBytes)
	fmt.Fprintf(w, "\n")
	if version != "" {
		fmt.Fprintf(w, "  Airvoice %s\n", version)
	}
	fmt.Fprintf(w, "  Relay: %s (E2E)\n", wsURL)
	fmt.Fprintf(w, "  Link: %s\n\n", uri)
	fmt.Fprintf(w, "  [airvoice] waiting for phone connection via relay...\n\n")
	return uri, nil
}

func printQRTo(w io.Writer, payload []byte) {
	cfg := qrterminal.Config{
		Level:          qrterminal.L,
		Writer:         w,
		QuietZone:      2,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
	}
	qrterminal.GenerateWithConfig(string(payload), cfg)
}
