---
title: 架构
description: 系统设计、仓库结构和通信协议。
order: 4
---

## 系统概览

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

**数据路径：** IME 写入 `TextEditor` → iOS 自动发送 → `{type:"text"}` → Go CLI → 剪贴板 + 模拟 Cmd/Ctrl+V → 光标位置。

## Go 包

| 包 | 职责 |
|---|------|
| `cli` (`main`) | 入口：`serve`, `version`；连接配对 + 服务器 + 粘贴 |
| `cli/protocol` | 入站/出站 JSON 消息类型 |
| `cli/pairing` | 局域网 IP、QR payload 序列化、终端 QR 渲染 |
| `cli/server` | `/health`, `/ws`, token 认证, hello/text/ping 处理器, 单客户端 hub |
| `cli/paste` | 平台粘贴：剪贴板 + 模拟粘贴键 |

## 粘贴后端

| 操作系统 | 检测方式 | 剪贴板 | 按键 | 依赖 |
|----------|----------|--------|------|------|
| macOS | `GOOS=darwin` | `pbcopy` | `osascript` Cmd+V | 辅助功能权限 |
| Linux X11 | not wayland | `xclip` | `xdotool ctrl+v` | `xclip`, `xdotool` |
| Linux Wayland | `XDG_SESSION_TYPE=wayland` 或 `WAYLAND_DISPLAY` | `wl-copy` | `ydotool` Ctrl+V 扫描码 | `wl-clipboard`, `ydotool`, `ydotoold` |

## 通信协议

传输：**WebSocket**，JSON 文本帧，仅限局域网。

### 配对 QR payload

```json
{ "v": 1, "ws": "ws://192.168.1.42:7383/ws", "token": "uuid" }
```

### 客户端 → 服务器

```json
{ "type": "hello", "device": "iPhone", "app": "0.1.0" }
{ "type": "text", "id": "uuid", "content": "你好\n世界", "ts": 1710000000 }
{ "type": "ping" }
```

### 服务器 → 客户端

```json
{ "type": "hello", "host": "my-mac", "version": "0.1.0" }
{ "type": "ack", "id": "uuid", "ok": true }
{ "type": "ack", "id": "uuid", "ok": false, "message": "paste failed: ..." }
{ "type": "pong" }
```

## 依赖

**Go：** `github.com/gorilla/websocket`, `github.com/google/uuid`, `github.com/mdp/qrterminal/v3`

**iOS：** SwiftUI, VisionKit (iOS 17+), 无第三方 WS 库。
