//go:build windows

package paste

import (
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW prevents a visible console when spawning PowerShell subprocesses.
const createNoWindow = 0x08000000

func configureCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
}
