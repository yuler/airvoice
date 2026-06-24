package paste

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Paster interface {
	Paste(text string) error
	Name() string
}

const pasteCommandTimeout = 4 * time.Second

// runCommand is a package-level variable that can be overridden in tests.
var runCommand = func(name string, stdin string, args ...string) error {
	_, err := runCommandCapture(name, stdin, args...)
	return err
}

func runCommandCapture(name string, stdin string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pasteCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", context.DeadlineExceeded
		}
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return msg, fmt.Errorf("%s: %w", msg, err)
		}
		return "", err
	}
	return "", nil
}
