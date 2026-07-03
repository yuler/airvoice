# Airvoice www UI Upgrade — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the Airvoice landing page and docs site UI to a premium, high-contrast design aligned with DESIGN.md.

**Architecture:** Modify existing React components and Astro layouts in `www/src/`. Add a new `downloads.ts` utility and `DocsPagination.tsx` component. No new routes or dependencies.

**Tech Stack:** Astro 5, React 18, Tailwind CSS v4, TypeScript — all already installed.

## Global Constraints

- Follow `DESIGN.md` tokens exactly: `#006efe` accent, `#000000`/`#ffffff` surfaces, `#eaeaea`/`#2e2e2e` borders, no gradients
- No new npm dependencies
- All files in `www/src/`
- Latest release tag: `v0.3.1` (read from `VERSION` file at root)
- All inline styles use CSS custom properties (`var(--...)`) except for the Hero dark-island and CTA-band which hardcode `#000000`
- `rounded-full` (9999px) for all pill buttons; `rounded-xl` (12px) for cards

---

## Task 1: Platform-aware Download Utility

**Files:**
- Create: `www/src/lib/downloads.ts`

**Interfaces:**
- Produces: `getDownloadUrls(): { cli: string; desktop: string; mobile: string }` — called client-side (React useEffect)

- [ ] Create `www/src/lib/downloads.ts`:

```typescript
const TAG = 'v0.3.1';
const BASE = `https://github.com/yuler/airvoice/releases/download/${TAG}`;
const RELEASES = 'https://github.com/yuler/airvoice/releases/latest';

export interface DownloadUrls {
  cli: string;
  desktop: string;
  mobile: string;
}

