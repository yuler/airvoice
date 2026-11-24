package upgrade

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type notice struct {
	current string
	latest  string
	hint    string
}

func TestBackgroundCheckSkipsDevAndEnvDisable(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Location", "/yuler/airvoice/releases/tag/v0.9.0")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	called := false
	done := StartBackgroundCheck(context.Background(), BackgroundOptions{
		CurrentVersion: "dev",
		CachePath:      filepath.Join(t.TempDir(), "update-check.json"),
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Notify:         func(string, string, string) { called = true },
	})
	<-done
	if called || hits.Load() != 0 {
		t.Fatalf("dev build should not check or notify: called=%v hits=%d", called, hits.Load())
	}

	t.Setenv(DisableEnvVar, "1")
	done = StartBackgroundCheck(context.Background(), BackgroundOptions{
		CurrentVersion: "0.4.7",
		CachePath:      filepath.Join(t.TempDir(), "update-check.json"),
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Notify:         func(string, string, string) { called = true },
	})
	<-done
	if called || hits.Load() != 0 {
		t.Fatalf("disabled env should not check or notify: called=%v hits=%d", called, hits.Load())
	}
}

func TestBackgroundCheckUsesFreshCacheWithoutNetwork(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Location", "/yuler/airvoice/releases/tag/v0.4.9")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	now := time.Date(2026, 11, 24, 17, 0, 0, 0, time.UTC)
	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	if err := SaveCache(cachePath, CacheEntry{
		CheckedAt:     now.Add(-time.Hour),
		LatestVersion: "0.4.8",
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var got []notice
	done := StartBackgroundCheck(ctx, BackgroundOptions{
		CurrentVersion: "0.4.7",
		ExecPath:       "/opt/homebrew/Cellar/airvoice/0.4.7/bin/airvoice",
		CachePath:      cachePath,
		Interval:       24 * time.Hour,
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Now:            func() time.Time { return now },
		Notify: func(current, latest, hint string) {
			got = append(got, notice{current, latest, hint})
		},
	})
	// Cache check is synchronous, so notification is already delivered before cancel.
	cancel()
	<-done

	if hits.Load() != 0 {
		t.Fatalf("fresh cache should not hit network, hits=%d", hits.Load())
	}
	if len(got) != 1 || got[0] != (notice{"0.4.7", "0.4.8", "brew upgrade airvoice"}) {
		t.Fatalf("got notices %+v", got)
	}
}

func TestBackgroundCheckRefreshesExpiredCacheAndDeduplicatesPeriodic(t *testing.T) {
	var mu sync.Mutex
	currentTag := "v0.4.8"
	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		mu.Lock()
		tag := currentTag
		mu.Unlock()
		w.Header().Set("Location", "/yuler/airvoice/releases/tag/"+tag)
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	// Pre-seed an expired cache with 0.4.8 so sync read notifies 0.4.8,
	// first background fetch sees 0.4.8 (deduped!), then ticker sees 0.4.9 (notified!).
	if err := SaveCache(cachePath, CacheEntry{
		CheckedAt:     time.Now().Add(-48 * time.Hour),
		LatestVersion: "0.4.8",
	}); err != nil {
		t.Fatal(err)
	}

	notifyCh := make(chan notice, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := StartBackgroundCheck(ctx, BackgroundOptions{
		CurrentVersion: "0.4.7",
		ExecPath:       "/home/user/.local/bin/airvoice",
		CachePath:      cachePath,
		Interval:       40 * time.Millisecond,
		Timeout:        time.Second,
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Notify: func(current, latest, hint string) {
			notifyCh <- notice{current, latest, hint}
		},
	})

	// First notice comes synchronously from cache (0.4.8).
	first := <-notifyCh
	if first != (notice{"0.4.7", "0.4.8", "airvoice upgrade"}) {
		t.Fatalf("first notice = %+v", first)
	}

	// Wait until at least one network hit occurs while tag is still v0.4.8.
	deadline := time.Now().Add(time.Second)
	for hits.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if hits.Load() == 0 {
		t.Fatal("expected background check to hit network for expired cache")
	}

	// Ensure 0.4.8 was not notified a second time.
	select {
	case dup := <-notifyCh:
		t.Fatalf("unexpected duplicate notice %+v", dup)
	default:
	}

	// Bump remote release to v0.4.9; periodic ticker should pick it up and notify.
	mu.Lock()
	currentTag = "v0.4.9"
	mu.Unlock()

	select {
	case second := <-notifyCh:
		if second != (notice{"0.4.7", "0.4.9", "airvoice upgrade"}) {
			t.Fatalf("second notice = %+v", second)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for periodic ticker notice for 0.4.9")
	}

	cancel()
	<-done

	cached, err := LoadCache(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if cached.LatestVersion != "0.4.9" {
		t.Fatalf("cached LatestVersion = %q, want 0.4.9", cached.LatestVersion)
	}
}

func TestBackgroundCheckIgnoresNetworkErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	called := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	done := StartBackgroundCheck(ctx, BackgroundOptions{
		CurrentVersion: "0.4.7",
		CachePath:      filepath.Join(t.TempDir(), "update-check.json"),
		Interval:       time.Hour,
		Timeout:        time.Second,
		Client:         &Client{BaseURL: srv.URL, HTTPClient: srv.Client()},
		Notify:         func(string, string, string) { called <- struct{}{} },
	})
	time.Sleep(30 * time.Millisecond)
	cancel()
	<-done

	select {
	case <-called:
		t.Fatal("should not notify on server error")
	default:
	}
}
