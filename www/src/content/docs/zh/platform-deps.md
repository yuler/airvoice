---
title: 平台依赖
description: 各平台的系统要求。
order: 4
---

# 平台依赖

## macOS

- **辅助功能 API 权限**：运行二进制文件的终端应用程序需要辅助功能权限。
  - 前往：*系统设置 -> 隐私与安全性 -> 辅助功能*。
  - 添加并启用你的终端应用程序（如 Terminal、iTerm2、ghostty）。

## Linux (X11)

需要 `xclip` 用于剪贴板访问，`xdotool` 用于键盘模拟：

```bash
sudo apt install xclip xdotool
```

## Linux (Wayland)

需要 `wl-clipboard` 用于剪贴板访问，`ydotool` 用于键盘模拟：

```bash
sudo apt install wl-clipboard ydotool
```

确保 `ydotoold` 服务已启用并运行：

```bash
systemctl --user enable --now ydotoold
```
