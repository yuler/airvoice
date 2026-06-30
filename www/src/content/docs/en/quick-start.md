---
title: Quick Start
description: Install Airvoice and get it running.
order: 2
---

Airvoice has two parts: a **desktop server** (CLI) and an **iOS client**. Both must be on the same Wi‑Fi.

## 1. Install the CLI

Install `mise` if you haven't:

```bash
curl https://mise.run | sh
```

Then install Airvoice:

```bash
mise trust
mise install
```

Or with Go directly:

```bash
go install github.com/yuler/airvoice/cli@latest
```

## 2. Start the Server

```bash
airvoice serve
```

A QR code appears in your terminal. Keep it running.

## 3. Install the iOS App

We don't have an Apple Developer Account, so you build from source with your Mac. It's free — just a regular Apple ID.

**What you need:**

- Mac with **Xcode 15+** (free from Mac App Store)
- Apple ID signed into Xcode
- iPhone with **iOS 17+**
- USB cable

### Step 1 — Enable Developer Mode on iPhone

On your iPhone: **Settings → Privacy & Security → Developer Mode** → turn it ON → restart when prompted.

> Don't see it? Connect your iPhone to your Mac with Xcode open first.

### Step 2 — Get the Source Code

```bash
git clone https://github.com/yuler/airvoice.git
cd airvoice
```

### Step 3 — Open in Xcode

```bash
open ios/Airvoice.xcodeproj
```

### Step 4 — Configure Signing

In Xcode:

1. Select the **Airvoice** project → **Airvoice** target
2. Under **Signing & Capabilities**: check **Automatically manage signing**
3. Set **Team** to your Apple ID (sign in if prompted)

### Step 5 — Build and Install

1. Connect iPhone via USB, unlock it, tap **Trust This Computer**
2. Select your iPhone from the device dropdown in Xcode
3. Press `Cmd + R` to build and install

### Step 6 — Trust the Developer Certificate

First time only on iPhone: **Settings → General → VPN & Device Management** → find your Apple ID → tap **Trust**.

### Step 7 — Connect

1. Make sure iPhone and Mac are on the same Wi‑Fi
2. Open Airvoice on iPhone
3. Scan the QR code from your terminal

Done. Speak on your phone — text appears at your cursor on the desktop.
