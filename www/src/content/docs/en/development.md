---
title: Development
description: Set up the development environment and contribute.
order: 3
---

## Prerequisites

- [mise](https://mise.run) — tool version manager
- Git

`mise` will install the correct versions of Go, Node, and other tools automatically.

## Setup

```bash
git clone https://github.com/yuler/airvoice.git
cd airvoice
mise trust
mise install
```

## Project Structure

```
airvoice/
├── cli/          # Go CLI — WebSocket server + keystroke injection
├── desktop/      # Desktop app (Wails)
├── android/      # Android Kotlin/Compose app
├── ios/          # iOS SwiftUI app
├── www/          # Documentation site (Astro)
├── scripts/      # Build and dev scripts
├── mise.toml     # Task runner + tool versions
└── go.mod
```

## Common Tasks

| Task | Command |
|------|---------|
| Build CLI | `mise run cli:build` |
| Run server | `mise run cli:dev` |
| Run tests | `mise run cli:test` |
| Build Android (debug APK) | `mise run android:build` |
| Install Android (USB) | `mise run android:install` |
| Build iOS (simulator) | `mise run ios:dev` |
| Build iOS (device) | `mise run ios:build` |
| Docs dev server | `mise run www:dev` |
| Build docs | `mise run www:build` |
| Bump version | `mise bump` |

## Android Development

Requires JDK (Android Studio JBR works). Build a debug APK:

```bash
mise run android:build
```

The APK is at `android/app/build/outputs/apk/debug/app-debug.apk`. Release builds are published on [GitHub Releases](https://github.com/yuler/airvoice/releases/latest) as `airvoice-android-*.apk`.

## iOS Development

Open the project in Xcode:

```bash
open ios/Airvoice.xcodeproj
```

The project uses [XcodeGen](https://github.com/yonaskolb/XcodeGen) to generate `project.yml`. If you modify project settings, edit `ios/project.yml` and regenerate with `./scripts/ios-xcodegen.sh` (exports `APP_VERSION` / `APP_BUILD` from `AIRVOICE_VERSION` or the latest `v*` Git tag). Do not run `xcodegen` alone.

```bash
./scripts/ios-xcodegen.sh
```

### Signing

For device builds, create your signing config:

```bash
cp ios/Signing.xcconfig.example ios/Signing.xcconfig
```

Edit `ios/Signing.xcconfig` and set `DEVELOPMENT_TEAM` to your Apple Team ID (find it in Xcode → Settings → Accounts).

## Version & Releases

Git tags (`vX.Y.Z`) are the canonical source of truth for product versions. Apps and CI resolve the version at build time via `AIRVOICE_VERSION` or the latest `v*` tag — do not copy it into `package.json` or source files.

Merges/pushes to `main` automatically cut a new release tag after CI passes: `patch` by default, `[minor]` or `[major]` in a commit message / PR title to bump higher, or `[skip release]` in the `HEAD` commit subject to skip. You can also dispatch the `Release` workflow manually or run `mise bump` locally to create and push a tag.

iOS `CFBundleVersion` and Android `versionCode` are derived as `major * 1000000 + minor * 1000 + patch`.
