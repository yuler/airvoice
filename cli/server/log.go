package server

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

var LogHook func(string)

func logStatus(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, " [airvoice] %s\n", msg)
	if LogHook != nil {
		LogHook(" [airvoice] " + msg)
	}
}

func previewText(s string, maxRunes int) string {
	s = strings.ReplaceAll(s, "\n", `\n`)
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes]) + "..."
}
