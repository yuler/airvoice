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

需要 Windows PowerShell 5.x（系统自带 `powershell`）：

- **剪贴板**：`Set-Clipboard`（通过 PowerShell）
- **按键**：`WScript.Shell` SendKeys 模拟 Ctrl+V（发送到当前前台窗口）

运行前请确保目标输入框已获得焦点。若按键注入失败，文本仍会写入剪贴板，可手动按 Ctrl+V。
