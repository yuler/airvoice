//go:build linux

package paste

import (
	"time"
)

type waylandPaster struct{}

func (w *waylandPaster) Paste(text string) error {
	if err := runCommand("wl-copy", text); err != nil {
		return err
	}
	// Run the keypress simulation asynchronously so that Wayland hangs/delays
	// do not block the clipboard copy success reply.
	go func() {
		time.Sleep(80 * time.Millisecond)
		_ = runCommand("ydotool", "", "key", "CTRL+v")
	}()
	return nil
}

func (w *waylandPaster) Name() string {
	return "wayland"
}
