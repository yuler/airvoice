# Airvoice www UI Upgrade — Design Spec

**Date:** 2026-07-02  
**Status:** Approved  
**Scope:** `www/` (Astro + React + Tailwind v4)

---

## Goal

Elevate the Airvoice landing page and docs site from a clean-but-plain layout to a premium, high-contrast design that matches the brand identity described in `DESIGN.md`: strictly monochrome surfaces, single `#006efe` brand accent, flat/border-driven depth, no gradients or colored backgrounds.

---

## Design Tokens (from DESIGN.md)

| Token | Light | Dark |
|---|---|---|
| `--primary-text` | `#171717` | `#ededed` |
| `--secondary-text` | `#666666` | `#a0a0a0` |
| `--muted-text` | `#888888` | `#666666` |
| `--background-primary` | `#ffffff` | `#000000` |
| `--background-secondary` | `#fafafa` | `#0d0e15` |
| `--border-default` | `#eaeaea` | `#2e2e2e` |
| `--accent-blue` | `#006efe` | `#006efe` |
| `--status-success` | `#28a948` | `#00ac3a` |
| `--status-warning` | `#ffae00` | `#ffae00` |
| `--status-error` | `#fc0035` | `#e2162a` |
| `--status-neutral` | `#8f8f8f` | `#8f8f8f` |

---

## Section 1: Header (`Header.tsx`)

**Changes:**
- Navigation links gain text labels: "Docs", "GitHub", "CLI" (add CLI link → `https://github.com/yuler/airvoice/releases/latest`)
- Header height stays 52px; light/dark toggle and lang toggle stay icon-only (compact)
- Mobile hamburger menu: add CLI link alongside Docs and GitHub

**No Download CTA in header** (user decision).

---

## Section 2: Hero (`Hero.tsx`)

### Dark-Island Hero

The Hero `<section>` always renders with a pure black background (`#000000`) regardless of the user's light/dark preference — a "dark island" that enforces brand identity on first load.

- `background: #000000` set directly on the `<section>` element
- All text inside uses dark-mode token values directly: `#ededed`, `#a0a0a0`
- Border and card surfaces inside use `#0d0e15` / `#2e2e2e`

### Headline Typography

- `h1` → `text-6xl` (60px) on desktop, `text-5xl` (48px) on mobile
- `letter-spacing: -0.04em` (tighter than current `-0.03em`)
- Font weight stays 700

### ALPHA Badge

Replace raw `<span>` with a bordered pill:
```html
<span style="border: 1px solid #006efe; color: #006efe; border-radius: 9999px; padding: 3px 10px; font-size: 11px; font-weight: 700; letter-spacing: 0.08em;">ALPHA</span>
```

### Terminal Window Mock (dark theme)

Replace the current white-background terminal mock with a dark terminal matching DESIGN.md CLI aesthetic:
- Background: `#000000`
- Border: `1px solid #2e2e2e`
- Title bar: `#0d0e15` with red/yellow/green traffic lights
- Monospace text: `#ededed` for values, `#666666` for labels
- "Connected" status text: `#00ac3a`
- Remove the `shadow-2xl` — use `border: 1px solid #2e2e2e` only (flat aesthetic)

### CTA Buttons

Update links for platform-aware downloads (see Section 7).  
Button sizes: primary `height: 48px`, secondary `height: 44px`, matching DESIGN.md component spec.

### Status Indicators

- The "Connected" dot uses `.av-pulse-dot` animation (already in `global.css`)
- Wrap all 4 indicators in a `flex-wrap` row, style unchanged (already correct)

### Scroll Fade-in

Add a lightweight `useScrollReveal` hook (pure Intersection Observer, no library) that adds `opacity-0 translate-y-4` → `opacity-100 translate-y-0` on first viewport intersection. Apply to: headline group, CTA group, hero visuals.

---

## Section 3: Features — Bento Grid (`Features.tsx`)

Replace the uniform 5-column grid with an asymmetric Bento grid on desktop.

### Desktop Grid Layout (≥ 768px)

```
CSS Grid: 3 columns × 2 rows

[col 1, rows 1-2]  Easy Pairing       — tall card (rowspan 2)
[col 2, row 1]     Real-Time Sync     — wide card (featured, larger icon + desc)  
[col 3, row 1]     Secure & Local
[col 2, row 2]     CLI Power
[col 3, row 2]     Desktop App
```

Grid definition:
```css
grid-template-columns: 1fr 1.6fr 1fr;
grid-template-rows: auto auto;
```

Featured card (Real-Time Sync): slightly larger padding, larger icon (36px), slightly larger title (`text-lg`), `background: var(--background-secondary)`.

All other cards: standard `p-6`, `text-base` title, icon 28px.

### Mobile: keep existing carousel (unchanged).

---

## Section 4: Works Everywhere (`WorksEverywhere.tsx`)

