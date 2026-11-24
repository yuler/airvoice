#!/usr/bin/env bash
# Manually create and push a release Git tag (vX.Y.Z).
# Normal releases are tagged automatically by GitHub Actions on push/merge to main;
# this script serves as a local fallback for manual tag-based releases.
#
# Usage:
#   ./scripts/bump.sh              # interactive (gum choose)
#   ./scripts/bump.sh patch        # non-interactive: bump patch
#   ./scripts/bump.sh minor        # non-interactive: bump minor
#   ./scripts/bump.sh major        # non-interactive: bump major
#   ./scripts/bump.sh 1.0.0        # non-interactive: set exact version

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

current=$(AIRVOICE_VERSION="" "$ROOT/scripts/resolve-version.sh")

# ── SemVer Helpers ──

semver_regex='^([0-9]+)\.([0-9]+)\.([0-9]+)(-([0-9A-Za-z.-]+))?$'

parse_semver() {
  local v="$1"
  if [[ ! "$v" =~ $semver_regex ]]; then
    gum_err "Invalid semver: $v (expected X.Y.Z or X.Y.Z-prerelease)"
    exit 1
  fi
  SEMVER_MAJOR="${BASH_REMATCH[1]}"
  SEMVER_MINOR="${BASH_REMATCH[2]}"
  SEMVER_PATCH="${BASH_REMATCH[3]}"
}

bump_semver() {
  local type="$1"
  parse_semver "$current"
  case "$type" in
    major) echo "$(( 10#$SEMVER_MAJOR + 1 )).0.0" ;;
    minor) echo "${SEMVER_MAJOR}.$(( 10#$SEMVER_MINOR + 1 )).0" ;;
    patch) echo "${SEMVER_MAJOR}.${SEMVER_MINOR}.$(( 10#$SEMVER_PATCH + 1 ))" ;;
  esac
}

# ── Choose version ──

if [[ $# -ge 1 ]]; then
  input="${1#v}"
  case "$input" in
    major|minor|patch)
      new_version=$(bump_semver "$input")
      ;;
    *)
      new_version="$input"
      ;;
  esac
else
  gum_header "Tag Release"
  gum style --foreground 212 "Current latest tag: v$current"
  choice=$(gum choose --header "Bump type" "patch" "minor" "major")
  new_version=$(bump_semver "$choice")
fi

# ── Validate ──

parse_semver "$new_version"
new_tag="v$new_version"

if git rev-parse --verify "refs/tags/$new_tag" >/dev/null 2>&1; then
  gum_err "Tag $new_tag already exists."
  exit 1
fi

gum style --margin "1 0" --foreground 212 --bold "Tagging HEAD: v$current → $new_tag"

# ── Confirm & Tag ──

if ! gum confirm "Create git tag $new_tag at HEAD?"; then
  gum_info "Aborted."
  exit 0
fi

git tag "$new_tag"
gum_info "Created tag $new_tag"

if gum confirm "Push tag $new_tag to origin (triggers Release workflow)?"; then
  git push origin "$new_tag"
  gum_info "Pushed $new_tag to origin."
else
  gum_warn "Skipped push. Push later with: git push origin $new_tag"
fi

echo ""
gum style --foreground 10 --bold "Release tag created: $new_tag"