export function getDownloadUrls(): DownloadUrls {
  if (typeof navigator === 'undefined') {
    return { cli: RELEASES, desktop: RELEASES, mobile: RELEASES };
  }

  const ua = navigator.userAgent;
  const platform = (navigator as Navigator & { userAgentData?: { platform: string } })
    .userAgentData?.platform ?? navigator.platform ?? '';

  const isWindows = /Win/i.test(platform) || /Windows/i.test(ua);
  const isMac = /Mac/i.test(platform) || /Macintosh/i.test(ua);
  const isLinux = /Linux/i.test(platform) && !/Android/i.test(ua);
  const isAndroid = /Android/i.test(ua);
  const isIOS = /iPhone|iPad|iPod/i.test(ua);
  const isARM = /arm64|aarch64/i.test(platform) || /arm64/i.test(ua);

  let cli = RELEASES;
  let desktop = RELEASES;
  let mobile = RELEASES;

  if (isWindows) {
    cli = `${BASE}/airvoice-cli-windows-amd64.exe`;
    desktop = `${BASE}/Airvoice-Desktop-${TAG}-Windows.zip`;
  } else if (isMac) {
    cli = isARM
      ? `${BASE}/airvoice-cli-darwin-arm64`
      : `${BASE}/airvoice-cli-darwin-amd64`;
    desktop = `${BASE}/Airvoice-Desktop-${TAG}-macOS.zip`;
  } else if (isLinux) {
    cli = `${BASE}/airvoice-cli-linux-amd64`;
    desktop = `${BASE}/Airvoice-Desktop-${TAG}-Linux.tar.gz`;
  }

  if (isAndroid) {
    mobile = `${BASE}/airvoice-android-${TAG}.apk`;
  } else if (isIOS) {
    mobile = 'https://github.com/yuler/airvoice';
  }

  return { cli, desktop, mobile };
}
```

- [ ] Commit:
```bash
git add www/src/lib/downloads.ts
git commit -m "feat(www): add platform-aware download URL utility"
```

---

## Task 2: Header — Add Text Labels & CLI Link

**Files:**
- Modify: `www/src/components/react/Header.tsx`

**Changes:**
- Add "CLI" nav link pointing to releases page
- Add text labels next to icons in desktop nav: "Docs", "GitHub", "CLI"
- Height stays 52px

- [ ] In `Header.tsx`, replace the desktop nav links section (the `<div className="hidden md:flex ...">` block) with:

```tsx
<div className="hidden md:flex items-center gap-5">
  <a
    href={loc('docs/background/')}
    className={`flex items-center gap-1.5 text-sm transition-colors ${active === 'docs' ? 'font-medium' : 'hover:opacity-80'}`}
    style={{ color: active === 'docs' ? 'var(--primary-text)' : 'var(--secondary-text)' }}
  >
    <BookIcon />
    {isZh ? '文档' : 'Docs'}
  </a>
  <a
    href="https://github.com/yuler/airvoice/releases/latest"
    target="_blank"
    rel="noopener noreferrer"
    className="flex items-center gap-1.5 text-sm transition-colors hover:opacity-80"
    style={{ color: 'var(--secondary-text)' }}
  >
    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="7 10 12 15 17 10" />
      <line x1="12" y1="15" x2="12" y2="3" />
    </svg>
    CLI
  </a>
  <a
    href="https://github.com/yuler/airvoice"
    target="_blank"
    rel="noopener noreferrer"
    className="flex items-center gap-1.5 text-sm transition-colors hover:opacity-80"
    style={{ color: 'var(--secondary-text)' }}
  >
    <GithubIcon />
    GitHub
  </a>

  {/* Lang Toggle */}
  <a
    href={getLangTogglePath()}
    className="flex items-center justify-center text-sm transition-colors hover:opacity-80 cursor-pointer"
    style={{ color: 'var(--secondary-text)' }}
    title={isZh ? 'Switch to English' : '切换至中文'}
  >
    {isZh ? '🇺🇸' : '🇨🇳'}
  </a>

  {/* Theme Toggle */}
  <button
    onClick={toggleTheme}
    className="flex items-center justify-center transition-colors hover:opacity-80 cursor-pointer bg-transparent border-none outline-none p-0"
    style={{ color: 'var(--secondary-text)' }}
    title={isZh ? '切换主题' : 'Toggle Theme'}
  >
    {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
  </button>
</div>
```

- [ ] Also update the mobile dropdown to include CLI link:

In the mobile dropdown `<div>`, add after the GitHub link:
```tsx
<a
  href="https://github.com/yuler/airvoice/releases/latest"
  target="_blank"
  rel="noopener noreferrer"
  className="flex items-center gap-2 text-sm transition-colors py-1 hover:opacity-80"
  style={{ color: 'var(--secondary-text)' }}
  onClick={() => setMobileOpen(false)}
>
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
    <polyline points="7 10 12 15 17 10" />
    <line x1="12" y1="15" x2="12" y2="3" />
  </svg>
  CLI
</a>
```

- [ ] Commit:
```bash
git add www/src/components/react/Header.tsx
git commit -m "feat(www): add text labels and CLI link to header nav"
```

---

## Task 3: Hero — Dark Island, Bigger Headline, Dark Terminal, Platform Links

**Files:**
- Modify: `www/src/components/react/Hero.tsx`

**Changes:**
1. `<section>` gets `background: '#000000'` hardcoded (dark island, always dark)
2. All text colors use dark-mode token values directly as hex strings
3. `h1` → `text-6xl` desktop, `text-5xl` mobile
4. ALPHA badge → bordered pill
5. `TerminalWindow` → dark theme (black bg, `#2e2e2e` border)
6. Status indicators adapt: use dark mode hex values since parent is always dark
7. CTA buttons use `getDownloadUrls()` via `useEffect` for platform links
8. Scroll reveal: `useScrollReveal` hook added, applied to left column and visuals

- [ ] Replace the entire `Hero.tsx` file with:

```tsx
import { useState, useEffect, useRef } from 'react';
import Lightbox from './Lightbox';
import { getDownloadUrls, type DownloadUrls } from '../../lib/downloads';

interface HeroProps {
  lang: 'en' | 'zh';
  base: string;
}

function useScrollReveal(threshold = 0.15) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.style.opacity = '0';
    el.style.transform = 'translateY(16px)';
    el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
    const obs = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.style.opacity = '1';
          el.style.transform = 'translateY(0)';
          obs.disconnect();
        }
      },
      { threshold }
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [threshold]);
  return ref;
}

function TerminalWindow() {
  return (
    <div
      className="rounded-2xl overflow-hidden w-[260px] sm:w-[270px]"
      style={{ background: '#000000', border: '1px solid #2e2e2e' }}
    >
      <div className="flex items-center gap-1.5 px-4 py-3 border-b" style={{ background: '#0d0e15', borderColor: '#2e2e2e' }}>
        <span className="w-2.5 h-2.5 rounded-full" style={{ background: '#ff5f57' }} />
        <span className="w-2.5 h-2.5 rounded-full" style={{ background: '#febc2e' }} />
        <span className="w-2.5 h-2.5 rounded-full" style={{ background: '#28c840' }} />
        <span className="ml-2 text-[10px] font-mono" style={{ color: '#666666' }}>airvoice</span>
      </div>
      <div className="p-4 flex flex-col items-start">
        <div className="mb-4">
          <img
            src={`${import.meta.env.BASE_URL || '/'}qrcode.svg`.replace(/\/+/g, '/')}
            alt="https://github.com/yuler/airvoice"
            className="w-[200px] h-[200px] block"
            style={{ filter: 'invert(1)' }}
          />
        </div>
        <div className="w-full font-mono text-[9px] leading-relaxed space-y-1 text-left">
          <div><span style={{ color: '#666666' }}>Token:</span>{' '}<span style={{ color: '#ededed' }}>277129e4-35ea-40af-a122-13a5839e5e1f</span></div>
          <div><span style={{ color: '#666666' }}>WebSocket URL:</span>{' '}<span style={{ color: '#ededed' }}>ws://192.168.20.189:7654/ws</span></div>
          <div><span style={{ color: '#666666' }}>Status:</span>{' '}<span style={{ color: '#00ac3a' }}>Connected</span></div>
          <div><span style={{ color: '#666666' }}>Device:</span>{' '}<span style={{ color: '#ededed' }}>Pixel 8 Pro</span></div>
          <div className="pt-2" style={{ color: '#666666' }}>[airvoice] waiting for phone connection...</div>
        </div>
      </div>
    </div>
  );
}

function DesktopWindow({ base }: { base: string }) {
  return (
    <div className="rounded-2xl overflow-hidden" style={{ width: '170px', border: '1px solid #2e2e2e' }}>
      <Lightbox
        src={`${base || '/'}desktop.png`.replace(/\/+/g, '/')}
        alt="Airvoice Desktop Client"
        className="w-full h-auto block"
      />
    </div>
  );
}

function MobilePhone({ base }: { base: string }) {
  return (
    <div className="rounded-2xl overflow-hidden" style={{ width: '170px', border: '1px solid #2e2e2e' }}>
      <Lightbox
        src={`${base || '/'}phone.jpg`.replace(/\/+/g, '/')}
        alt="Airvoice App Screenshot"
        className="w-full h-auto block"
      />
    </div>
  );
}

function MobileHeroVisual({ activeTab, base }: { activeTab: 'cli' | 'desktop'; base: string }) {
  const leftItemLeft = activeTab === 'cli' ? 'left-[-12px]' : 'left-[15px]';
  return (
    <div className="relative mx-auto h-[360px] w-[310px]">
      <div className={`absolute ${leftItemLeft} top-8 z-10 scale-[0.75] origin-top-left transition-all duration-300`}>
        {activeTab === 'cli' ? <TerminalWindow /> : <DesktopWindow base={base} />}
      </div>
      <div className="absolute right-[-12px] top-[20px] z-20 scale-[0.8] origin-top-right transition-all duration-300">
        <MobilePhone base={base} />
      </div>
    </div>
  );
}

function DesktopHeroVisual({ activeTab, base }: { activeTab: 'cli' | 'desktop'; base: string }) {
  const leftItemLeft = activeTab === 'cli' ? 'left-[10px]' : 'left-[60px]';
  const phoneLeft = activeTab === 'cli' ? 'left-[320px]' : 'left-[290px]';
  return (
    <div className="relative flex items-center justify-center h-[420px] w-full max-w-[540px] mx-auto">
      <div className={`absolute ${leftItemLeft} top-12 z-10 transition-all duration-300`}>
        {activeTab === 'cli' ? <TerminalWindow /> : <DesktopWindow base={base} />}
      </div>
      <div className={`absolute ${phoneLeft} top-4 z-20 transition-all duration-300`}>
        <MobilePhone base={base} />
      </div>
    </div>
  );
}

function DownloadIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="7 10 12 15 17 10" />
      <line x1="12" y1="15" x2="12" y2="3" />
    </svg>
  );
}

function MonitorIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <path d="M8 21h8M12 17v4" />
    </svg>
  );
}

export default function Hero({ lang, base }: HeroProps) {
  const [activeTab, setActiveTab] = useState<'cli' | 'desktop'>('cli');
  const [urls, setUrls] = useState<DownloadUrls>({
    cli: 'https://github.com/yuler/airvoice/releases/latest',
    desktop: 'https://github.com/yuler/airvoice/releases/latest',
    mobile: 'https://github.com/yuler/airvoice/releases/latest',
  });
  const isZh = lang === 'zh';

  useEffect(() => {
    setUrls(getDownloadUrls());
  }, []);

  const leftRef = useScrollReveal(0.1);
  const rightRef = useScrollReveal(0.1);

  const statuses = [
    { color: '#00ac3a', label: isZh ? '已连接' : 'Connected', pulse: true },
    { color: '#ffae00', label: isZh ? '连接中...' : 'Connecting...' },
    { color: '#e2162a', label: isZh ? '错误' : 'Error' },
    { color: '#8f8f8f', label: isZh ? '离线' : 'Offline' },
  ];

  return (
    <section style={{ background: '#000000' }} className="py-14 md:py-20 lg:py-24">
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <div className="flex flex-col gap-10 lg:flex-row lg:items-center lg:gap-16">

          {/* Left column */}
          <div ref={leftRef} className="flex-1 min-w-0">
            {/* ALPHA badge */}
            <div className="mb-5">
              <span style={{
                border: '1px solid #006efe',
                color: '#006efe',
                borderRadius: '9999px',
                padding: '3px 10px',
                fontSize: '11px',
                fontWeight: 700,
                letterSpacing: '0.08em',
                display: 'inline-block',
              }}>
                ALPHA
              </span>
            </div>

            {/* Headline */}
            <h1
              className="text-5xl font-bold leading-tight md:text-6xl"
              style={{ color: '#ededed', letterSpacing: '-0.04em' }}
            >
              {isZh ? (
                <>与你的设备沟通。<br />简单直接。</>
              ) : (
                <>Talk to your<br />devices. Simply.</>
              )}
            </h1>

            <p className="mt-5 text-base leading-relaxed" style={{ color: '#a0a0a0', maxWidth: '440px' }}>
              {isZh
                ? 'Airvoice 是跨设备通信的统一桥梁。我们提供简洁的移动端应用与强大的命令行 CLI，同时也为不习惯命令行的用户准备了直观易用的桌面客户端。'
                : 'Airvoice is a unified bridge for cross-device communication, offering a clean mobile app, a powerful CLI, and a sleek desktop client for those who prefer a GUI.'}
            </p>

            {/* CTA buttons */}
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <a
                href={urls.cli}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 rounded-full px-6 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90"
                style={{ background: '#006efe', height: '48px' }}
              >
                <DownloadIcon />
                {isZh ? '下载 CLI' : 'Download CLI'}
              </a>
              <a
                href={urls.desktop}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 rounded-full px-6 py-3 text-sm font-semibold transition-colors hover:bg-white/10"
                style={{ border: '1px solid #2e2e2e', color: '#ededed', height: '44px' }}
              >
                <MonitorIcon />
                {isZh ? '下载桌面版' : 'Download Desktop'}
              </a>
            </div>

            {/* Status indicators */}
            <div className="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2">
              {statuses.map((s) => (
                <span key={s.label} className="flex items-center gap-1.5 text-sm" style={{ color: '#a0a0a0' }}>
                  {s.pulse
                    ? <span className="av-pulse-dot" style={{ background: s.color }} />
                    : <span className="w-2 h-2 rounded-full flex-shrink-0" style={{ background: s.color }} />
                  }
                  {s.label}
                </span>
              ))}
            </div>
          </div>

          {/* Right column */}
          <div ref={rightRef} className="flex-1 min-w-0 flex flex-col items-center justify-center">
            {/* Tab switcher */}
            <div className="w-full max-w-[540px] flex justify-center lg:justify-end mb-8">
              <div className="inline-flex rounded-full p-1" style={{ background: '#0d0e15', border: '1px solid #2e2e2e' }}>
                {(['cli', 'desktop'] as const).map((tab) => (
                  <button
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className="rounded-full px-5 py-2 text-xs font-bold transition-all duration-200 cursor-pointer border-none outline-none"
                    style={
                      activeTab === tab
                        ? { backgroundColor: '#1a1a1a', color: '#ededed' }
                        : { backgroundColor: 'transparent', color: '#666666' }
                    }
                  >
                    {tab === 'cli' ? (isZh ? '命令行 CLI' : 'CLI') : (isZh ? '桌面端 Desktop' : 'Desktop')}
                  </button>
                ))}
              </div>
            </div>

            <div className="hidden lg:block w-full">
              <DesktopHeroVisual activeTab={activeTab} base={base} />
            </div>
            <div className="lg:hidden">
              <MobileHeroVisual activeTab={activeTab} base={base} />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
```

- [ ] Commit:
```bash
git add www/src/components/react/Hero.tsx
git commit -m "feat(www): dark island hero with bigger headline and platform download links"
```

---

## Task 4: Features — Bento Grid Layout

**Files:**
- Modify: `www/src/components/react/Features.tsx`

**Changes:**
- Desktop: 3-col × 2-row bento grid via `grid-template-areas`
- Easy Pairing: tall card (row-span 2)
- Real-Time Sync: featured card (slightly larger, `background-secondary`)
- Mobile carousel: unchanged

- [ ] Replace the desktop grid section in `Features.tsx`. The `features` array and mobile carousel remain unchanged. Only the desktop grid block changes:

Replace `{/* Desktop View: Grid layout */}` block with:

```tsx
{/* Desktop View: Bento Grid */}
<div
  className="hidden sm:grid gap-4"
  style={{
    gridTemplateColumns: '1fr 1.6fr 1fr',
    gridTemplateRows: 'auto auto',
    gridTemplateAreas: `
      "pairing sync secure"
      "pairing cli desktop"
    `,
  }}
>
  {/* Easy Pairing — tall */}
  <div
    className="rounded-xl p-6 flex flex-col"
    style={{
      gridArea: 'pairing',
      background: 'var(--background-primary)',
      border: '1px solid var(--border-default)',
    }}
  >
    <div className="mb-4 inline-flex items-center justify-center rounded-xl p-2.5 self-start"
      style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
      {features[0].icon}
    </div>
    <h3 className="mb-2 text-base font-semibold" style={{ color: 'var(--primary-text)' }}>{features[0].title}</h3>
    <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[0].desc}</p>
  </div>

  {/* Real-Time Sync — featured */}
  <div
    className="rounded-xl p-7 flex flex-col"
    style={{
      gridArea: 'sync',
      background: 'var(--background-secondary)',
      border: '1px solid var(--border-default)',
    }}
  >
    <div className="mb-4 inline-flex items-center justify-center rounded-xl p-3 self-start"
      style={{ background: 'var(--background-primary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
      <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" aria-hidden="true">
        <line x1="4" y1="10" x2="4" y2="14" />
        <line x1="8" y1="7" x2="8" y2="17" />
        <line x1="12" y1="4" x2="12" y2="20" />
        <line x1="16" y1="7" x2="16" y2="17" />
        <line x1="20" y1="10" x2="20" y2="14" />
      </svg>
    </div>
    <h3 className="mb-2 text-lg font-semibold" style={{ color: 'var(--primary-text)' }}>{features[1].title}</h3>
    <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[1].desc}</p>
  </div>

  {/* Secure & Local */}
  <div
    className="rounded-xl p-6"
    style={{ gridArea: 'secure', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
  >
    <div className="mb-4 inline-flex items-center justify-center rounded-xl p-2.5"
      style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
      {features[2].icon}
    </div>
    <h3 className="mb-1.5 text-base font-semibold" style={{ color: 'var(--primary-text)' }}>{features[2].title}</h3>
    <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[2].desc}</p>
  </div>

  {/* CLI Power */}
  <div
    className="rounded-xl p-6"
    style={{ gridArea: 'cli', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
  >
    <div className="mb-4 inline-flex items-center justify-center rounded-xl p-2.5"
      style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
      {features[3].icon}
    </div>
    <h3 className="mb-1.5 text-base font-semibold" style={{ color: 'var(--primary-text)' }}>{features[3].title}</h3>
    <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[3].desc}</p>
  </div>

  {/* Desktop App */}
  <div
    className="rounded-xl p-6"
    style={{ gridArea: 'desktop', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
  >
    <div className="mb-4 inline-flex items-center justify-center rounded-xl p-2.5"
      style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
      {features[4].icon}
    </div>
    <h3 className="mb-1.5 text-base font-semibold" style={{ color: 'var(--primary-text)' }}>{features[4].title}</h3>
    <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[4].desc}</p>
  </div>
</div>
```

- [ ] Commit:
```bash
git add www/src/components/react/Features.tsx
git commit -m "feat(www): bento grid layout for features section"
```

---

## Task 5: WorksEverywhere — Monochrome Icons

**Files:**
- Modify: `www/src/components/react/WorksEverywhere.tsx`

**Changes:**
- Replace all coloured SVG icons with monochrome `currentColor` icons
- Remove `shadow-sm hover:shadow-md` — use `hover:border-color` instead (flat aesthetic)

- [ ] Replace the entire `WorksEverywhere.tsx`:

```tsx
interface WorksEverywhereProps {
  lang: 'en' | 'zh';
}

function AppleIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701" />
    </svg>
  );
}

function AndroidIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M17.6 9.48l1.84-3.18c.16-.31.04-.69-.26-.85a.637.637 0 0 0-.83.22l-1.88 3.24C14.79 8.33 13.44 8 12 8s-2.79.33-4.47.91L5.65 5.67a.637.637 0 0 0-.83-.22c-.3.16-.42.54-.26.85L6.4 9.48C3.3 11.25 1.28 14.44 1 18h22c-.28-3.56-2.3-6.75-5.4-8.52zM7 15.25a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5zm10 0a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5z" />
    </svg>
  );
}

function MacOSIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="4" width="20" height="13" rx="2" />
      <path d="M8 21h8" />
      <path d="M12 17v4" />
    </svg>
  );
}

function WindowsIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.549H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801" />
    </svg>
  );
}

function LinuxIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <circle cx="12" cy="8" r="4" />
      <path d="M6 20v-2a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v2" />
    </svg>
  );
}

export default function WorksEverywhere({ lang }: WorksEverywhereProps) {
  const isZh = lang === 'zh';

  const platforms = [
    { icon: <AppleIcon />, name: 'iOS' },
    { icon: <AndroidIcon />, name: 'Android' },
    { icon: <MacOSIcon />, name: 'macOS' },
    { icon: <WindowsIcon />, name: 'Windows' },
    { icon: <LinuxIcon />, name: 'Linux' },
  ];

  return (
    <section className="py-12 md:py-16" style={{ background: 'var(--background-secondary)', borderTop: '1px solid var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <h2
          className="text-center text-2xl font-bold mb-10"
          style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}
        >
          {isZh ? '多平台运行' : 'Works everywhere'}
        </h2>

        <div className="flex flex-wrap justify-center gap-4">
          {platforms.map((p) => (
            <div
              key={p.name}
              className="flex items-center gap-3 rounded-xl px-6 py-4 min-w-[130px] transition-colors"
              style={{
                background: 'var(--background-primary)',
                border: '1px solid var(--border-default)',
                color: 'var(--primary-text)',
              }}
            >
              <div className="flex-shrink-0">{p.icon}</div>
              <span className="text-sm font-semibold">{p.name}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
```

- [ ] Commit:
```bash
git add www/src/components/react/WorksEverywhere.tsx
git commit -m "feat(www): monochrome platform icons in WorksEverywhere"
```

---

## Task 6: GetStarted — Platform-Aware Download Links

**Files:**
- Modify: `www/src/components/react/GetStarted.tsx`

**Changes:**
- Import and use `getDownloadUrls()` via `useEffect` for all "Download" href values
- Replace the two hardcoded `https://github.com/yuler/airvoice/releases` hrefs with the dynamic URLs
- Android mobile card: use `urls.mobile` for the "GitHub Releases" button

- [ ] Add the import and state at the top of `GetStarted`:

Remove `import { ArrowSquareOut } from '@phosphor-icons/react';` and replace it with:
```tsx
import { useState, useEffect } from 'react';
import { ArrowSquareOut } from '@phosphor-icons/react';
import { getDownloadUrls, type DownloadUrls } from '../../lib/downloads';
```

Add state after `const isZh = lang === 'zh';`:
```tsx
const [urls, setUrls] = useState<DownloadUrls>({
  cli: 'https://github.com/yuler/airvoice/releases/latest',
  desktop: 'https://github.com/yuler/airvoice/releases/latest',
  mobile: 'https://github.com/yuler/airvoice/releases/latest',
});

useEffect(() => {
  setUrls(getDownloadUrls());
}, []);
```

- [ ] Update the CLI download `<a>` href from `"https://github.com/yuler/airvoice/releases"` to `{urls.cli}`.

- [ ] Update the Desktop download `<a>` href from `"https://github.com/yuler/airvoice/releases"` to `{urls.desktop}`.

- [ ] Update the mobile "GitHub Releases" `<a>` href from `"https://github.com/yuler/airvoice/releases"` to `{urls.mobile}`.

- [ ] Commit:
```bash
git add www/src/components/react/GetStarted.tsx
git commit -m "feat(www): platform-aware download links in GetStarted"
```

---

## Task 7: Footer — Add Quick Links Columns

**Files:**
- Modify: `www/src/components/react/Footer.tsx`

**Changes:**
- Expand from 2-column (logo + links) to 3-column layout: branding | product links | resources
- Add `SoundwaveIcon` inline (same SVG as in Header) next to the "Airvoice" wordmark

- [ ] Replace the entire `Footer.tsx`:

```tsx
interface FooterProps {
  lang: 'en' | 'zh';
}

function SoundwaveIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M2 10v4" /><path d="M6 6v12" /><path d="M10 3v18" /><path d="M14 6v12" /><path d="M18 10v4" /><path d="M22 12v0" />
    </svg>
  );
}

export default function Footer({ lang }: FooterProps) {
  const isZh = lang === 'zh';
  const currentYear = new Date().getFullYear();
  const base = import.meta.env.BASE_URL || '/';
  const loc = (path: string) => isZh ? `${base}zh/${path}`.replace(/\/+/g, '/') : `${base}${path}`.replace(/\/+/g, '/');

  return (
    <footer className="border-t" style={{ backgroundColor: 'var(--background-secondary)', borderColor: 'var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6 py-12">
        <div className="grid grid-cols-1 gap-10 sm:grid-cols-3">

          {/* Branding */}
          <div>
            <div className="flex items-center gap-2 mb-3" style={{ color: 'var(--primary-text)' }}>
              <SoundwaveIcon />
              <span className="font-semibold text-sm">Airvoice</span>
            </div>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--muted-text)', maxWidth: '200px' }}>
              {isZh ? '跨设备通信的统一桥梁。' : 'A unified bridge for cross-device communication.'}
            </p>
            <p className="mt-4 text-xs" style={{ color: 'var(--muted-text)' }}>
              © {currentYear} Airvoice.{' '}
              {isZh ? '保留所有权利。' : 'All rights reserved.'}
            </p>
          </div>

          {/* Product */}
          <div>
            <h4 className="mb-4 text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--muted-text)' }}>
              {isZh ? '产品' : 'Product'}
            </h4>
            <ul className="space-y-3">
              {[
                { href: 'https://github.com/yuler/airvoice/releases/latest', label: isZh ? '下载 CLI' : 'Download CLI', external: true },
                { href: 'https://github.com/yuler/airvoice/releases/latest', label: isZh ? '下载桌面版' : 'Download Desktop', external: true },
                { href: 'https://github.com/yuler/airvoice/releases/latest', label: isZh ? '移动端应用' : 'Mobile App', external: true },
                { href: 'https://github.com/yuler/airvoice/releases', label: isZh ? '版本发布' : 'Releases', external: true },
              ].map((l) => (
                <li key={l.label}>
                  <a
                    href={l.href}
                    target={l.external ? '_blank' : undefined}
                    rel={l.external ? 'noopener noreferrer' : undefined}
                    className="text-sm hover:underline transition-all"
                    style={{ color: 'var(--secondary-text)' }}
                  >
                    {l.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h4 className="mb-4 text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--muted-text)' }}>
              {isZh ? '资源' : 'Resources'}
            </h4>
            <ul className="space-y-3">
              {[
                { href: loc('docs/background/'), label: isZh ? '文档' : 'Documentation' },
                { href: loc('docs/quick-start/'), label: isZh ? '快速开始' : 'Quick Start' },
                { href: loc('docs/architecture/'), label: isZh ? '架构' : 'Architecture' },
                { href: 'https://github.com/yuler/airvoice', label: 'GitHub', external: true },
              ].map((l) => (
                <li key={l.label}>
                  <a
                    href={l.href}
                    target={(l as { external?: boolean }).external ? '_blank' : undefined}
                    rel={(l as { external?: boolean }).external ? 'noopener noreferrer' : undefined}
                    className="text-sm hover:underline transition-all"
                    style={{ color: 'var(--secondary-text)' }}
                  >
                    {l.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </footer>
  );
}
```

- [ ] Commit:
```bash
git add www/src/components/react/Footer.tsx
git commit -m "feat(www): expand footer with 3-column layout and quick links"
```

---

## Task 8: Docs Sidebar — Remove Kumo Dep, New Active State

**Files:**
- Modify: `www/src/components/react/DocsSidebar.tsx`

**Changes:**
- Remove `import { List, X } from '@phosphor-icons/react'` and `import { Button } from '@cloudflare/kumo'`
- Inline SVG hamburger and close icons (same style as Header)
- Active link: left accent border `border-left: 2px solid #006efe`, no bg color
- Inactive links: `var(--secondary-text)` color
- Sidebar background: `var(--background-secondary)`, border: `var(--border-default)`
- Add soundwave + "Airvoice" logo link at top

- [ ] Replace entire `DocsSidebar.tsx`:

```tsx
import { useState } from 'react';

interface SidebarProps {
  lang: 'en' | 'zh';
  base: string;
  currentPath: string;
}

function HamburgerIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

function SoundwaveIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M2 10v4" /><path d="M6 6v12" /><path d="M10 3v18" /><path d="M14 6v12" /><path d="M18 10v4" /><path d="M22 12v0" />
    </svg>
  );
}

export default function DocsSidebar({ lang, base, currentPath }: SidebarProps) {
  const [open, setOpen] = useState(false);
  const isZh = lang === 'zh';
  const normalizedBase = base.endsWith('/') ? base : `${base}/`;
  const loc = (path: string) => lang === 'en' ? `${normalizedBase}${path}` : `${normalizedBase}zh/${path}`;

  const sections = [
    {
      title: isZh ? '入门' : 'Overview',
      items: [
        { href: loc('docs/background/'), label: isZh ? '背景' : 'Background' },
        { href: loc('docs/quick-start/'), label: isZh ? '快速开始' : 'Quick Start' },
      ],
    },
    {
      title: isZh ? '指南' : 'Guide',
      items: [
        { href: loc('docs/development/'), label: isZh ? '开发指南' : 'Development' },
        { href: loc('docs/architecture/'), label: isZh ? '架构' : 'Architecture' },
        { href: loc('docs/platform-deps/'), label: isZh ? '平台依赖' : 'Platform Deps' },
      ],
    },
  ];

  return (
    <>
      {/* Mobile toggle button */}
      <button
        className="fixed left-4 top-3.5 z-50 md:hidden flex items-center justify-center rounded-md p-1.5 transition-colors"
        style={{ color: 'var(--secondary-text)', background: 'transparent', border: 'none', cursor: 'pointer' }}
        onClick={() => setOpen(!open)}
        aria-label={isZh ? '切换菜单' : 'Toggle menu'}
      >
        {open ? <CloseIcon /> : <HamburgerIcon />}
      </button>

      {/* Backdrop */}
      {open && (
        <div
          className="fixed inset-0 z-30 md:hidden"
          style={{ background: 'rgba(0,0,0,0.5)' }}
          onClick={() => setOpen(false)}
        />
      )}

      {/* Sidebar panel */}
      <aside
        className={`fixed top-0 left-0 z-40 h-full w-64 overflow-auto transition-transform duration-300 ${
          open ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
        }`}
        style={{
          background: 'var(--background-secondary)',
          borderRight: '1px solid var(--border-default)',
        }}
      >
        {/* Logo */}
        <div className="flex h-[52px] items-center border-b px-4" style={{ borderColor: 'var(--border-default)' }}>
          <a
            href={loc('')}
            className="flex items-center gap-2 transition-opacity hover:opacity-80"
            style={{ color: 'var(--primary-text)', textDecoration: 'none' }}
          >
            <SoundwaveIcon />
            <span className="text-sm font-semibold">Airvoice</span>
          </a>
        </div>

        {/* Nav */}
        <nav className="p-4">
          {sections.map((section) => (
            <div key={section.title} className="mb-6">
              <h3
                className="mb-2 text-[11px] font-semibold uppercase tracking-wider"
                style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}
              >
                {section.title}
              </h3>
              <ul className="space-y-0.5">
                {section.items.map((item) => {
                  const isActive = currentPath === item.href ||
                    currentPath === item.href.replace(/\/$/, '') ||
                    (item.href !== normalizedBase && currentPath.startsWith(item.href));
                  return (
                    <li key={item.href}>
                      <a
                        href={item.href}
                        onClick={() => setOpen(false)}
                        className="block rounded-md px-3 py-2 text-sm transition-colors"
                        style={
                          isActive
                            ? {
                                borderLeft: '2px solid #006efe',
                                borderRadius: '0 6px 6px 0',
                                paddingLeft: '10px',
                                color: 'var(--primary-text)',
                                fontWeight: 500,
                                background: 'transparent',
                                textDecoration: 'none',
                              }
                            : {
                                color: 'var(--secondary-text)',
                                textDecoration: 'none',
                              }
                        }
                      >
                        {item.label}
                      </a>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>
      </aside>
    </>
  );
}
```

- [ ] Commit:
```bash
git add www/src/components/react/DocsSidebar.tsx
git commit -m "feat(www): redesign docs sidebar with custom icons and blue active border"
```

---

## Task 9: TableOfContents — Blue Active, Back to Top, Lower Threshold

**Files:**
- Modify: `www/src/components/react/TableOfContents.tsx`
- Modify: `www/src/layouts/DocsLayout.astro` (change `xl:block` → `lg:block`)

**Changes:**
- Active heading: `color: #006efe` (accent blue)
- Add "Back to top" link at bottom
- In `DocsLayout.astro`: `xl:block` → `lg:block` for the TOC aside

- [ ] Replace `TableOfContents.tsx`:

```tsx
import { useState, useEffect, useMemo } from 'react';

interface TocProps {
  headings: { depth: number; slug: string; text: string }[];
  lang: 'en' | 'zh';
}

export default function TableOfContents({ headings, lang }: TocProps) {
  const filtered = useMemo(() => headings.filter((h) => h.depth <= 3), [headings]);
  const [activeId, setActiveId] = useState<string>('');

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) setActiveId(entry.target.id);
        }
      },
      { rootMargin: '-80px 0px -70% 0px', threshold: 0 }
    );
    for (const h of filtered) {
      const el = document.getElementById(h.slug);
      if (el) observer.observe(el);
    }
    return () => observer.disconnect();
  }, [filtered]);

  if (filtered.length === 0) return null;

  return (
    <nav>
      <h4
        className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
        style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}
      >
        {lang === 'en' ? 'On this page' : '本页目录'}
      </h4>
      <ul className="space-y-1">
        {filtered.map((h) => {
          const isActive = activeId === h.slug;
          return (
            <li key={h.slug}>
              <a
                href={`#${h.slug}`}
                className="block text-sm transition-colors"
                style={{
                  paddingLeft: h.depth === 3 ? '16px' : undefined,
                  color: isActive ? '#006efe' : 'var(--muted-text)',
                  fontWeight: isActive ? 500 : undefined,
                  textDecoration: 'none',
                }}
              >
                {h.text}
              </a>
            </li>
          );
        })}
      </ul>
      <div className="mt-6 pt-4" style={{ borderTop: '1px solid var(--border-default)' }}>
        <a
          href="#top"
          className="text-xs transition-colors hover:opacity-80"
          style={{ color: 'var(--muted-text)', textDecoration: 'none' }}
          onClick={(e) => { e.preventDefault(); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
        >
          ↑ {lang === 'en' ? 'Back to top' : '回到顶部'}
        </a>
      </div>
    </nav>
  );
}
```

- [ ] In `DocsLayout.astro`, change `hidden w-56 shrink-0 xl:block` → `hidden w-56 shrink-0 lg:block`.

- [ ] Commit:
```bash
git add www/src/components/react/TableOfContents.tsx www/src/layouts/DocsLayout.astro
git commit -m "feat(www): blue active TOC, back-to-top, lower lg threshold"
```

---

## Task 10: DocsPagination — New Component + DocsLayout Integration

**Files:**
- Create: `www/src/components/react/DocsPagination.tsx`
- Modify: `www/src/layouts/DocsLayout.astro`
- Modify: `www/src/pages/docs/[...slug].astro`
- Modify: `www/src/pages/zh/docs/[...slug].astro`

**Nav order (both en and zh):**
1. background
2. quick-start
3. development
4. architecture
5. platform-deps

- [ ] Create `www/src/components/react/DocsPagination.tsx`:

```tsx
interface PaginationProps {
  lang: 'en' | 'zh';
  currentSlug: string;
  base: string;
}

