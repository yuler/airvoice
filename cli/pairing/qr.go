package pairing

import (
	"os"

	"github.com/mdp/qrterminal/v3"
)

// PrintQR generates a QR code from the payload and prints it to os.Stderr using half-blocks,
// enforcing a QuietZone of 2 and terminal default colors (preventing inverted scanner issues).
func PrintQR(payload []byte) {
	config := qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     os.Stderr,
		QuietZone:  2,
		HalfBlocks: true,
	}
	qrterminal.GenerateWithConfig(string(payload), config)
}
