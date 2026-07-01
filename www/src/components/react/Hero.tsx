import { useState, useEffect } from 'react';
import Lightbox from './Lightbox';

interface HeroProps {
  lang: 'en' | 'zh';
  base: string;
}

// ── Terminal window mock ──────────────────────────────────────────────────────
function TerminalWindow() {
  const bg = '#ffffff';
  const fg = '#374151';
  const labelColor = '#6b7280';

  return (
    <div
      className="rounded-2xl overflow-hidden shadow-2xl w-[260px] sm:w-[270px]"
      style={{ background: bg, border: '1px solid #e5e7eb' }}
    >
      {/* Title bar */}
      <div className="flex items-center gap-1.5 px-4 py-3 border-b" style={{ borderColor: '#e5e7eb' }}>
        <span className="w-2 h-2 rounded-full" style={{ background: '#ff5f57' }} />
        <span className="w-2 h-2 rounded-full" style={{ background: '#febc2e' }} />
        <span className="w-2 h-2 rounded-full" style={{ background: '#28c840' }} />
        <span className="ml-2 text-[9px] text-gray-400 font-mono">airvoice</span>
      </div>
      
      {/* Body with QR and info vertically stacked */}
      <div className="p-4 flex flex-col items-start">
        {/* QR Code Container */}
        <div className="mb-4">
          <img
            src={`${import.meta.env.BASE_URL || '/'}qrcode.svg`.replace(/\/+/g, '/')}
            alt="https://github.com/yuler/airvoice"
            className="w-[200px] h-[200px] block"
          />
        </div>

        {/* Text Contents */}
        <div className="w-full font-mono text-[9px] leading-relaxed space-y-1 text-left">
          <div>
            <span style={{ color: labelColor }}>Token:</span>{' '}
            <span style={{ color: fg }}>277129e4-35ea-40af-a122-13a5839e5e1f</span>
          </div>
          <div>
            <span style={{ color: labelColor }}>WebSocket URL:</span>{' '}
            <span style={{ color: fg }}>ws://192.168.20.189:7654/ws</span>
          </div>
          <div className="pt-2" style={{ color: labelColor }}>
            [airvoice] waiting for phone connection...
          </div>
        </div>
      </div>
    </div>
  );
}

// ── Desktop client mockup ─────────────────────────────────────────────────────
function DesktopWindow() {
  return (
    <div
      className="rounded-2xl overflow-hidden shadow-2xl"
      style={{
        background: '#fff',
        width: '170px',
      }}
    >
      <Lightbox
        src={`${import.meta.env.BASE_URL || '/'}desktop.png`.replace(/\/+/g, '/')}
        alt="Airvoice Desktop Client"
        className="w-full h-auto block"
      />
    </div>
  );
}

// ── Mobile phone mock ─────────────────────────────────────────────────────────
function MobilePhone() {
  return (
    <div
      className="rounded-2xl overflow-hidden shadow-2xl"
      style={{
        width: '170px',
      }}
    >
      <Lightbox
        src={`${import.meta.env.BASE_URL || '/'}phone.jpg`.replace(/\/+/g, '/')}
        alt="Airvoice App Screenshot"
        className="w-full h-auto block"
      />
    </div>
  );
}

// ── Mobile hero visuals (stacked phone + desktop client overlay) ───────────────
function MobileHeroVisual({ activeTab }: { activeTab: 'cli' | 'desktop' }) {
  const leftItemLeft = activeTab === 'cli' ? 'left-[-12px]' : 'left-[15px]';
  const phoneRight = 'right-[-12px]';

  return (
    <div className="relative mx-auto h-[360px] w-[310px]">
      {/* 1. CLI/Desktop (Left / Background) */}
      <div className={`absolute ${leftItemLeft} top-8 z-10 scale-[0.75] origin-top-left transition-all duration-300`}>
        {activeTab === 'cli' ? <TerminalWindow /> : <DesktopWindow />}
      </div>

      {/* 2. iOS App (Right / Foreground) */}
      <div className={`absolute ${phoneRight} top-[20px] z-20 scale-[0.8] origin-top-right transition-all duration-300`}>
        <MobilePhone />
      </div>
    </div>
  );
}

// ── Desktop hero visual ───────────────────────────────────────────────────────
function DesktopHeroVisual({ activeTab }: { activeTab: 'cli' | 'desktop' }) {
  const leftItemLeft = activeTab === 'cli' ? 'left-[10px]' : 'left-[60px]';
  const phoneLeft = activeTab === 'cli' ? 'left-[320px]' : 'left-[290px]';

  return (
    <div className="relative flex items-center justify-center h-[420px] w-full max-w-[540px] mx-auto">
      {/* 1. CLI/Desktop (Left / Background) */}
      <div className={`absolute ${leftItemLeft} top-12 z-10 transition-all duration-300`}>
        {activeTab === 'cli' ? <TerminalWindow /> : <DesktopWindow />}
      </div>

      {/* 2. iOS App GUI (Right / Foreground) */}
      <div className={`absolute ${phoneLeft} top-4 z-20 transition-all duration-300`}>
        <MobilePhone />
      </div>
    </div>
  );
}

