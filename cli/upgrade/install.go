package upgrade

import (
	"fmt"
	"path/filepath"
	"strings"
)

// AssetName returns the GitHub Release asset filename for the given platform.
func AssetName(goos, goarch string) (string, error) {
	switch goos {
	case "darwin", "linux":
		switch goarch {
		case "arm64", "amd64":
			return fmt.Sprintf("airvoice-cli-%s-%s", goos, goarch), nil
		}
	case "windows":
		if goarch == "amd64" {
			return "airvoice-cli-windows-amd64.exe", nil
		}
	}
	return "", fmt.Errorf("unsupported platform %s/%s", goos, goarch)
}

// ResolveExecutablePath follows symlinks when possible, falling back to execPath.
func ResolveExecutablePath(execPath string) string {
	if execPath == "" {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return execPath
	}
	return resolved
}

// IsHomebrewPath reports whether path points inside a Homebrew Cellar for airvoice.
func IsHomebrewPath(path string) bool {
	if path == "" {
		return false
	}
	normalized := strings.ToLower(filepath.ToSlash(path))
	return strings.Contains(normalized, "/cellar/airvoice/")
}

// IsHomebrewInstall reports whether execPath or its symlink target lives in Homebrew Cellar.
func IsHomebrewInstall(execPath string) bool {
	if IsHomebrewPath(execPath) {
		return true
	}
	return IsHomebrewPath(ResolveExecutablePath(execPath))
}

// UpgradeHint returns the recommended upgrade command for the current installation.
func UpgradeHint(execPath string) string {
	if IsHomebrewInstall(execPath) {
		return "brew upgrade airvoice"
	}
	return "airvoice upgrade"
}
