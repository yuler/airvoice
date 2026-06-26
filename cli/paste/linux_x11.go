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
	// Run the keypress simulation asynchronously so that X11 hangs/delays
	// do not block the clipboard copy success reply.
	go func() {
		time.Sleep(80 * time.Millisecond)
		_ = runCommand("xdotool", "", "key", "ctrl+v")
	}()
	return nil
}

func (x *x11Paster) Name() string {
	return "x11"
}
