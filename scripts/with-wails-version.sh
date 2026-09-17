#!/usr/bin/env bash
# Run a Wails command with productVersion taken from the repo-root VERSION file.
# Restores desktop/wails.json afterwards so the placeholder is not committed.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
load_app_version

cd "$ROOT/desktop"

backup=$(mktemp)
cp wails.json "$backup"
restore() { cp "$backup" wails.json; rm -f "$backup"; }
trap restore EXIT

python() {
  if command -v python3 >/dev/null 2>&1; then
    command python3 "$@"
  else
    command python "$@"
  fi
}

python - "$APP_VERSION" <<'PY'
import json, pathlib, sys

version = sys.argv[1]
path = pathlib.Path("wails.json")
data = json.loads(path.read_text())
data.setdefault("info", {})["productVersion"] = version
path.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n")
PY

"$@"
