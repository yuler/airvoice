//go:build linux

package paste

import (
	"fmt"
	"os"
	"time"
)

type waylandPaster struct{}

func (w *waylandPaster) Paste(text string) error {
	if err := runCommand("wl-copy", text); err != nil {
		return err
	}
	time.Sleep(80 * time.Millisecond)

	// If YDOTOOL_SOCKET is not set, search for common paths and set it for the process
	if os.Getenv("YDOTOOL_SOCKET") == "" {
		if _, err := os.Stat("/tmp/.ydotool_socket"); err == nil {
			os.Setenv("YDOTOOL_SOCKET", "/tmp/.ydotool_socket")
		} else {
			uidSocket := fmt.Sprintf("/run/user/%d/ydotool/socket", os.Getuid())
			if _, err := os.Stat(uidSocket); err == nil {
				os.Setenv("YDOTOOL_SOCKET", uidSocket)
			}
		}
	}

	return runCommand("ydotool", "", "key", "CTRL+v")
}

func (w *waylandPaster) Name() string {
	return "wayland"
}
