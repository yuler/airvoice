# Airvoice 官网实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建 Airvoice 静态官网（着陆页 + 文档站），使用 Astro 6 + Tailwind CSS 4，部署到 GitHub Pages。

**Architecture:** Astro 6 输出零 JS 静态页面，Tailwind CSS 4 处理样式，Content Collections 管理文档，astro-i18n-aut 实现中英双语，GitHub Actions 部署到 GH Pages。

**Tech Stack:** Astro 6, Tailwind CSS 4, TypeScript, astro-i18n-aut

## Global Constraints

- 项目位于仓库根目录 `www/`
- 所有设计 Token 遵循 `DESIGN.md` 色彩/排版/间距规范
- 中英双语：路径式 `/en/...`, `/zh/...`
- 输出零客户端 JS（纯静态 HTML/CSS）
- Node.js >= 18

---

## File Structure

```
www/
├── package.json
├── astro.config.mjs
├── tsconfig.json
├── tailwind.config.mjs
├── src/
│   ├── styles/
│   │   └── global.css              # DESIGN.md token → CSS 变量 + Tailwind
│   ├── layouts/
│   │   ├── BaseLayout.astro        # HTML shell, meta, lang switch
│   │   └── DocsLayout.astro        # 文档页布局（侧边栏 + TOC）
│   ├── components/
│   │   ├── Hero.astro
│   │   ├── Features.astro
│   │   ├── HowItWorks.astro
│   │   ├── PlatformSupport.astro
│   │   ├── Download.astro
│   │   ├── Footer.astro
│   │   ├── LangSwitch.astro
│   │   └── ThemeToggle.astro
│   ├── content/
│   │   ├── config.ts               # Content Collection schema
│   │   └── docs/
│   │       ├── en/
│   │       │   ├── quick-start.md
│   │       │   ├── background.md
│   │       │   ├── architecture.md
│   │       │   └── platform-deps.md
│   │       └── zh/
│   │           ├── quick-start.md
│   │           ├── background.md
│   │           ├── architecture.md
│   │           └── platform-deps.md
│   └── pages/
│       ├── en/
│       │   ├── index.astro         # 着陆页
│       │   └── docs/
│       │       └── [...slug].astro # 文档动态路由
│       └── zh/
│           ├── index.astro
│           └── docs/
│               └── [...slug].astro
└── public/
    └── images/                     # 截图、logo 等静态资源
```

---

### Task 1: 初始化 Astro 6 项目

**Files:**
- Create: `www/package.json`
- Create: `www/astro.config.mjs`
- Create: `www/tsconfig.json`

- [ ] **Step 1: 创建 www 目录并初始化**

```bash
mkdir -p www
cd www
npm init -y
```

- [ ] **Step 2: 安装 Astro 6 及依赖**

```bash
npm install astro@latest @astrojs/tailwind@latest tailwindcss@latest
npm install astro-i18n-aut
```

- [ ] **Step 3: 创建 astro.config.mjs**

```js
import { defineConfig } from 'astro/config';
import tailwind from '@astrojs/tailwind';
import { i18n, filterSitemap } from 'astro-i18n-aut/integration';

const defaultLocale = 'en';
const locales = {
  en: 'en',
  zh: 'zh',
};

export default defineConfig({
  site: 'https://airvoice.dev',
  integrations: [
    tailwind(),
    i18n({
      locales,
      defaultLocale,
      redirectDefaultLocale: false,
    }),
  ],
  trailingSlash: 'always',
});
```

- [ ] **Step 4: 创建 tsconfig.json**

```json
{
  "extends": "astro/tsconfigs/strict"
}
```

- [ ] **Step 5: 更新 package.json scripts**

```json
{
  "scripts": {
    "dev": "astro dev",
    "build": "astro build",
    "preview": "astro preview"
  }
}
```

- [ ] **Step 6: 验证项目初始化**

```bash
cd www && npm run dev
```

Expected: Astro dev server 启动在 `http://localhost:4321`

- [ ] **Step 7: Commit**

```bash
git add www/
git commit -m "chore: init astro 6 project in www/"
```

---

### Task 2: 配置 Tailwind CSS 4 + 设计 Token

**Files:**
- Create: `www/tailwind.config.mjs`
- Create: `www/src/styles/global.css`

