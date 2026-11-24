package main

import (
	"runtime/debug"
	"testing"
)

func TestProductVersionPrefersLdflags(t *testing.T) {
	orig := version
	version = "0.4.6"
	t.Cleanup(func() { version = orig })

	if got := productVersion(); got != "0.4.6" {
		t.Fatalf("productVersion() = %q, want 0.4.6", got)
	}
}

func TestProductVersionUsesModuleVersion(t *testing.T) {
	origVersion := version
	origRead := readBuildInfo
	version = "dev"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.4.6"}}, true
	}
	t.Cleanup(func() {
		version = origVersion
		readBuildInfo = origRead
	})

	if got := productVersion(); got != "0.4.6" {
		t.Fatalf("productVersion() = %q, want 0.4.6", got)
	}
}

func TestProductVersionDevWhenUntagged(t *testing.T) {
	origVersion := version
	origRead := readBuildInfo
	version = "dev"
	readBuildInfo = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	}
	t.Cleanup(func() {
		version = origVersion
		readBuildInfo = origRead
	})

	if got := productVersion(); got != "dev" {
		t.Fatalf("productVersion() = %q, want dev", got)
	}
}

func TestNormalizeModuleVersion(t *testing.T) {
	cases := map[string]string{
		"v0.4.6":   "0.4.6",
		"0.4.6":    "0.4.6",
		"(devel)":  "dev",
		"":         "dev",
		"  v1.2.3": "1.2.3",
	}
	for in, want := range cases {
		if got := normalizeModuleVersion(in); got != want {
			t.Fatalf("normalizeModuleVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
