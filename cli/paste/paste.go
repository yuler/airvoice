package paste

import (
	"os/exec"
	"strings"
)

type Paster interface {
	Paste(text string) error
	Name() string
}

// runCommand is a package-level variable that can be overridden in tests.
var runCommand = func(name string, stdin string, args ...string) error {
	cmd := exec.Command(name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	return cmd.Run()
}
