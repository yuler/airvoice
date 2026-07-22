# Desktop Linux AppImage Release

## Goal

Ship the Linux Desktop client as a self-contained AppImage so users do not need system `webkit2gtk-4.0`. Replace the current `Airvoice-Desktop-*-Linux.tar.gz` release asset and point website download links at the AppImage.

## Non-goals

- Flatpak packaging or Flathub publishing
- Migrating Desktop from Wails v2 to Wails v3
- Changing Desktop application behavior (UI, tray, paste)
- Changing macOS / Windows release artifacts

## Context

- Desktop is Wails v2.12 (`github.com/wailsapp/wails/v2`).
- Current Linux release is a naked binary in a tarball, linked against system WebKitGTK.
- Many newer distros only provide `webkit2gtk-4.1` (or none of 4.0), so the tarball fails at runtime.
- Wails v2 has no first-class AppImage generator; packaging uses `linuxdeploy` (+ GTK plugin) and `appimagetool`.

## Architecture

Keep building on `ubuntu-22.04` in `.github/workflows/desktop-release.yml` for older glibc compatibility.

```text
wails build (linux/amd64)
        │
        ▼
desktop/scripts/package-appimage.sh
  · stage AppDir (binary, .desktop, icon)
  · linuxdeploy + plugin-gtk (bundle GTK/WebKit libs)
  · deploy WebKit helper process deps
  · set WEBKIT_EXEC_PATH for relocatable helpers
  · appimagetool → .AppImage
        │
        ▼
GitHub Release asset
Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage
```

## Packaging script

**Path:** `desktop/scripts/package-appimage.sh`

**Inputs:** path to `build/bin/Airvoice`, version string, output directory.

**Assets (committed):**

- `desktop/build/linux/Airvoice.desktop` — Name/Exec/Icon/Categories for AppImage
- App icon — reuse `desktop/tray_icon.png` or a dedicated PNG under `desktop/build/linux/`

**AppDir layout:**

- `usr/bin/Airvoice`
- `usr/share/applications/Airvoice.desktop`
- Icon under hicolor (or linuxdeploy-compatible icon path)

**Tooling (CI):**

- Download fixed versions of `linuxdeploy`, `linuxdeploy-plugin-gtk`, and `appimagetool` for `x86_64`
- Run `linuxdeploy` with `--plugin=gtk`
- Additionally `--deploy-deps-only` for WebKit helpers (`WebKitNetworkProcess`, `WebKitWebProcess`, `WebKitGPUProcess`, and related binaries as needed)
- Ensure runtime can find helpers via `WEBKIT_EXEC_PATH` (AppRun wrapper or equivalent) so WebKit does not look under hardcoded `/usr/lib/.../webkit2gtk-4.0/`

**Output:** `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage` (executable bit set).

Local verification: `wails build` then run the same script.

## Release workflow changes

**File:** `.github/workflows/desktop-release.yml`

- After Linux `wails build`, invoke `package-appimage.sh`
- Collect only the `.AppImage` for Linux (stop uploading `tar.gz` and opportunistic `.deb` copies)
- macOS / Windows matrix entries unchanged

## Website / sync changes

**File:** `www/src/lib/downloads.ts`

- Linux desktop URL → `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage`

**File:** `.github/workflows/www-downloads-link.yml`

- Required asset list: replace `Airvoice-Desktop-${VERSION}-Linux.tar.gz` with `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage`

## Docs (light)

Where README or platform-deps mention Linux Desktop needing system `webkit2gtk`, update to prefer the AppImage (no system WebKit required for Desktop). CLI paste deps (`xclip` / `ydotool` / etc.) remain unchanged.

## Success criteria

1. Tag release produces `Airvoice-Desktop-*-Linux-x86_64.AppImage` and does not produce Linux `tar.gz`.
2. Website Linux desktop download points at that AppImage.
3. AppImage runs on a distro without `webkit2gtk-4.0` installed (e.g. Arch / Ubuntu 24.04), without installing WebKit system packages.
4. macOS and Windows release assets remain as today.