const ORDER_EN = [
  { slug: 'background', label: 'Background' },
  { slug: 'quick-start', label: 'Quick Start' },
  { slug: 'development', label: 'Development' },
  { slug: 'architecture', label: 'Architecture' },
  { slug: 'platform-deps', label: 'Platform Deps' },
];

const ORDER_ZH = [
  { slug: 'background', label: '背景' },
  { slug: 'quick-start', label: '快速开始' },
  { slug: 'development', label: '开发指南' },
  { slug: 'architecture', label: '架构' },
  { slug: 'platform-deps', label: '平台依赖' },
];

export default function DocsPagination({ lang, currentSlug, base }: PaginationProps) {
  const order = lang === 'zh' ? ORDER_ZH : ORDER_EN;
  const normalizedBase = base.endsWith('/') ? base : `${base}/`;
  const loc = (slug: string) =>
    lang === 'en'
      ? `${normalizedBase}docs/${slug}/`.replace(/\/+/g, '/')
      : `${normalizedBase}zh/docs/${slug}/`.replace(/\/+/g, '/');

  const idx = order.findIndex((p) => p.slug === currentSlug);
  if (idx === -1) return null;

  const prev = idx > 0 ? order[idx - 1] : null;
  const next = idx < order.length - 1 ? order[idx + 1] : null;

  if (!prev && !next) return null;

  return (
    <nav
      className="mt-12 grid grid-cols-1 gap-3 sm:grid-cols-2"
      style={{ borderTop: '1px solid var(--border-default)', paddingTop: '24px' }}
    >
      {prev ? (
        <a
          href={loc(prev.slug)}
          className="flex flex-col rounded-xl p-4 transition-colors hover:border-opacity-80"
          style={{
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            textDecoration: 'none',
          }}
        >
          <span className="text-xs mb-1" style={{ color: 'var(--muted-text)' }}>
            ← {lang === 'en' ? 'Previous' : '上一篇'}
          </span>
          <span className="text-sm font-medium" style={{ color: 'var(--primary-text)' }}>
            {prev.label}
          </span>
        </a>
      ) : <div />}

      {next ? (
        <a
          href={loc(next.slug)}
          className="flex flex-col rounded-xl p-4 transition-colors sm:text-right"
          style={{
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            textDecoration: 'none',
          }}
        >
          <span className="text-xs mb-1" style={{ color: 'var(--muted-text)' }}>
            {lang === 'en' ? 'Next' : '下一篇'} →
          </span>
          <span className="text-sm font-medium" style={{ color: 'var(--primary-text)' }}>
            {next.label}
          </span>
        </a>
      ) : <div />}
    </nav>
  );
}
```

- [ ] In `DocsLayout.astro`, add `currentSlug` to the Props interface and the `Astro.props` destructure:

```astro
interface Props {
  title: string;
  description: string;
  lang: 'en' | 'zh';
  headings: { depth: number; slug: string; text: string }[];
  currentSlug: string;
}

