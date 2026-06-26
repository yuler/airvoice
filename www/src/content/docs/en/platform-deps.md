---
title: Platform Dependencies
description: System requirements for each platform.
order: 4
---

# Platform Dependencies

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
