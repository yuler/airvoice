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
├── VERSION       # Canonical product version (semver)
├── cli/          # Go CLI — WebSocket server + keystroke injection
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

The project uses [XcodeGen](https://github.com/yonaskolb/XcodeGen) to generate `project.yml`. If you modify project settings, edit `ios/project.yml` and regenerate:

```bash
xcodegen generate
```

### Signing

For device builds, create your signing config:

```bash
cp ios/Signing.xcconfig.example ios/Signing.xcconfig
```

Edit `ios/Signing.xcconfig` and set `DEVELOPMENT_TEAM` to your Apple Team ID (find it in Xcode → Settings → Accounts).

## Version

The product version lives only in the repo-root `VERSION` file. Apps and CI read it at build time — do not copy it into `package.json` or other source files. Release with `mise bump`.
