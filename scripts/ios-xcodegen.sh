#!/usr/bin/env bash
# Generate ios/Airvoice.xcodeproj from project.yml using VERSION.

set -euo pipefail

source "$(dirname "$0")/lib.sh"
generate_xcode_project
