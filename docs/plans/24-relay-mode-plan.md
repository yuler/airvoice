# 24-relay-mode-plan

**Spec:** [`docs/specs/24-relay-mode-design.md`](../specs/24-relay-mode-design.md)

**Goal:** Deliver end-to-end encrypted cross-LAN Relay Mode (official quota-limited relay on Aliyun `ssh 99` + self-hostable Docker/binary) across four incremental, reviewable Pull Requests.

**Tech Stack:** Go 1.25 (`crypto/hkdf`, `crypto/aes`, `gorilla/websocket`), Wails + Vue 3 (Desktop), SwiftUI + CryptoKit (iOS 17+), Kotlin + Jetpack Compose + JCA (Android API 26+), Docker + Caddy.

---

## Phase Overview

| Phase | Scope                                                       | Target PR Deliverables                                                                |
| :---- | :---------------------------------------------------------- | :------------------------------------------------------------------------------------ |
| 1     | Shared E2E Crypto + Relay Server + Deployment Templates     | `cli/e2e/`, `relay/`, `cmd/airvoice-relay/`, `deploy/relay/`, Aliyun 99 live smoke    |
| 2     | Host Support (CLI `--relay` + Desktop GUI Relay Mode)       | `cli/config/`, `cli/pairing/`, `cli/relayclient/`, `cli/main.go`, `desktop/`          |
| 3     | Mobile Clients (iOS & Android v2 Pairing, Paste, E2E, UX)   | `ios/Airvoice/`, `android/app/` with shared cross-platform E2E test vectors           |
| 4     | Docs, Website & Release Automation                          | `README.md`, `docs/architecture.md`, `www/`, `.github/workflows/release.yml`          |

---

## Phase 1: E2E Core, Relay Server & Aliyun 99 Deployment

**Objective:** Build and verify the zero-knowledge relay server, shared Go E2E crypto package, deterministic cross-platform test vectors, and self-hosting / official deployment templates, then deploy and smoke-test on Aliyun `ssh 99`.

### Files

| Path                                  | Responsibility                                                                                 |
| :------------------------------------ | :--------------------------------------------------------------------------------------------- |
| `cli/e2e/keys.go`                     | `GenerateKey`, HKDF-SHA256 derivation (`roomId`, `hostAuth`, `clientAuth`, `keyC2H`, `keyH2C`) |
| `cli/e2e/codec.go`                    | AES-256-GCM `Envelope` seal/open, AAD formatting, `SessionGuard` replay check                  |
| `cli/e2e/e2e_test.go`                 | Unit tests + deterministic test vectors (`cli/e2e/testdata/vectors.json`)                      |
| `relay/config.go`                     | Flag/env config for limits, timeouts, optional `AUTH_KEY`, `TRUST_PROXY`                       |
| `relay/limiter.go`                    | In-memory per-room token bucket, per-room/IP daily quota, per-IP room counter                  |
| `relay/hub.go`                        | Room map, `host`/`client` slot ownership (`SHA-256(auth)`), `peer_status`                      |
| `relay/server.go`                     | `/health`, `/ws` upgrade, frame size guard, ping/pong, idle eviction loop                      |
| `relay/server_test.go`                | End-to-end WebSocket tests covering auth, forwarding, limits, and timeouts                     |
| `cmd/airvoice-relay/main.go`          | Standalone server entrypoint with graceful SIGINT/SIGTERM shutdown                             |
| `deploy/relay/Dockerfile`             | Multi-stage static Linux binary image                                                          |
| `deploy/relay/docker-compose.yml`     | Local-bound relay container + optional Caddy profile                                           |
| `deploy/relay/Caddyfile`              | Auto-TLS reverse proxy config for `relay.airvoice.yuler.cc`                                    |

### Tasks

