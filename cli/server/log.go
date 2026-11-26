package server

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"
)

var (
	logHookMu sync.RWMutex
	logHook   func(string)
)

// SetLogHook sets an optional thread-safe callback for server status logs.
func SetLogHook(fn func(string)) {
	logHookMu.Lock()
	logHook = fn
	logHookMu.Unlock()
}

// LogStatus formats and writes a host status log line to stderr and the optional log hook.
func LogStatus(format string, args ...any) {
	logStatus(format, args...)
}

func logStatus(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, " [airvoice] %s\n", msg)
	logHookMu.RLock()
	hook := logHook
	logHookMu.RUnlock()
	if hook != nil {
		hook(" [airvoice] " + msg)
	}
}

func previewText(s string, maxRunes int) string {
	s = strings.ReplaceAll(s, "\n", `\n`)
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes]) + "..."
}
