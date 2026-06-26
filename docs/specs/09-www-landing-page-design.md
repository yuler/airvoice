# Airvoice 官网设计文档

## 概述

为 Airvoice 项目构建静态官网，包含产品着陆页和文档站。使用 Astro 6 + Tailwind CSS 4 构建，部署到 GitHub Pages。

## 技术选型

| 层 | 选择 | 理由 |
|---|------|------|
| 框架 | Astro 6 | 零 JS 默认输出，岛屿架构按需交互 |
| 样式 | Tailwind CSS 4 | 实用优先，配合 DESIGN.md token 映射 |
| 内容 | Astro Content Collections | Markdown/MDX 驱动文档，类型安全 |
| i18n | astro-i18n-aut | 路径式多语言（`/en/...`, `/zh/...`） |
| 部署 | GitHub Pages | GitHub Actions 自动部署，免费 |

## 目录结构

```
www/
├── src/
│   ├── components/
│   │   ├── Hero.astro
│   │   ├── Features.astro
│   │   ├── HowItWorks.astro
│   │   ├── PlatformSupport.astro
│   │   ├── Download.astro
│   │   └── Footer.astro
│   ├── layouts/
│   │   ├── BaseLayout.astro
│   │   └── DocsLayout.astro
│   ├── content/
│   │   ├── docs/
│   │   │   ├── en/
│   │   │   │   ├── background.md
│   │   │   │   ├── architecture.md
│   │   │   │   ├── quick-start.md
│   │   │   │   └── platform-deps.md
│   │   │   └── zh/
│   │   │       ├── background.md
│   │   │       ├── architecture.md
│   │   │       ├── quick-start.md
│   │   │       └── platform-deps.md
│   │   └── config.ts
│   ├── pages/
│   │   ├── en/
│   │   │   ├── index.astro
│   │   │   └── docs/
│   │   │       └── [...slug].astro
│   │   └── zh/
│   │       ├── index.astro
│   │       └── docs/
│   │           └── [...slug].astro
│   └── styles/
│       └── global.css
├── public/
│   └── images/
├── astro.config.mjs
├── tailwind.config.mjs
└── package.json
```

## 着陆页板块

### 1. Hero

- Slogan: "Speak on your phone, type on your Mac or Linux PC — via LAN, no cloud."
- 副标题: "Airvoice is a bridge, not a speech engine."
- 主按钮: 安装命令（`mise install` 或 `go install`）
- 次按钮: GitHub 仓库链接
- 视觉: iOS 截图 + 终端 QR 码示意

### 2. Features

4 个核心卖点卡片：

| 卖点 | 图标 | 说明 |
|------|------|------|
| 局域网传输 | Wifi | 数据不离开本地网络 |
| 零云端 | Shield | 无需账号、无服务器、无订阅 |
| 跨平台 | Monitor | macOS + Linux (X11/Wayland) |
| 开源 | Code | MIT License，完全透明 |

### 3. How It Works

三步流程：

1. **安装 CLI** — `mise install` 或 `go install github.com/user/airvoice/cli@latest`
2. **扫码配对** — 运行 `airvoice serve`，iOS 扫描终端 QR 码
3. **说话打字** — iPhone 上用豆包/微信输入法口述，文字自动出现在电脑光标处

### 4. Platform Support

三个平台卡片，列出依赖：

- **macOS**: Accessibility 权限
- **Linux X11**: `xclip`, `xdotool`
- **Linux Wayland**: `wl-clipboard`, `ydotool`

### 5. Download / Install

多行安装命令：

```bash
# mise (推荐)
mise trust && mise install

# Go
go install github.com/user/airvoice/cli@latest

# GitHub Releases
# 链接到 releases 页面
```

### 6. Footer

- GitHub 链接
- License (MIT)
- 项目简介

## 文档页面

### 内容来源

从现有 `docs/` 移植，拆分为 4 篇：

| 文档 | 来源 | 内容 |
|------|------|------|
| Quick Start | README.md | 安装、运行、iOS 配对 |
| Background | docs/background.md | 动机、对比、设计决策 |
| Architecture | docs/architecture.md | 系统架构、协议、包结构 |
| Platform Deps | README.md §Platform Dependencies | 各平台依赖安装 |

### 文档布局

- 左侧导航栏（文档列表）
- 右侧目录（当前页 TOC）
- 上方语言切换（EN/ZH）
- 暗色模式切换

## 设计 Token 映射

复用 `DESIGN.md` 中定义的色彩/排版/间距，映射为 CSS 变量：

```css
:root {
  --color-bg: #ffffff;
  --color-bg-secondary: #fafafa;
  --color-text: #171717;
  --color-text-secondary: #666666;
  --color-text-muted: #888888;
  --color-accent: #006efe;
  --color-border: #eaeaea;
  --color-success: #28a948;
  --color-warning: #ffae00;
  --color-error: #fc0035;
  --radius-sm: 6px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-full: 9999px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
}

@media (prefers-color-scheme: dark) {
  :root {
    --color-bg: #000000;
    --color-bg-secondary: #0d0e15;
    --color-text: #ededed;
    --color-text-secondary: #a0a0a0;
    --color-text-muted: #666666;
    --color-border: #2e2e2e;
    --color-success: #00ac3a;
    --color-error: #e2162a;
  }
}
```

## 部署

### GitHub Actions

- 触发：push 到 `main` 分支的 `www/` 目录
- 构建：`npm run build`
- 部署：推送到 `gh-pages` 分支

### 环境变量

- `SITE`: `https://<username>.github.io/airvoice` 或自定义域名

## 范围外

- 博客系统
- 搜索功能
- 用户认证
- 动态内容
- 自定义域名（初始阶段）
