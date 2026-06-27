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
| 构建 iOS（模拟器） | `mise run ios:dev` |
| 构建 iOS（真机） | `mise run ios:build` |
| 文档开发服务器 | `mise run www:dev` |
| 构建文档 | `mise run www:build` |

## iOS 开发

在 Xcode 中打开项目：

```bash
open ios/Airvoice.xcodeproj
```

项目使用 [XcodeGen](https://github.com/yonaskolb/XcodeGen) 生成 `project.yml`。修改项目设置后，编辑 `ios/project.yml` 并重新生成：

```bash
xcodegen generate
```

### 代码签名

真机构建需要创建签名配置：

```bash
cp ios/Signing.xcconfig.example ios/Signing.xcconfig
```

编辑 `ios/Signing.xcconfig`，将 `DEVELOPMENT_TEAM` 设置为你的 Apple Team ID（在 Xcode → 设置 → 账户中查看）。
