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

### 1. Build the Go CLI
Compile the desktop server application:
```bash
go build -o bin/airvoice ./cli
```

### 2. Run the Server
Start the server process:
```bash
./bin/airvoice serve
```
This will start the local server and print a QR code in your terminal.

### 3. Connect the iOS Client
1. Open the iOS application (`ios/Airvoice.xcodeproj`).
2. Run the application on an iOS device connected to the same local network (Wi-Fi) as your computer.
3. Scan the terminal QR code using the iOS application to establish the connection.

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
