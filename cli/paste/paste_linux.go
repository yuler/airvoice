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
		if _, err := lookPath("wl-copy"); err != nil {
			return nil, fmt.Errorf("wl-clipboard (wl-copy) is not installed. Please install it first (e.g. 'sudo apt install wl-clipboard')")
		}
		if _, err := lookPath("ydotool"); err != nil {
			return nil, fmt.Errorf("ydotool is not installed. Please install it first (e.g. 'sudo apt install ydotool' or 'yay -S ydotool')")
		}
		return &waylandPaster{}, nil
	case SessionX11:
		if _, err := lookPath("xclip"); err != nil {
			return nil, fmt.Errorf("xclip is not installed. Please install it first (e.g. 'sudo apt install xclip')")
		}
		if _, err := lookPath("xdotool"); err != nil {
			return nil, fmt.Errorf("xdotool is not installed. Please install it first (e.g. 'sudo apt install xdotool')")
		}
		return &x11Paster{}, nil
	default:
		return nil, fmt.Errorf("unsupported or undetected session type: %s", session)
	}
}
