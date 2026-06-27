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
  # Prefer xcodebuild destinations — matches what the build step uses.
  if [[ -d "$IOS_PROJECT" ]] && command -v xcodebuild >/dev/null 2>&1; then
    local from_xcodebuild
    from_xcodebuild=$(
      xcodebuild -showdestinations -project "$IOS_PROJECT" -scheme "$IOS_SCHEME" 2>/dev/null \
        | grep '{ platform:iOS, arch:arm64, id:' \
        | sed -E 's/.*id:([^,]+), name:([^}]+) \}.*/\2 (\1)/' \
        | sed '/^$/d' || true
    )
    if [[ -n "$from_xcodebuild" ]]; then
      echo "$from_xcodebuild"
      return 0
    fi
  fi

  # devicectl JSON (Xcode 15+); xctrace often crashes on newer Xcode builds.
  if command -v xcrun >/dev/null 2>&1 && xcrun devicectl help list devices >/dev/null 2>&1; then
    local tmp json_devices
    tmp=$(mktemp)
    if xcrun devicectl list devices --json-output "$tmp" >/dev/null 2>&1; then
      json_devices=$(
        python3 - "$tmp" <<'PY'
import json, sys

with open(sys.argv[1], encoding="utf-8") as f:
    data = json.load(f)

for device in data.get("result", {}).get("devices", []):
    hw = device.get("hardwareProperties", {})
    if hw.get("platform") != "iOS":
        continue
    props = device.get("deviceProperties", {})
    conn = device.get("connectionProperties", {})
    if conn.get("pairingState") not in (None, "paired"):
        continue
    name = props.get("name", "iPhone")
    osv = props.get("osVersionNumber", "")
    udid = hw.get("udid") or device.get("identifier", "")
    if not udid:
        continue
    if osv:
        print(f"{name} ({osv}) ({udid})")
    else:
        print(f"{name} ({udid})")
PY
      )
      rm -f "$tmp"
      if [[ -n "$json_devices" ]]; then
        echo "$json_devices"
        return 0
      fi
    else
      rm -f "$tmp"
    fi
  fi

  # Legacy fallback.
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

# Look up CoreDevice metadata for a hardware UDID (identifier, tunnel, OS version).
lookup_device_metadata() {
  local udid="$1"
  if ! command -v xcrun >/dev/null 2>&1 || ! xcrun devicectl help list devices >/dev/null 2>&1; then
    return 1
  fi
  local tmp
  tmp=$(mktemp)
  if ! xcrun devicectl list devices --json-output "$tmp" >/dev/null 2>&1; then
    rm -f "$tmp"
    return 1
  fi
  python3 - "$udid" "$tmp" <<'PY'
import json, sys

udid, path = sys.argv[1], sys.argv[2]
with open(path, encoding="utf-8") as f:
    data = json.load(f)

for device in data.get("result", {}).get("devices", []):
    hw = device.get("hardwareProperties", {})
    if hw.get("platform") != "iOS":
        continue
    hardware_udid = hw.get("udid", "")
    core_id = device.get("identifier", "")
    if udid not in (hardware_udid, core_id):
        continue
    props = device.get("deviceProperties", {})
    conn = device.get("connectionProperties", {})
    print(json.dumps({
        "udid": hardware_udid or udid,
        "identifier": core_id,
        "name": props.get("name", "iPhone"),
        "osVersion": props.get("osVersionNumber", ""),
        "pairingState": conn.get("pairingState", ""),
        "tunnelState": conn.get("tunnelState", ""),
        "transportType": conn.get("transportType", ""),
    }))
    break
PY
  rm -f "$tmp"
}

hint_ios_build_failure() {
  local log="$1"
  gum_err "Build failed — app was not installed."
  if [[ -f "$log" ]] && grep -q 'PLA Update available' "$log"; then
    gum_info "Apple 开发者协议未接受：请团队 Account Holder（Wei Shufang）登录 developer.apple.com 同意最新协议。"
  elif [[ -f "$log" ]] && grep -q 'No Accounts' "$log"; then
    gum_info "请在 Xcode → Settings → Accounts 登录 Apple ID。"
  elif [[ -f "$log" ]] && grep -qE 'No profiles for|No code signature|requires a development team|Signing for' "$log"; then
    gum_info "真机安装需要签名：在 ios/Signing.xcconfig 填写 DEVELOPMENT_TEAM（参考 Signing.xcconfig.example），然后重新 xcodegen generate。"
  elif [[ -f "$log" ]] && grep -qE 'not installed|Unable to find a destination' "$log"; then
    gum_info "Xcode 缺少 iOS 平台支持：xcodebuild -downloadPlatform iOS"
    gum_info "若 iPhone 系统比 Xcode 新，请升级 Xcode。"
  fi
  gum_info "Build log: $log"
}

run_logged() {
  local log="$1"
  shift
  "$@" 2>&1 | tee -a "$log"
  return "${PIPESTATUS[0]}"
}

ensure_ios_platform() {
  gum_warn "iOS platform support missing in Xcode — downloading (may take a few minutes)…"
  if gum spin --spinner dot --title "Downloading iOS platform…" -- xcodebuild -downloadPlatform iOS; then
    return 0
  fi
  gum_warn "Platform download failed. Try manually: xcodebuild -downloadAllPlatforms"
  return 1
}

