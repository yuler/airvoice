package upgrade

import (
	"context"
	"os"
	"strings"
	"time"
)

const (
	// DefaultCheckInterval is the TTL and periodic refresh cadence for background checks.
	DefaultCheckInterval = 24 * time.Hour

	// DefaultCheckTimeout bounds each background network request.
	DefaultCheckTimeout = 5 * time.Second

	// DisableEnvVar opts out of background update checks when set to a truthy value.
	DisableEnvVar = "AIRVOICE_NO_UPDATE_CHECK"
)

// BackgroundOptions configures StartBackgroundCheck.
type BackgroundOptions struct {
	CurrentVersion string
	ExecPath       string
	CachePath      string
	Interval       time.Duration
	Timeout        time.Duration
	Client         *Client
	Now            func() time.Time
	Notify         func(current, latest, hint string)
}

// IsCheckDisabled reports whether background update checks are disabled via environment.
func IsCheckDisabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(DisableEnvVar))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// StartBackgroundCheck immediately checks the local cache (synchronously, <1ms)
// and launches a background goroutine tied to ctx that refreshes expired caches
// and re-checks periodically while the server runs.
//
// It returns a channel that closes when the background goroutine exits.
func StartBackgroundCheck(ctx context.Context, opts BackgroundOptions) <-chan struct{} {
	done := make(chan struct{})

	if opts.ExecPath != "" {
		CleanupStaleBackup(opts.ExecPath)
	}

	current := Normalize(opts.CurrentVersion)
	if IsCheckDisabled() || opts.Notify == nil {
		close(done)
		return done
	}
	if _, err := Parse(current); err != nil {
		// Skip dev and non-semver builds.
		close(done)
		return done
	}

	interval := opts.Interval
	if interval <= 0 {
		interval = DefaultCheckInterval
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultCheckTimeout
	}
	nowFn := opts.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	cachePath := opts.CachePath
	if cachePath == "" {
		cachePath = CachePath()
	}
	client := opts.Client
	if client == nil {
		client = NewClient("airvoice/" + current)
	}
	hint := UpgradeHint(opts.ExecPath)

	lastNotified := ""
	var initialDelay time.Duration

	if entry, err := LoadCache(cachePath); err == nil {
		if IsNewer(current, entry.LatestVersion) {
			lastNotified = entry.LatestVersion
			opts.Notify(current, entry.LatestVersion, hint)
		}
		now := nowFn()
		if IsFresh(entry, now, interval) {
			remaining := interval - now.Sub(entry.CheckedAt)
			if remaining > 0 {
				initialDelay = remaining
			}
		}
	}

	checkOnce := func() {
		reqCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		latest, err := client.FetchLatestVersion(reqCtx)
		if err != nil {
			return
		}
		_ = SaveCache(cachePath, CacheEntry{
			CheckedAt:     nowFn().UTC(),
			LatestVersion: latest,
		})
		if IsNewer(current, latest) && latest != lastNotified {
			lastNotified = latest
			opts.Notify(current, latest, hint)
		}
	}

	go func() {
		defer close(done)

		if initialDelay > 0 {
			timer := time.NewTimer(initialDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		} else {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		checkOnce()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				checkOnce()
			}
		}
	}()

	return done
}
