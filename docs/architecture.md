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
│  airvoice CLI (Go)          │  Desktop GUI (Wails + Vue 3) │
│  QR pairing │ WS server    │  QR code │ status │ history  │
│  token auth │ paste inject │  settings │ system tray       │
└────────────────────────────────────────────────────────────┘
        ▼
   Focused app on macOS / Linux (any text field)
```

**Data path:** IME writes into `TextEditor` → iOS auto-send → `{type:"text"}` → Go CLI → clipboard + synthetic Cmd/Ctrl+V → cursor.

## Repository layout

```text
airvoice/
├── README.md
├── AGENTS.md
├── go.mod
├── go.sum
├── cli/
│   ├── main.go                  # serve, version
│   ├── protocol/
│   │   ├── messages.go          # JSON structs
│   │   └── messages_test.go
│   ├── pairing/
│   │   ├── payload.go           # QR JSON {v, ws, token}
│   │   ├── lanip.go             # pick LAN IPv4
│   │   ├── qr.go                # terminal QR
│   │   └── lanip_test.go
│   ├── server/
│   │   ├── server.go            # HTTP mux, ListenAndServe
│   │   ├── handler.go           # /ws upgrade, message dispatch
│   │   ├── hub.go               # single-client connection
│   │   └── server_test.go
│   └── paste/
│       ├── paste.go             # Paster interface, New()
│       ├── detect.go            # darwin / x11 / wayland
│       ├── darwin.go            # //go:build darwin
│       ├── linux_x11.go         # //go:build linux
│       ├── linux_wayland.go     # //go:build linux
│       └── detect_test.go
├── ios/
│   ├── Airvoice.xcodeproj
│   └── Airvoice/
│       ├── AirvoiceApp.swift
│       ├── Models/
│       │   ├── PairingPayload.swift
│       │   └── ProtocolMessage.swift
│       ├── Services/
│       │   ├── ConnectionManager.swift
│       │   └── AutoSendController.swift
│       ├── Views/
│       │   ├── ContentView.swift
│       │   ├── OnboardingView.swift
│       │   ├── QRScannerView.swift
│       │   └── HomeView.swift
│       ├── Utilities/
│       │   └── Toast.swift
│       └── Info.plist
└── docs/
    ├── background.md
    ├── architecture.md
    └── plans/
        └── 00-mvp-plan.md
```

## Go packages

| Package | Responsibility |
|---------|----------------|
| `cli` (`main`) | Entry: `serve`, `version`; wires pairing + server + paste |
| `cli/protocol` | Inbound/outbound JSON message types |
| `cli/pairing` | LAN IP, QR payload marshal, terminal QR render |
| `cli/server` | `/health`, `/ws`, token auth, hello/text/ping handlers, single-client hub |
| `cli/paste` | Platform paste: clipboard + synthetic paste key |

## Desktop GUI modules

| Module | Responsibility |
|--------|----------------|
| `desktop/main.go` | Wails entry point, app bootstrap |
| `desktop/app.go` | Go backend: WS server, QR generation, connection management, bindings for frontend |
| `desktop/history.go` | SQLite-based message history store |
| `desktop/tray.go` | System tray icon and menu |
| `desktop/frontend/` | Vue 3 + TailwindCSS SPA: QR code, status badge, history list, settings panel, i18n |

### Paste backends

| OS | Detection | Clipboard | Keystroke | Deps |
|----|-----------|-----------|-----------|------|
| macOS | `GOOS=darwin` | `pbcopy` | `osascript` Cmd+V | Accessibility permission |
| Linux X11 | not wayland | `xclip` | `xdotool ctrl+v` | `xclip`, `xdotool` |
| Linux Wayland | `XDG_SESSION_TYPE=wayland` or `WAYLAND_DISPLAY` | `wl-copy` | `ydotool` Ctrl+V scancodes | `wl-clipboard`, `ydotool`, `ydotoold` |

80 ms delay between clipboard set and key event. Multi-line text preserved in clipboard.

## iOS modules

| Unit | Responsibility |
|------|----------------|
| `PairingPayload` | Decode QR JSON |
| `ProtocolMessage` | Encode/decode WS frames |
| `ConnectionManager` | `URLSessionWebSocketTask`, hello/text send, ack handling |
| `AutoSendController` | 1.5 s idle debounce, keyboard-dismiss trigger, dedup, in-flight lock |
| `OnboardingView` | 豆包 / 微信 install guide |
| `QRScannerView` | VisionKit barcode scan |
| `HomeView` | Status, TextEditor, 说话 button, wire auto-send → connection |

## Wire protocol

Transport: **WebSocket**, JSON text frames, LAN only.

### Pairing QR payload

```json
{ "v": 1, "ws": "ws://192.168.1.42:7383/ws", "token": "uuid" }
```

iOS connects to `ws` with `?token=<token>` query param. Token rotates each `airvoice serve` session.

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

### HTTP

- `GET /health` → `200 ok`
- `GET /ws?token=...` → WebSocket upgrade; `401` if token mismatch

### Auto-send rules (iOS)

Send when non-empty text and either:

1. Keyboard hides, or
2. No text change for **1.5 s**

Guards: trim-empty skip, dedup last-acked content, in-flight lock until ack or 5 s timeout, reset debounce on refocus.

On `ack ok`: haptic, toast「已发送到电脑」, clear field. On error: keep text, toast error.

## Dependencies

**Go:** `github.com/gorilla/websocket`, `github.com/google/uuid`, `github.com/mdp/qrterminal/v3`

**iOS:** SwiftUI, VisionKit (iOS 17+), no third-party WS library.

**System (desktop):**

- macOS: Accessibility for Terminal / `airvoice` binary
- Linux X11: `sudo apt install xclip xdotool`
- Linux Wayland: `wl-clipboard`, `ydotool`; `systemctl --user enable --now ydotoold`

## Security (MVP)

- One-time token per `serve` session
- Single active phone connection
- No TLS on LAN (trusted network assumption)
- No cloud, no accounts
