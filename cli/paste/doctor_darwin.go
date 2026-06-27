//go:build darwin

package paste

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintDoctor writes macOS paste permission diagnostics to w.
func PrintDoctor(w io.Writer) int {
	fmt.Fprintln(w, "Airvoice paste diagnostics (macOS)")
	fmt.Fprintln(w)

	host := detectHostTerminalApp()
	if host.Path != "" {
		fmt.Fprintf(w, "Host app: %s (%s)\n", host.Label, host.Path)
	} else {
		fmt.Fprintln(w, "Host app: (could not detect — check parent terminal in Accessibility)")
	}

	if exe, err := os.Executable(); err == nil {
		fmt.Fprintf(w, "Airvoice binary: %s\n", exe)
	}
	fmt.Fprintln(w)

	fmt.Fprint(w, "Clipboard (pbcopy): ")
	if err := runCommand("pbcopy", "airvoice-doctor-test"); err != nil {
		fmt.Fprintf(w, "FAILED (%v)\n", err)
	} else {
		fmt.Fprintln(w, "OK")
	}

	fmt.Fprint(w, "Keystroke (osascript): ")
	_, err := runCommandCapture("osascript", "", "-e", pasteKeystrokeScript)
	if err != nil {
		fmt.Fprintf(w, "FAILED\n  %v\n", err)
	} else {
		fmt.Fprintln(w, "OK")
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "Enable all of these in System Settings → Privacy & Security → Accessibility:")
	for _, t := range darwinAccessibilityTargets() {
		fmt.Fprintf(w, "  • %s\n    %s\n", t.Label, t.Path)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "After changing permissions:")
	fmt.Fprintln(w, "  1. Fully quit the host app (Cmd+Q), not just close the terminal tab")
	fmt.Fprintln(w, "  2. Reopen and run: mise run dev")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Tip: Enabling only Terminal.app does not help if you run serve from Cursor's integrated terminal.")

	if err != nil {
		return 1
	}
	return 0
}

func IsKeystrokePermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not allowed to send keystrokes") ||
		strings.Contains(msg, "(1002)") ||
		strings.Contains(msg, "assistive access") ||
		strings.Contains(msg, "辅助功能")
}