const { title, description, lang, headings, currentSlug } = Astro.props;
```

Import `DocsPagination`:
```astro
import DocsPagination from '../components/react/DocsPagination';
```

After `<slot />` inside the `<article>`, add:
```astro
<DocsPagination lang={lang} currentSlug={currentSlug} base={base} client:load />
```

- [ ] In `www/src/pages/docs/[...slug].astro`, pass `currentSlug={entry.slug}` to `<DocsLayout>`.

- [ ] In `www/src/pages/zh/docs/[...slug].astro`, pass `currentSlug={entry.slug}` to `<DocsLayout>`.

- [ ] Commit:
```bash
git add www/src/components/react/DocsPagination.tsx www/src/layouts/DocsLayout.astro www/src/pages/docs/[...slug].astro www/src/pages/zh/docs/[...slug].astro
git commit -m "feat(www): add prev/next pagination to docs pages"
```

---

## Task 11: global.css — Scroll Behaviour

**Files:**
- Modify: `www/src/styles/global.css`

**Changes:**
- Add `scroll-behavior: smooth` to `html`

- [ ] Add to `global.css` after the `body` block:

```css
html {
  scroll-behavior: smooth;
}
```

- [ ] Commit:
```bash
git add www/src/styles/global.css
git commit -m "feat(www): smooth scroll behaviour"
```

---

## Self-Review Checklist

- [x] Platform download utility (`downloads.ts`) — Task 1
- [x] Header text labels + CLI link — Task 2
- [x] Hero dark island, big headline, dark terminal, platform links, fade-in — Task 3
- [x] Features bento grid — Task 4
- [x] WorksEverywhere monochrome icons — Task 5
- [x] GetStarted platform links — Task 6
- [x] Footer 3-column — Task 7
- [x] DocsSidebar kumo removed, blue active — Task 8
- [x] TOC blue active + back to top + lg threshold — Task 9
- [x] DocsPagination component + layout integration — Task 10
- [x] Smooth scroll — Task 11
