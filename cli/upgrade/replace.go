package upgrade

import (
	"errors"
	"fmt"
	"os"
	"runtime"
)

// ReplaceExecutable atomically replaces targetPath with newBinaryPath.
func ReplaceExecutable(targetPath, newBinaryPath string) error {
	return replaceExecutableForOS(runtime.GOOS, targetPath, newBinaryPath)
}

func replaceExecutableForOS(goos, targetPath, newBinaryPath string) error {
	mode := os.FileMode(0o755)
	if info, err := os.Stat(targetPath); err == nil {
		if m := info.Mode().Perm(); m != 0 {
			mode = m
		}
	}
	if goos != "windows" {
		if err := os.Chmod(newBinaryPath, mode); err != nil {
			return wrapPermissionError(goos, targetPath, err)
		}
		if err := os.Rename(newBinaryPath, targetPath); err != nil {
			return wrapPermissionError(goos, targetPath, err)
		}
		return nil
	}

	// On Windows, a running executable cannot be overwritten directly,
	// but it can be renamed aside before moving the replacement into place.
	backupPath := targetPath + ".old"
	_ = os.Remove(backupPath)

	hadExisting := false
	if _, err := os.Stat(targetPath); err == nil {
		hadExisting = true
		if err := os.Rename(targetPath, backupPath); err != nil {
			return wrapPermissionError(goos, targetPath, err)
		}
	}

	if err := os.Rename(newBinaryPath, targetPath); err != nil {
		if hadExisting {
			_ = os.Rename(backupPath, targetPath)
		}
		return wrapPermissionError(goos, targetPath, err)
	}

	// Best-effort removal; if still locked by the running process,
	// CleanupStaleBackup removes it on the next invocation.
	_ = os.Remove(backupPath)
	return nil
}

// CleanupStaleBackup removes a leftover "<binary>.old" file from a previous Windows upgrade.
func CleanupStaleBackup(execPath string) {
	resolved := ResolveExecutablePath(execPath)
	if resolved == "" {
		return
	}
	_ = os.Remove(resolved + ".old")
}

func wrapPermissionError(goos, targetPath string, err error) error {
	if !errors.Is(err, os.ErrPermission) {
		return err
	}
	if goos == "windows" {
		return fmt.Errorf("permission denied writing %s; try running from an Administrator terminal", targetPath)
	}
	return fmt.Errorf("permission denied writing %s; try: sudo airvoice upgrade", targetPath)
}