- [ ] **Step 1: 创建 tailwind.config.mjs**

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  darkMode: 'media',
  theme: {
    extend: {
      colors: {
        bg: {
          DEFAULT: 'var(--color-bg)',
          secondary: 'var(--color-bg-secondary)',
        },
        text: {
          DEFAULT: 'var(--color-text)',
          secondary: 'var(--color-text-secondary)',
          muted: 'var(--color-text-muted)',
        },
        accent: 'var(--color-accent)',
        border: 'var(--color-border)',
        success: 'var(--color-success)',
        warning: 'var(--color-warning)',
        error: 'var(--color-error)',
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
        full: 'var(--radius-full)',
      },
      spacing: {
        sm: 'var(--spacing-sm)',
        md: 'var(--spacing-md)',
        lg: 'var(--spacing-lg)',
      },
    },
  },
  plugins: [],
};
```

- [ ] **Step 2: 创建 global.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

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

@layer base {
  body {
    @apply bg-bg text-text;
  }
}
```

- [ ] **Step 3: 验证 Tailwind 生效**

在 `www/src/pages/index.astro` 中添加测试内容：

```astro
---
---
<html>
<body>
  <h1 class="text-4xl text-accent font-bold">Airvoice</h1>
</body>
</html>
```

```bash
cd www && npm run dev
```

Expected: 页面显示蓝色 "Airvoice" 标题

- [ ] **Step 4: Commit**

```bash
git add www/tailwind.config.mjs www/src/styles/global.css
git commit -m "feat: configure tailwind with design tokens"
```

---

### Task 3: 创建基础布局 + i18n

**Files:**
- Create: `www/src/layouts/BaseLayout.astro`
- Create: `www/src/components/LangSwitch.astro`
- Create: `www/src/components/ThemeToggle.astro`

- [ ] **Step 1: 创建 BaseLayout.astro**

```astro
---
import '../styles/global.css';

interface Props {
  title: string;
  description: string;
  lang: 'en' | 'zh';
}

const { title, description, lang } = Astro.props;
const currentPath = Astro.url.pathname;
const otherLang = lang === 'en' ? 'zh' : 'en';
const otherPath = currentPath.replace(`/${lang}/`, `/${otherLang}/`);
---

<!doctype html>
<html lang={lang}>
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="description" content={description} />
    <title>{title}</title>
  </head>
  <body class="min-h-screen bg-bg text-text">
    <nav class="fixed top-0 w-full bg-bg/80 backdrop-blur-sm border-b border-border z-50">
      <div class="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
        <a href={`/${lang}/`} class="text-xl font-bold">Airvoice</a>
        <div class="flex items-center gap-4">
          <a href={`/${lang}/docs/quick-start/`} class="text-sm text-text-secondary hover:text-text">
            {lang === 'en' ? 'Docs' : '文档'}
          </a>
          <a href={otherPath} class="text-sm text-text-secondary hover:text-text">
            {otherLang.toUpperCase()}
          </a>
        </div>
      </div>
    </nav>
    <main class="pt-16">
      <slot />
    </main>
  </body>
</html>
```

- [ ] **Step 2: 验证布局**

在 `www/src/pages/en/index.astro` 中使用布局：

```astro
---
import BaseLayout from '../../layouts/BaseLayout.astro';
---

<BaseLayout title="Airvoice" description="Speak on your phone, type on your PC" lang="en">
  <h1 class="text-4xl font-bold text-center py-20">Airvoice</h1>
</BaseLayout>
```

```bash
cd www && npm run dev
```

Expected: 页面显示导航栏和标题，导航栏有 Docs 和语言切换链接

- [ ] **Step 3: Commit**

```bash
git add www/src/layouts/BaseLayout.astro www/src/pages/
git commit -m "feat: add base layout with nav and i18n"
```

---

### Task 4: 创建着陆页组件 — Hero + Features

**Files:**
- Create: `www/src/components/Hero.astro`
- Create: `www/src/components/Features.astro`

- [ ] **Step 1: 创建 Hero.astro**

