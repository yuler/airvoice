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
| Build iOS (simulator) | `mise run ios:dev` |
| Build iOS (device) | `mise run ios:build` |
| Docs dev server | `mise run www:dev` |
| Build docs | `mise run www:build` |

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
