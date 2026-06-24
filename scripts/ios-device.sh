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

IOS_SIGN_TEAM=""
if ! xcodebuild -project "$IOS_PROJECT" -scheme "$IOS_SCHEME" -showBuildSettings 2>/dev/null \
  | grep -qE 'DEVELOPMENT_TEAM = [A-Z0-9]{10}'; then
  gum_warn "No development team configured in the Xcode project."
  IOS_SIGN_TEAM=$(gum input --placeholder "Apple Team ID (10 chars, from developer.apple.com)" --width 40)
  if [[ -z "$IOS_SIGN_TEAM" ]]; then
    gum_err "Team ID required for device signing. Set DEVELOPMENT_TEAM in Xcode → Signing & Capabilities."
    exit 1
  fi
fi
export IOS_SIGN_TEAM

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
gum_info "On your Mac, run: mise run serve"
gum_info "Then scan the terminal QR code from the iOS app."
