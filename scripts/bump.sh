#!/usr/bin/env bash
# Bump Airvoice version. The repo-root VERSION file is the only source of truth.
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

VERSION_FILE="$ROOT/VERSION"
if [[ ! -f "$VERSION_FILE" ]]; then
  gum_err "VERSION file not found. Create it with: echo 0.1.0 > VERSION"
  exit 1
fi

current=$(tr -d '[:space:]' < "$VERSION_FILE")

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
    major) echo "$(( SEMVER_MAJOR + 1 )).0.0" ;;
    minor) echo "${SEMVER_MAJOR}.$(( SEMVER_MINOR + 1 )).0" ;;
    patch) echo "${SEMVER_MAJOR}.${SEMVER_MINOR}.$(( SEMVER_PATCH + 1 ))" ;;
  esac
}

# ── Choose version ──

if [[ $# -ge 1 ]]; then
  input="$1"
  case "$input" in
    major|minor|patch)
      new_version=$(bump_semver "$input")
      ;;
    *)
      new_version="$input"
      ;;
  esac
else
  gum_header "Bump Version"
  gum style --foreground 212 "Current: $current"
  choice=$(gum choose --header "Bump type" "patch" "minor" "major")
  new_version=$(bump_semver "$choice")
fi

# ── Validate ──

parse_semver "$new_version"

if [[ "$new_version" == "$current" ]]; then
  gum_warn "Version unchanged ($current). Nothing to do."
  exit 0
fi

gum style --margin "1 0" --foreground 212 --bold "Bumping $current → $new_version"

# ── Confirm ──

if ! gum confirm "Apply version $new_version?"; then
  gum_info "Aborted."
  exit 0
fi

# ── Update VERSION file ──

echo "$new_version" > "$VERSION_FILE"
gum_info "Updated VERSION ($new_version)"

# ── Git tag ──

echo ""
if gum confirm "Create git commit and tag v$new_version?"; then
  # Commit only VERSION — leave any unrelated staged changes staged.
  git commit -m "🚀 [release] Bump version to $new_version" -- VERSION
  git tag "v$new_version"
  gum_info "Created commit & tag v$new_version"
  if gum confirm "Push commit and tag to origin?"; then
    git push && git push origin "v$new_version"
    gum_info "Pushed to origin."
  fi
else
  gum_warn "Skipped git tag. Changes are unstaged — review with: git diff"
fi

echo ""
gum style --foreground 10 --bold "Version bumped successfully: $current → $new_version"