```astro
---
const lang = Astro.props.lang as 'en' | 'zh';
---

<section class="relative py-24 md:py-32">
  <div class="max-w-6xl mx-auto px-6 text-center">
    <h1 class="text-4xl md:text-6xl font-bold tracking-tight mb-6">
      {lang === 'en' ? (
        <>Speak on your phone,<br />type on your PC</>
      ) : (
        <>手机说话，<br />电脑打字</>
      )}
    </h1>
    <p class="text-lg md:text-xl text-text-secondary max-w-2xl mx-auto mb-10">
      {lang === 'en'
        ? 'Airvoice is a LAN bridge — no cloud, no accounts, no data leaves your network.'
        : 'Airvoice 是局域网语音桥接工具 — 无云端、无账号、数据不出本地网络。'}
    </p>
    <div class="flex flex-col sm:flex-row gap-4 justify-center">
      <code class="inline-block bg-bg-secondary border border-border rounded-md px-6 py-3 font-mono text-sm">
        mise install
      </code>
      <a
        href="https://github.com/anthropics/airvoice"
        class="inline-flex items-center justify-center px-6 py-3 border border-border rounded-full text-sm hover:bg-bg-secondary transition-colors"
      >
        GitHub →
      </a>
    </div>
  </div>
</section>
```

- [ ] **Step 2: 创建 Features.astro**

```astro
---
const lang = Astro.props.lang as 'en' | 'zh';

const features = lang === 'en' ? [
  { icon: '📡', title: 'LAN Only', desc: 'Data never leaves your local network. No cloud, no servers.' },
  { icon: '🔒', title: 'Zero Cloud', desc: 'No accounts, no subscriptions, no tracking.' },
  { icon: '💻', title: 'Cross-Platform', desc: 'macOS + Linux (X11 and Wayland).' },
  { icon: '📖', title: 'Open Source', desc: 'MIT License. Fully transparent.' },
] : [
  { icon: '📡', title: '纯局域网', desc: '数据不离开本地网络，无云端中转。' },
  { icon: '🔒', title: '零云端', desc: '无需账号、无订阅、无追踪。' },
  { icon: '💻', title: '跨平台', desc: '支持 macOS 和 Linux（X11 / Wayland）。' },
  { icon: '📖', title: '开源', desc: 'MIT License，完全透明。' },
];
---

<section class="py-20 bg-bg-secondary">
  <div class="max-w-6xl mx-auto px-6">
    <h2 class="text-3xl font-bold text-center mb-12">
      {lang === 'en' ? 'Why Airvoice?' : '为什么选择 Airvoice？'}
    </h2>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
      {features.map(f => (
        <div class="bg-bg border border-border rounded-lg p-6">
          <div class="text-3xl mb-4">{f.icon}</div>
          <h3 class="font-semibold mb-2">{f.title}</h3>
          <p class="text-sm text-text-secondary">{f.desc}</p>
        </div>
      ))}
    </div>
  </div>
</section>
```

- [ ] **Step 3: 在着陆页中组装组件**

更新 `www/src/pages/en/index.astro`：

```astro
---
import BaseLayout from '../../layouts/BaseLayout.astro';
import Hero from '../../components/Hero.astro';
import Features from '../../components/Features.astro';
---

<BaseLayout
  title="Airvoice — Speak on your phone, type on your PC"
  description="LAN voice bridge for Mac and Linux. No cloud, no accounts."
  lang="en"
>
  <Hero lang="en" />
  <Features lang="en" />
</BaseLayout>
```

- [ ] **Step 4: 验证**

```bash
cd www && npm run dev
```

Expected: 着陆页显示 Hero 区域和 4 个特性卡片

- [ ] **Step 5: Commit**

```bash
git add www/src/components/Hero.astro www/src/components/Features.astro www/src/pages/
git commit -m "feat: add hero and features components"
```

---

### Task 5: 创建着陆页组件 — HowItWorks + PlatformSupport

**Files:**
- Create: `www/src/components/HowItWorks.astro`
- Create: `www/src/components/PlatformSupport.astro`

- [ ] **Step 1: 创建 HowItWorks.astro**

