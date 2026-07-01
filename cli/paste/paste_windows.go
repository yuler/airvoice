//go:build windows

package paste

import (
	"encoding/base64"
	"fmt"
	"time"
)

// Keystroke via WScript.Shell SendKeys (targets the active foreground window).
const pasteKeystrokeScript = `(New-Object -ComObject WScript.Shell).SendKeys('^v')`

type windowsPaster struct{}

func New() (Paster, error) {
	session := DetectSession()
	if session != SessionWindows {
		return nil, fmt.Errorf("unsupported or undetected session type: %s", session)
	}
	return &windowsPaster{}, nil
}

func (w *windowsPaster) Paste(text string) error {
	if text == "" {
		return nil
	}

	// Base64-encode the text to completely bypass any Windows/PowerShell console code page encoding issues.
	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	// Write base64 string to powershell via stdin, decode it in-memory as UTF-8,
	// and set it to the clipboard in a single PowerShell invocation.
	psCmd := "$ErrorActionPreference = 'Stop'; $b = [Console]::In.ReadToEnd().Trim(); if ($b) { Set-Clipboard -Value ([System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($b))) }"
	if err := runCommand("powershell", encoded, "-NoProfile", "-Command", psCmd); err != nil {
		return fmt.Errorf("clipboard (Set-Clipboard) failed: %w", err)
	}
	time.Sleep(80 * time.Millisecond)

	if err := runCommand("powershell", "", "-NoProfile", "-Command", pasteKeystrokeScript); err != nil {
		return formatWindowsKeystrokeError(err)
	}
	return nil
}

func (w *windowsPaster) Name() string {
	return "windows"
}

func formatWindowsKeystrokeError(err error) error {
	return fmt.Errorf("keystroke (SendKeys) failed: %w；文本已复制到剪贴板，可手动按 Ctrl+V", err)
}