- [ ] **1.1 E2E crypto package (`cli/e2e`)**: Implement HKDF-SHA256 derivation, AES-256-GCM `Envelope` seal/open with AAD, and `SessionGuard` (requiring each new `sid` to open with a valid `seq == 1` `hello` whose `ts` is within ±5 min before accepting `text`/app frames, plus strictly monotonic `seq` and retired `sid` ring). Export deterministic test vectors to `cli/e2e/testdata/vectors.json` for Phase 3 iOS/Android parity tests, including negative tests for unauthenticated-`sid` `text` replay and stale `hello.ts`.
- [ ] **1.2 Relay hub & slot security (`relay/hub.go`)**: Implement 1-host + 1-client rooms keyed by `roomId`, binding `SHA-256(auth)` per role so matching auth cleanly replaces stale sockets (`4001`) and mismatched auth receives `403`. Emit `peer_status` on join/leave.
- [ ] **1.3 Multi-dimensional memory limiter (`relay/limiter.go`)**: Implement per-IP active room cap (`5`), global room cap (`500`), max frame size (`16 KiB`), per-room rate bucket (`30/min`, burst `10`), and daily counter (`1,000/day` per room and per IP, UTC rollover + periodic GC).
- [ ] **1.4 Relay HTTP/WS server & CLI (`relay/server.go`, `cmd/airvoice-relay`)**: Wire `/health`, `/ws` with header-only auth (`X-Airvoice-Room`/`Role`/`Auth` + optional `Authorization: Bearer`, no query-string auth), trusted proxy IP extraction, 30s/90s ping/pong, 15m orphan idle timeout, 2h session idle timeout, and graceful shutdown.
- [ ] **1.5 Deployment templates (`deploy/relay/`)**: Write `Dockerfile`, `docker-compose.yml`, and `Caddyfile`.
- [ ] **1.6 Verify & Deploy to Aliyun `ssh 99`**:
  - Run `go test -race ./cli/e2e/... ./relay/... ./cmd/airvoice-relay/...`.
  - Inspect existing services/ports on `ssh 99`, deploy `airvoice-relay` behind `wss://relay.airvoice.yuler.cc/ws`, and verify `/health` + two-peer encrypted WebSocket round-trip.

---

## Phase 2: Host Support (CLI `--relay` & Desktop GUI)

**Objective:** Allow both `airvoice` CLI and the Wails Desktop app to run in Relay Mode with E2E encryption, persistent `relayKey`, copyable pairing links, and live token/key rotation.

### Files

| Path                                                | Responsibility                                                             |
| :-------------------------------------------------- | :------------------------------------------------------------------------- |
| `cli/config/config.go`                              | Add `RelayKey`, `ConnectionMode`, `RelayURL`, `RelayAuthKey`; dual rotate  |
| `cli/pairing/payload.go`                            | Add `PayloadV2`, `PairingLink()` (`airvoice://pair?...`), terminal render  |
| `cli/server/handler.go`                             | Extract shared message processor so LAN WS & Relay client share paste flow |
| `cli/relayclient/client.go`                         | Outbound WS client to relay, E2E decrypt/encrypt, backoff reconnect        |
| `cli/main.go`                                       | Add `--relay`, `--relay-url`, and `--relay-auth-key` flags; keep bare LAN  |
| `desktop/app.go`                                    | Start/stop LAN vs Relay mode, expose `GetPairingLink` URI & relay settings |
| `desktop/frontend/src/components/SettingsPanel.vue` | Mode switch (`LAN / Relay`), Relay URL & optional Auth Key fields          |
| `desktop/frontend/src/components/QRCode.vue`        | `Relay · E2E` badge and "Copy Pairing Link" action                         |
| `desktop/frontend/src/components/StatusBadge.vue`   | Show LAN vs Relay·E2E and relay reconnecting states                        |
| `desktop/frontend/src/i18n/*`                       | Chinese and English copy for relay settings, status, and errors            |

### Tasks

- [ ] **2.1 Config & key persistence (`cli/config`)**: Add `relayKey`, `connectionMode`, `relayUrl`, `relayAuthKey` to `Settings` (preserving all fields on read-modify-write). Add `EnsureRelayKey` (lazy 32-byte key) and update `RotateToken` to rotate both `token` and `relayKey`; extend `WatchToken` to detect `relayKey` changes.
- [ ] **2.2 Pairing v2 & URI (`cli/pairing`)**: Add `PayloadV2` JSON marshaling, `BuildPairingURI` (`airvoice://pair?...`), and `PrintRelayPairing` terminal output showing QR, relay URL, and copyable pairing link.
- [ ] **2.3 Reusable paste/message core + Relay host client (`cli/relayclient`)**: Refactor inbound `hello`/`text`/`ping` processing out of `cli/server/handler.go` into a transport-agnostic processor; implement `relayclient.Client` connecting as `role=host` with E2E envelope wrapping, `peer_status` callbacks, quota error logging, and exponential-backoff reconnect.
- [ ] **2.4 CLI flags (`cli/main.go`)**: Support standard Go `flag` options `--relay` (bool, reusing `settings.json` `relayUrl` or defaulting to `wss://relay.airvoice.yuler.cc/ws`), `--relay-url <url>` (implicitly enables relay mode), and `--relay-auth-key <key>`. Skip local port binding in Relay mode.
- [ ] **2.5 Desktop GUI (`desktop/`)**: Wire Relay mode lifecycle into `App` (switching mode cleanly stops LAN server or Relay client and regenerates QR + pairing link), update Vue Settings/QR/Status components following `DESIGN.md`, and add i18n strings.
- [ ] **2.6 Verify**: Run `go test -race ./cli/... ./desktop/...` and `cd desktop/frontend && npm run build`; manually verify CLI and Desktop against `wss://relay.airvoice.yuler.cc/ws`.

---

## Phase 3: Mobile Clients (iOS & Android)

