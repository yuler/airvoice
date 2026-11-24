package main

import (
	"runtime/debug"
	"strings"
)

// readBuildInfo is replaced in tests.
var readBuildInfo = debug.ReadBuildInfo

// productVersion prefers the ldflags override used by release builds.
// go install does not pass those flags, so it falls back to the module version
// recorded in the binary (for example v0.4.6 from @latest).
func productVersion() string {
	if v := strings.TrimSpace(version); v != "" && v != "dev" {
		return v
	}
	info, ok := readBuildInfo()
	if !ok {
		return "dev"
	}
	return normalizeModuleVersion(info.Main.Version)
}

func normalizeModuleVersion(raw string) string {
	v := strings.TrimSpace(raw)
	v = strings.TrimPrefix(v, "v")
	if v == "" || v == "(devel)" {
		return "dev"
	}
	return v
}
