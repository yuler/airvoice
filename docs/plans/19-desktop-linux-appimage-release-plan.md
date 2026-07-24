# Desktop Linux AppImage Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Linux Desktop `tar.gz` release with a self-contained AppImage so users do not need system `webkit2gtk-4.0`, and point website download links at that AppImage.

**Architecture:** Keep Wails v2 builds on `ubuntu-22.04`. After `wails build`, run `desktop/scripts/package-appimage.sh` (pinned linuxdeploy + GTK plugin, WebKit helper copy, `WEBKIT_EXEC_PATH` in custom AppRun — no `.so` path rewrite). Upload only the `.AppImage` for Linux. Update `downloads.ts` and the WWW sync required-assets list.

**Tech Stack:** Wails v2.12, linuxdeploy `1-alpha-20251107-1`, appimagetool `1.9.0`, linuxdeploy-plugin-gtk (pinned commit), GitHub Actions, bash.

## Global Constraints

- Linux release asset name: `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage`
- Stop shipping `Airvoice-Desktop-*-Linux.tar.gz` and Linux `.deb` artifacts
- Do not migrate to Wails v3 or add Flatpak
- Do not change macOS / Windows release artifacts
- Spec: `docs/specs/19-desktop-linux-appimage-release-design.md`
- Plans live only under `docs/plans/`

## File Structure

| File | Responsibility |
|------|----------------|
| `desktop/packaging/linux/Airvoice.desktop` | FreeDesktop metadata for AppImage |
| `desktop/packaging/linux/airvoice.png` | App icon (copy of tray icon) |
| `desktop/scripts/package-appimage.sh` | Stage AppDir, bundle deps, emit AppImage |
| `.github/workflows/desktop-release.yml` | Call packager; collect `.AppImage` only |
| `www/src/lib/downloads.ts` | Linux desktop download URL |
| `.github/workflows/www-downloads-link.yml` | Required release asset name |
| `www/src/content/docs/{en,zh}/platform-deps.md` | Document AppImage / no system WebKit for Desktop |

---

### Task 1: Desktop metadata assets

**Files:**
- Create: `desktop/packaging/linux/Airvoice.desktop`
- Create: `desktop/packaging/linux/airvoice.png` (copied from `desktop/tray_icon.png`)

**Interfaces:**
- Produces: desktop file with `Name=Airvoice`, `Exec=Airvoice`, `Icon=airvoice`; PNG named `airvoice.png`

- [ ] **Step 1: Create the desktop file**

Create `desktop/packaging/linux/Airvoice.desktop`:

```desktop
[Desktop Entry]
Type=Application
Name=Airvoice
Comment=Speak on your phone, type on your desktop
Exec=Airvoice
Icon=airvoice
Categories=Utility;Network;
Terminal=false
StartupNotify=true
```

- [ ] **Step 2: Copy the app icon**

```bash
mkdir -p desktop/packaging/linux
cp desktop/tray_icon.png desktop/packaging/linux/airvoice.png
```

Note: do **not** use `desktop/build/linux/` — `desktop/build/` is gitignored (Wails build output).

- [ ] **Step 3: Commit**

```bash
git add desktop/packaging/linux/Airvoice.desktop desktop/packaging/linux/airvoice.png
git commit -m "$(cat <<'EOF'
📦 [desktop]: Add Linux AppImage desktop entry and icon

EOF
)"
```

---

### Task 2: AppImage packaging script

**Files:**
- Create: `desktop/scripts/package-appimage.sh`

**Interfaces:**
- Consumes: binary path, version string, output directory; assets from Task 1
- Produces: `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage` in the output directory (executable)

- [ ] **Step 1: Write `desktop/scripts/package-appimage.sh`**

Create the file with this content (deploy AppDir → set `WEBKIT_EXEC_PATH` in AppRun → pack with appimagetool).

> **Note (post-review):** Do **not** sed-rewrite `/usr` → `././` inside bundled WebKit `.so` files — that produced incorrect helper paths. Relocatability is handled by `WEBKIT_EXEC_PATH` in AppRun. Packaging tools are pinned to versioned releases with sha256 verification (not `continuous`).

The authoritative script lives at `desktop/scripts/package-appimage.sh` (do not paste a stale copy here). Key behaviors:

- Stage AppDir (`usr/bin/Airvoice`, `.desktop`, PNG icon)
- Copy WebKit helpers from known `/usr/lib*` roots only
- Download + sha256-verify pinned `linuxdeploy`, `appimagetool`, and `linuxdeploy-plugin-gtk`
- Run linuxdeploy `--plugin gtk`, restore custom AppRun, pack with appimagetool
- Emit `Airvoice-Desktop-${VERSION}-Linux-${ARCH}.AppImage`

Make executable:

```bash
chmod +x desktop/scripts/package-appimage.sh
```

