//go:build windows

package paste

import (
	"os/exec"
	"syscall"
)

// CREATE_NO_WINDOW prevents child processes from opening a console window.
const createNoWindow = 0x08000000

func configureCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
}
