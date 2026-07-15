#!/usr/bin/env bash
# Sync VERSION file into platform build configs.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

VERSION_FILE="$ROOT/VERSION"
if [[ ! -f "$VERSION_FILE" ]]; then
  gum_err "VERSION file not found at $VERSION_FILE"
  exit 1
fi
canonical=$(tr -d '[:space:]' < "$VERSION_FILE")

# Helper for portable in-place sed
sedi() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    sed -i "" "$@"
  else
    sed -i "$@"
  fi
}

# iOS project.yml
sedi "s/MARKETING_VERSION: \".*\"/MARKETING_VERSION: \"$canonical\"/" "$ROOT/ios/project.yml"
gum_info "iOS project.yml → $canonical"

# Android build.gradle.kts is read at build time (see build.gradle.kts)
gum_info "Android build.gradle.kts reads VERSION at build time"

# CLI
sedi "s/const version = \".*\"/const version = \"$canonical\"/" "$ROOT/cli/main.go"
gum_info "CLI main.go → $canonical"

gum_ok "All platforms synced to $canonical"
