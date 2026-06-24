//go:build darwin

package paste

import (
	"fmt"
	"time"
)

// Keystroke via System Events (requires Accessibility for the parent terminal / osascript).
const pasteKeystrokeScript = `tell application "System Events" to keystroke "v" using command down`

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
		return fmt.Errorf("clipboard (pbcopy) failed: %w", err)
	}
	time.Sleep(80 * time.Millisecond)

	if err := runCommand("osascript", "", "-e", pasteKeystrokeScript); err != nil {
		return formatDarwinKeystrokeError(err)
	}
	return nil
}

func (d *darwinPaster) Name() string {
	return "darwin"
}

func formatDarwinKeystrokeError(err error) error {
	if IsKeystrokePermissionError(err) {
		return fmt.Errorf("%s；文本已复制到剪贴板，可手动按 Cmd+V", darwinAccessibilityHintText())
	}
	return fmt.Errorf("keystroke (osascript) failed: %w", err)
}