```astro
---
const lang = Astro.props.lang as 'en' | 'zh';

const steps = lang === 'en' ? [
  { num: '01', title: 'Install CLI', desc: 'Run mise install or go install to get the airvoice binary.' },
  { num: '02', title: 'Scan QR', desc: 'Run airvoice serve, scan the QR code with your iPhone.' },
  { num: '03', title: 'Speak & Type', desc: 'Dictate on your phone using any voice keyboard. Text appears at your cursor.' },
] : [
  { num: '01', title: '安装 CLI', desc: '运行 mise install 或 go install 获取 airvoice 二进制文件。' },
  { num: '02', title: '扫码配对', desc: '运行 airvoice serve，用 iPhone 扫描终端中的 QR 码。' },
  { num: '03', title: '说话打字', desc: '在手机上用任意语音键盘口述，文字自动出现在电脑光标处。' },
];
---

<section class="py-20">
  <div class="max-w-6xl mx-auto px-6">
    <h2 class="text-3xl font-bold text-center mb-16">
      {lang === 'en' ? 'How It Works' : '工作原理'}
    </h2>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-12">
      {steps.map(s => (
        <div class="text-center">
          <div class="text-5xl font-bold text-accent mb-4">{s.num}</div>
          <h3 class="text-xl font-semibold mb-2">{s.title}</h3>
          <p class="text-text-secondary">{s.desc}</p>
        </div>
      ))}
    </div>
  </div>
</section>
```

- [ ] **Step 2: 创建 PlatformSupport.astro**

```astro
---
const lang = Astro.props.lang as 'en' | 'zh';

const platforms = lang === 'en' ? [
  { name: 'macOS', deps: ['Accessibility permission for Terminal'], cmd: null },
  { name: 'Linux X11', deps: ['xclip', 'xdotool'], cmd: 'sudo apt install xclip xdotool' },
  { name: 'Linux Wayland', deps: ['wl-clipboard', 'ydotool'], cmd: 'sudo apt install wl-clipboard ydotool' },
] : [
  { name: 'macOS', deps: ['终端的辅助功能权限'], cmd: null },
  { name: 'Linux X11', deps: ['xclip', 'xdotool'], cmd: 'sudo apt install xclip xdotool' },
  { name: 'Linux Wayland', deps: ['wl-clipboard', 'ydotool'], cmd: 'sudo apt install wl-clipboard ydotool' },
];
---

<section class="py-20 bg-bg-secondary">
  <div class="max-w-6xl mx-auto px-6">
    <h2 class="text-3xl font-bold text-center mb-12">
      {lang === 'en' ? 'Platform Support' : '平台支持'}
    </h2>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      {platforms.map(p => (
        <div class="bg-bg border border-border rounded-lg p-6">
          <h3 class="text-xl font-semibold mb-4">{p.name}</h3>
          <ul class="text-sm text-text-secondary space-y-2 mb-4">
            {p.deps.map(d => <li>• {d}</li>)}
          </ul>
          {p.cmd && (
            <code class="block bg-bg-secondary border border-border rounded-md p-3 text-xs font-mono">
              {p.cmd}
            </code>
          )}
        </div>
      ))}
    </div>
  </div>
</section>
```

- [ ] **Step 3: 更新着陆页组装**

更新 `www/src/pages/en/index.astro`，添加新组件：

```astro
---
import BaseLayout from '../../layouts/BaseLayout.astro';
import Hero from '../../components/Hero.astro';
import Features from '../../components/Features.astro';
import HowItWorks from '../../components/HowItWorks.astro';
import PlatformSupport from '../../components/PlatformSupport.astro';
---

<BaseLayout
  title="Airvoice — Speak on your phone, type on your PC"
  description="LAN voice bridge for Mac and Linux. No cloud, no accounts."
  lang="en"
>
  <Hero lang="en" />
  <Features lang="en" />
  <HowItWorks lang="en" />
  <PlatformSupport lang="en" />
</BaseLayout>
```

- [ ] **Step 4: 验证**

```bash
cd www && npm run dev
```

Expected: 着陆页显示全部 4 个板块

- [ ] **Step 5: Commit**

```bash
git add www/src/components/ www/src/pages/
git commit -m "feat: add how-it-works and platform support components"
```

---

### Task 6: 创建着陆页组件 — Download + Footer

**Files:**
- Create: `www/src/components/Download.astro`
- Create: `www/src/components/Footer.astro`

- [ ] **Step 1: 创建 Download.astro**

