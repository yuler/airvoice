---
title: 开发指南
description: 搭建开发环境并参与贡献。
order: 3
---

## 前提条件

- [mise](https://mise.run) — 工具版本管理器
- Git

`mise` 会自动安装正确版本的 Go、Node 和其他工具。

## 初始化

```bash
git clone https://github.com/yuler/airvoice.git
cd airvoice
mise trust
mise install
```

## 项目结构

```
airvoice/
├── cli/          # Go CLI — WebSocket 服务器 + 按键注入
├── desktop/      # 桌面应用（Wails）
├── android/      # Android Kotlin/Compose 应用
├── ios/          # iOS SwiftUI 应用
├── www/          # 文档站点（Astro）
├── scripts/      # 构建和开发脚本
├── mise.toml     # 任务运行器 + 工具版本
└── go.mod
```

## 常用任务

| 任务 | 命令 |
|------|------|
| 构建 CLI | `mise run cli:build` |
| 运行服务器 | `mise run cli:dev` |
| 运行测试 | `mise run cli:test` |
| 构建 Android（调试 APK） | `mise run android:build` |
| 安装 Android（USB） | `mise run android:install` |
| 构建 iOS（模拟器） | `mise run ios:dev` |
| 构建 iOS（真机） | `mise run ios:build` |
| 文档开发服务器 | `mise run www:dev` |
| 构建文档 | `mise run www:build` |
| 升版本 | `mise bump` |

## Android 开发

需要 JDK（Android Studio 自带的 JBR 即可）。构建调试版 APK：

```bash
mise run android:build
```

APK 输出路径：`android/app/build/outputs/apk/debug/app-debug.apk`。正式版会发布到 [GitHub Releases](https://github.com/yuler/airvoice/releases/latest)，文件名为 `airvoice-android-*.apk`。

## iOS 开发

在 Xcode 中打开项目：

```bash
open ios/Airvoice.xcodeproj
```

项目使用 [XcodeGen](https://github.com/yonaskolb/XcodeGen) 生成 `project.yml`。修改项目设置后，编辑 `ios/project.yml`，用 `./scripts/ios-xcodegen.sh` 重新生成（会从 `AIRVOICE_VERSION` 或最新 `v*` Git Tag 导出 `APP_VERSION` / `APP_BUILD`）。不要单独跑 `xcodegen`。

```bash
./scripts/ios-xcodegen.sh
```

### 代码签名

真机构建需要创建签名配置：

```bash
cp ios/Signing.xcconfig.example ios/Signing.xcconfig
```

编辑 `ios/Signing.xcconfig`，将 `DEVELOPMENT_TEAM` 设置为你的 Apple Team ID（在 Xcode → 设置 → 账户中查看）。

## 版本与发布

产品版本以 **Git Tag（`vX.Y.Z`）** 为唯一来源，各端在构建时通过 `AIRVOICE_VERSION` 环境变量或最新 `v*` Tag 自动解析，不要在 `package.json` 或源码里硬编码版本号。

代码推送或 PR 合并到 `main` 且 CI 通过后，GitHub Actions 会自动打新 Tag 并发布：默认升级 `patch`；提交信息或 PR 标题包含 `[minor]` / `[major]` 时对应升级；`HEAD` 提交包含 `[skip release]` 时跳过发版。也可以通过 GitHub Actions 的 `Release` 工作流手动触发，或本地运行 `mise bump` 手动打 Tag。

iOS `CFBundleVersion` 和 Android `versionCode` 由 `major * 1000000 + minor * 1000 + patch` 算出。
