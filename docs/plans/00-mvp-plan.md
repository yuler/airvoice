# Airvoice MVP — Oneshot Build Plan

> **For agents:** Implement the full MVP from this document in one pass.  
> Context: [background.md](../background.md) · Structure & protocol: [architecture.md](../architecture.md)  
> Status: **in-progress**

## Progress

| Task | Status |
|------|--------|
| A1. `go.mod` | ✅ completed |
| A2. `cli/protocol/messages.go` | ✅ completed |
| A3. `cli/pairing/` | ✅ completed |
| A4. `cli/paste/` | ✅ completed |
| A5. `cli/server/` | ✅ completed |
| A6. `cli/main.go` | ✅ completed |
| A7. Go verify (`go test ./cli/... && go build`) | ✅ completed |
| B1. iOS project shell | ✅ completed |
| B2. iOS models | ✅ completed |
| B3. `ConnectionManager.swift` | ✅ completed |
| B4. `AutoSendController.swift` | ✅ completed |
| B5. iOS views | ✅ completed |
| B6. `Info.plist` | ✅ completed |
| C. README update | ✅ completed |

## Goal

Deliver a working Airvoice MVP:

1. `airvoice serve` — Go CLI prints QR, accepts one iOS client, pastes text at cursor (macOS, Linux X11, Wayland).
2. iOS app — scan QR, dictate via 豆包/微信 IME in `TextEditor`, auto-send to PC, show ack feedback.

## Success criteria

- [ ] QR pair on same Wi‑Fi in < 30 s
- [ ] 50+ Chinese chars via 豆包输入法 → PC cursor within 2 s of IME confirm
- [ ] Multiline `\n` preserved
- [ ] macOS + Linux X11 + Linux Wayland
- [ ] No cloud, no accounts, no custom STT

## Build order

```text
1. Go: cli/protocol → cli/pairing → cli/paste → cli/server → cli/main.go
2. Go: go test ./cli/... && go build -o bin/airvoice ./cli
3. Go: smoke with curl + websocat
4. iOS: Xcode project → models → services → views
5. E2E: serve + scan + dictate + paste
```

---

## Part A — Go CLI

### A1. `go.mod`

```go
module github.com/airvoice/airvoice

go 1.22

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/mdp/qrterminal/v3 v3.2.0
)
```

### A2. `cli/protocol/messages.go`

```go
package protocol

import "encoding/json"

type Inbound struct {
	Type    string `json:"type"`
	Device  string `json:"device,omitempty"`
	App     string `json:"app,omitempty"`
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	TS      int64  `json:"ts,omitempty"`
}

type Outbound struct {
	Type    string `json:"type"`
	Host    string `json:"host,omitempty"`
	Version string `json:"version,omitempty"`
	ID      string `json:"id,omitempty"`
	OK      bool   `json:"ok,omitempty"`
	Message string `json:"message,omitempty"`
}

func ParseInbound(data []byte) (Inbound, error) {
	var m Inbound
	return m, json.Unmarshal(data, &m)
}

func (o Outbound) Bytes() ([]byte, error) { return json.Marshal(o) }
```

Test: unmarshal `{"type":"text","id":"abc","content":"你好\n世界","ts":1}` preserves newlines.

### A3. `cli/pairing/`

- `payload.go` — `Payload{Version int \`json:"v"\`, WS, Token string}` + `Marshal() []byte`
- `lanip.go` — `LocalIPv4()` walks interfaces, returns first non-loopback IPv4
- `qr.go` — `PrintQR(payload []byte)` via `qrterminal.Generate(..., os.Stderr)`

### A4. `cli/paste/`

**`paste.go`**

```go
type Paster interface {
	Paste(text string) error
	Name() string
}
func New() (Paster, error) // switch on DetectSession()
```

**`detect.go`** — `SessionDarwin | SessionX11 | SessionWayland | SessionUnknown`  
Linux wayland if `XDG_SESSION_TYPE==wayland` OR `WAYLAND_DISPLAY` is set.

**`darwin.go`** (`//go:build darwin`)

```go
func (d *darwin) Paste(text string) error {
	// pbcopy stdin ← text; sleep 80ms; osascript keystroke "v" using command down
}
```

**`linux_x11.go`** — `xclip -selection clipboard` → sleep 80ms → `xdotool key ctrl+v`

**`linux_wayland.go`** — `wl-copy` → sleep 80ms → `ydotool key 29:1 47:1 47:0 29:0`

### A5. `cli/server/`

**`hub.go`** — mutex; `Set(conn)` closes previous; `Clear(conn)` if same.

**`handler.go`** — on WS message:

| `type` | Action |
|--------|--------|
| `hello` | reply `{type:"hello", host, version:"0.1.0"}` |
| `text` | `paster.Paste(content)` → `{type:"ack", id, ok:true}` or `ok:false, message:err` |
| `ping` | `{type:"pong"}` |

**`server.go`**

```go
type Config struct {
	Addr, Token, Hostname, Version string
	Paster paste.Paster
}
// Handler(): GET /health → "ok"; GET /ws → upgrade
// Token: r.URL.Query().Get("token") must match Config.Token else 401
// ListenAndServe on 0.0.0.0:port
```

Test: bad token → 401 on WS dial.

### A6. `cli/main.go`

```go
// airvoice serve [--port 7383]
//   token := uuid.NewString()
//   ip := pairing.LocalIPv4()
//   wsURL := fmt.Sprintf("ws://%s:%d/ws", ip, port)
//   PrintQR(payload{v:1, ws, token})
//   paste.New(); server.ListenAndServe
// airvoice version → "airvoice 0.1.0"
```

