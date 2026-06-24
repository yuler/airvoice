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

Longer term the same transport layer can support PC-triggered recording, Apple Watch, or local AI cleanup. MVP optimizes for shipping fast.

## Why not existing tools?

| Approach | Examples | Limitation |
|----------|----------|------------|
| Virtual microphone | [WO Mic](https://wolicheng.com), [AudioRelay](https://audiorelay.net) | Kernel drivers / audio plumbing; not a text bridge |
| Cross-screen IME | [讯飞输入法](https://srf.xunfei.cn) | Cloud-only, IME lock-in, poor Linux support |
| Remote control | [Unified Remote](https://www.unifiedremote.com) | Remote-first, not focused typing bridge |

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

## Interaction model

```text
[PC]  airvoice serve  →  QR in terminal
[iOS] Scan QR  →  WebSocket connected
[iOS] Tap 说话  →  focus TextEditor, show IME keyboard
[iOS] Tap mic on 豆包/微信  →  dictate  →  confirm on IME
[iOS] Auto-send when idle / keyboard hides
[PC]  Paste at cursor  →  ack
[iOS] Haptic + toast, clear field
```

**MVP constraint:** trigger is phone-side. PC cannot wake the iOS mic while app is backgrounded (iOS policy). App should stay foreground during use.

## Risks

| Risk | Mitigation |
|------|------------|
| iOS background limits | Phone-triggered flow only |
| Paste blocked (terminal, secure fields) | `error` ack + CLI log |
| Linux Wayland fragmentation | `ydotool` + docs; test GNOME/KDE |
| IME streams partial text | 1.5 s debounce + dedup |
| Auto-send while editing | Reset debounce on text change / refocus |
| LAN IP change | Re-scan QR |

## Out of scope (MVP)

PC hotkey trigger, Apple Watch, real-time char streaming, cloud relay, Windows client, GUI desktop app, custom STT, E2E encryption.
