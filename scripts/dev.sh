#!/usr/bin/env bash
# Interactive development menu powered by gum.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
cd "$ROOT"

gum_header "Airvoice Dev"

choice=$(gum choose --header "What do you want to do?" \
  "Build CLI" \
  "Run tests" \
  "Start server (QR pairing)" \
  "iOS device debug" \
  "Run setup again" \
  "Quit")

case "$choice" in
  "Build CLI")
    build_cli
    ;;
  "Run tests")
    gum spin --spinner dot --title "Running Go tests…" -- go test ./cli/...
    gum style --foreground 10 "Tests passed."
    ;;
  "Start server (QR pairing)")
    build_cli
    gum_info "Starting airvoice serve — scan the QR code from your iPhone."
    gum style --foreground 212 "Press Ctrl+C to stop."
    ./bin/airvoice serve
    ;;
  "iOS device debug")
    exec "$ROOT/scripts/ios-device.sh"
    ;;
  "Run setup again")
    exec "$ROOT/scripts/setup.sh"
    ;;
  "Quit")
  ;;
esac
