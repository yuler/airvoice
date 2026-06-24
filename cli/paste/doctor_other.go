//go:build !darwin

package paste

import (
	"fmt"
	"io"
)

// PrintDoctor writes paste diagnostics for non-macOS platforms.
func PrintDoctor(w io.Writer) int {
	fmt.Fprintln(w, "doctor is only available on macOS")
	return 1
}

func IsKeystrokePermissionError(err error) bool {
	return false
}
