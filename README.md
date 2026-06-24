# Airvoice

Speak on your phone, type on your Mac or Linux PC — via LAN, no cloud.

**Airvoice is a bridge, not a speech engine.** Use 豆包 or 微信输入法 on iOS for recognition; the desktop CLI pastes text at your cursor.

## Docs

| Document | Description |
|----------|-------------|
| [Background](docs/background.md) | Why Airvoice exists, competitors, decisions |
| [Architecture](docs/architecture.md) | System design, repo layout, wire protocol |
| [MVP build plan](docs/plans/00-mvp-plan.md) | Oneshot implementation spec for agents |

## Quick start (after implementation)

```bash
# Desktop
go build -o bin/airvoice ./cli
./bin/airvoice serve
# Scan terminal QR with iOS app

# iOS
# Open ios/Airvoice.xcodeproj → run on device (same Wi‑Fi)
```

## Status

Pre-MVP — spec approved, implementation pending. See [00-mvp-plan.md](docs/plans/00-mvp-plan.md).

## License

TBD
