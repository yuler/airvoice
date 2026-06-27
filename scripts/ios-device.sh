#!/usr/bin/env bash
# Build Airvoice iOS app and install on a connected physical device.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

require_xcode

gum_header "iOS Device Debug"

devices=$(list_physical_ios_devices || true)
if [[ -z "$devices" ]]; then
  gum_err "No physical iOS devices found."
  gum_info "Connect your iPhone via USB, unlock it, and tap Trust This Computer."
  exit 1
fi

selected=$(echo "$devices" | gum choose --header "Select device:")
udid=$(echo "$selected" | parse_device_udid)

gum_info "Device: $selected"
gum_info "UDID: $udid"

if [[ ! -f ios/Signing.xcconfig ]]; then
  cp ios/Signing.xcconfig.example ios/Signing.xcconfig
  gum_warn "Created ios/Signing.xcconfig — edit DEVELOPMENT_TEAM before building."
fi

team=$(grep -E '^DEVELOPMENT_TEAM\s*=' ios/Signing.xcconfig 2>/dev/null | sed -E 's/^DEVELOPMENT_TEAM[[:space:]]*=[[:space:]]*//' | tr -d '[:space:]' || true)
if [[ -z "$team" ]]; then
  gum_warn "DEVELOPMENT_TEAM is empty in ios/Signing.xcconfig."
  team=$(gum input --placeholder "Apple Team ID (10 chars)" --width 40)
  if [[ -z "$team" ]]; then
    gum_err "Team ID required for device install. Set DEVELOPMENT_TEAM in ios/Signing.xcconfig."
    exit 1
  fi
  export IOS_SIGN_TEAM="$team"
else
  export IOS_SIGN_TEAM="$team"
fi

mkdir -p "$IOS_DERIVED"

if ! build_ios_app_for_device "$udid"; then
  hint_ios_build_failure "$IOS_DERIVED/xcodebuild.log"
  exit 1
fi

app="$IOS_DERIVED/Build/Products/Debug-iphoneos/Airvoice.app"
if [[ ! -d "$app" ]]; then
  gum_err "Build succeeded but app bundle not found: $app"
  exit 1
fi

if ! install_ios_app_on_device "$udid" "$app"; then
  exit 1
fi

gum style --margin "1 0 0 0" --bold "Desktop server"
gum_info "On your Mac, run: mise run dev"
gum_info "Then scan the terminal QR code from the iOS app."