```astro
---
const lang = Astro.props.lang as 'en' | 'zh';
---

<section class="py-20">
  <div class="max-w-6xl mx-auto px-6">
    <h2 class="text-3xl font-bold text-center mb-12">
      {lang === 'en' ? 'Get Started' : '开始使用'}
    </h2>
    <div class="max-w-2xl mx-auto space-y-6">
      <div class="bg-bg-secondary border border-border rounded-lg p-6">
        <h3 class="font-semibold mb-3">mise (recommended)</h3>
        <code class="block bg-bg border border-border rounded-md p-3 text-sm font-mono">
          mise trust && mise install
        </code>
      </div>
      <div class="bg-bg-secondary border border-border rounded-lg p-6">
        <h3 class="font-semibold mb-3">Go</h3>
        <code class="block bg-bg border border-border rounded-md p-3 text-sm font-mono">
          go install github.com/anthropics/airvoice/cli@latest
        </code>
      </div>
      <div class="text-center">
        <a
          href="https://github.com/anthropics/airvoice/releases"
          class="text-accent hover:underline text-sm"
        >
          {lang === 'en' ? 'All releases on GitHub →' : '所有版本见 GitHub →'}
        </a>
      </div>
    </div>
  </div>
</section>
```

- [ ] **Step 2: 创建 Footer.astro**

```astro
---
const lang = Astro.props.lang as 'en' | 'zh';
---

<footer class="py-12 border-t border-border">
  <div class="max-w-6xl mx-auto px-6 flex flex-col md:flex-row items-center justify-between gap-4">
    <div class="text-sm text-text-muted">
      © 2024 Airvoice. MIT License.
    </div>
    <div class="flex gap-6">
      <a href="https://github.com/anthropics/airvoice" class="text-sm text-text-secondary hover:text-text">
        GitHub
      </a>
      <a href={`/${lang}/docs/quick-start/`} class="text-sm text-text-secondary hover:text-text">
        {lang === 'en' ? 'Docs' : '文档'}
      </a>
    </div>
  </div>
</footer>
```

- [ ] **Step 3: 更新着陆页组装全部组件**

更新 `www/src/pages/en/index.astro`：

```astro
---
import BaseLayout from '../../layouts/BaseLayout.astro';
import Hero from '../../components/Hero.astro';
import Features from '../../components/Features.astro';
import HowItWorks from '../../components/HowItWorks.astro';
import PlatformSupport from '../../components/PlatformSupport.astro';
import Download from '../../components/Download.astro';
import Footer from '../../components/Footer.astro';
---

<BaseLayout
  title="Airvoice — Speak on your phone, type on your PC"
  description="LAN voice bridge for Mac and Linux. No cloud, no accounts."
  lang="en"
>
  <Hero lang="en" />
  <Features lang="en" />
  <HowItWorks lang="en" />
  <PlatformSupport lang="en" />
  <Download lang="en" />
  <Footer lang="en" />
</BaseLayout>
```

- [ ] **Step 4: 验证**

```bash
cd www && npm run dev
```

Expected: 完整着陆页，包含所有 6 个板块

- [ ] **Step 5: Commit**

```bash
git add www/src/components/ www/src/pages/
git commit -m "feat: add download and footer components"
```

---

### Task 7: 创建中文着陆页

**Files:**
- Create: `www/src/pages/zh/index.astro`

- [ ] **Step 1: 创建中文着陆页**

```astro
---
import BaseLayout from '../../layouts/BaseLayout.astro';
import Hero from '../../components/Hero.astro';
import Features from '../../components/Features.astro';
import HowItWorks from '../../components/HowItWorks.astro';
import PlatformSupport from '../../components/PlatformSupport.astro';
import Download from '../../components/Download.astro';
import Footer from '../../components/Footer.astro';
---

<BaseLayout
  title="Airvoice — 手机说话，电脑打字"
  description="局域网语音桥接工具，支持 Mac 和 Linux。无云端、无账号。"
  lang="zh"
>
  <Hero lang="zh" />
  <Features lang="zh" />
  <HowItWorks lang="zh" />
  <PlatformSupport lang="zh" />
  <Download lang="zh" />
  <Footer lang="zh" />
</BaseLayout>
```

- [ ] **Step 2: 验证中英文切换**

```bash
cd www && npm run dev
```

Expected: 访问 `/en/` 和 `/zh/` 分别显示英文和中文着陆页，导航栏语言切换链接正常

- [ ] **Step 3: Commit**

```bash
git add www/src/pages/zh/
git commit -m "feat: add chinese landing page"
```

---

