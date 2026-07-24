#!/usr/bin/env bash
# Bundle the Wails Linux binary into a relocatable AppImage with WebKitGTK.
set -euo pipefail

usage() {
  echo "Usage: $0 --binary PATH --version VERSION --output-dir DIR" >&2
  exit 1
}

BINARY=""
VERSION=""
OUTPUT_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --binary) BINARY="${2:-}"; shift 2 ;;
    --version) VERSION="${2:-}"; shift 2 ;;
    --output-dir) OUTPUT_DIR="${2:-}"; shift 2 ;;
    *) usage ;;
  esac
done

[[ -n "$BINARY" && -n "$VERSION" && -n "$OUTPUT_DIR" ]] || usage
[[ -f "$BINARY" ]] || { echo "binary not found: $BINARY" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DESKTOP_DIR="$(cd "$SCRIPT_DIR/../packaging/linux" && pwd)"
DESKTOP_FILE="$DESKTOP_DIR/Airvoice.desktop"
ICON_FILE="$DESKTOP_DIR/airvoice.png"
[[ -f "$DESKTOP_FILE" ]] || { echo "missing $DESKTOP_FILE" >&2; exit 1; }
[[ -f "$ICON_FILE" ]] || { echo "missing $ICON_FILE" >&2; exit 1; }

BINARY="$(cd "$(dirname "$BINARY")" && pwd)/$(basename "$BINARY")"
OUTPUT_DIR="$(mkdir -p "$OUTPUT_DIR" && cd "$OUTPUT_DIR" && pwd)"

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH=x86_64 ;;
  aarch64|arm64) ARCH=aarch64 ;;
  *) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

# Pinned packaging tooling (versioned releases + sha256).
LINUXDEPLOY_VERSION="1-alpha-20251107-1"
APPIMAGETOOL_VERSION="1.9.0"
LINUXDEPLOY_PLUGIN_GTK_SHA="7a3fbc31a9e5075073ff8790f26effbac5f84453"
case "$ARCH" in
  x86_64)
    LINUXDEPLOY_SHA256="c20cd71e3a4e3b80c3483cef793cda3f4e990aca14014d23c544ca3ce1270b4d"
    APPIMAGETOOL_SHA256="46fdd785094c7f6e545b61afcfb0f3d98d8eab243f644b4b17698c01d06083d1"
    ;;
  aarch64)
    LINUXDEPLOY_SHA256="620095110d693282b8ebeb244a95b5e911cf8f65f76c88b4b47d16ae6346fcff"
    APPIMAGETOOL_SHA256="04f45ea45b5aa07bb2b071aed9dbf7a5185d3953b11b47358c1311f11ea94a96"
    ;;
esac
LINUXDEPLOY_PLUGIN_GTK_SHA256="b0f4cbc684a0103a9651f0955b635eaea0096b3a66c0f5a2c2aa337960375171"

download_verify() {
  local url="$1" dest="$2" sha="$3" label="$4"
  curl -fsSL -o "$dest" "$url" || { echo "failed to download $label" >&2; exit 1; }
  echo "$sha  $dest" | sha256sum -c - >/dev/null \
    || { echo "checksum mismatch for $label" >&2; exit 1; }
}

WORK="$(mktemp -d "${TMPDIR:-/tmp}/airvoice-appimage.XXXXXX")"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

APPDIR="$WORK/Airvoice-$ARCH.AppDir"
mkdir -p "$APPDIR/usr/bin"
cp "$BINARY" "$APPDIR/usr/bin/Airvoice"
chmod +x "$APPDIR/usr/bin/Airvoice"
cp "$DESKTOP_FILE" "$APPDIR/Airvoice.desktop"
cp "$ICON_FILE" "$APPDIR/airvoice.png"
cp "$ICON_FILE" "$APPDIR/.DirIcon"

write_apprun() {
  cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/bash
set -euo pipefail
HERE="$(dirname "$(readlink -f "$0")")"
cd "$HERE"
export PATH="$HERE/usr/bin:${PATH:-}"
export LD_LIBRARY_PATH="$HERE/usr/lib:$HERE/usr/lib/x86_64-linux-gnu:$HERE/usr/lib/aarch64-linux-gnu:${LD_LIBRARY_PATH:-}"
# Prefer webkit2gtk-4.0 (Wails v2 link target), then 4.1.
WEBKIT_EXEC_PATH=""
for candidate in \
  "$HERE/usr/lib/x86_64-linux-gnu/webkit2gtk-4.0" \
  "$HERE/usr/lib/aarch64-linux-gnu/webkit2gtk-4.0" \
  "$HERE/usr/lib/webkit2gtk-4.0" \
  "$HERE/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1" \
  "$HERE/usr/lib/aarch64-linux-gnu/webkit2gtk-4.1" \
  "$HERE/usr/lib/webkit2gtk-4.1"
do
  if [[ -d "$candidate" ]]; then
    WEBKIT_EXEC_PATH="$candidate"
    break
  fi
done
export WEBKIT_EXEC_PATH
export APPDIR="$HERE"
exec "$HERE/usr/bin/Airvoice" "$@"
EOF
  chmod +x "$APPDIR/AppRun"
}

