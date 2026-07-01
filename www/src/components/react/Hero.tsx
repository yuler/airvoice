import { ArrowRight } from '@phosphor-icons/react';

interface HeroProps {
  lang: 'en' | 'zh';
  base: string;
}

// Platform SVG icons (inline, no external deps)
function WindowsIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-label="Windows">
      <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.549H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801" />
    </svg>
  );
}

function AppleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-label="macOS">
      <path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701" />
    </svg>
  );
}

function LinuxIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-label="Linux">
      {/* Tux penguin — simplified clean version */}
      <path d="M12 2C9.243 2 7 4.243 7 7v2c0 .738.203 1.43.556 2.022C6.613 12.482 6 14.207 6 16c0 2.21 1.343 3 3 3h6c1.657 0 3-.79 3-3 0-1.793-.613-3.518-1.556-4.978C16.797 10.43 17 9.738 17 9V7c0-2.757-2.243-5-5-5zm-1.5 6a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5zm3 0a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5zM12 18c-1.5 0-2-.5-2-2s.895-3.5 2-3.5 2 1.5 2 3.5-.5 2-2 2z" />
    </svg>
  );
}

function AndroidIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-label="Android">
      {/* Material Design Android robot icon */}
      <path d="M17.6 9.48l1.84-3.18c.16-.31.04-.69-.26-.85a.637.637 0 0 0-.83.22l-1.88 3.24C14.79 8.33 13.44 8 12 8s-2.79.33-4.47.91L5.65 5.67a.637.637 0 0 0-.83-.22c-.3.16-.42.54-.26.85L6.4 9.48C3.3 11.25 1.28 14.44 1 18h22c-.28-3.56-2.3-6.75-5.4-8.52zM7 15.25a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5zm10 0a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5z" />
    </svg>
  );
}

function IosIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-label="iOS">
      <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.8-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z" />
    </svg>
  );
}

export default function Hero({ lang, base }: HeroProps) {
  const isZh = lang === 'zh';
  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  return (
    <section className="py-24 md:py-36">
      <div className="mx-auto max-w-3xl px-6 text-center">

        {/* Live "LAN Only" badge */}
        <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-kumo-hairline px-3 py-1.5 text-xs font-medium text-kumo-subtle">
          <span className="av-pulse-dot" />
          {isZh ? '局域网专属 · 无需联网' : 'LAN Only · No Internet Required'}
        </div>

        <h1 className="text-4xl font-bold tracking-tight text-kumo-default md:text-5xl lg:text-6xl" style={{ letterSpacing: '-0.04em' }}>
          {isZh
            ? <>手机说话，<br />电脑打字</>
            : <>Speak on your phone,<br />type on your PC</>
          }
        </h1>

        <p className="mt-6 text-base text-kumo-subtle md:text-lg" style={{ maxWidth: '480px', margin: '1.5rem auto 0' }}>
          {isZh
            ? '局域网语音桥接。无云端、无账号、无追踪——所有数据留在本地。'
            : 'Voice bridge over your local network. No cloud, no accounts, no data leaves your device.'}
        </p>

        {/* Platform support row */}
        <div className="mt-8 flex flex-col items-center gap-3">
          <p className="text-xs text-kumo-inactive uppercase tracking-widest">
            {isZh ? '支持平台' : 'Supported Platforms'}
          </p>
          <div className="flex items-center gap-1">
            {/* Desktop platforms */}
            <div className="flex items-center gap-1">
              <span className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2.5 py-1.5 text-kumo-subtle" title="Windows">
                <WindowsIcon />
                <span className="text-xs font-medium">Windows</span>
              </span>
              <span className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2.5 py-1.5 text-kumo-subtle" title="macOS">
                <AppleIcon />
                <span className="text-xs font-medium">macOS</span>
              </span>
              <span className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2.5 py-1.5 text-kumo-subtle" title="Linux">
                <LinuxIcon />
                <span className="text-xs font-medium">Linux</span>
              </span>
            </div>
            {/* Divider */}
            <span className="mx-2 text-kumo-inactive text-sm select-none">+</span>
            {/* Mobile platforms */}
            <div className="flex items-center gap-1">
              <span className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2.5 py-1.5 text-kumo-subtle" title="Android">
                <AndroidIcon />
                <span className="text-xs font-medium">Android</span>
              </span>
              <span className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2.5 py-1.5 text-kumo-subtle" title="iOS">
                <IosIcon />
                <span className="text-xs font-medium">iOS</span>
              </span>
            </div>
          </div>
        </div>

        {/* CTA buttons */}
        <div className="mt-10 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <a
            href="#get-started"
            className="inline-flex items-center gap-2 rounded-full px-6 py-3 text-sm font-medium text-white transition-opacity hover:opacity-90"
            style={{ backgroundColor: 'var(--accent-blue, #006efe)', height: '44px' }}
          >
            {isZh ? '快速开始' : 'Get Started'}
            <ArrowRight size={14} />
          </a>
          <a
            href={loc('docs/background/')}
            className="inline-flex items-center gap-2 rounded-full border border-kumo-hairline px-6 py-3 text-sm font-medium text-kumo-default transition-colors hover:bg-kumo-control"
            style={{ height: '44px' }}
          >
            {isZh ? '了解更多' : 'Learn more'}
          </a>
        </div>
      </div>
    </section>
  );
}