### Task 8: 设置 Content Collections + 移植文档

**Files:**
- Create: `www/src/content/config.ts`
- Create: `www/src/content/docs/en/quick-start.md`
- Create: `www/src/content/docs/en/background.md`
- Create: `www/src/content/docs/en/architecture.md`
- Create: `www/src/content/docs/en/platform-deps.md`
- Create: `www/src/content/docs/zh/quick-start.md`
- Create: `www/src/content/docs/zh/background.md`
- Create: `www/src/content/docs/zh/architecture.md`
- Create: `www/src/content/docs/zh/platform-deps.md`

- [ ] **Step 1: 创建 Content Collection schema**

```ts
import { defineCollection, z } from 'astro:content';

const docs = defineCollection({
  type: 'content',
  schema: z.object({
    title: z.string(),
    description: z.string(),
    order: z.number().optional(),
  }),
});

export const collections = { docs };
```

- [ ] **Step 2: 创建英文文档 — quick-start.md**

从 `README.md` 移植 Quick Start 内容：

```markdown
---
title: Quick Start
description: Get Airvoice up and running in minutes.
order: 1
---

# Quick Start

## 1. Setup

```bash
mise trust            # first time in this repo
mise install          # install pinned tools
mise run setup        # build CLI + check platform deps
```

## 2. Run the Server

```bash
mise run dev
```

This builds `bin/airvoice` and starts the server, printing a QR code in your terminal.

## 3. Connect the iOS Client

**Simulator or Xcode:** open `ios/Airvoice.xcodeproj` and run on a device/simulator.

**Physical device (macOS):**

```bash
mise run ios:device   # gum picker → build → install on USB-connected iPhone
```

Then on the same Wi‑Fi:

1. Run `mise run dev` on your Mac.
2. Open Airvoice on the iPhone and scan the terminal QR code.

## mise Tasks

| Task | Description |
|------|-------------|
| `mise run setup` | Install tools, check deps, build CLI |
| `mise run dev` | Build + start server (default) |
| `mise run menu` | Interactive gum menu |
| `mise run build` | Build Go CLI |
| `mise run test` | `go test ./cli/...` |
| `mise run serve` | Alias for `dev` |
| `mise run ios:device` | Build & install on physical iOS device (macOS) |
```

- [ ] **Step 3: 创建英文文档 — background.md**

从 `docs/background.md` 移植，保持原内容结构。

- [ ] **Step 4: 创建英文文档 — architecture.md**

从 `docs/architecture.md` 移植，保持原内容结构。

- [ ] **Step 5: 创建英文文档 — platform-deps.md**

从 `README.md` Platform Dependencies 部分移植：

```markdown
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
```

- [ ] **Step 6: 创建中文文档**

为每篇英文文档创建对应的中文翻译版本，放在 `docs/zh/` 目录下。翻译标题和描述，技术内容保持原样。

- [ ] **Step 7: Commit**

```bash
git add www/src/content/
git commit -m "docs: port docs to content collections (en + zh)"
```

---

### Task 9: 创建文档布局 + 文档页面

**Files:**
- Create: `www/src/layouts/DocsLayout.astro`
- Create: `www/src/pages/en/docs/[...slug].astro`
- Create: `www/src/pages/zh/docs/[...slug].astro`

- [ ] **Step 1: 创建 DocsLayout.astro**

