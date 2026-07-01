#!/usr/bin/env bash
# Show versions for all Airvoice platforms.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

# ── Canonical version from VERSION file ──
VERSION_FILE="$ROOT/VERSION"
if [[ ! -f "$VERSION_FILE" ]]; then
  gum_err "VERSION file not found at $VERSION_FILE"
  exit 1
fi
canonical=$(tr -d '[:space:]' < "$VERSION_FILE")

gum_header "Airvoice Versions"
gum style --margin "1 0 0 0" --foreground 212 --bold "Canonical version: $canonical"

# ── Read platform versions ──

declare -A versions
MISSING="N/A"

# CLI (Go)
cli_ver=$(sed -n 's/^const version = "\(.*\)"/\1/p' cli/main.go) || true
versions[CLI]="${cli_ver:-$MISSING}"

# iOS (Info.plist CFBundleShortVersionString)
ios_ver=$(awk '/CFBundleShortVersionString/{f=1;next} f&&/<string>/{gsub(/.*<string>|<\/string>.*/,"");print;exit}' ios/Airvoice/Info.plist)
versions[iOS]="${ios_ver:-$MISSING}"

# Android
android_ver=$(sed -n 's/.*versionName = "\(.*\)"/\1/p' android/app/build.gradle.kts)
versions[Android]="${android_ver:-$MISSING}"

# Desktop (wails.json)
desktop_ver=$(python3 -c "import json; print(json.load(open('desktop/wails.json'))['info']['productVersion'])")
versions[Desktop]="${desktop_ver:-$MISSING}"

# Desktop Frontend
desktop_fe_ver=$(python3 -c "import json; print(json.load(open('desktop/frontend/package.json'))['version'])")
versions[Desktop FE]="${desktop_fe_ver:-$MISSING}"

# WWW
www_ver=$(python3 -c "import json; print(json.load(open('www/package.json'))['version'])")
versions[WWW]="${www_ver:-$MISSING}"

# ── Print table ──

printf '\n'
printf '  \033[1;35m%-16s %-10s %s\033[0m\n' "Platform" "Version" "Status"
printf '  %-16s %-10s %s\n' "────────────────" "──────────" "──────"
for platform in CLI iOS Android Desktop "Desktop FE" WWW; do
  ver="${versions[$platform]}"
  if [[ "$ver" == "$MISSING" ]]; then
    printf '  \033[0;90m%-16s %-10s missing\033[0m\n' "$platform" "$ver"
  elif [[ "$ver" == "$canonical" ]]; then
    printf '  \033[1;37m%-16s %-10s \033[1;32mOK\033[0m\n' "$platform" "$ver"
  else
    printf '  \033[1;33m%-16s %-10s \033[1;31mmismatch\033[0m\n' "$platform" "$ver"
  fi
done
printf '\n'

# ── Git tag ──

if git describe --tags --abbrev=0 >/dev/null 2>&1; then
  latest_tag=$(git describe --tags --abbrev=0)
  if [[ "$latest_tag" == "v$canonical" ]]; then
    gum_info "Latest git tag: $latest_tag (matches VERSION)"
  else
    gum_warn "Latest git tag: $latest_tag (VERSION is $canonical)"
  fi
else
  gum_info "No git tags found."
fi
