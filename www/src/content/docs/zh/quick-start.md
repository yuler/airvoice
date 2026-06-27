---
title: 快速开始
description: 安装 Airvoice 并开始使用。
order: 2
---

Airvoice 由两部分组成：**桌面端服务器**（CLI）和 **iOS 客户端**。两者必须在同一 Wi‑Fi 下。

## 1. 安装 CLI

如果还没有安装 `mise`：

```bash
curl https://mise.run | sh
```

然后安装 Airvoice：

```bash
mise trust
mise install
```

或者直接用 Go 安装：

```bash
go install github.com/yuler/airvoice/cli@latest
```

## 2. 启动服务器

```bash
airvoice serve
```

终端会打印一个二维码。保持此进程运行。

## 3. 安装 iOS 应用

我们没有 Apple 开发者账号，所以你需要通过 Mac 从源代码构建。这是免费的 — 只需要一个普通的 Apple ID。

**前提条件：**

- 安装了 **Xcode 15+** 的 Mac（Mac App Store 免费下载）
- 在 Xcode 中登录了 Apple ID
- 运行 **iOS 17+** 的 iPhone
- USB 数据线

### 第一步 — 在 iPhone 上开启开发者模式

在 iPhone 上：**设置 → 隐私与安全性 → 开发者模式** → 打开 → 提示时重启。

> 没看到选项？先把 iPhone 连接到打开了 Xcode 的 Mac。

### 第二步 — 获取源代码

```bash
git clone https://github.com/yuler/airvoice.git
cd airvoice
```

### 第三步 — 在 Xcode 中打开

```bash
open ios/Airvoice.xcodeproj
```

### 第四步 — 配置签名

在 Xcode 中：

1. 选择 **Airvoice** 项目 → **Airvoice** Target
2. 在 **Signing & Capabilities** 下：勾选 **Automatically manage signing**
3. **Team** 选择你的 Apple ID（如果没有登录会提示）

### 第五步 — 构建并安装

1. USB 连接 iPhone，解锁，点击 **信任此电脑**
2. 在 Xcode 顶部选择你的 iPhone
3. 按 `Cmd + R` 构建并安装

### 第六步 — 信任开发者证书

首次在 iPhone 上：**设置 → 通用 → VPN 与设备管理** → 找到你的 Apple ID → 点击 **信任**。

### 第七步 — 连接

1. 确保 iPhone 和 Mac 在同一 Wi‑Fi
2. 在 iPhone 上打开 Airvoice
3. 扫描终端中的二维码

完成。在手机上说话，文字出现在电脑光标处。
