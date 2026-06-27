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
	time.Sleep(80 * time.Millisecond)
	return runCommand("ydotool", "", "key", "CTRL+v")
}

func (w *waylandPaster) Name() string {
	return "wayland"
}