- [ ] **Step 2: Sanity-check script syntax**

Run:

```bash
bash -n desktop/scripts/package-appimage.sh
```

Expected: exit 0, no output.

- [ ] **Step 3: Commit**

```bash
git add desktop/scripts/package-appimage.sh
git commit -m "$(cat <<'EOF'
📦 [desktop]: Add Linux AppImage packaging script

EOF
)"
```

---

### Task 3: Wire packaging into desktop release workflow

**Files:**
- Modify: `.github/workflows/desktop-release.yml`

**Interfaces:**
- Consumes: `package-appimage.sh` from Task 2
- Produces: Linux release artifact `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage` only (no `tar.gz` / `.deb`)

- [ ] **Step 1: After Linux build, package AppImage**

In `.github/workflows/desktop-release.yml`, after the `Build` step, add:

```yaml
      - name: Package Linux AppImage
        if: matrix.platform == 'linux'
        working-directory: desktop
        run: |
          ./scripts/package-appimage.sh \
            --binary build/bin/Airvoice \
            --version "${{ steps.version.outputs.version }}" \
            --output-dir release
```

- [ ] **Step 2: Replace Linux artifact collection**

Replace the `linux)` branch inside `Collect artifacts` with:

```bash
            linux)
              # AppImage is produced by package-appimage.sh into release/
              if ! ls release/Airvoice-Desktop-*-Linux-*.AppImage >/dev/null 2>&1; then
                echo "Expected AppImage missing in release/" >&2
                exit 1
              fi
              ;;
```

Remove the previous `tar.gz` / `.deb` copy logic for Linux. Keep the outer `mkdir -p release` and the darwin/windows branches unchanged.

Ensure the AppImage packaging step writes into `desktop/release` **before** Collect runs, and Collect does not delete existing files. Because packaging already writes to `release/`, the linux branch only verifies presence.

Full Collect step for clarity after edit:

```yaml
      - name: Collect artifacts
        shell: bash
        working-directory: desktop
        run: |
          VERSION="${{ steps.version.outputs.version }}"
          PLATFORM="${{ matrix.platform }}"
          ARCH="${{ matrix.arch }}"
          mkdir -p release

          case "$PLATFORM" in
            darwin)
              [ -d "build/bin/Airvoice.app" ] && zip -r "release/Airvoice-Desktop-${VERSION}-macOS.zip" build/bin/Airvoice.app
              [ -f "build/bin/Airvoice.dmg" ] && cp build/bin/Airvoice.dmg "release/Airvoice-Desktop-${VERSION}-macOS.dmg"
              ;;
            linux)
              if ! ls release/Airvoice-Desktop-*-Linux-*.AppImage >/dev/null 2>&1; then
                echo "Expected AppImage missing in release/" >&2
                exit 1
              fi
              ;;
            windows)
              [ -f "build/bin/Airvoice.exe" ] && tar czf "release/Airvoice-Desktop-${VERSION}-Windows.zip" -C build/bin Airvoice.exe
              [ -f "build/bin/Airvoice-amd64.exe" ] && cp build/bin/Airvoice-amd64.exe "release/Airvoice-Desktop-${VERSION}-Windows-amd64.exe"
              [ -f "build/bin/Airvoice-${VERSION}-amd64.exe" ] && cp build/bin/Airvoice-${VERSION}-amd64.exe "release/Airvoice-Desktop-${VERSION}-Windows-amd64.exe"
              ;;
          esac

          echo "--- Released files ---"
          find release -type f -exec ls -lh {} \;
```

- [ ] **Step 3: Validate workflow YAML**

Run:

```bash
ruby -e "require 'yaml'; YAML.load_file('.github/workflows/desktop-release.yml'); puts 'ok'"
```

Expected: `ok`

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/desktop-release.yml
git commit -m "$(cat <<'EOF'
👷 [ci]: Publish Linux Desktop as AppImage instead of tar.gz

EOF
)"
```

---

### Task 4: Website download links and sync gate

**Files:**
- Modify: `www/src/lib/downloads.ts`
- Modify: `.github/workflows/www-downloads-link.yml`

**Interfaces:**
- Consumes: AppImage asset name from Global Constraints
- Produces: Linux desktop URLs and sync required-asset list pointing at AppImage

- [ ] **Step 1: Update `getDesktopDownloadUrl` and `getDownloadUrls`**

In `www/src/lib/downloads.ts`, change both Linux branches from:

```ts
`${BASE}/Airvoice-Desktop-${VERSION}-Linux.tar.gz`
```

to:

```ts
`${BASE}/Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage`
```

(two call sites: `getDesktopDownloadUrl` `case 'linux'`, and `getDownloadUrls` `else if (isLinux)`.)

- [ ] **Step 2: Update required assets in WWW sync workflow**

In `.github/workflows/www-downloads-link.yml`, replace:

```bash
            "Airvoice-Desktop-${VERSION}-Linux.tar.gz"
