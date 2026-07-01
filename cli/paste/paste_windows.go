//go:build windows

package paste

import (
	"encoding/base64"
	"fmt"
	"strings"
)

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

	// Combine clipboard and keystroke into a single PowerShell invocation to minimize startup overhead.
	// Uses try/catch with error tokens so Go can distinguish failure modes.
	psCmd := "$ErrorActionPreference = 'Stop'; $b = [Console]::In.ReadToEnd().Trim(); if ($b) { " +
		"try { Set-Clipboard -Value ([System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($b))) } catch { throw 'CLIPBOARD_FAILED' }; " +
		"Start-Sleep -Milliseconds 80; " +
		"try { (New-Object -ComObject WScript.Shell).SendKeys('^v') } catch { throw 'KEYSTROKE_FAILED' } " +
		"}"
	if err := runCommand("powershell", encoded, "-NoProfile", "-Command", psCmd); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "CLIPBOARD_FAILED") {
			return fmt.Errorf("clipboard (Set-Clipboard) failed: %w", err)
		}
		if strings.Contains(errStr, "KEYSTROKE_FAILED") {
			return formatWindowsKeystrokeError(err)
		}
		return fmt.Errorf("paste via PowerShell failed: %w", err)
	}
	return nil
}

func (w *windowsPaster) Name() string {
	return "windows"
}

func formatWindowsKeystrokeError(err error) error {
	return fmt.Errorf("keystroke (SendKeys) failed: %w; text copied to clipboard, please press Ctrl+V manually", err)
}
