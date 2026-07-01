#!/usr/bin/env bash
# Bump Airvoice version across all platforms.
#
# Usage:
#   ./scripts/bump.sh              # interactive (gum choose)
#   ./scripts/bump.sh patch        # non-interactive: bump patch
#   ./scripts/bump.sh minor        # non-interactive: bump minor
#   ./scripts/bump.sh major        # non-interactive: bump major
#   ./scripts/bump.sh 2.0.0        # non-interactive: set exact version

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

VERSION_FILE="$ROOT/VERSION"
if [[ ! -f "$VERSION_FILE" ]]; then
  gum_err "VERSION file not found. Create it with: echo 0.1.0 > VERSION"
  exit 1
fi

current=$(tr -d '[:space:]' < "$VERSION_FILE")

# ── Helpers ──

semver_regex='^([0-9]+)\.([0-9]+)\.([0-9]+)$'

parse_semver() {
  local v="$1"
  if [[ ! "$v" =~ $semver_regex ]]; then
    gum_err "Invalid semver: $v (expected X.Y.Z)"
    exit 1
  fi
  SEMVER_MAJOR="${BASH_REMATCH[1]}"
  SEMVER_MINOR="${BASH_REMATCH[2]}"
  SEMVER_PATCH="${BASH_REMATCH[3]}"
}

calc_version_code() {
  # Android versionCode: major*10000 + minor*100 + patch, minimum 1
  local code=$(( SEMVER_MAJOR * 10000 + SEMVER_MINOR * 100 + SEMVER_PATCH ))
  [[ "$code" -lt 1 ]] && code=1
  echo "$code"
}

bump_semver() {
  local type="$1"
  parse_semver "$current"
  case "$type" in
    major) echo "$(( SEMVER_MAJOR + 1 )).0.0" ;;
    minor) echo "${SEMVER_MAJOR}.$(( SEMVER_MINOR + 1 )).0" ;;
    patch) echo "${SEMVER_MAJOR}.${SEMVER_MINOR}.$(( SEMVER_PATCH + 1 ))" ;;
    *)     echo "$type" ;;
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
version_code=$(calc_version_code)

if [[ "$new_version" == "$current" ]]; then
  gum_warn "Version unchanged ($current). Nothing to do."
  exit 0
fi

gum style --margin "1 0" --foreground 212 --bold "Bumping $current → $new_version"

# ── Confirm ──

if ! gum confirm "Apply version $new_version to all platforms?"; then
  gum_info "Aborted."
  exit 0
fi

# ── Update all files ──

# 1. VERSION file
echo "$new_version" > "$VERSION_FILE"
gum_info "VERSION"

# 2. CLI (Go)
sed -i "s/const version = \"$current\"/const version = \"$new_version\"/" cli/main.go
gum_info "cli/main.go"

# 3. Android
sed -i "s/versionCode = [0-9]*/versionCode = $version_code/" android/app/build.gradle.kts
sed -i "s/versionName = \"$current\"/versionName = \"$new_version\"/" android/app/build.gradle.kts
gum_info "android/app/build.gradle.kts  (versionCode=$version_code)"

# 4. iOS Info.plist
awk -v old="$current" -v new="$new_version" '
  /CFBundleShortVersionString/ { found=1 }
  found && /<string>/ {
    sub(/<string>[^<]*<\/string>/, "<string>" new "</string>")
    found=0
  }
  { print }
' ios/Airvoice/Info.plist > /tmp/airvoice_info.plist && mv /tmp/airvoice_info.plist ios/Airvoice/Info.plist
gum_info "ios/Airvoice/Info.plist  (CFBundleShortVersionString)"

# 5. iOS Info.plist CFBundleVersion (auto-increment build number)
old_build=$(awk '/CFBundleVersion/{found=1; next} found && /<string>/{gsub(/.*<string>|<\/string>.*/,""); print; exit}' ios/Airvoice/Info.plist)
new_build=$(( ${old_build:-0} + 1 ))
awk -v old="$old_build" -v new="$new_build" '
  /CFBundleVersion/ { found=1 }
  found && /<string>/ {
    sub(/<string>[^<]*<\/string>/, "<string>" new "</string>")
    found=0
  }
  { print }
' ios/Airvoice/Info.plist > /tmp/airvoice_info.plist && mv /tmp/airvoice_info.plist ios/Airvoice/Info.plist
gum_info "ios/Airvoice/Info.plist  (CFBundleVersion build $new_build)"

# 6. iOS project.yml
sed -i "s/MARKETING_VERSION: \"$current\"/MARKETING_VERSION: \"$new_version\"/" ios/project.yml
gum_info "ios/project.yml"

# 7. Desktop wails.json
python3 -c "
import json
with open('desktop/wails.json') as f:
    data = json.load(f)
data['info']['productVersion'] = '$new_version'
with open('desktop/wails.json', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
gum_info "desktop/wails.json"

# 8. Desktop frontend package.json
python3 -c "
import json
with open('desktop/frontend/package.json') as f:
    data = json.load(f)
data['version'] = '$new_version'
with open('desktop/frontend/package.json', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
gum_info "desktop/frontend/package.json"

# 9. WWW package.json
python3 -c "
import json
with open('www/package.json') as f:
    data = json.load(f)
data['version'] = '$new_version'
with open('www/package.json', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
gum_info "www/package.json"

# 10. Homebrew formula
perl -pi -e "s/version \"$current\"/version \"$new_version\"/" Formula/airvoice.rb
perl -pi -e 's/sha256 "[^"]+"/sha256 "PLACEHOLDER"/' Formula/airvoice.rb
gum_info "Formula/airvoice.rb"

# ── Git tag ──

echo ""
if gum confirm "Create git tag v$new_version?"; then
  git add -A
  git commit -m "$(cat <<EOF
🔖 [release]: Bump version to $new_version
EOF
)"
  git tag "v$new_version"
  gum_info "Tagged v$new_version"
  if gum confirm "Push commit and tag to origin?"; then
    git push && git push --tags
    gum_info "Pushed to origin."
  fi
else
  gum_warn "Skipped git tag. Changes are unstaged — review with: git diff"
fi

echo ""
gum style --foreground 10 --bold "Version bumped: $current → $new_version"
