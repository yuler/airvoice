---
title: 平台依赖
description: 各平台的系统要求。
order: 5
---

## macOS

- **辅助功能权限**：运行二进制文件的终端应用需要辅助功能权限。
  - 前往：*系统设置 → 隐私与安全性 → 辅助功能*。
  - 添加并启用你的终端应用（如 Terminal、iTerm2、ghostty）。

## Linux（X11）

需要 `xclip` 访问剪贴板，`xdotool` 模拟键盘输入：

```bash
sudo apt install xclip xdotool
```

## Linux（Wayland）

需要 `wl-clipboard` 访问剪贴板，`ydotool` 模拟键盘输入：

```bash
sudo apt install wl-clipboard ydotool
```

确保 `ydotoold` 服务已启用并运行：

```bash
systemctl --user enable --now ydotoold
```

## Windows

CLI 可通过 `go build` 在 Windows 上编译。WebSocket 服务器正常工作，但剪贴板粘贴和按键注入尚未实现。

当前状态：**实验性，未经测试**。欢迎贡献。