**Objective:** Add `v: 2` Relay + E2E support, pairing link paste, official vs custom relay confirmation, and `waitingForHost` UX across both iOS and Android with full test vector parity.

### Files

| Platform | Paths                                                                                                                                                                                                                                                                                                                 |
| :------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| iOS      | `ios/Airvoice/Models/PairingPayload.swift`, `ios/Airvoice/Services/RelayCrypto.swift`, `ConnectionManager.swift`, `KeychainTokenStore.swift`, `Views/QRScannerView.swift`, `Views/HomeView.swift`, `Views/SettingsSheetView.swift`, `Utilities/AppStrings.swift`, `ios/AirvoiceTests/*`                               |
| Android  | `android/app/src/main/java/cc/yuler/airvoice/models/PairingPayload.kt`, `services/RelayCrypto.kt`, `services/ConnectionManager.kt`, `services/StorageManager.kt`, `ui/screens/QRScannerScreen.kt`, `ui/screens/HomeScreen.kt`, `ui/screens/SettingsBottomSheet.kt`, `ui/i18n/AppStrings.kt`, `android/app/src/test/*` |

### Tasks

- [ ] **3.1 Cross-platform E2E crypto (`RelayCrypto.swift` & `RelayCrypto.kt`)**: Implement HKDF-SHA256 derivation, AES-256-GCM seal/open with AAD, and `SessionGuard` replay + `seq == 1` `hello` (`ts` ±5 min) gating. Copy `vectors.json` test cases into both XCTest and JUnit suites and assert byte-for-byte parity with Go.
- [ ] **3.2 Pairing v2 & URI parsing**: Extend `PairingPayload` on both platforms to parse `v: 1` (LAN), `v: 2` (Relay JSON), and `airvoice://pair?...` URIs. Persist `relayKey` and optional `relayAuthKey` via Keychain (iOS) and `TokenCipher` (Android).
- [ ] **3.3 Paste pairing link UI**: Add a bottom "Paste Pairing Link" button on both QR scanner screens (auto-reads clipboard; falls back to a paste dialog if clipboard is empty or not a valid payload).
- [ ] **3.4 ConnectionManager Relay transport & states**:
  - Send `X-Airvoice-Room`, `X-Airvoice-Role: client`, `X-Airvoice-Auth`, and optional `Authorization: Bearer <auth>`.
  - Handle `peer_status(online: false)` → `waitingForHost` (`等待电脑上线` / `Waiting for PC`, disabling send); on `peer_status(online: true)`, rotate client `sid`, reset `seq`, send encrypted `hello` (`seq == 1`, `ts`), and enter `connected` after validating the host's encrypted `hello` reply (`seq == 1`, `ts`).
  - Encrypt outgoing `hello`/`text`/`ping` with `keyC2H` and decrypt incoming frames with `keyH2C`.
  - Surface localized toasts for `rate_limited`, `quota_exceeded`, `message_too_large`.
- [ ] **3.5 Confirmation & Settings UI**: Connect directly to `relay.airvoice.yuler.cc` with `Relay · E2E` badge; show lightweight domain confirmation for custom relay hosts. Display `Relay (E2E): <host>` in Settings sheet.
- [ ] **3.6 Verify**:
  - Android: `cd android && ./gradlew testDebugUnitTest assembleDebug`
  - iOS: run `./scripts/ios-xcodegen.sh` and verify via GitHub Actions `ios-test` (or `xcodebuild test` on macOS).
  - Manual E2E test: phone on cellular 4G/5G dictating to PC over `wss://relay.airvoice.yuler.cc/ws`.

---

## Phase 4: Documentation, Website & Release Automation

**Objective:** Document Relay Mode and self-hosting for users, and automate `airvoice-relay` release artifacts.

### Files

| Path                              | Responsibility                                                        |
| :-------------------------------- | :-------------------------------------------------------------------- |
| `docs/architecture.md`            | Update system diagram, wire protocol v2, and E2E security section     |
| `README.md`                       | Add cross-LAN Relay quick start and self-hosting summary              |
| `www/src/content/docs/*`          | Add Relay Mode & Self-Hosting guide (zh + en)                         |
| `.github/workflows/release.yml`   | Build and attach `airvoice-relay-linux-{amd64,arm64}` release assets  |

### Tasks

- [ ] **4.1 Docs & Website**: Document `--relay` / `--relay-url`, Desktop mode toggle, official free relay quotas, privacy/E2E guarantees, and one-command Docker self-hosting (`deploy/relay/`).
- [ ] **4.2 Release CI**: Add `airvoice-relay` build targets (`linux/amd64`, `linux/arm64`) to `.github/workflows/release.yml` using existing tag-based version resolution (`scripts/resolve-version.sh`).
- [ ] **4.3 Verify**: Run `cd www && npm run build` and validate workflow YAML syntax.
