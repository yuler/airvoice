# 24-relay-mode-design

## Overview

Add an optional, end-to-end encrypted (E2E) **Relay Mode** to Airvoice so phones and computers can connect across LANs, cellular networks, or isolated corporate Wi-Fi — using either a free, quota-limited official relay hosted at `wss://relay.airvoice.yuler.cc/ws` or any self-hosted instance.

LAN mode remains the default with zero external network traffic. Relay mode is explicitly opt-in, zero-knowledge (the relay never sees plaintext or encryption keys), and requires no user accounts.

---

## Design Decisions

| Area                 | Decision                                                                                                                                                     |
| :------------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Connection Modes     | Explicit mode switch: LAN by default (zero outbound traffic); CLI enables Relay via `--relay [url]`, Desktop switches `LAN / Relay` in Settings              |
| Room & Device Model  | 1 room = 1 Host (PC) + at most 1 active Phone (new phone replaces previous phone); mobile stays paired to 1 host at a time                                   |
| Pairing Entry        | QR code primary + copyable/pasteable pairing link (`airvoice://pair?...`); mobile QR scanner screen adds a bottom "Paste Pairing Link" action                |
| E2E Encryption       | 32-byte random PSK (`relayKey`) in QR/link + HKDF-SHA256 subkeys + AES-256-GCM (0-RTT, per-direction keys, session replay guard, zero 3rd-party crypto libs) |
| Server Packaging     | Standalone Go binary in this repo (`relay/` + `cmd/airvoice-relay`) + official Docker image + `deploy/relay/` (Docker Compose + Caddy self-host template)    |
| Official Relay Host  | `wss://relay.airvoice.yuler.cc/ws` deployed on Aliyun 99 ECS (2C2G / 3 Mbps), reverse-proxied behind Caddy or existing host proxy                            |
| Official Free Quotas | No accounts; in-memory multi-dimensional quotas (16 KiB frame, 30 msg/min, 1,000 msg/day, 5 rooms/IP, 500 global rooms, 15m orphan timeout); all tunable     |
| Self-Host Auth       | Optional `--auth-key` on relay server enables `Bearer` access control; included optionally in `v: 2` QR/link so phones need no manual entry                  |
| Host Config Behavior | Bare `airvoice` always starts LAN; `airvoice --relay` reuses `relayUrl`/`relayAuthKey` from `~/.airvoice/settings.json` (falling back to official URL)       |
| Mobile Confirmation  | Official host `relay.airvoice.yuler.cc` connects directly with `Relay · E2E` badge; custom relay domains prompt a lightweight domain confirmation dialog     |
| Host Offline State   | Relay retains room up to 15m orphan timeout and emits `peer_status(online: false)`; phone shows "Waiting for PC" and auto-resumes when host reconnects       |
| Backwards Compat     | LAN `v: 1` payload and plaintext LAN transport remain unchanged; Relay uses `v: 2` payload so older mobile builds cleanly reject with `unsupportedVersion`   |

---

## System Architecture

```text
┌───────────────────────────┐              ┌──────────────────────────────────┐              ┌───────────────────────────┐
│  Mobile (iOS / Android)   │              │  airvoice-relay (Zero-Knowledge) │              │  Host (CLI / Desktop)     │
│                           │  WSS + E2E   │                                  │  WSS + E2E   │                           │
│  holds: relayKey (32B)    │─────────────▶│  sees only: roomId, role, auth   │◀─────────────│  holds: relayKey (32B)    │
│  derives: roomId, keys    │  ciphertext  │  forwards opaque e2e frames      │  ciphertext  │  derives: roomId, keys    │
│  encrypts/decrypts JSON   │◀─────────────│  enforces rate / quota / idle    │─────────────▶│  decrypts -> Paste()      │
└───────────────────────────┘              └──────────────────────────────────┘              └───────────────────────────┘
```

