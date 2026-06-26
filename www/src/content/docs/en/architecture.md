---
title: Architecture
description: System design, repository structure, and wire protocol.
order: 3
---

# Architecture

## System overview

```text
┌────────────────────────────────────────────────────────────┐
│  iOS (SwiftUI)                                             │
│  Onboarding → QR Scanner → Home (TextEditor + 豆包/微信 IME) │
│       │                                                    │
│       │  WebSocket / JSON (LAN)                            │
└───────┼────────────────────────────────────────────────────┘
        ▼
┌────────────────────────────────────────────────────────────┐
│  airvoice CLI (Go)                                         │
│  QR pairing │ WS server │ token auth │ paste injection     │
└────────────────────────────────────────────────────────────┘
        ▼
   Focused app on macOS / Linux (any text field)
```

**Data path:** IME writes into `TextEditor` → iOS auto-send → `{type:"text"}` → Go CLI → clipboard + synthetic Cmd/Ctrl+V → cursor.

## Go packages

| Package | Responsibility |
|---------|----------------|
| `cli` (`main`) | Entry: `serve`, `version`; wires pairing + server + paste |
| `cli/protocol` | Inbound/outbound JSON message types |
| `cli/pairing` | LAN IP, QR payload marshal, terminal QR render |
| `cli/server` | `/health`, `/ws`, token auth, hello/text/ping handlers, single-client hub |
| `cli/paste` | Platform paste: clipboard + synthetic paste key |

## Paste backends

| OS | Detection | Clipboard | Keystroke | Deps |
|----|-----------|-----------|-----------|------|
| macOS | `GOOS=darwin` | `pbcopy` | `osascript` Cmd+V | Accessibility permission |
| Linux X11 | not wayland | `xclip` | `xdotool ctrl+v` | `xclip`, `xdotool` |
| Linux Wayland | `XDG_SESSION_TYPE=wayland` or `WAYLAND_DISPLAY` | `wl-copy` | `ydotool` Ctrl+V scancodes | `wl-clipboard`, `ydotool`, `ydotoold` |

## Wire protocol

Transport: **WebSocket**, JSON text frames, LAN only.

### Pairing QR payload

```json
{ "v": 1, "ws": "ws://192.168.1.42:7383/ws", "token": "uuid" }
```

### Client → server

```json
{ "type": "hello", "device": "iPhone", "app": "0.1.0" }
{ "type": "text", "id": "uuid", "content": "你好\n世界", "ts": 1710000000 }
{ "type": "ping" }
```

### Server → client

```json
{ "type": "hello", "host": "my-mac", "version": "0.1.0" }
{ "type": "ack", "id": "uuid", "ok": true }
{ "type": "ack", "id": "uuid", "ok": false, "message": "paste failed: ..." }
{ "type": "pong" }
```

## Dependencies

**Go:** `github.com/gorilla/websocket`, `github.com/google/uuid`, `github.com/mdp/qrterminal/v3`

**iOS:** SwiftUI, VisionKit (iOS 17+), no third-party WS library.
