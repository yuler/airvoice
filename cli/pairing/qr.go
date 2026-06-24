package pairing

import (
	"os"

	"github.com/mdp/qrterminal/v3"
)

// PrintQR generates a QR code from the payload and prints it to os.Stderr using half-blocks.
func PrintQR(payload []byte) {
	qrterminal.GenerateHalfBlock(string(payload), qrterminal.L, os.Stderr)
}
