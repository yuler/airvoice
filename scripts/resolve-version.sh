#!/usr/bin/env bash
# Resolve the canonical Airvoice semver version.
# Priority:
#   1. AIRVOICE_VERSION environment variable (explicit CI/build override)
#   2. Highest vX.Y.Z Git tag in the repository
#   3. Fallback "0.0.0"

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SEMVER_REGEX='^([0-9]+)\.([0-9]+)\.([0-9]+)(-([0-9A-Za-z.-]+))?$'

clean_version() {
  local raw="$1"
  raw="${raw#v}"
  printf '%s' "$raw" | tr -d '[:space:]'
}

if [[ -n "${AIRVOICE_VERSION:-}" ]]; then
  candidate="$(clean_version "$AIRVOICE_VERSION")"
  if [[ ! "$candidate" =~ $SEMVER_REGEX ]]; then
    echo "Invalid AIRVOICE_VERSION: ${AIRVOICE_VERSION} (expected X.Y.Z or X.Y.Z-prerelease)" >&2
    exit 1
  fi
  printf '%s\n' "$candidate"
  exit 0
fi

if command -v git >/dev/null 2>&1 && git -C "$ROOT" rev-parse --git-dir >/dev/null 2>&1; then
  tag="$(git -C "$ROOT" tag --list 'v*' --sort=-v:refname 2>/dev/null | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n 1 || true)"
  if [[ -n "$tag" ]]; then
    printf '%s\n' "${tag#v}"
    exit 0
  fi
fi

printf '0.0.0\n'
