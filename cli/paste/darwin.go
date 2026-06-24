//go:build darwin

package paste

import (
	"fmt"
	"time"
)

type darwinPaster struct{}

func New() (Paster, error) {
	session := DetectSession()
	if session != SessionDarwin {
		return nil, fmt.Errorf("unsupported or undetected session type: %s", session)
	}
	return &darwinPaster{}, nil
}

func (d *darwinPaster) Paste(text string) error {
	if err := runCommand("pbcopy", text); err != nil {
		return err
	}
	time.Sleep(80 * time.Millisecond)
	return runCommand("osascript", "", "-e", `tell application "System Events" to keystroke "v" using command down`)
}

func (d *darwinPaster) Name() string {
	return "darwin"
}
