#!/usr/bin/env bash
# Generate ios/Airvoice.xcodeproj from project.yml using the resolved version (AIRVOICE_VERSION or latest v* Git tag).

set -euo pipefail

source "$(dirname "$0")/lib.sh"
generate_xcode_project
