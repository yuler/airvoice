---
title: 快速开始
description: 几分钟内启动并运行 Airvoice。
order: 1
---

# 快速开始

## 1. 安装

```bash
mise trust            # 首次在此仓库运行
mise install          # 安装固定版本的工具
mise run setup        # 构建 CLI + 检查平台依赖
```

## 2. 运行服务器

```bash
mise run dev
```

此命令构建 `bin/airvoice` 并启动服务器，在终端中打印 QR 码。

## 3. 连接 iOS 客户端

**模拟器或 Xcode：** 打开 `ios/Airvoice.xcodeproj`，在设备/模拟器上运行。

**物理设备（macOS）：**

```bash
mise run ios:device   # gum 选择器 → 构建 → 安装到 USB 连接的 iPhone
```

然后在同一 Wi-Fi 下：

1. 在 Mac 上运行 `mise run dev`。
2. 在 iPhone 上打开 Airvoice 并扫描终端中的 QR 码。

## mise 任务

| 任务 | 描述 |
|------|------|
| `mise run setup` | 安装工具、检查依赖、构建 CLI |
| `mise run dev` | 构建 + 启动服务器（默认） |
| `mise run menu` | 交互式 gum 菜单 |
| `mise run build` | 构建 Go CLI |
| `mise run test` | `go test ./cli/...` |
| `mise run serve` | `dev` 的别名 |
| `mise run ios:device` | 构建并安装到物理 iOS 设备（macOS） |
