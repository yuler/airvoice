# Airvoice

Speak on your phone, type on your Mac or Linux PC — via LAN, no cloud.

**Airvoice is a bridge, not a speech engine.** Use dictation on your iOS device; the desktop CLI acts as a server that receives and types the recognized text at your cursor.

---

## Documentation

Explore the detailed documentation to understand the project better:

*   **[Background](docs/background.md)**: Project motivation, comparisons, and core decisions.
*   **[Architecture](docs/architecture.md)**: System design, repository structure, and wire protocol.
*   **[MVP Build Plan](docs/plans/00-mvp-plan.md)**: Detailed step-by-step implementation plan.

---

## Quick Start

This project uses [mise](https://mise.jdx.dev/) to pin **Go**, **Swift**, and **gum**, and to run common dev tasks.

### 1. Setup

```bash
mise trust            # first time in this repo
mise install          # install pinned tools
mise run setup        # build CLI + check platform deps
```

### 2. Run the Server

```bash
mise run dev
```

This builds `bin/airvoice` and starts the server, printing a QR code in your terminal.

For an interactive menu (build, test, iOS device, etc.):

```bash
mise run menu
```

### 3. Desktop GUI (Optional)

For a native GUI instead of CLI:

```bash
cd desktop
wails dev
```

Or build a production binary:

```bash
cd desktop && wails build
./build/bin/Airvoice.app/Contents/MacOS/Airvoice
```

### 4. Connect the iOS Client

**Simulator or Xcode:** open `ios/Airvoice.xcodeproj` and run on a device/simulator.

**Physical device (macOS):**

```bash
mise run ios:device   # gum picker → build → install on USB-connected iPhone
```

Then on the same Wi‑Fi:

1. Run `mise run dev` on your Mac.
2. Open Airvoice on the iPhone and scan the terminal QR code.

### mise tasks

| Task | Description |
|------|-------------|
| `mise run setup` | Install tools, check deps, build CLI |
| `mise run dev` | Build + start server (default) |
| `mise run menu` | Interactive gum menu |
| `mise run build` | Build Go CLI |
| `mise run test` | `go test ./cli/...` |
| `mise run serve` | Alias for `dev` |
| `mise run ios:device` | Build & install on physical iOS device (macOS) |

Scripts live in `scripts/` (gum UI); `mise.toml` wires them as tasks.

---

## Platform Dependencies

The desktop application simulates keyboard input and interacts with the clipboard. Make sure the dependencies for your platform are met:

### macOS
*   **Accessibility API Permission**: The terminal application running the binary (or the compiled binary itself) requires Accessibility permission.
    *   Go to: *System Settings -> Privacy & Security -> Accessibility*.
    *   Add and enable your terminal application (e.g., Terminal, iTerm2, ghostty) or the compiled binary.

### Linux (X11)
Requires `xclip` for clipboard access and `xdotool` for keyboard emulation:
```bash
sudo apt install xclip xdotool
```

### Linux (Wayland)
Requires `wl-clipboard` for clipboard access and `ydotool` for keyboard emulation:
*   Install the dependencies (e.g., `sudo apt install wl-clipboard ydotool`).
*   Ensure the `ydotoold` service is enabled and running, and your user has permission to write to `/dev/uinput` (or run `ydotool` with appropriate privileges).

---

## License

TBD