- **LAN mode** (`v: 1`): unchanged direct WebSocket from Phone → Host on local port (`7654` CLI / `7655` Desktop).
- **Relay mode** (`v: 2`): both Host and Phone open outbound WebSockets to the relay. In Relay mode, the Host does **not** bind a local LAN listening port (eliminating port conflicts and LAN-IP detection failures on cellular hotspots).

---

## Pairing Payload v2 & Pairing Link

### QR JSON Payload (`v: 2`)

```json
{
  "v": 2,
  "mode": "relay",
  "ws": "wss://relay.airvoice.yuler.cc/ws",
  "key": "<43-char base64url raw 32-byte relayKey>",
  "auth": "<optional self-hosted relay auth key>"
}
```

- `v: 1` remains strictly for LAN (`{v, ws, token}`).
- `v: 2` requires `mode == "relay"`, non-empty `ws`, and valid 32-byte base64url-nopad `key`. `auth` is omitted when empty.

### Copyable Pairing Link

```text
airvoice://pair?v=2&mode=relay&ws=wss%3A%2F%2Frelay.airvoice.yuler.cc%2Fws&key=<base64url>[&auth=<urlencoded>]
```

Mobile paste parsing accepts both `airvoice://pair?...` URIs and raw JSON payloads (`v: 1` or `v: 2`).

### Transport Security Validation

Reuse the existing `isLocalHost` rule across iOS and Android:

- Public hosts MUST use `wss://` (TLS + E2E defense in depth).
- Private/local hosts (`localhost`, RFC 1918, `.local`, Tailscale `100.64.0.0/10` / `fc00::/7`) permit `ws://` for private self-hosted deployments while still enforcing mandatory E2E encryption.

---

## End-to-End Encryption (E2E) Specification

### Master Secret & Rotation

`~/.airvoice/settings.json` adds `relayKey` (32 random bytes from `crypto/rand`, encoded as unpadded base64url).

- Existing `token` (UUID) is preserved so LAN pairings never break on upgrade.
- `relayKey` is lazily generated the first time Relay mode is used.
- `airvoice token refresh` and the Desktop Refresh button rotate **both** `token` and `relayKey`, disconnecting active LAN or Relay sessions immediately.

### Key Derivation (HKDF-SHA256)

Using RFC 5869 HKDF-SHA256 with `ikm = raw 32-byte relayKey`, `salt = nil` (32 zero bytes), and distinct `info` strings:

| Derived Value | `info` String                 | Length | Sent to Relay? | Purpose                                        |
| :------------ | :---------------------------- | :----- | :------------- | :--------------------------------------------- |
| `roomId`      | `airvoice-relay-room-v1`      | 16 B   | Yes (b64url)   | Opaque room routing identifier (22 chars)      |
| `hostAuth`    | `airvoice-relay-host-auth-v1` | 32 B   | Yes (b64url)   | Proves Host ownership of room slot             |
| `clientAuth`  | `airvoice-relay-client-v1`    | 32 B   | Yes (b64url)   | Proves Client ownership of room slot           |
| `keyC2H`      | `airvoice-e2e-c2h-v1`         | 32 B   | **Never**      | AES-256-GCM key for Phone → Host frames        |
| `keyH2C`      | `airvoice-e2e-h2c-v1`         | 32 B   | **Never**      | AES-256-GCM key for Host → Phone frames        |

**Zero third-party crypto dependencies:**

- **Go 1.25**: `crypto/hkdf`, `crypto/aes`, `crypto/cipher`, `crypto/sha256`
- **iOS 17+**: `CryptoKit` (`HKDF<SHA256>`, `AES.GCM`)
- **Android (`minSdk 26`)**: `javax.crypto` (`HmacSHA256` for 25-line RFC 5869 HKDF, `AES/GCM/NoPadding`)

### Encrypted Wire Envelope & Replay Protection

Every application message (`hello`, `text`, `ack`, `ping`, `pong`) is JSON-encoded as plaintext and wrapped in an `e2e` envelope:

