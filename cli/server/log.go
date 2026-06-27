package server

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func logStatus(format string, args ...any) {
	fmt.Fprintf(os.Stderr, " [airvoice] "+format+"\n", args...)
}

func previewText(s string, maxRunes int) string {
	s = strings.ReplaceAll(s, "\n", `\n`)
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes]) + "..."
}