```astro
---
import '../../styles/global.css';
import BaseLayout from './BaseLayout.astro';

interface Props {
  title: string;
  description: string;
  lang: 'en' | 'zh';
  headings: { depth: number; slug: string; text: string }[];
}

const { title, description, lang, headings } = Astro.props;

const docsNav = lang === 'en' ? [
  { href: '/en/docs/quick-start/', label: 'Quick Start' },
  { href: '/en/docs/background/', label: 'Background' },
  { href: '/en/docs/architecture/', label: 'Architecture' },
  { href: '/en/docs/platform-deps/', label: 'Platform Dependencies' },
] : [
  { href: '/zh/docs/quick-start/', label: '快速开始' },
  { href: '/zh/docs/background/', label: '背景' },
  { href: '/zh/docs/architecture/', label: '架构' },
  { href: '/zh/docs/platform-deps/', label: '平台依赖' },
];
---

<BaseLayout title={title} description={description} lang={lang}>
  <div class="max-w-6xl mx-auto px-6 py-12 flex gap-12">
    <aside class="w-64 shrink-0 hidden lg:block">
      <nav class="sticky top-24 space-y-1">
        {docsNav.map(link => (
          <a
            href={link.href}
            class:list={[
              'block px-4 py-2 text-sm rounded-md transition-colors',
              Astro.url.pathname === link.href
                ? 'bg-bg-secondary text-text font-medium'
                : 'text-text-secondary hover:text-text hover:bg-bg-secondary',
            ]}
          >
            {link.label}
          </a>
        ))}
      </nav>
    </aside>
    <article class="prose prose-neutral dark:prose-invert max-w-3xl">
      <slot />
    </article>
    {headings.length > 0 && (
      <aside class="w-48 shrink-0 hidden xl:block">
        <nav class="sticky top-24 space-y-1">
          <p class="text-xs font-semibold text-text-muted uppercase mb-3">
            {lang === 'en' ? 'On this page' : '本页目录'}
          </p>
          {headings.filter(h => h.depth <= 3).map(h => (
            <a
              href={`#${h.slug}`}
              class:list={[
                'block text-sm text-text-secondary hover:text-text transition-colors',
                h.depth === 3 && 'pl-4',
              ]}
            >
              {h.text}
            </a>
          ))}
        </nav>
      </aside>
    )}
  </div>
</BaseLayout>
```

- [ ] **Step 2: 创建英文文档动态路由**

```astro
---
import { getCollection } from 'astro:content';
import DocsLayout from '../../../layouts/DocsLayout.astro';

export async function getStaticPaths() {
  const docs = await getCollection('docs', ({ id }) => id.startsWith('en/'));
  return docs.map(doc => ({
    params: { slug: doc.id.replace('en/', '') },
    props: { doc },
  }));
}

const { doc } = Astro.props;
const { Content, headings } = await doc.render();
---

<DocsLayout
  title={doc.data.title}
  description={doc.data.description}
  lang="en"
  headings={headings}
>
  <Content />
</DocsLayout>
```

- [ ] **Step 3: 创建中文文档动态路由**

同上，将 `en/` 替换为 `zh/`。

- [ ] **Step 4: 验证文档页面**

```bash
cd www && npm run dev
```

Expected: 访问 `/en/docs/quick-start/` 显示文档页面，有侧边栏导航和 TOC

- [ ] **Step 5: Commit**

```bash
git add www/src/layouts/DocsLayout.astro www/src/pages/
git commit -m "feat: add docs layout and dynamic routes"
```

---

### Task 10: 配置 GitHub Pages 部署

**Files:**
- Create: `.github/workflows/deploy-www.yml`

- [ ] **Step 1: 创建 GitHub Actions workflow**

```yaml
name: Deploy to GitHub Pages

on:
  push:
    branches: [main]
    paths: ['www/**']
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: true

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
      - run: npm ci
        working-directory: www
      - run: npm run build
        working-directory: www
      - uses: actions/upload-pages-artifact@v3
        with:
          path: www/dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 2: 更新 astro.config.mjs site 配置**

根据实际仓库名更新 `site` 字段：

```js
site: 'https://<username>.github.io/airvoice',
```

- [ ] **Step 3: 验证构建**

```bash
cd www && npm run build
```

Expected: 构建成功，`www/dist/` 目录生成静态文件

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/deploy-www.yml www/astro.config.mjs
git commit -m "ci: add github pages deployment for www"
```

---

### Task 11: 最终验证

- [ ] **Step 1: 本地预览完整站点**

```bash
cd www && npm run build && npm run preview
```

- [ ] **Step 2: 检查所有页面**

- [ ] `/en/` — 英文着陆页完整
- [ ] `/zh/` — 中文着陆页完整
- [ ] `/en/docs/quick-start/` — 英文文档
- [ ] `/zh/docs/quick-start/` — 中文文档
- [ ] 语言切换链接正常
- [ ] 暗色模式正常

- [ ] **Step 3: 检查构建输出**

```bash
ls -la www/dist/
```

Expected: 生成 `en/` 和 `zh/` 目录，每个目录下有 `index.html` 和 `docs/` 子目录

- [ ] **Step 4: 最终 Commit**

```bash
git add www/
git commit -m "feat: complete airvoice website with landing page and docs"
```
