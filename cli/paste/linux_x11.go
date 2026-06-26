//go:build linux

package paste

import (
	"time"
)

type x11Paster struct{}

func (x *x11Paster) Paste(text string) error {
	if err := runCommand("xclip", text, "-selection", "clipboard"); err != nil {
		return err
	}
	time.Sleep(80 * time.Millisecond)
	return runCommand("xdotool", "", "key", "ctrl+v")
}

func (x *x11Paster) Name() string {
	return "x11"
}
