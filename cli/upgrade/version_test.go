package upgrade

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAndCompare(t *testing.T) {
	valid := []struct {
		raw  string
		want Semver
	}{
		{"0.4.8", Semver{0, 4, 8}},
		{"v0.4.8", Semver{0, 4, 8}},
		{"  v1.20.300  ", Semver{1, 20, 300}},
	}
	for _, tc := range valid {
		got, err := Parse(tc.raw)
		if err != nil {
			t.Fatalf("Parse(%q) unexpected err: %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("Parse(%q) = %+v, want %+v", tc.raw, got, tc.want)
		}
	}

	for _, bad := range []string{"", "dev", "0.4", "0.4.8-rc1", "v1.2.x", "-1.0.0"} {
		if _, err := Parse(bad); err == nil {
			t.Fatalf("Parse(%q) expected error, got nil", bad)
		}
	}

	if !IsNewer("0.4.7", "0.4.8") {
		t.Fatal("expected 0.4.8 > 0.4.7")
	}
	if !IsNewer("0.4.9", "0.10.0") {
		t.Fatal("expected 0.10.0 > 0.4.9")
	}
	if IsNewer("0.4.8", "0.4.8") {
		t.Fatal("equal versions should not be newer")
	}
	if IsNewer("0.5.0", "0.4.9") {
		t.Fatal("older version should not be newer")
	}
	if IsNewer("dev", "0.4.8") {
		t.Fatal("dev build should not report IsNewer")
	}
}

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "airvoice-cli-darwin-arm64"},
		{"darwin", "amd64", "airvoice-cli-darwin-amd64"},
		{"linux", "arm64", "airvoice-cli-linux-arm64"},
		{"linux", "amd64", "airvoice-cli-linux-amd64"},
		{"windows", "amd64", "airvoice-cli-windows-amd64.exe"},
	}
	for _, tc := range cases {
		got, err := AssetName(tc.goos, tc.goarch)
		if err != nil {
			t.Fatalf("AssetName(%s,%s) err: %v", tc.goos, tc.goarch, err)
		}
		if got != tc.want {
			t.Fatalf("AssetName(%s,%s) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
	if _, err := AssetName("windows", "arm64"); err == nil {
		t.Fatal("expected error for unsupported windows/arm64")
	}
	if _, err := AssetName("freebsd", "amd64"); err == nil {
		t.Fatal("expected error for unsupported freebsd/amd64")
	}
}

func TestHomebrewDetection(t *testing.T) {
	brewPaths := []string{
		"/opt/homebrew/Cellar/airvoice/0.4.8/bin/airvoice",
		"/usr/local/Cellar/airvoice/0.4.8/bin/airvoice",
		"/home/linuxbrew/.linuxbrew/Cellar/airvoice/0.4.8/bin/airvoice",
	}
	for _, p := range brewPaths {
		if !IsHomebrewPath(p) {
			t.Fatalf("expected Homebrew path for %q", p)
		}
		if got := UpgradeHint(p); got != "brew upgrade airvoice" {
			t.Fatalf("UpgradeHint(%q) = %q, want brew upgrade airvoice", p, got)
		}
	}

	nonBrewPaths := []string{
		"/usr/local/bin/airvoice",
		"/home/user/.local/bin/airvoice",
		`C:\Users\user\bin\airvoice.exe`,
	}
	for _, p := range nonBrewPaths {
		if IsHomebrewPath(p) {
			t.Fatalf("did not expect Homebrew path for %q", p)
		}
		if got := UpgradeHint(p); got != "airvoice upgrade" {
			t.Fatalf("UpgradeHint(%q) = %q, want airvoice upgrade", p, got)
		}
	}

	// Verify symlink resolution detects /opt/homebrew/bin/airvoice -> .../Cellar/airvoice/...
	root := t.TempDir()
	cellarBin := filepath.Join(root, "Cellar", "airvoice", "0.4.8", "bin", "airvoice")
	if err := os.MkdirAll(filepath.Dir(cellarBin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cellarBin, []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(root, "bin", "airvoice")
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(cellarBin, linkPath); err != nil {
		t.Skipf("symlink creation not permitted in this environment: %v", err)
	}
	if !IsHomebrewInstall(linkPath) {
		t.Fatalf("expected symlink %q -> %q to be detected as Homebrew", linkPath, cellarBin)
	}
}
