# 20-persistent-pairing-token-design

## Description

Pairing QR codes currently encode a new UUID on every CLI process and every desktop launch. Phones that already saved `ws` + `token` must scan again after a host restart.

This design keeps **one pairing token per machine** in the existing local config file, shows that token on CLI and desktop, and lets the user rotate it when they want. After a refresh, any connected phone is disconnected and must pair again.

The QR payload and WebSocket auth query do not change: `{ "v": 1, "ws": "ws://…/ws", "token": "uuid" }` and `GET /ws?token=…`.

---

## Proposed Solution

### Single config file

Path: `~/.airvoice/settings.json` (directory mode `0700`, file mode `0600`). No `pairing.json`.

```json
{
  "token": "<uuid>",
  "port": 7655,
  "autoStart": false,
  "language": "zh-CN"
}
```

- **`token`**: shared pairing secret for CLI and desktop on this machine. Load-or-create: if the file is missing or `token` is empty/whitespace, mint a UUID and write it back without wiping other fields.
- **`port` / `autoStart` / `language`**: desktop settings as today. CLI does **not** use `port`; it listens on `--port` / `-p` (default `7654`).
- Every write is read-modify-write so a token refresh cannot drop desktop fields, and `SaveSettings` cannot drop `token`.

Source of truth is the file. In-memory `Server.SetToken` is a cache of the last successfully loaded token.

### File watch

CLI (while serving) and desktop poll `settings.json` about once a second and compare the `token` field. No new filesystem-watch dependency. When the **token string** changes:

1. `SetToken(newToken)`
2. `DisconnectClients()`

Writes that only change `port` / `autoStart` / `language` must **not** kick the phone.

The process that performs `token refresh` writes the file, then applies SetToken + disconnect itself. Other running hosts apply via the watcher. If the newly loaded token equals the in-memory token, skip disconnect.

### CLI command surface

There is no `serve` subcommand. Bare `airvoice` starts the server.

```text
airvoice [--port 7654]
airvoice settings
airvoice token refresh
airvoice doctor
airvoice version
```

| Invocation | Behavior |
|---|---|
| `airvoice` / `airvoice --port 7654` | Load-or-create token, print QR + `Token:` + WebSocket URL, listen. Token is **not** rotated on start. |
| `airvoice settings` | Load-or-create, then print `settings.json` (pretty JSON) to stdout. No server required. |
| `airvoice token refresh` | Mint a new UUID, persist (preserving other fields), print the new token. Running `airvoice` or desktop pick it up via watch and disconnect the phone. |
| `airvoice doctor` / `airvoice version` | Unchanged. |

Unknown first arguments that look like flags (`-p`, `--port`) belong to the default server command. Known subcommands are `settings`, `token`, `doctor`, `version`. `token` requires `refresh`; other `token` args print usage and exit `1`. A leftover `airvoice serve` is an unknown command (usage, exit `1`), not an alias.

### Desktop

- Under the QR, show the token as selectable/copyable text, taken from the existing pairing payload (`GetPairingLink` JSON `token` field — same UUID the QR encodes).
- Refresh control remains. It is allowed **while connected**: persist new token, SetToken, disconnect. Then reload QR + token label.
- Remove the “cannot refresh pairing while connected” guard.
- `SaveSettings` must preserve `token`. Startup loads token from the same file instead of `uuid.New()` every launch.

### Phone

No protocol or client change. Saved connection reconnects until the host token is refreshed; then the old token gets `401` and the user scans again.

### Shared module

Add package `cli/config` used by CLI and desktop:

- Config path (`$HOME/.airvoice/settings.json`)
- Load / save / load-or-create token
- Read-modify-write token vs other settings fields
- `WatchToken(ctx, onChange)` — poll ~1s, call `onChange` only when `token` changes

Desktop `App.loadSettings` / `SaveSettings` go through this helper so CLI and GUI cannot diverge on path or merge rules.

### Error handling

- **Unreadable file on first start:** mint a token and write a valid file; copy over any JSON fields that did parse.
- **Unreadable file after start:** keep last good in-memory token; log; watcher retries.
- **Unwritable file on refresh:** fail the CLI command / desktop UI; do **not** `SetToken` in memory (disk stays source of truth).
- **Watcher failure:** log and retry; keep serving with the last loaded token.
- **Stale phone token:** existing `401 Unauthorized`.

---

## Architectural & Code Changes

### Files to add

- [`cli/config/`](../../cli/config/): helper + tests (path, load-or-create, preserve fields, token-change poll).

### Files to modify

- [`cli/main.go`](../../cli/main.go): default command = server; add `settings` and `token refresh`; remove `serve` as the required subcommand.
- [`cli/pairing/`](../../cli/pairing/): persist token; `PrintPairingWithToken` stays for QR/text output.
- [`cli/server/server.go`](../../cli/server/server.go): already has `SetToken` / `DisconnectClients`; no auth protocol change.
- [`desktop/app.go`](../../desktop/app.go): load token from settings; `RefreshPairing` always allowed; persist token; start watcher.
- [`desktop/frontend/src/components/QRCode.vue`](../../desktop/frontend/src/components/QRCode.vue): show token text under the QR.
- Docs that say `airvoice serve` (README, www quick-start, architecture): default invocation is `airvoice`.

### Out of scope

- TLS on LAN
- Short pairing codes instead of UUID
- Unifying CLI and desktop listen ports
- Mobile UI / payload version bump
- Keeping `airvoice serve` as a documented command

---

## Verification Plan

### Automated tests

- Load-or-create: missing file writes a token; existing token is reused; empty token is replaced.
- Read-modify-write: token refresh preserves `port` / `language` / `autoStart`; settings save preserves `token`.
- Token poll: token change fires callback; language-only write does not.
- Server: after SetToken, old query token is rejected; new token connects; disconnect is invoked on refresh.
- Desktop `RefreshPairing` succeeds while status is `connected` and persists the new token.
- CLI against temp `HOME`: default command still serves (tested via flag parse / wiring where practical); `settings` prints JSON including `token`; `token refresh` changes token on disk.

Run `mise x -- go test ./...` (plus existing desktop/frontend tests if they cover QR bindings).

### Manual verification

1. Delete `~/.airvoice/settings.json`. Run `airvoice`. Confirm QR + Token line. Run `airvoice settings` in another terminal; token matches. Restart `airvoice`; token is unchanged; phone reconnects without scanning.
2. With CLI serving and a phone connected, run `airvoice token refresh`. Phone drops. New QR/token on the serving process (after watch). Old token cannot reconnect; scan works.
3. Desktop: token visible under QR; refresh while connected kicks the phone and updates QR + label. CLI `airvoice settings` shows the same token.
4. Change desktop language/port in settings; phone stays connected. Token field still present in the file.
5. `airvoice doctor` and `airvoice version` still work. `airvoice serve` is not required (and is not listed in usage).
