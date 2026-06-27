---
title: Background
description: Why Airvoice exists, how it works, and when to use it.
order: 1
---

## Introduction

## Why This Exists

I work at a desk with a tower PC and a Mac mini — neither has a built-in microphone. Bluetooth headsets are unreliable, and I didn't want to buy a dedicated mic just for occasional voice input.

My phone already has an excellent microphone with great noise cancellation. And I already use voice keyboards like 豆包输入法 (Doubao IME) on my phone every day. The question was: **how do I get the text from my phone's voice keyboard into whatever app has focus on my desktop?**

That's all Airvoice does. It's a bridge.

## How It Works

```
iPhone (voice keyboard)  →  LAN WebSocket  →  Desktop (paste at cursor)
```

1. **Desktop:** A Go CLI starts a WebSocket server on your local network and prints a QR code.
2. **iPhone:** The iOS app connects by scanning the QR code. It provides a text field where you use any voice keyboard (豆包输入法, 微信输入法, etc.) to dictate.
3. **When you dismiss the keyboard** (or after 1.5 seconds of idle), the text is sent over LAN and pasted at your cursor position.

No cloud. No accounts. No data leaves your local network. The speech recognition is done entirely by the voice keyboard on your phone — Airvoice doesn't include its own STT engine.

## When to Use Airvoice

Airvoice is for a **very specific scenario**:

- You use a desktop without a built-in microphone (tower PC, Mac mini, headless server)
- You want to dictate text into any app (editor, terminal, browser, chat) using your phone
- You care about privacy and want LAN-only, no-cloud communication
- You're on Linux where most commercial voice input tools don't work well

**If this isn't your situation, you probably don't need Airvoice.**

## Alternatives

For most people, these tools are a better fit:

| Tool | Description | Best for |
|------|-------------|----------|
| **[TypeLess](https://typeless.com)** | AI-powered voice-to-text for Mac | Mac users who want polished, all-in-one voice typing |
| **豆包输入法** (Doubao IME) | Voice keyboard with excellent Chinese & English recognition | Mobile-first voice input, works great standalone |
| **微信输入法** | WeChat's voice keyboard | WeChat ecosystem users |
| **讯飞输入法** | iFlyte's voice keyboard with strong Chinese STT | Chinese language heavy users |
| **Whisper-based tools** | Open-source local STT (e.g., whisper.cpp) | Users who want fully local, self-hosted speech recognition |

## Design Decisions

| Topic | Decision |
|-------|----------|
| Product name | **Airvoice** |
| Mobile | iOS SwiftUI, iOS 17+ |
| Desktop | Go CLI (`airvoice serve`) |
| Pairing | QR code with one-time token |
| Send mode | Auto-send when keyboard dismisses or text idle 1.5s |
| Multi-line | Paste as-is, preserve `\n` |
| Linux | X11 + Wayland supported |
| STT | Third-party IME only — we don't do speech recognition |
