#!/usr/bin/env bash
# Shared helpers for Airvoice scripts.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IOS_PROJECT="$ROOT/ios/Airvoice.xcodeproj"
IOS_SCHEME="Airvoice"
IOS_DERIVED="$ROOT/.build/ios"

export ROOT IOS_PROJECT IOS_SCHEME IOS_DERIVED

gum_header() {
  gum style --border double --padding "0 2" --border-foreground 212 "$1"
}

gum_info() {
  gum log --level info "$@"
}

gum_warn() {
  gum log --level warn "$@"
}

gum_err() {
  gum log --level error "$@"
}

require_command() {
  local cmd="$1"
  local hint="${2:-}"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    gum_err "$cmd not found."
    [[ -n "$hint" ]] && gum_err "$hint"
    exit 1
  fi
}

require_macos() {
  if [[ "$(uname -s)" != "Darwin" ]]; then
    gum_err "This script requires macOS (found $(uname -s))."
    gum_info "iOS device builds need Xcode on a Mac."
    exit 1
  fi
}

require_xcode() {
  require_macos
  if ! command -v xcodebuild >/dev/null 2>&1; then
    gum_err "xcodebuild not found. Install Xcode from the App Store."
    exit 1
  fi
  if [[ ! -d "$IOS_PROJECT" ]]; then
    gum_err "Xcode project not found: $IOS_PROJECT"
    gum_info "Create ios/Airvoice.xcodeproj or open the iOS sources in Xcode first."
    exit 1
  fi
}

list_physical_ios_devices() {
  # Output: one device per line as "Name (OS) (UDID)"
  xcrun xctrace list devices 2>/dev/null \
    | awk '
        /^== Devices ==$/ { in_devices=1; next }
        /^== / { in_devices=0 }
        in_devices && /\([0-9A-Fa-f-]{20,}\)/ && !/Simulator/ { print }
      '
}

parse_device_udid() {
  # Last parenthesized token is the device UDID.
  sed -E 's/.*\(([0-9A-Fa-f-]+)\)[[:space:]]*$/\1/'
}

build_cli() {
  gum spin --spinner dot --title "Building Go CLI…" -- \
    go build -o "$ROOT/bin/airvoice" "$ROOT/cli"
  gum_info "Built $ROOT/bin/airvoice"
}
