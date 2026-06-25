//go:build linux

package paste

import (
	"fmt"
	"os/exec"
)

var lookPath = exec.LookPath

func New() (Paster, error) {
	session := DetectSession()
	switch session {
	case SessionWayland:
		if _, err := lookPath("ydotool"); err != nil {
			return nil, fmt.Errorf("ydotool is not installed. Please install it first (e.g. 'sudo apt install ydotool' or 'yay -S ydotool')")
		}
		return &waylandPaster{}, nil
	case SessionX11:
		return &x11Paster{}, nil
	default:
		return nil, fmt.Errorf("unsupported or undetected session type: %s", session)
	}
}