- Replace coloured platform logos with monochrome SVG icons (inline `currentColor`), respecting DESIGN.md's "strictly monochrome" rule
- Section background: `var(--background-secondary)` (`#fafafa` light / `#0d0e15` dark) — creates alternating rhythm
- Platform list: iOS, Android, macOS, Windows, Linux (same 5)

SVG sources: simple outline/fill icons, all rendered at 28px, using `currentColor` so they adapt to theme.

---

## Section 5: GetStarted / CTA Banner (`GetStarted.tsx`)

The existing "Open source and community driven" footer banner becomes a full-bleed dark band:

- Background: always `#000000` (same dark-island as Hero), regardless of theme
- Text: white (`#ededed`), subtext `#a0a0a0`
- "View on GitHub" button: bordered pill (secondary style, white border/text on dark bg)
- Full-width `w-full`, `py-16`

---

## Section 6: Footer (`Footer.tsx`)

Add a proper footer with three columns:

| Product | Resources | Connect |
|---|---|---|
| Download CLI | Docs | GitHub |
| Download Desktop | Quick Start | CLI on GitHub |
| Get Mobile App | Architecture | Releases |

Below: copyright line — `© 2026 Airvoice. All rights reserved.`

Footer background: `var(--background-secondary)`, top border `var(--border-default)`.

---

## Section 7: Platform-Aware Download Links

Create `www/src/lib/downloads.ts` — a small utility that detects the user's OS and returns the correct GitHub Releases download URL.

Latest release tag: dynamically read from `VERSION` file at build time (currently `v0.3.1`).  
Base URL pattern: `https://github.com/yuler/airvoice/releases/download/{tag}/{filename}`

| Platform detected | CLI file | Desktop file |
|---|---|---|
| macOS ARM64 (`Mac` + ARM) | `airvoice-cli-darwin-arm64` | `Airvoice-Desktop-{tag}-macOS.zip` |
| macOS x64 | `airvoice-cli-darwin-amd64` | `Airvoice-Desktop-{tag}-macOS.zip` |
| Linux | `airvoice-cli-linux-amd64` | `Airvoice-Desktop-{tag}-Linux.tar.gz` |
| Windows | `airvoice-cli-windows-amd64.exe` | `Airvoice-Desktop-{tag}-Windows.zip` |
| Android (mobile UA) | — | `airvoice-android-{tag}.apk` |
| Fallback | `https://github.com/yuler/airvoice/releases/latest` | same |

Detection uses `navigator.userAgent` + `navigator.platform` (client-side, React `useEffect`).  
Server-side (Astro SSG): always render the fallback releases page URL, hydrate client-side.

Mobile "Get Mobile App" button:
- Android UA → direct `.apk` link
- iOS UA → link to `https://github.com/yuler/airvoice` (no App Store yet)
- Default → releases page

---

## Section 8: Docs Layout Full Upgrade

### DocsSidebar (`DocsSidebar.tsx`)

- Remove `@cloudflare/kumo` Button import — replace hamburger with a pure SVG button styled with inline CSS matching the Header pattern
- Active link style: remove `bg-kumo-control`, replace with left border accent: `border-left: 2px solid #006efe; background: transparent; color: var(--primary-text)`
- Inactive links: `color: var(--secondary-text)`, hover: `color: var(--primary-text)`
- Section headings: `font-size: 11px; letter-spacing: 0.08em; text-transform: uppercase; color: var(--muted-text)` — same style as ALPHA badge tracking
- Sidebar background: `var(--background-secondary)`, right border: `1px solid var(--border-default)`
- Logo link at top of sidebar: add the soundwave SVG icon (same as Header), consistent branding

### DocsLayout (`DocsLayout.astro`)

- Add breadcrumb bar below header: `Overview > Background` (derived from current slug)
- Main content `px-6 py-10 md:px-12` stays, but `max-w-3xl` → `max-w-2xl` (narrower for reading comfort, ~65ch)
- Below `<slot />`: add prev/next navigation component

### TableOfContents (`TableOfContents.tsx`)

- Show from `lg:` (currently `xl:`) — threshold lowered
- Active heading highlighted with `color: #006efe`
- Add "Back to top" link at bottom of TOC panel (scrolls to `#top`)
- Smooth scroll behavior: `scroll-behavior: smooth` on `<html>` (already handled by Astro or add to global.css)

### Prev/Next Navigation (new component: `DocsPagination.tsx`)

New file: `www/src/components/react/DocsPagination.tsx`

Renders two cards side by side (or stacked on mobile):
- Left card: `← Previous: <title>`
- Right card: `Next: <title> →`

Navigation order follows sidebar section order:
1. Background
2. Quick Start
3. Development
4. Architecture
5. Platform Deps

Cards styled with `border: 1px solid var(--border-default)`, `border-radius: 12px`, `padding: 16px 20px`.

---

## Non-Goals

- No changes to content (copy) in docs pages
- No new pages or routes
- No Astro version upgrade
- No animation library (Framer Motion etc.) — only Intersection Observer
- No App Store integration
