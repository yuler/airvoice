package upgrade

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Options configures a foreground `airvoice upgrade` execution.
type Options struct {
	CurrentVersion string
	TargetVersion  string
	CheckOnly      bool
	ExecPath       string
	CachePath      string
	GOOS           string
	GOARCH         string
	Client         *Client
	Out            io.Writer
	Now            func() time.Time
}

// Run executes `airvoice upgrade` (or `airvoice upgrade --check`).
func Run(ctx context.Context, opts Options) error {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	execPath := opts.ExecPath
	if execPath == "" {
		var err error
		execPath, err = os.Executable()
		if err != nil {
			return fmt.Errorf("locate current executable: %w", err)
		}
	}
	cachePath := opts.CachePath
	if cachePath == "" {
		cachePath = CachePath()
	}
	current := Normalize(opts.CurrentVersion)
	if current == "" {
		current = "dev"
	}
	client := opts.Client
	if client == nil {
		client = NewClient("airvoice/" + current)
	}

	CleanupStaleBackup(execPath)

	target := ""
	explicitVersion := strings.TrimSpace(opts.TargetVersion) != ""
	if explicitVersion {
		if _, err := Parse(opts.TargetVersion); err != nil {
			return fmt.Errorf("invalid --version %q (expected X.Y.Z)", opts.TargetVersion)
		}
		target = Normalize(opts.TargetVersion)
	} else {
		latest, err := client.FetchLatestVersion(ctx)
		if err != nil {
			return err
		}
		target = latest
		_ = SaveCache(cachePath, CacheEntry{
			CheckedAt:     nowFn().UTC(),
			LatestVersion: latest,
		})
	}

	hint := UpgradeHint(execPath)

	if opts.CheckOnly {
		return printCheckStatus(out, current, target, hint, explicitVersion)
	}

	if IsHomebrewInstall(execPath) {
		return fmt.Errorf("airvoice is installed via Homebrew; upgrade with:\n  brew upgrade airvoice")
	}

	if current == target {
		fmt.Fprintf(out, "airvoice v%s is already up to date.\n", target)
		return nil
	}

	asset, err := AssetName(goos, goarch)
	if err != nil {
		return err
	}

	expectedSHA, err := client.FetchChecksum(ctx, target, asset)
	if err != nil {
		return err
	}

	realTarget := ResolveExecutablePath(execPath)
	destDir := filepath.Dir(realTarget)

	fmt.Fprintf(out, "Downloading airvoice v%s (%s)...\n", target, asset)
	tmpPath, err := client.DownloadVerified(ctx, target, asset, expectedSHA, destDir)
	if err != nil {
		return wrapPermissionError(goos, realTarget, err)
	}

	if err := replaceExecutableForOS(goos, realTarget, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	action := "Upgraded"
	if cmp, err := Compare(current, target); err == nil && cmp > 0 {
		action = "Downgraded"
	}
	fmt.Fprintf(out, "%s airvoice: %s -> v%s\n", action, formatVersionLabel(current), target)
	return nil
}

func printCheckStatus(out io.Writer, current, target, hint string, explicitVersion bool) error {
	if explicitVersion {
		if current == target {
			fmt.Fprintf(out, "airvoice is already on v%s.\n", target)
			return nil
		}
		fmt.Fprintf(out, "Current: %s, target: v%s\n", formatVersionLabel(current), target)
		return nil
	}

	if IsNewer(current, target) {
		fmt.Fprintf(out, "Update available: v%s -> v%s\nRun: %s\n", current, target, hint)
		return nil
	}
	if _, err := Parse(current); err != nil {
		fmt.Fprintf(out, "Current version: %s (latest release: v%s)\n", current, target)
		return nil
	}
	fmt.Fprintf(out, "airvoice v%s is already up to date.\n", current)
	return nil
}

func formatVersionLabel(v string) string {
	if _, err := Parse(v); err == nil {
		return "v" + v
	}
	return v
}
