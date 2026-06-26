---
title: Quick Start
description: Get Airvoice up and running in minutes.
order: 1
---

# Quick Start

## 1. Setup

```bash
mise trust            # first time in this repo
mise install          # install pinned tools
mise run setup        # build CLI + check platform deps
```

## 2. Run the Server

```bash
mise run dev
```

This builds `bin/airvoice` and starts the server, printing a QR code in your terminal.

## 3. Connect the iOS Client

**Simulator or Xcode:** open `ios/Airvoice.xcodeproj` and run on a device/simulator.

**Physical device (macOS):**

```bash
mise run ios:device   # gum picker → build → install on USB-connected iPhone
```

Then on the same Wi‑Fi:

1. Run `mise run dev` on your Mac.
2. Open Airvoice on the iPhone and scan the terminal QR code.

## mise Tasks

| Task | Description |
|------|-------------|
| `mise run setup` | Install tools, check deps, build CLI |
| `mise run dev` | Build + start server (default) |
| `mise run menu` | Interactive gum menu |
| `mise run build` | Build Go CLI |
| `mise run test` | `go test ./cli/...` |
| `mise run serve` | Alias for `dev` |
| `mise run ios:device` | Build & install on physical iOS device (macOS) |