Imports in `cli/main.go` use `github.com/airvoice/airvoice/cli/pairing`, `cli/paste`, `cli/server`, etc.

Log paste backend + ws URL to **stderr**. QR to stderr.

### A7. Go verify

```bash
go test ./cli/...
go build -o bin/airvoice ./cli
./bin/airvoice serve &
curl -s localhost:7383/health   # ok
# websocat "ws://127.0.0.1:7383/ws?token=<TOKEN>"
# → {"type":"hello","device":"test","app":"0.1.0"}
# → {"type":"text","id":"1","content":"hello","ts":1}
```

---

## Part B — iOS App

Target: **iOS 17+**, SwiftUI, Bundle ID `com.airvoice.app`.

### B1. Project shell

- `AirvoiceApp.swift` — `@main` → `ContentView`
- `ContentView` — `@AppStorage("hasSeenOnboarding")` → `OnboardingView` or `HomeView`

### B2. Models

**`PairingPayload.swift`**

```swift
struct PairingPayload: Codable {
    let v: Int; let ws: String; let token: String
    static func decode(from string: String) throws -> PairingPayload
}
```

**`ProtocolMessage.swift`**

```swift
struct OutboundHello: Encodable { let type = "hello"; let device, app: String }
struct OutboundText: Encodable { let type = "text"; let id, content: String; let ts: Int }
struct InboundMessage: Decodable {
    let type: String; let host, message: String?; let id: String?; let ok: Bool?
}
```

### B3. `ConnectionManager.swift`

- `@Published state: disconnected | connecting | connected | error(String)`
- `@Published hostName: String?`
- `connect(payload:)` — append `?token=` to ws URL, `URLSessionWebSocketTask`
- On open: send `OutboundHello(device: UIDevice.current.model, app: "0.1.0")`
- `sendText(id:content:)` — `OutboundText`
- Receive loop → `hello` sets hostName; `ack` calls `onAck?(id, ok, message)`
- `disconnect()`

### B4. `AutoSendController.swift`

```swift
@MainActor final class AutoSendController {
    var onSend: ((String) -> Void)?
    private let idleSeconds = 1.5
    private var debounceTask: Task<Void, Never>?
    private var lastAcked: String?
    private(set) var inFlight = false

    func textDidChange(_ text: String) { /* cancel + sleep idleSeconds + attemptSend */ }
    func keyboardDidHide(currentText: String) { Task { await attemptSend(currentText) } }
    func resetOnFocus() { debounceTask?.cancel() }
    func markAcked(_ content: String) { lastAcked = content; inFlight = false }
    func clearInFlight() { inFlight = false }

    private func attemptSend(_ raw: String) async {
        // trim non-empty; !inFlight; raw != lastAcked; inFlight=true; onSend?(raw) — keep newlines
    }
}
```

Unit tests: debounce fires; dedup skips same content after `markAcked`.

### B5. Views

**`OnboardingView`** (Chinese)

- 推荐：豆包输入法（主）· 备选：微信输入法
- 设置 → 通用 → 键盘 → 添加键盘
- 无需「完全访问」（只读 App 内输入框）
- Button「开始使用」

**`QRScannerView`** — VisionKit `DataScannerViewController` wrapper; on QR string → `PairingPayload.decode` → callback.

**`HomeView`**

- Status bar: `已连接 {host}` / `未连接` + scan button
- `@FocusState` + `TextEditor` — `onChange` → `autoSend.textDidChange`
- Button「说话」→ `focused = true`
- `keyboardWillHideNotification` → `autoSend.keyboardDidHide(currentText: text)`
- Wire `onSend` → `connection.sendText`; `onAck` ok → `markAcked`, `text=""`, haptic, toast; error → `clearInFlight`, toast
- `isIdleTimerDisabled = true` when connected
- 5 s send timeout → `clearInFlight`, toast「超时，请重试」

**`Toast.swift`** — bottom banner, auto-dismiss 2 s.

### B6. Info.plist

- `NSCameraUsageDescription` = `扫描二维码以连接电脑`
- `MinimumOSVersion` = 17.0

### B7. iOS verify

1. Mac/Linux: `./bin/airvoice serve`
2. iPhone same Wi‑Fi: scan QR
3. Tap 说话 → 豆包语音 → confirm
4. Text at PC cursor; iOS clears + toast

---

## Part C — README update

Update root `README.md`:

- Link `docs/background.md`, `docs/architecture.md`, `docs/plans/00-mvp-plan.md`
- Quick start: build CLI, `airvoice serve`, iOS scan
- Platform deps (Accessibility, xclip/xdotool, ydotoold)
- Remove Tauri references

---

## E2E checklist

| # | Check |
|---|--------|
| 1 | `go test ./cli/...` passes |
| 2 | `airvoice serve` shows QR + paste backend on stderr |
| 3 | Invalid token → 401 |
| 4 | Valid `text` message → paste at cursor (multiline) |
| 5 | iOS onboarding → scan → connected badge |
| 6 | Auto-send on idle + keyboard hide |
| 7 | Ack failure shows toast, keeps text |
| 8 | 微信输入法 works as fallback |

## Non-goals (do not implement)

PC hotkey trigger, Apple Watch, streaming deltas, mDNS, Windows, GUI desktop, custom STT, settings screen, manual「发送到 PC」button.
