import { useState } from 'react';
import Lightbox from './Lightbox';
import type { HeroMessages } from '../../i18n/messages';

interface HeroProps {
  m: HeroMessages;
  docsUrl: string;
  base?: string;
}

function TerminalWindow() {
  const bg = '#ffffff';
  const fg = '#374151';
  const labelColor = '#6b7280';

  return (
    <div
      className="rounded-2xl overflow-hidden w-[210px] sm:w-[270px]"
      style={{ background: bg, border: '1px solid #e5e7eb' }}
    >
      <div className="flex items-center gap-1 sm:gap-1.5 px-2.5 py-2 sm:px-4 sm:py-3 border-b" style={{ borderColor: '#e5e7eb' }}>
        <span className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full" style={{ background: '#ff5f57' }} />
        <span className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full" style={{ background: '#febc2e' }} />
        <span className="w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full" style={{ background: '#28c840' }} />
        <span className="ml-1 sm:ml-2 text-[7px] sm:text-[9px] text-gray-400 font-mono">airvoice</span>
      </div>
      <div className="p-2.5 sm:p-4 flex flex-col items-start">
        <div className="mb-2.5 sm:mb-4 w-full flex justify-center sm:justify-start">
          <img
            src={`${import.meta.env.BASE_URL || '/'}qrcode.svg`.replace(/\/+/g, '/')}
            alt="https://github.com/yuler/airvoice"
            className="w-[140px] h-[140px] sm:w-[200px] sm:h-[200px] block"
          />
        </div>
        <div className="w-full font-mono text-[7px] sm:text-[9px] leading-relaxed space-y-0.5 sm:space-y-1 text-left truncate-children">
          <div className="truncate">
            <span style={{ color: labelColor }}>Token:</span>{' '}
            <span style={{ color: fg }}>277129e4-35ea-40af-a122-13a5839e5e1f</span>
          </div>
          <div className="truncate">
            <span style={{ color: labelColor }}>WebSocket URL:</span>{' '}
            <span style={{ color: fg }}>ws://192.168.20.189:7654/ws</span>
          </div>
          <div className="pt-1 sm:pt-2 truncate" style={{ color: labelColor }}>
            [airvoice] waiting for phone connection...
          </div>
        </div>
      </div>
    </div>
  );
}

function DesktopWindow({ base }: { base: string }) {
  return (
    <div
      className="rounded-2xl overflow-hidden w-[150px] sm:w-[170px]"
      style={{ border: '1px solid var(--border-default)' }}
    >
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
    <div
      className="rounded-2xl overflow-hidden w-[150px] sm:w-[170px] lg:w-[220px]"
      style={{ border: '1px solid var(--border-default)' }}
    >
      <Lightbox
        src={`${base || '/'}phone.jpg`.replace(/\/+/g, '/')}
        alt="Airvoice App Screenshot"
        className="w-full h-auto block"
      />
    </div>
  );
}

function MobileHeroVisual({ activeTab, base }: { activeTab: 'cli' | 'desktop'; base: string }) {
  return (
    <div className="flex items-center justify-center gap-5 h-[400px] w-full max-w-[400px] mx-auto">
      <div className="transition-all duration-300 flex-shrink-0">
        {activeTab === 'cli' ? <TerminalWindow /> : <DesktopWindow base={base} />}
      </div>
      <div className="transition-all duration-300 flex-shrink-0">
        <MobilePhone base={base} />
      </div>
    </div>
  );
}

function DesktopHeroVisual({ activeTab, base }: { activeTab: 'cli' | 'desktop'; base: string }) {
  return (
    <div className="flex items-center justify-center gap-10 h-[420px] lg:h-[480px] w-full max-w-[540px] lg:max-w-[600px] mx-auto">
      <div className="transition-all duration-300 flex-shrink-0">
        {activeTab === 'cli' ? <TerminalWindow /> : <DesktopWindow base={base} />}
      </div>
      <div className="transition-all duration-300 flex-shrink-0">
        <MobilePhone base={base} />
      </div>
    </div>
  );
}

function ChevronDownIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="12" y1="5" x2="12" y2="19"></line>
      <polyline points="19 12 12 19 5 12"></polyline>
    </svg>
  );
}

function BookOpenIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
      <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
    </svg>
  );
}

export default function Hero({ m, docsUrl, base = import.meta.env.BASE_URL || '/' }: HeroProps) {
  const [activeTab, setActiveTab] = useState<'cli' | 'desktop'>('cli');

  const statuses = [
    { color: 'var(--status-success)', label: m.statusConnected, pulse: true },
    { color: 'var(--status-warning)', label: m.statusConnecting },
    { color: 'var(--status-error)', label: m.statusError },
    { color: 'var(--status-neutral)', label: m.statusOffline },
  ];

  return (
    <section
      className="py-14 md:py-20 lg:py-24 border-b"
      style={{ background: 'var(--background-primary)', borderColor: 'var(--border-default)' }}
    >
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <div className="flex flex-col gap-10 lg:flex-row lg:items-center lg:gap-16">
          <div className="flex-1 min-w-0">
            <div className="mb-5">
              <span style={{
                border: '1px solid #006efe',
                color: '#006efe',
                borderRadius: '9999px',
                padding: '3px 12px',
                fontSize: '11px',
                fontWeight: 700,
                letterSpacing: '0.08em',
                display: 'inline-block',
              }}>
                ALPHA
              </span>
            </div>
            <h1
              className="text-5xl font-bold leading-tight md:text-6xl"
              style={{ color: 'var(--primary-text)', letterSpacing: '-0.04em' }}
            >
              {m.title}
            </h1>
            <p className="mt-5 text-base leading-relaxed" style={{ color: 'var(--secondary-text)', maxWidth: '440px' }}>
              {m.subtitle}
            </p>
            <div className="mt-8 flex flex-row gap-3">
              <a
                href="#get-started"
                className="inline-flex items-center justify-center gap-1.5 rounded-full px-4 text-xs font-bold text-white transition-opacity hover:opacity-90 cursor-pointer"
                style={{ background: '#006efe', height: '40px', minWidth: '136px' }}
              >
                <ChevronDownIcon />
                {m.cta}
              </a>
              <a
                href={docsUrl}
                className="inline-flex items-center justify-center gap-1.5 rounded-full border px-4 text-xs font-bold transition-colors hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer"
                style={{ borderColor: 'var(--border-default)', color: 'var(--primary-text)', height: '40px', minWidth: '136px' }}
              >
                <BookOpenIcon />
                {m.docs}
              </a>
            </div>
            <div className="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2">
              {statuses.map((s) => (
                <span key={s.label} className="flex items-center gap-1.5 text-sm" style={{ color: 'var(--secondary-text)' }}>
                  {s.pulse
                    ? <span className="av-pulse-dot" style={{ background: s.color }} />
                    : <span className="w-2 h-2 rounded-full flex-shrink-0" style={{ background: s.color }} />
                  }
                  {s.label}
                </span>
              ))}
            </div>
          </div>
          <div className="flex-1 min-w-0 flex flex-col items-center justify-center">
            <div className="hidden lg:block w-full">
              <DesktopHeroVisual activeTab={activeTab} base={base} />
            </div>
            <div className="lg:hidden">
              <MobileHeroVisual activeTab={activeTab} base={base} />
            </div>
            <div className="w-full max-w-[540px] flex justify-center mt-6">
              <div
                className="inline-flex rounded-full p-1"
                style={{ background: 'var(--background-secondary)', border: '1px solid var(--border-default)' }}
              >
                {(['cli', 'desktop'] as const).map((tab) => (
                  <button
                    key={tab}
                    onClick={() => setActiveTab(tab)}
                    className="rounded-full px-5 py-2 text-xs font-bold transition-all duration-200 cursor-pointer border-none outline-none"
                    style={
                      activeTab === tab
                        ? { backgroundColor: '#006efe', color: '#ffffff', boxShadow: '0 1px 3px rgba(0,110,254,0.35)' }
                        : { backgroundColor: 'transparent', color: 'var(--secondary-text)' }
                    }
                  >
                    {tab === 'cli' ? m.tabCli : m.tabDesktop}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
