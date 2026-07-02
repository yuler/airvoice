import { useState, useEffect, useRef } from 'react';
import Lightbox from './Lightbox';
import { getDownloadUrls, type DownloadUrls } from '../../lib/downloads';

interface HeroProps {
  lang: 'en' | 'zh';
  base: string;
}

function useScrollReveal(threshold = 0.1) {
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.style.opacity = '0';
    el.style.transform = 'translateY(20px)';
    el.style.transition = 'opacity 0.7s ease, transform 0.7s ease';
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
      {/* Title bar */}
      <div
        className="flex items-center gap-1.5 px-4 py-3 border-b"
        style={{ background: '#0d0e15', borderColor: '#2e2e2e' }}
      >
        <span className="w-2.5 h-2.5 rounded-full" style={{ background: '#ff5f57' }} />
        <span className="w-2.5 h-2.5 rounded-full" style={{ background: '#febc2e' }} />
        <span className="w-2.5 h-2.5 rounded-full" style={{ background: '#28c840' }} />
        <span className="ml-2 text-[10px] font-mono" style={{ color: '#666666' }}>airvoice</span>
      </div>

      {/* Body */}
      <div className="p-4 flex flex-col items-start">
        <div className="mb-4">
          <img
            src={`${import.meta.env.BASE_URL || '/'}qrcode.svg`.replace(/\/+/g, '/')}
            alt="Airvoice QR code"
            className="w-[200px] h-[200px] block"
            style={{ filter: 'invert(1)' }}
          />
        </div>
        <div className="w-full font-mono text-[9px] leading-relaxed space-y-1 text-left">
          <div>
            <span style={{ color: '#666666' }}>Token:</span>{' '}
            <span style={{ color: '#ededed' }}>277129e4-35ea-40af-a122-13a5839e5e1f</span>
          </div>
          <div>
            <span style={{ color: '#666666' }}>WebSocket URL:</span>{' '}
            <span style={{ color: '#ededed' }}>ws://192.168.20.189:7654/ws</span>
          </div>
          <div>
            <span style={{ color: '#666666' }}>Status:</span>{' '}
            <span style={{ color: '#00ac3a' }}>Connected</span>
          </div>
          <div>
            <span style={{ color: '#666666' }}>Device:</span>{' '}
            <span style={{ color: '#ededed' }}>Pixel 8 Pro</span>
          </div>
          <div className="pt-2" style={{ color: '#666666' }}>
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
      className="rounded-2xl overflow-hidden"
      style={{ width: '170px', border: '1px solid #2e2e2e' }}
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
      className="rounded-2xl overflow-hidden"
      style={{ width: '170px', border: '1px solid #2e2e2e' }}
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
  const rightRef = useScrollReveal(0.05);

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
                padding: '3px 12px',
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
                className="inline-flex items-center justify-center gap-2 rounded-full px-6 text-sm font-semibold text-white transition-opacity hover:opacity-90"
                style={{ background: '#006efe', height: '48px' }}
              >
                <DownloadIcon />
                {isZh ? '下载 CLI' : 'Download CLI'}
              </a>
              <a
                href={urls.desktop}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center gap-2 rounded-full px-6 text-sm font-semibold transition-colors hover:bg-white/10"
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
              <div
                className="inline-flex rounded-full p-1"
                style={{ background: '#0d0e15', border: '1px solid #2e2e2e' }}
              >
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
                    {tab === 'cli'
                      ? (isZh ? '命令行 CLI' : 'CLI')
                      : (isZh ? '桌面端 Desktop' : 'Desktop')}
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