```json
{
  "type": "e2e",
  "v": 1,
  "sid": "<11-char base64url 8-byte connection session ID>",
  "seq": 1,
  "iv": "<16-char base64url 12-byte random GCM nonce>",
  "ct": "<base64url AES-256-GCM ciphertext || 16-byte auth tag>"
}
```

- **AAD**: UTF-8 bytes of `airvoice-e2e-v1|<dir>|<roomId>|<sid>|<seq>` where `<dir>` is `c2h` (Phone → Host) or `h2c` (Host → Phone).
- **Per-connection session (`sid` + `seq`)**: Each peer generates a fresh random 8-byte `sid` when a connection opens or `peer_status(online: true)` transitions to online, starting `seq = 1` and incrementing by 1 per frame.
- **Anti-replay**: Receivers track the peer's active `sid` and require `seq > lastSeq`. Previously retired `sid`s seen during the process lifetime are rejected (bounded ring/set of last 256 `sid`s). `hello` includes `ts` (Unix seconds) checked within ±5 minutes to block cross-restart replay of captured sessions.
- **Failure handling**: Any authentication tag failure or replay violation drops the frame and closes/resets the session cleanly; no plaintext ever leaks on error.

---

## Relay Server Protocol & Lifecycle

### Endpoints & Connection Handshake

- `GET /health` → `200 ok`
- `GET /ws` → WebSocket upgrade. Headers (with query parameter fallback for diagnostics):

| Header               | Query Fallback | Required        | Description                                  |
| :------------------- | :------------- | :-------------- | :------------------------------------------- |
| `X-Airvoice-Room`    | `room`         | Yes             | Derived `roomId` (22-char base64url)         |
| `X-Airvoice-Role`    | `role`         | Yes             | `host` or `client`                           |
| `X-Airvoice-Auth`    | `auth`         | Yes             | Derived `hostAuth` or `clientAuth`           |
| `Authorization`      | —              | If `--auth-key` | `Bearer <relayAuthKey>` for private relays   |

### Room Ownership & Replacement Rules

- A room holds at most two slots: `host` and `client`.
- The first valid connection for a role stores `SHA-256(auth)` on that slot for the room's lifetime.
- A subsequent connection for the same role with the **same** `SHA-256(auth)` replaces the previous socket (closing the old socket with code `4001 Replaced`).
- A subsequent connection for the same role with a **different** `SHA-256(auth)` is rejected with HTTP `403 Forbidden`, preventing slot hijacking even if `roomId` were exposed.

### Relay Control Frames

The relay emits only two plaintext JSON control types and forwards all `type: "e2e"` text frames verbatim without inspection:

```json
{ "type": "peer_status", "online": true }
{ "type": "peer_status", "online": false }
{ "type": "error", "code": "rate_limited", "message": "too many messages; try again shortly" }
```

| Error Code           | Trigger                                                              |
| :------------------- | :------------------------------------------------------------------- |
| `peer_offline`       | Endpoint sent an `e2e` frame while the opposite peer is not online   |
| `message_too_large`  | Frame exceeded `MAX_MESSAGE_BYTES` (default 16 KiB)                  |
| `rate_limited`       | Room exceeded per-minute token bucket (default 30/min, burst 10)     |
| `quota_exceeded`     | Room or IP exceeded daily message limit (default 1,000/day)          |
| `room_limit_reached` | Handshake rejected (`503` / `429`) when per-IP or global cap is full |

### Keepalive & Idle Eviction

- Relay sends WebSocket `Ping` every `30s`, expects `Pong` within `90s`.
- **Orphan idle timeout** (`15m` default): if only one peer is connected for 15 minutes, the relay closes the connection with code `4002 Idle Timeout` and deletes the room.
- **Session idle timeout** (`2h` default): if neither peer sends any `e2e` frame for 2 hours, the relay closes both sockets and reclaims the room.

---

## Official Free Relay Quotas & Deployment

### Configurable Limits (`cmd/airvoice-relay`)

All limits live in memory (zero database, zero message logging) and are configurable via CLI flags or `AIRVOICE_RELAY_*` environment variables:

