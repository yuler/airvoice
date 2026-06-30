#!/usr/bin/env bash
# Install mise-managed tools and all platform dependencies.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

gum_header "Airvoice Setup"

# ── mise tools ──
gum spin --spinner dot --title "Installing mise tools (go, node, gum, swift)…" -- mise install

# ── CLI (Go) ──
mkdir -p bin
gum spin --spinner dot --title "Building CLI…" -- go build -o bin/airvoice ./cli

# ── WWW (Astro) ──
if [[ -d "$ROOT/www" ]]; then
  gum spin --spinner dot --title "Installing www dependencies…" -- bash -c "cd www && npm install"
fi

# ── Desktop (Wails) ──
if [[ -d "$ROOT/desktop" ]]; then
  if ! command -v wails >/dev/null 2>&1; then
    gum spin --spinner dot --title "Installing Wails CLI…" -- go install github.com/wailsapp/wails/v2/cmd/wails@latest
  else
    gum_info "Wails CLI already installed: $(wails version 2>/dev/null || echo 'unknown')"
  fi
  gum spin --spinner dot --title "Installing desktop frontend dependencies…" -- bash -c "cd desktop/frontend && npm install"

  # Linux: webkit2gtk-4.0 shim for Ubuntu 24.04+ (ships 4.1 only)
  if [[ "$(uname -s)" == "Linux" ]] && ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
    if pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
      _pc_dir="$ROOT/.mise/pkgconfig"
      mkdir -p "$_pc_dir"
      if [[ ! -f "$_pc_dir/webkit2gtk-4.0.pc" ]]; then
        cat > "$_pc_dir/webkit2gtk-4.0.pc" << 'PCEOF'
prefix=/usr
exec_prefix=${prefix}
libdir=/usr/lib/x86_64-linux-gnu
includedir=${prefix}/include

Name: WebKitGTK
Description: Web content engine for GTK (compat shim for 4.1)
URL: https://webkitgtk.org
Version: 2.52.3
Requires: glib-2.0 gtk+-3.0 libsoup-3.0 javascriptcoregtk-4.1
Libs: -L${libdir} -lwebkit2gtk-4.1
Cflags: -I${includedir}/webkitgtk-4.1
PCEOF
        gum_info "Created webkit2gtk-4.0 compat shim (mapped to 4.1)"
      fi
    else
      gum_warn "webkit2gtk not found — desktop builds will fail. Install: sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev"
    fi
  fi
fi

# ── iOS (macOS only) ──
if [[ "$(uname -s)" == "Darwin" ]]; then
  if [[ ! -d "$IOS_PROJECT" ]]; then
    gum spin --spinner dot --title "Generating Xcode project…" -- \
      xcodegen generate --spec ios/project.yml --project ios/
  fi

  if [[ ! -f ios/Signing.xcconfig ]]; then
    cp ios/Signing.xcconfig.example ios/Signing.xcconfig
    gum_warn "Created ios/Signing.xcconfig — set DEVELOPMENT_TEAM for iPhone device installs."
  fi
fi

# ── Android ──
if [[ -d "$ROOT/android" ]]; then
  gum_info "Checking Android SDK…"
  _java_ok=false
  if [[ -n "${JAVA_HOME:-}" ]] && [[ -x "$JAVA_HOME/bin/java" ]]; then
    _java_ok=true
  elif command -v java >/dev/null 2>&1; then
    _java_ok=true
  fi
  if [[ "$_java_ok" == "true" ]]; then
    gum_info "  ✓ JDK found"
    if [[ -n "${ANDROID_HOME:-}" ]] || [[ -d "$HOME/Android/Sdk" ]]; then
      gum_info "  ✓ Android SDK found"
    else
      gum_warn "  Android SDK not found — Android builds may fail"
    fi
  else
    gum_warn "  JDK 17 not found — Android builds will fail. Install: brew install openjdk@17 (macOS) or sudo apt install openjdk-17-jdk (Linux)"
  fi
fi

# ── Tool versions ──
gum_info "Tool versions:"
mise current 2>/dev/null | while read -r line; do
  gum style --foreground 212 "  $line"
done

# ── Platform notes ──
gum style --margin "1 0 0 0" --bold "Platform dependencies"

case "$(uname -s)" in
  Darwin)
    gum_info "macOS: grant Accessibility for paste injection (run ./bin/airvoice doctor to verify)."
    gum_info "  System Settings → Privacy & Security → Accessibility"
    gum_info "  Enable your host app (Cursor/Terminal/iTerm), bin/airvoice, and osascript — then fully quit and reopen the app."
  ;;
  Linux)
  if [[ "${XDG_SESSION_TYPE:-}" == "wayland" || -n "${WAYLAND_DISPLAY:-}" ]]; then
    for pkg in wl-copy ydotool; do
      if command -v "$pkg" >/dev/null 2>&1; then
        gum style --foreground 10 "  ✓ $pkg"
      else
        gum style --foreground 9 "  ✗ $pkg — install: sudo apt install wl-clipboard ydotool"
      fi
    done
  else
    for pkg in xclip xdotool; do
      if command -v "$pkg" >/dev/null 2>&1; then
        gum style --foreground 10 "  ✓ $pkg"
      else
        gum style --foreground 9 "  ✗ $pkg — install: sudo apt install xclip xdotool"
      fi
    done
  fi
  ;;
esac

gum style --margin "1 0 0 0" --foreground 10 "Setup complete."
gum_info "Next: mise run cli:dev | mise run www:dev | mise run desktop:dev"
