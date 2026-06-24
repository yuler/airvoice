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

sign_team=""
if ! xcodebuild -project "$IOS_PROJECT" -scheme "$IOS_SCHEME" -showBuildSettings 2>/dev/null \
  | grep -qE 'DEVELOPMENT_TEAM = [A-Z0-9]{10}'; then
  gum_warn "No development team configured in the Xcode project."
  sign_team=$(gum input --placeholder "Apple Team ID (10 chars, from developer.apple.com)" --width 40)
  if [[ -z "$sign_team" ]]; then
    gum_err "Team ID required for device signing. Set DEVELOPMENT_TEAM in Xcode → Signing & Capabilities."
    exit 1
  fi
fi

mkdir -p "$IOS_DERIVED"
build_args=(
  -project "$IOS_PROJECT"
  -scheme "$IOS_SCHEME"
  -configuration Debug
  -destination "id=$udid"
  -derivedDataPath "$IOS_DERIVED"
  -allowProvisioningUpdates
  build
)
if [[ -n "$sign_team" ]]; then
  build_args+=(DEVELOPMENT_TEAM="$sign_team")
fi

gum spin --spinner dot --title "Building for device…" -- xcodebuild "${build_args[@]}"

app="$IOS_DERIVED/Build/Products/Debug-iphoneos/Airvoice.app"
if [[ ! -d "$app" ]]; then
  gum_err "Build succeeded but app bundle not found: $app"
  exit 1
fi

if command -v xcrun >/dev/null 2>&1 && xcrun devicectl help device install app >/dev/null 2>&1; then
  gum spin --spinner dot --title "Installing on device…" -- \
    xcrun devicectl device install app --device "$udid" "$app"
  gum style --foreground 10 "Installed on device. Open Airvoice on your iPhone."
elif command -v ios-deploy >/dev/null 2>&1; then
  gum spin --spinner dot --title "Installing via ios-deploy…" -- \
    ios-deploy --id "$udid" --bundle "$app" --justlaunch
  gum style --foreground 10 "Launched on device."
else
  gum_warn "Built $app but no installer found."
  gum_info "Install manually: open Xcode → Window → Devices and Simulators → drag the .app"
  gum_info "Or install ios-deploy: brew install ios-deploy"
  open -R "$app"
fi

gum style --margin "1 0 0 0" --bold "Desktop server"
gum_info "On your Mac, run: mise run serve"
gum_info "Then scan the terminal QR code from the iOS app."
