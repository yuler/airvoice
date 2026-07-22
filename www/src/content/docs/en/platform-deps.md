---
title: Platform Dependencies
description: System requirements for each platform.
order: 5
---

## Android

- **Camera permission**: Required for QR code scanning during pairing.
- **Install from APK**: Download from [GitHub Releases](https://github.com/yuler/airvoice/releases/latest). Allow installation from unknown sources if prompted.

## Desktop (Linux AppImage)

Prefer the `Airvoice-Desktop-*-Linux-x86_64.AppImage` from [GitHub Releases](https://github.com/yuler/airvoice/releases/latest). It bundles WebKitGTK, so you do **not** need system `webkit2gtk-4.0` / `webkit2gtk-4.1` for the Desktop app.

Make it executable and run:

```bash
chmod +x Airvoice-Desktop-*-Linux-x86_64.AppImage
./Airvoice-Desktop-*-Linux-x86_64.AppImage
```

Clipboard / paste automation still needs the X11 or Wayland tools below when using Desktop or CLI.

## macOS

- **Accessibility API Permission**: The terminal application running the binary requires Accessibility permission.
  - Go to: *System Settings -> Privacy & Security -> Accessibility*.
  - Add and enable your terminal application (e.g., Terminal, iTerm2, ghostty).

## Linux (X11)

Requires `xclip` for clipboard access and `xdotool` for keyboard emulation:

```bash
sudo apt install xclip xdotool
```

## Linux (Wayland)

Requires `wl-clipboard` for clipboard access and `ydotool` for keyboard emulation:

```bash
sudo apt install wl-clipboard ydotool
```

Ensure the `ydotoold` service is enabled and running:

```bash
systemctl --user enable --now ydotoold
```

## Windows

Requires Windows PowerShell 5.x (built-in `powershell`):

- **Clipboard**: `Set-Clipboard` (via PowerShell)
- **Keystroke**: `WScript.Shell` SendKeys simulates Ctrl+V (sent to the foreground window)

Ensure the target text field has focus before sending. If keystroke injection fails, text remains on the clipboard — press Ctrl+V manually.

Current status: **experimental, untested**. Contributions welcome.