# Copy WebKit helpers / injected bundle from known lib paths only.
WEBKIT_SEARCH_ROOTS=(
  /usr/lib/x86_64-linux-gnu
  /usr/lib/aarch64-linux-gnu
  /usr/lib
  /usr/lib64
)
found=()
for root in "${WEBKIT_SEARCH_ROOTS[@]}"; do
  [[ -d "$root" ]] || continue
  while IFS= read -r -d '' path; do
    found+=("$path")
  done < <(find "$root" -type f \( \
    -name 'WebKitWebProcess' -o \
    -name 'WebKitNetworkProcess' -o \
    -name 'WebKitGPUProcess' -o \
    -name 'libwebkit2gtkinjectedbundle.so' \
  \) -print0 2>/dev/null || true)
done

for want in WebKitWebProcess WebKitNetworkProcess libwebkit2gtkinjectedbundle.so; do
  ok=0
  for path in "${found[@]+"${found[@]}"}"; do
    [[ "$(basename "$path")" == "$want" ]] && ok=1 && break
  done
  [[ "$ok" -eq 1 ]] || { echo "required WebKit file not found: $want" >&2; exit 1; }
done

for path in "${found[@]}"; do
  rel="${path#/}"
  dest="$APPDIR/$rel"
  mkdir -p "$(dirname "$dest")"
  cp -a "$path" "$dest"
done

write_apprun

TOOLS="$WORK/tools"
mkdir -p "$TOOLS"
LINUXDEPLOY="$TOOLS/linuxdeploy-$ARCH.AppImage"
APPIMAGETOOL="$TOOLS/appimagetool-$ARCH.AppImage"
download_verify \
  "https://github.com/linuxdeploy/linuxdeploy/releases/download/${LINUXDEPLOY_VERSION}/linuxdeploy-$ARCH.AppImage" \
  "$LINUXDEPLOY" \
  "$LINUXDEPLOY_SHA256" \
  "linuxdeploy"
download_verify \
  "https://github.com/AppImage/appimagetool/releases/download/${APPIMAGETOOL_VERSION}/appimagetool-$ARCH.AppImage" \
  "$APPIMAGETOOL" \
  "$APPIMAGETOOL_SHA256" \
  "appimagetool"
chmod +x "$LINUXDEPLOY" "$APPIMAGETOOL"
download_verify \
  "https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/${LINUXDEPLOY_PLUGIN_GTK_SHA}/linuxdeploy-plugin-gtk.sh" \
  "$TOOLS/linuxdeploy-plugin-gtk.sh" \
  "$LINUXDEPLOY_PLUGIN_GTK_SHA256" \
  "linuxdeploy-plugin-gtk"
chmod +x "$TOOLS/linuxdeploy-plugin-gtk.sh"

export DEPLOY_GTK_VERSION=3
export NO_STRIP=1
export PATH="$TOOLS:$PATH"
export APPIMAGE_EXTRACT_AND_RUN=1

# Bundle shared library deps into AppDir only (no AppImage yet).
(
  cd "$WORK"
  "$LINUXDEPLOY" --appimage-extract-and-run \
    --appdir "$APPDIR" \
    --executable "$APPDIR/usr/bin/Airvoice" \
    --desktop-file "$APPDIR/Airvoice.desktop" \
    --icon-file "$APPDIR/airvoice.png" \
    --plugin gtk
)

# WebKit helper discovery: WEBKIT_EXEC_PATH (set in AppRun) directs WebKit to
# find WebKitWebProcess / WebKitNetworkProcess / WebKitGPUProcess inside the
# AppDir, so no binary patching of .so files is needed.

# linuxdeploy may overwrite AppRun; restore ours before packing.
write_apprun

FINAL="$OUTPUT_DIR/Airvoice-Desktop-${VERSION}-Linux-${ARCH}.AppImage"
APPIMAGE_EXTRACT_AND_RUN=1 "$APPIMAGETOOL" "$APPDIR" "$FINAL"
chmod +x "$FINAL"
echo "Wrote $FINAL"
ls -lh "$FINAL"