run_xcodebuild_ios() {
  local udid="$1"
  local log="$2"
  shift 2
  local -a args=("$@")
  local -a destinations=()
  local dest meta name

  if [[ -n "$udid" ]]; then
    destinations+=("id=$udid" "platform=iOS,id=$udid")
    if meta=$(lookup_device_metadata "$udid" 2>/dev/null); then
      name=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("name",""))' <<<"$meta")
      if [[ -n "$name" ]]; then
        destinations+=("platform=iOS,name=$name")
      fi
    fi
    for dest in "${destinations[@]}"; do
      if run_logged "$log" xcodebuild "${args[@]}" -sdk iphoneos -destination "$dest" build; then
        return 0
      fi
      if ! grep -qE 'Unable to find a destination|not installed|Timed out waiting' "$log"; then
        return 1
      fi
    done
  fi

  run_logged "$log" xcodebuild "${args[@]}" -sdk iphoneos build
}

needs_ios_platform_retry() {
  local log="$1"
  [[ -f "$log" ]] && grep -qE 'not installed|Unable to find a destination|Timed out waiting' "$log"
}

hint_ios_deploy_failure() {
  gum_err "Install failed."
  gum_info "Checklist:"
  gum_info "  • iPhone unlocked, USB connected, Trust This Computer accepted"
  gum_info "  • Settings → Privacy & Security → Developer Mode enabled (reboot if prompted)"
  gum_info "  • ios/Signing.xcconfig: DEVELOPMENT_TEAM set (see Signing.xcconfig.example)"
  gum_info "  • Xcode version supports your iPhone iOS (try: xcodebuild -downloadAllPlatforms)"
  gum_info "  • Or open ios/Airvoice.xcodeproj in Xcode and Run on your device once"
}

install_ios_app_on_device() {
  local udid="$1"
  local app="$2"
  local log="$IOS_DERIVED/install.log"
  local meta core_id tunnel bundle_id="com.airvoice.app"
  local -a device_refs

  device_refs=("$udid")
  if meta=$(lookup_device_metadata "$udid" 2>/dev/null); then
    core_id=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("identifier",""))' <<<"$meta")
    [[ -n "$core_id" && "$core_id" != "$udid" ]] && device_refs=("$core_id" "$udid")
    tunnel=$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("tunnelState",""))' <<<"$meta")
    if [[ "$tunnel" == "disconnected" ]]; then
      gum_warn "Device tunnel is disconnected — unlock iPhone and keep the USB cable connected."
    fi
  fi

  mkdir -p "$IOS_DERIVED"
  : >"$log"

  if command -v xcrun >/dev/null 2>&1 && xcrun devicectl help device install app >/dev/null 2>&1; then
    local ref
    for ref in "${device_refs[@]}"; do
      gum_info "Installing via devicectl (device ${ref})..."
      if run_logged "$log" xcrun devicectl device install app --device "$ref" "$app"; then
        gum_info "Launching Airvoice on device..."
        if run_logged "$log" xcrun devicectl device process launch --device "$ref" "$bundle_id"; then
          gum style --foreground 10 "Installed and launched on device."
        else
          gum_style_installed_manual_launch
        fi
        return 0
      fi
    done
  fi

  if command -v ios-deploy >/dev/null 2>&1; then
    gum_info "Retrying install via ios-deploy..."
    if run_logged "$log" ios-deploy --id "$udid" --bundle "$app" --justlaunch; then
      gum style --foreground 10 "Installed and launched on device."
      return 0
    fi
  fi

  hint_ios_deploy_failure
  gum_info "Install log: $log"
  open -R "$app" 2>/dev/null || true
  return 1
}

gum_style_installed_manual_launch() {
  gum style --foreground 10 "Installed on device."
  gum_info "Open the Airvoice app on your iPhone (home screen or App Library)."
}

build_ios_app_for_device() {
  local udid="$1"
  local log="$IOS_DERIVED/xcodebuild.log"
  local -a base_args sign_args

  if [[ -z "${IOS_SIGN_TEAM:-}" ]] && [[ -f "$ROOT/ios/Signing.xcconfig" ]]; then
    IOS_SIGN_TEAM=$(grep -E '^DEVELOPMENT_TEAM\s*=' "$ROOT/ios/Signing.xcconfig" | head -1 | sed -E 's/^DEVELOPMENT_TEAM[[:space:]]*=[[:space:]]*//' | tr -d '[:space:]')
  fi

  mkdir -p "$IOS_DERIVED"
  base_args=(
    -project "$IOS_PROJECT"
    -scheme "$IOS_SCHEME"
    -configuration Debug
    -derivedDataPath "$IOS_DERIVED"
    -allowProvisioningUpdates
  )
  sign_args=()
  if [[ -n "${IOS_SIGN_TEAM:-}" ]]; then
    sign_args+=(DEVELOPMENT_TEAM="$IOS_SIGN_TEAM")
  fi

  : >"$log"

  # Use -sdk iphoneos; -showdestinations often omits USB devices even when builds work.
  gum_info "Building for device ${udid}..."
  if run_xcodebuild_ios "$udid" "$log" "${base_args[@]}" "${sign_args[@]}"; then
    return 0
  fi

  if needs_ios_platform_retry "$log"; then
    if ensure_ios_platform && run_xcodebuild_ios "$udid" "$log" "${base_args[@]}" "${sign_args[@]}"; then
      return 0
    fi
    gum_warn "Device destination unavailable; building with iphoneos SDK only…"
    if run_xcodebuild_ios "" "$log" "${base_args[@]}" "${sign_args[@]}"; then
      return 0
    fi
  fi

  return 1
}

build_cli() {
  gum spin --spinner dot --title "Building Go CLI…" -- \
    go build -o "$ROOT/bin/airvoice" "$ROOT/cli"
  gum_info "Built $ROOT/bin/airvoice"
}
