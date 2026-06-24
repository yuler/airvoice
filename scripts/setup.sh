#!/usr/bin/env bash
# Install mise-managed tools and check platform dependencies.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

gum_header "Airvoice Setup"

gum spin --spinner dot --title "Installing mise tools (go, gum, swift)…" -- mise install

if [[ "$(uname -s)" == "Darwin" && ! -d "$IOS_PROJECT" ]]; then
  gum spin --spinner dot --title "Generating Xcode project…" -- \
    xcodegen generate --spec ios/project.yml --project ios/
fi

mkdir -p bin

gum_info "Tool versions:"
mise current 2>/dev/null | while read -r line; do
  gum style --foreground 212 "  $line"
done

gum style --margin "1 0 0 0" --bold "Platform dependencies"

case "$(uname -s)" in
  Darwin)
    gum_info "macOS: grant Accessibility to your terminal (or bin/airvoice) for paste injection."
    gum_info "  System Settings → Privacy & Security → Accessibility"
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

gum spin --spinner dot --title "Building CLI…" -- go build -o bin/airvoice ./cli

gum style --margin "1 0 0 0" --foreground 10 "Setup complete."
gum_info "Next: mise run dev"
