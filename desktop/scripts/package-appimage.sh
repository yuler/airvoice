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
export WEBKIT_EXEC_PATH="$(find "$HERE/usr" -type d \( -name 'webkit2gtk-4.0' -o -name 'webkit2gtk-4.1' \) 2>/dev/null | head -1 || true)"
export APPDIR="$HERE"
exec "$HERE/usr/bin/Airvoice" "$@"
EOF
  chmod +x "$APPDIR/AppRun"
}

# Copy WebKit helpers / injected bundle, preserving /usr/... layout under AppDir.
found=()
while IFS= read -r -d '' path; do
  found+=("$path")
done < <(find /usr -type f \( \
  -name 'WebKitWebProcess' -o \
  -name 'WebKitNetworkProcess' -o \
  -name 'WebKitGPUProcess' -o \
  -name 'libwebkit2gtkinjectedbundle.so' \
\) -print0 2>/dev/null || true)

for want in WebKitWebProcess WebKitNetworkProcess libwebkit2gtkinjectedbundle.so; do
  ok=0
  for path in "${found[@]:-}"; do
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
curl -fsSL -o "$LINUXDEPLOY" \
  "https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-$ARCH.AppImage"
curl -fsSL -o "$APPIMAGETOOL" \
  "https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-$ARCH.AppImage"
chmod +x "$LINUXDEPLOY" "$APPIMAGETOOL"
curl -fsSL -o "$TOOLS/linuxdeploy-plugin-gtk.sh" \
  "https://raw.githubusercontent.com/linuxdeploy/linuxdeploy-plugin-gtk/master/linuxdeploy-plugin-gtk.sh"
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

# Same-length rewrite: hardcoded "/usr/..." helper paths become "././..." relative to AppDir.
# Requires AppRun to `cd` into AppDir before exec.
while IFS= read -r -d '' so; do
  sed -i 's|/usr|././|g' "$so"
done < <(find "$APPDIR" -type f \( -name 'libwebkit2gtk-*.so*' -o -name 'libjavascriptcoregtk-*.so*' \) -print0)

# linuxdeploy may overwrite AppRun; restore ours before packing.
write_apprun

FINAL="$OUTPUT_DIR/Airvoice-Desktop-${VERSION}-Linux-${ARCH}.AppImage"
APPIMAGE_EXTRACT_AND_RUN=1 "$APPIMAGETOOL" "$APPDIR" "$FINAL"
chmod +x "$FINAL"
echo "Wrote $FINAL"
ls -lh "$FINAL"
