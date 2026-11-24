# 22-cli-upgrade-design

## Description

Airvoice CLI currently has no built-in self-upgrade command or update notification mechanism. Users who install the CLI via `install.sh`, Homebrew, or manual GitHub Release downloads must manually check GitHub Releases and reinstall whenever a new version ships.

This design adds:

1. A pure-Go `airvoice upgrade` subcommand with SHA-256 verification and atomic binary replacement across macOS, Linux, and Windows.
2. A non-blocking background update check for the long-running `airvoice` server with a 24-hour local cache and periodic refresh.

---

## Proposed Solution

### 1. CLI Command Surface

```text
airvoice upgrade [--check] [--version X.Y.Z]
```

| Invocation                            | Behavior                                                                                   |
| ------------------------------------- | ------------------------------------------------------------------------------------------ |
| `airvoice upgrade`                    | Fetch latest release, verify SHA-256 against `checksums-cli.txt`, atomically replace self. |
| `airvoice upgrade --check`            | Check latest release, refresh cache, print whether an update is available without writing. |
| `airvoice upgrade --version 0.4.8`    | Install or roll back to a specific `vX.Y.Z` release (`v` prefix optional).                 |
| `airvoice upgrade` (Homebrew install) | Refuse direct Cellar overwrite and print `brew upgrade airvoice`.                          |

### 2. Version Discovery & Verification

- **Latest version check**: `GET https://github.com/yuler/airvoice/releases/latest` with redirect following disabled (`http.ErrUseLastResponse`). Extract `X.Y.Z` from the `Location` header (`/releases/tag/vX.Y.Z`). This avoids GitHub REST API's 60 req/hr unauthenticated rate limit while still honoring `HTTP_PROXY` / `HTTPS_PROXY`.
- **Release assets**: Maps `runtime.GOOS` / `runtime.GOARCH` to the existing CI release filenames (`airvoice-cli-darwin-arm64`, `airvoice-cli-darwin-amd64`, `airvoice-cli-linux-arm64`, `airvoice-cli-linux-amd64`, `airvoice-cli-windows-amd64.exe`).
- **Integrity verification**: Downloads `checksums-cli.txt` for the target tag, streams the binary into a temporary file inside the target executable's directory (ensuring same-filesystem rename and fail-fast permission checks), computes SHA-256 while streaming, and aborts on mismatch before touching the installed binary.
- **Atomic replacement**:
  - **macOS / Linux**: Preserves executable permissions (`0755` fallback) and performs POSIX atomic `os.Rename(tmp, target)`.
  - **Windows**: Renames the running `airvoice.exe` to `airvoice.exe.old`, moves the verified binary into `airvoice.exe` (rolling back on error), and removes stale `.old` files on subsequent launches/upgrades.
- **Permission errors**: If the target directory is not writable (e.g. `/usr/local/bin`), fails before downloading the binary with an actionable hint (`sudo airvoice upgrade`).

### 3. Homebrew Detection

Resolves symlinks on `os.Executable()` via `filepath.EvalSymlinks`. If the resolved path contains `/Cellar/airvoice/` (case-insensitive, slash-normalized), the install is treated as Homebrew-managed:

- `airvoice upgrade` refuses to overwrite the Cellar binary and directs the user to run `brew upgrade airvoice`.
- Background update notices and `airvoice upgrade --check` suggest `brew upgrade airvoice` instead of `airvoice upgrade`.

### 4. Background Update Check Strategy

Triggered **only** when starting the long-running `airvoice` server (not on short subcommands like `version`, `doctor`, `settings`, or `token refresh`, keeping their output clean and instant).

- **Cache file**: `~/.airvoice/update-check.json` (directory `0700`, file `0600`, atomic temp-file rename):
  ```json
  {
    "checked_at": "2026-11-24T17:00:00Z",
    "latest_version": "0.4.8"
  }
  ```
- **TTL / Interval**: 24 hours (`DefaultCheckInterval = 24 * time.Hour`), network timeout 5 seconds (`DefaultCheckTimeout = 5 * time.Second`).
- **Skip conditions**:
  - `dev` builds (or any non-semver `productVersion()`).
  - `AIRVOICE_NO_UPDATE_CHECK` set to a truthy value (`1`, `true`, `yes`, `on`).
- **Startup & long-running lifecycle**:
  1. **Synchronous cache read (< 1ms)**: On server startup, read `~/.airvoice/update-check.json`. If cached `latest_version > current_version`, immediately print a single notice line to `stderr`:
     `  [airvoice] update available: v0.4.7 -> v0.4.8 (run \`airvoice upgrade\`)`
  2. **Background goroutine**:
     - If the cache is missing or older than 24h, fetch the latest version immediately in the background; otherwise wait for the remaining TTL before the first refresh.
     - Continue checking on a 24-hour ticker while the server stays running.
     - Network, timeout, and cache I/O errors are silently ignored so offline usage is never disrupted.
     - Deduplicates notices in memory: each newer version is printed at most once per server process.

---

## Architectural & Code Changes

### Files to add

- `cli/upgrade/version.go`: Semver normalization, parsing, and comparison.
- `cli/upgrade/install.go`: Platform asset mapping, symlink resolution, and Homebrew detection.
- `cli/upgrade/cache.go`: `~/.airvoice/update-check.json` atomic persistence and TTL freshness checks.
- `cli/upgrade/release.go`: GitHub Release redirect lookup, `checksums-cli.txt` parsing, and SHA-256 verified downloads.
- `cli/upgrade/replace.go`: Atomic cross-platform executable replacement and Windows `.old` cleanup.
- `cli/upgrade/check.go`: Non-blocking server background checker with startup cache read and 24h ticker.
- `cli/upgrade/upgrade.go`: `airvoice upgrade` orchestration (`--check`, `--version`, Homebrew guard).
- `cli/upgrade/*_test.go`: Unit tests covering semver, Homebrew detection, cache I/O, checksum/download verification, atomic replace (Unix & Windows paths), background ticker/dedup/opt-out, and end-to-end upgrade flow with `httptest.Server`.

### Files to modify

- `cli/main.go`: Register `airvoice upgrade` subcommand, update usage output, and invoke `upgrade.StartBackgroundCheck` in `runServer`.
- `cli/main_test.go`: Verify `upgrade` in help output and flag/dispatch wiring.
- `README.md`: Document `airvoice upgrade` under Installation / Quick Start.

### Out of scope

- Silent automatic binary replacement without user confirmation.
- Desktop GUI (Wails) auto-updater.
- Changes to `www/public/install.sh` or GitHub Release workflow artifacts.

---

## Verification Plan

1. Run `go test -v -race ./cli/...` locally.
2. Verify CI `cli-test` passes across `ubuntu-latest`, `macos-latest`, and `windows-latest`.
3. Manual smoke check: `airvoice upgrade --check` against live GitHub Releases and `airvoice help` output.