| Env Var / Flag                         | Official Default | Self-Host Guidance        |
| :------------------------------------- | :--------------- | :------------------------ |
| `AIRVOICE_RELAY_ADDR`                  | `:8787`          | Bind `127.0.0.1:8787`     |
| `AIRVOICE_RELAY_AUTH_KEY`              | `""` (open)      | Set secret for private    |
| `AIRVOICE_RELAY_MAX_MESSAGE_BYTES`     | `16384` (16 KiB) | Raise if needed           |
| `AIRVOICE_RELAY_RATE_PER_MINUTE`       | `30`             | `0` disables rate limit   |
| `AIRVOICE_RELAY_RATE_BURST`            | `10`             | Token bucket burst        |
| `AIRVOICE_RELAY_DAILY_MESSAGE_LIMIT`   | `1000`           | `0` disables daily cap    |
| `AIRVOICE_RELAY_MAX_ROOMS_PER_IP`      | `5`              | Per client IP cap         |
| `AIRVOICE_RELAY_MAX_GLOBAL_ROOMS`      | `500`            | Memory/FD ceiling         |
| `AIRVOICE_RELAY_ORPHAN_IDLE_TIMEOUT`   | `15m`            | Single-peer wait limit    |
| `AIRVOICE_RELAY_SESSION_IDLE_TIMEOUT`  | `2h`             | Inactive room eviction    |
| `AIRVOICE_RELAY_TRUST_PROXY`           | `false`          | `true` behind Caddy/Nginx |

### Repository Layout & Deployment Templates

```text
airvoice/
├── cli/
│   ├── e2e/                     # Shared HKDF-SHA256 + AES-256-GCM envelope & replay guard
│   ├── relayclient/             # Host outbound WebSocket relay transport -> server message handler
│   └── ...
├── relay/                       # Relay server hub, room auth, rate/quota limiter, HTTP/WS server
├── cmd/
│   └── airvoice-relay/          # Standalone relay binary entrypoint
└── deploy/
    └── relay/
        ├── Dockerfile           # Multi-stage static Go build
        ├── docker-compose.yml   # Relay service (+ optional Caddy profile)
        └── Caddyfile            # Auto-TLS reverse proxy template for relay.airvoice.yuler.cc
```

- On Aliyun `ssh 99`: run `airvoice-relay` container bound to `127.0.0.1:8787` with `AIRVOICE_RELAY_TRUST_PROXY=true`. If 80/443 is already managed by an existing Nginx/Caddy instance on the host, add a WebSocket reverse proxy vhost for `relay.airvoice.yuler.cc` pointing to `127.0.0.1:8787`; otherwise enable the bundled Caddy profile.
- GitHub Releases workflow will additionally build `airvoice-relay-linux-{amd64,arm64}` (and publish a GHCR image if desired).

---

## Client & UX Specification

### Shared Settings Schema (`~/.airvoice/settings.json`)

```json
{
  "token": "<uuid>",
  "relayKey": "<43-char base64url>",
  "port": 7655,
  "autoStart": false,
  "language": "zh-CN",
  "connectionMode": "lan",
  "relayUrl": "",
  "relayAuthKey": ""
}
```

### CLI (`cli/`)

```text
airvoice [--port 7654]
airvoice --relay [wss://...] [--relay-auth-key <key>]
airvoice token refresh
airvoice settings
```

- Bare `airvoice` always starts LAN mode (never unexpectedly connects to the internet even if Desktop is set to Relay).
- `airvoice --relay` resolves URL as: explicit CLI value → `settings.json` `relayUrl` → `wss://relay.airvoice.yuler.cc/ws`. Same precedence for `--relay-auth-key`.
- Prints terminal QR (`v: 2`), `Relay: <url> (E2E)`, copyable `Link: airvoice://pair?...`, and live status logs (`relay connected`, `peer online`, `peer offline`, `quota/rate warnings`).
- Automatic exponential-backoff reconnect to the relay (2s → 30s) when network blips occur; stops on `401`/`403` auth errors.

