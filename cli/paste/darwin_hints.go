//go:build darwin

package paste

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AccessibilityTarget is an app or binary that may need Accessibility on macOS.
type AccessibilityTarget struct {
	Label string
	Path  string
}

func darwinAccessibilityTargets() []AccessibilityTarget {
	seen := map[string]bool{}
	var out []AccessibilityTarget

	add := func(label, path string) {
		path = strings.TrimSpace(path)
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		out = append(out, AccessibilityTarget{Label: label, Path: path})
	}

	if host := detectHostTerminalApp(); host.Path != "" {
		add(host.Label, host.Path)
	}

	exe, err := os.Executable()
	if err == nil {
		if abs, err := filepath.EvalSymlinks(exe); err == nil {
			exe = abs
		}
		add("airvoice", exe)
	}

	add("osascript", "/usr/bin/osascript")

	return out
}

func detectHostTerminalApp() AccessibilityTarget {
	if host := parentAppBundle(); host.Path != "" {
		return host
	}

	switch tp := os.Getenv("TERM_PROGRAM"); tp {
	case "Apple_Terminal":
		return AccessibilityTarget{Label: "Terminal", Path: findAppBundle("Terminal.app")}
	case "iTerm.app", "iTerm2":
		return AccessibilityTarget{Label: "iTerm", Path: findAppBundle("iTerm.app")}
	case "WarpTerminal":
		return AccessibilityTarget{Label: "Warp", Path: findAppBundle("Warp.app")}
	case "ghostty":
		return AccessibilityTarget{Label: "Ghostty", Path: findAppBundle("Ghostty.app")}
	case "vscode":
		if path := findAppBundle("Cursor.app"); path != "" {
			return AccessibilityTarget{Label: "Cursor", Path: path}
		}
		return AccessibilityTarget{Label: "Visual Studio Code", Path: findAppBundle("Visual Studio Code.app")}
	default:
		if strings.Contains(strings.ToLower(tp), "cursor") {
			return AccessibilityTarget{Label: "Cursor", Path: findAppBundle("Cursor.app")}
		}
	}

	return AccessibilityTarget{}
}

func findAppBundle(name string) string {
	candidates := []string{
		filepath.Join("/Applications", name),
		filepath.Join(os.Getenv("HOME"), "Applications", name),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return ""
}

func parentAppBundle() AccessibilityTarget {
	pid := os.Getppid()
	for i := 0; i < 12 && pid > 1; i++ {
		path := processExecutablePath(pid)
		if app := appBundleFromPath(path); app.Path != "" {
			return app
		}
		pid = parentPID(pid)
	}
	return AccessibilityTarget{}
}

func appBundleFromPath(path string) AccessibilityTarget {
	idx := strings.Index(path, ".app/")
	if idx < 0 {
		return AccessibilityTarget{}
	}
	appPath := path[:idx+4]
	name := strings.TrimSuffix(filepath.Base(appPath), ".app")
	return AccessibilityTarget{Label: name, Path: appPath}
}

func processExecutablePath(pid int) string {
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "comm=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func parentPID(pid int) int {
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "ppid=").Output()
	if err != nil {
		return 0
	}
	var ppid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &ppid); err != nil {
		return 0
	}
	return ppid
}

func darwinAccessibilityHintText() string {
	targets := darwinAccessibilityTargets()
	var b strings.Builder
	b.WriteString("系统设置 → 隐私与安全性 → 辅助功能，启用以下全部项后完全退出并重启应用：")
	for _, t := range targets {
		b.WriteString(fmt.Sprintf("\n• %s（%s）", t.Label, t.Path))
	}
	b.WriteString("\n运行 ./bin/airvoice doctor 检查；若只勾选了 Terminal 但用 Cursor 终端运行，仍会失败")
	return b.String()
}
