---
title: Background
description: Project motivation, comparisons, and core decisions.
order: 2
---

# Background

## Problem

Desktop setups often lack a reliable microphone: no mic on a tower PC, flaky Bluetooth headsets, or you want to pace around and dictate while your eyes stay on the monitor. A phone is already a better mic with stronger noise cancellation. The gap is getting recognized **text** into whatever app has focus on Mac or Linux.

## Inspiration

This project started from a [Gemini conversation](https://gemini.google.com/share/62257304d6bd) exploring remote voice input: phone as mic, PC as receiver, solving scenarios where wiring a mic into the desktop is awkward.

## Product idea

**Airvoice is a bridge, not a speech engine.**

- **Mobile:** iOS app with a text field; user dictates via a third-party voice keyboard (**豆包输入法** primary, **微信输入法** fallback).
- **Desktop:** Go CLI on macOS / Linux; receives text over LAN WebSocket and pastes at the cursor.
- **No custom STT in MVP** — leverage existing IME quality instead of training or hosting models.

## Why not existing tools?

| Approach | Examples | Limitation |
|----------|----------|------------|
| Virtual microphone | WO Mic, AudioRelay | Kernel drivers / audio plumbing; not a text bridge |
| Cross-screen IME | 讯飞输入法 | Cloud-only, IME lock-in, poor Linux support |
| Remote control | Unified Remote | Remote-first, not focused typing bridge |

**Airvoice differentiation:** LAN-only, Linux-friendly, no custom STT, use the keyboard you already trust.

## Design decisions (locked)

| Topic | Decision |
|-------|----------|
| Product name | **Airvoice** |
| Mobile | iOS SwiftUI, iOS 17+ |
| Desktop | Go CLI (`airvoice serve`) |
| Pairing | QR code with one-time token (no mDNS in MVP) |
| Send mode | Auto-send when keyboard dismisses or text idle 1.5 s |
| Multi-line | Paste as-is, preserve `\n` |
| Linux | X11 + Wayland in MVP |
| STT | Third-party IME only |
