package paste

import (
	"context"
	"fmt"
	"os"
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

// runCommandWithEnv runs a command with optional additional KEY=VALUE environment entries
// without mutating the current process environment.
var runCommandWithEnv = func(extraEnv []string, name string, stdin string, args ...string) error {
	if len(extraEnv) == 0 {
		return runCommand(name, stdin, args...)
	}
	_, err := runCommandCaptureEnv(extraEnv, name, stdin, args...)
	return err
}

func runCommandCapture(name string, stdin string, args ...string) (string, error) {
	return runCommandCaptureEnv(nil, name, stdin, args...)
}

func runCommandCaptureEnv(extraEnv []string, name string, stdin string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pasteCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	configureCmd(cmd)
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stderr strings.Builder
	isClipboardCmd := name == "wl-copy" || name == "xclip" || name == "xsel"
	if !isClipboardCmd {
		// Capture stderr for non-clipboard commands.
		// We avoid capturing stderr for clipboard commands because tools like xclip/xsel
		// may daemonize and keep the stderr pipe open, causing cmd.Run() to hang indefinitely.
		cmd.Stderr = &stderr
	}

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", context.DeadlineExceeded
		}
		if !isClipboardCmd {
			if msg := strings.TrimSpace(stderr.String()); msg != "" {
				return msg, fmt.Errorf("%s: %w", msg, err)
			}
		}
		return "", err
	}
	return "", nil
}