// ── Status indicators ─────────────────────────────────────────────────────────
function StatusIndicators({ isZh }: { isZh: boolean }) {
  const statuses = [
    { color: 'var(--status-success)', label: isZh ? '已连接' : 'Connected' },
    { color: 'var(--status-warning)', label: isZh ? '连接中...' : 'Connecting...' },
    { color: 'var(--status-error)', label: isZh ? '错误' : 'Error' },
    { color: 'var(--status-neutral)', label: isZh ? '离线' : 'Offline' },
  ];

  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
      {statuses.map((s) => (
        <span key={s.label} className="flex items-center gap-1.5 text-sm" style={{ color: 'var(--secondary-text)' }}>
          <span className="w-2 h-2 rounded-full flex-shrink-0" style={{ background: s.color }} />
          {s.label}
        </span>
      ))}
    </div>
  );
}

// ── Download icon ─────────────────────────────────────────────────────────────
function DownloadIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="7 10 12 15 17 10" />
      <line x1="12" y1="15" x2="12" y2="3" />
    </svg>
  );
}

function MonitorIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <path d="M8 21h8M12 17v4" />
    </svg>
  );
}

// ── Main Hero ─────────────────────────────────────────────────────────────────
export default function Hero({ lang, base }: HeroProps) {
  const [activeTab, setActiveTab] = useState<'cli' | 'desktop'>('cli');
  const isZh = lang === 'zh';

  return (
    <section className="py-12 md:py-16 lg:py-20">
      <div className="mx-auto max-w-6xl px-4 md:px-6">

        {/* ── Desktop layout: 2-col ── */}
        <div className="flex flex-col gap-10 lg:flex-row lg:items-center lg:gap-12">

          {/* Left column */}
          <div className="flex-1 min-w-0">
            {/* ALPHA badge */}
            <div className="mb-4">
              <span
                className="text-xs font-bold uppercase tracking-widest"
                style={{ color: '#006efe' }}
              >
                ALPHA
              </span>
            </div>

            {/* Headline */}
            <h1
              className="text-4xl font-bold leading-tight tracking-tight md:text-5xl"
              style={{ color: 'var(--primary-text)', letterSpacing: '-0.03em' }}
            >
              {isZh ? (
                <>与你的设备沟通。<br />简单直接。</>
              ) : (
                <>Talk to your<br />devices. Simply.</>
              )}
            </h1>

            <p className="mt-4 text-base leading-relaxed" style={{ color: 'var(--secondary-text)', maxWidth: '440px' }}>
              {isZh
                ? 'Airvoice 是跨设备通信的统一桥梁。我们提供简洁的移动端应用与强大的命令行 CLI，同时也为不习惯命令行的用户准备了直观易用的桌面客户端。'
                : 'Airvoice is a unified bridge for cross-device communication, offering a clean mobile app, a powerful CLI, and a sleek desktop client for those who prefer a GUI.'}
            </p>

            {/* CTA buttons */}
            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              {/* Primary */}
              <a
                href="https://github.com/yuler/airvoice/releases"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 rounded-full px-6 py-3 text-sm font-semibold text-white transition-opacity hover:opacity-90"
                style={{ background: '#006efe' }}
              >
                <DownloadIcon />
                {isZh ? '下载 CLI' : 'Download CLI'}
              </a>
              {/* Secondary */}
              <a
                href="https://github.com/yuler/airvoice/releases"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 rounded-full border px-6 py-3 text-sm font-semibold transition-colors hover:bg-black/5 dark:hover:bg-white/10"
                style={{ borderColor: 'var(--border-default)', color: 'var(--primary-text)' }}
              >
                <MonitorIcon />
                {isZh ? '下载桌面版' : 'Download Desktop'}
              </a>
            </div>

            {/* Status indicators */}
            <div className="mt-6">
              <StatusIndicators isZh={isZh} />
            </div>
          </div>

          {/* Right column — desktop: terminal + phone, mobile: stacked phone */}
          <div className="flex-1 min-w-0 flex flex-col items-center justify-center">
            {/* Tab switcher */}
            <div className="w-full max-w-[540px] flex justify-center lg:justify-end mb-8">
              <div className="inline-flex rounded-full p-1" style={{ background: 'var(--background-secondary)', border: '1px solid var(--border-default)' }}>
                <button
                  onClick={() => setActiveTab('cli')}
                  className={`rounded-full px-5 py-2 text-xs font-bold transition-all duration-200 cursor-pointer border-none outline-none ${
                    activeTab === 'cli'
                      ? 'shadow-sm'
                      : 'hover:opacity-85'
                  }`}
                  style={
                    activeTab === 'cli'
                      ? { backgroundColor: 'var(--background-primary)', color: 'var(--primary-text)' }
                      : { backgroundColor: 'transparent', color: 'var(--secondary-text)' }
                  }
                >
                  {isZh ? '命令行 CLI' : 'CLI'}
                </button>
                <button
                  onClick={() => setActiveTab('desktop')}
                  className={`rounded-full px-5 py-2 text-xs font-bold transition-all duration-200 cursor-pointer border-none outline-none ${
                    activeTab === 'desktop'
                      ? 'shadow-sm'
                      : 'hover:opacity-85'
                  }`}
                  style={
                    activeTab === 'desktop'
                      ? { backgroundColor: 'var(--background-primary)', color: 'var(--primary-text)' }
                      : { backgroundColor: 'transparent', color: 'var(--secondary-text)' }
                  }
                >
                  {isZh ? '桌面端 Desktop' : 'Desktop'}
                </button>
              </div>
            </div>

            {/* Show different layout based on screen */}
            <div className="hidden lg:block w-full">
              <DesktopHeroVisual activeTab={activeTab} />
            </div>
            <div className="lg:hidden">
              <MobileHeroVisual activeTab={activeTab} />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