```

with:

```bash
            "Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage"
```

- [ ] **Step 3: Commit**

```bash
git add www/src/lib/downloads.ts .github/workflows/www-downloads-link.yml
git commit -m "$(cat <<'EOF'
🔗 [www]: Point Linux desktop downloads at AppImage

EOF
)"
```

---

### Task 5: Docs — Desktop AppImage note

**Files:**
- Modify: `www/src/content/docs/en/platform-deps.md`
- Modify: `www/src/content/docs/zh/platform-deps.md`

**Interfaces:**
- Produces: docs stating Desktop Linux AppImage bundles WebKit; CLI paste deps unchanged

- [ ] **Step 1: Add Desktop section (English)**

In `www/src/content/docs/en/platform-deps.md`, insert before `## macOS`:

````markdown
## Desktop (Linux AppImage)

Prefer the `Airvoice-Desktop-*-Linux-x86_64.AppImage` from [GitHub Releases](https://github.com/yuler/airvoice/releases/latest). It bundles WebKitGTK, so you do **not** need system `webkit2gtk-4.0` / `webkit2gtk-4.1` for the Desktop app.

Make it executable and run:

```bash
chmod +x Airvoice-Desktop-*-Linux-x86_64.AppImage
./Airvoice-Desktop-*-Linux-x86_64.AppImage
```

Clipboard / paste automation still needs the X11 or Wayland tools below when using Desktop or CLI.
````

- [ ] **Step 2: Add Desktop section (Chinese)**

In `www/src/content/docs/zh/platform-deps.md`, insert before `## macOS`:

````markdown
## Desktop（Linux AppImage）

优先从 [GitHub Releases](https://github.com/yuler/airvoice/releases/latest) 下载 `Airvoice-Desktop-*-Linux-x86_64.AppImage`。AppImage 已打包 WebKitGTK，**不需要**系统安装 `webkit2gtk-4.0` / `webkit2gtk-4.1`。

赋予可执行权限后运行：

```bash
chmod +x Airvoice-Desktop-*-Linux-x86_64.AppImage
./Airvoice-Desktop-*-Linux-x86_64.AppImage
```

使用 Desktop 或 CLI 时，剪贴板 / 模拟按键仍需要下方 X11 或 Wayland 工具。
````

- [ ] **Step 3: Commit**

```bash
git add www/src/content/docs/en/platform-deps.md www/src/content/docs/zh/platform-deps.md
git commit -m "$(cat <<'EOF'
📝 [docs]: Document Linux Desktop AppImage WebKit bundling

EOF
)"
```

---

### Task 6: Verify packaging on CI (workflow_dispatch)

**Files:**
- None (manual verification)

**Interfaces:**
- Consumes: Tasks 1–5 merged to a branch or `main`

- [ ] **Step 1: Trigger Desktop Release via workflow_dispatch**

In GitHub Actions UI, run **Desktop Release** (workflow_dispatch). Or:

```bash
gh workflow run desktop-release.yml
```

- [ ] **Step 2: Confirm Linux artifact**

```bash
gh run list --workflow=desktop-release.yml --limit 1
# after success:
gh run download <run-id> --name desktop-linux-amd64
ls -lh Airvoice-Desktop-*-Linux-x86_64.AppImage
```

Expected: one executable AppImage; no `tar.gz`.

- [ ] **Step 3: Smoke-test AppImage on a machine without webkit2gtk-4.0**

```bash
chmod +x Airvoice-Desktop-*-Linux-x86_64.AppImage
./Airvoice-Desktop-*-Linux-x86_64.AppImage
```

Expected: window opens without `libwebkit2gtk-4.0` / `WebKitNetworkProcess` missing errors.

If helpers still resolve to system `/usr/...`, re-check that AppRun sets `WEBKIT_EXEC_PATH` to the bundled `webkit2gtk-4.0` / `4.1` directory inside the AppDir. CI also runs an extract-based check (`--appimage-extract`) for AppRun, `Airvoice`, and WebKit helpers after packaging.

---

## Self-review checklist

| Spec requirement | Task |
|------------------|------|
| AppImage via linuxdeploy + WebKit helpers | Task 2 |
| Asset name `Airvoice-Desktop-${VERSION}-Linux-x86_64.AppImage` | Tasks 2–4 |
| Stop `tar.gz` / `.deb` | Task 3 |
| `downloads.ts` + WWW sync gate | Task 4 |
| Docs note | Task 5 |
| Runs without system webkit2gtk-4.0 | Task 2 `WEBKIT_EXEC_PATH` + Task 6 smoke test + CI extract verify |
| No Flatpak / no Wails v3 / macOS+Windows unchanged | Global Constraints + Task 3 |