### Desktop GUI (`desktop/`)

- **Settings panel**: adds Connection Mode toggle (`局域网 / 中继` | `LAN / Relay`), Relay Server URL input (placeholder `wss://relay.airvoice.yuler.cc/ws`), optional Relay Auth Key input.
- **QR card**: displays a compact `Relay · E2E` pill when in Relay mode and a "Copy Pairing Link" button next to Refresh.
- **Status badge**: distinguishes `Waiting for phone`, `Connected (LAN)`, `Connected (Relay · E2E)`, and `Relay disconnected (reconnecting...)`.

### Mobile (iOS & Android)

- **Pairing**:
  - `PairingPayload` parses both `v: 1` (LAN) and `v: 2` (Relay + E2E) from QR JSON or `airvoice://pair?...` URI.
  - QR Scanner bottom bar adds a **Paste Pairing Link** button (reads clipboard and validates immediately, or opens a text input dialog if clipboard is empty/invalid).
  - `relayKey` and optional `relayAuthKey` are stored in iOS Keychain / Android `TokenCipher` (encrypted), never plaintext preferences.
- **Confirmation**:
  - Official host `relay.airvoice.yuler.cc` connects directly without an alarming warning dialog.
  - Custom self-hosted relay hosts display a concise confirmation: *"Connect via custom relay `<host>` (end-to-end encrypted)?"*
- **Runtime States & Errors**:
  - Connected to relay, `peer_status(online: false)` → state `waitingForHost` (`等待电脑上线` / `Waiting for PC`), send button disabled, no error toast spam.
  - `peer_status(online: true)` → phone resets its E2E `sid`/`seq`, sends encrypted `hello`, and transitions to `connected` (`已连接 · E2E`) upon receiving encrypted `hello` reply.
  - Relay `error` frames (`rate_limited`, `quota_exceeded`, `message_too_large`) surface clear localized toasts without wiping saved pairing credentials.
  - Settings sheet displays `Relay (E2E): <host>` under Connection & Persistence.

---

## Security Properties & Limitations

- **Confidentiality & Integrity**: All voice text, hostnames, device names, and acks are AES-256-GCM encrypted end-to-end; the relay only observes connection timing, room IDs, and ciphertext length.
- **Known Trade-off (Static PSK)**: By design (for 0-RTT simplicity and zero extra handshake states), Relay mode uses a static pairing PSK without per-session forward secrecy. Users can rotate `relayKey` at any time via `airvoice token refresh` or Desktop Refresh.
- **Quota Abuse Resistance**: Because daily limits are counted per room and per IP in memory with a strict 16 KiB frame cap, the free official relay cannot be exploited as a bulk file tunnel or high-bandwidth proxy on the 3 Mbps Aliyun ECS.

---

## Verification Plan

1. **Cross-platform E2E test vectors**: Shared JSON test vectors verified in Go (`cli/e2e`), Swift (`ios/AirvoiceTests`), and Kotlin (`android/.../test`) proving identical HKDF derivation, AAD construction, AES-256-GCM seal/open, and replay rejection.
2. **Relay unit & integration tests**: Room pairing, `hostAuth`/`clientAuth` slot protection, `peer_status` transitions, per-minute rate limit, daily quota enforcement, 16 KiB frame rejection, orphan & session idle timeouts, and optional `--auth-key` auth.
3. **Host & Desktop tests**: `cli/config` `relayKey` persistence & dual rotation on refresh, `cli/relayclient` reconnect & paste pipeline, Desktop settings switch & QR/link generation.
4. **Mobile tests**: `PairingPayload` v1/v2 & URI parsing, official vs custom relay confirmation gating, `waitingForHost` state transitions, encrypted round-trip against mock WebSocket server.
5. **Live deployment smoke test**: Deploy `airvoice-relay` to Aliyun `ssh 99` behind `wss://relay.airvoice.yuler.cc/ws` and verify cross-network dictation from cellular phone to PC.
