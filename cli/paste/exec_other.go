//go:build !windows

package paste

import "os/exec"

func configureCmd(cmd *exec.Cmd) {}
