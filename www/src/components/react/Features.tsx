import { useState } from 'react';

interface FeaturesProps {
  lang: 'en' | 'zh';
}

// Icons matching the design exactly
function QRIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="3" y="3" width="7" height="7" rx="1" />
      <rect x="14" y="3" width="7" height="7" rx="1" />
      <rect x="3" y="14" width="7" height="7" rx="1" />
      <rect x="5" y="5" width="3" height="3" fill="currentColor" stroke="none" />
      <rect x="16" y="5" width="3" height="3" fill="currentColor" stroke="none" />
      <rect x="5" y="16" width="3" height="3" fill="currentColor" stroke="none" />
      <path d="M14 14h2v2h-2zM18 14h3M14 18h2M18 18h3M20 14v3M14 20h6" />
    </svg>
  );
}

// Custom soundwave icon matching the visual design diagram:
// It has a series of vertical rounded bars (like a symmetric wave/pulse)
function SyncIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" aria-hidden="true">
      <line x1="4" y1="10" x2="4" y2="14" />
      <line x1="8" y1="7" x2="8" y2="17" />
      <line x1="12" y1="4" x2="12" y2="20" />
      <line x1="16" y1="7" x2="16" y2="17" />
      <line x1="20" y1="10" x2="20" y2="14" />
    </svg>
  );
}

function LockIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="5" y="11" width="14" height="11" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
      <circle cx="12" cy="16" r="1.5" fill="currentColor" stroke="none" />
    </svg>
  );
}

function TerminalIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="4" width="20" height="16" rx="2" />
      <path d="M7 9l3 3-3 3" />
      <path d="M13 15h4" />
    </svg>
  );
}

function DesktopIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <line x1="8" y1="21" x2="16" y2="21" />
      <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}

interface FeatureCard {
  icon: React.ReactNode;
  title: string;
  desc: string;
}

export default function Features({ lang }: FeaturesProps) {
  const [activeIdx, setActiveIdx] = useState(0);
  const isZh = lang === 'zh';

  const features: FeatureCard[] = [
    {
      icon: <QRIcon />,
      title: isZh ? '轻松配对' : 'Easy Pairing',
      desc: isZh
        ? '扫描二维码或手动输入配对码，秒级完成连接。'
        : 'Pair in seconds with QR code or manual code.',
    },
    {
      icon: <SyncIcon />,
      title: isZh ? '实时同步' : 'Real-Time Sync',
      desc: isZh
        ? '通过 WebSocket 即时传输，可靠回退机制保障稳定。'
        : 'Instant delivery over WebSocket with reliable fallback.',
    },
    {
      icon: <LockIcon />,
      title: isZh ? '安全且本地' : 'Secure & Local',
      desc: isZh
        ? '连接保留在局域网内，无云端，无追踪。'
        : 'Connections stay on your LAN. No cloud. No tracking.',
    },
    {
      icon: <TerminalIcon />,
      title: isZh ? 'CLI 强力支持' : 'CLI Power',
      desc: isZh
        ? '使用 Airvoice 命令行工具自动化你的工作流程。'
        : 'Automate workflows with the Airvoice command line.',
    },
    {
      icon: <DesktopIcon />,
      title: isZh ? '桌面端 App' : 'Desktop App',
      desc: isZh
        ? '提供可视化状态面板、连接配置与历史记录查看。'
        : 'Visual status panel, connection configuration and history viewer.',
    },
  ];

  return (
    <section className="border-t py-16 md:py-20" style={{ borderColor: 'var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        {/* Section title */}
        <h2
          className="mb-10 text-center text-2xl font-bold"
          style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}
        >
          {isZh ? '您所需的一切' : 'Everything you need'}
        </h2>

        {/* Desktop View: Grid layout */}
        <div className="hidden sm:grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
          {features.map((f) => (
            <div
              key={f.title}
              className="rounded-xl p-6 transition-shadow hover:shadow-sm"
              style={{
                background: 'var(--background-primary)',
                border: '1px solid var(--border-default)',
              }}
            >
              {/* Icon */}
              <div
                className="mb-4 inline-flex items-center justify-center rounded-xl p-2.5"
                style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}
              >
                {f.icon}
              </div>

              {/* Title */}
              <h3 className="mb-1.5 text-base font-semibold" style={{ color: 'var(--primary-text)' }}>
                {f.title}
              </h3>

              {/* Description */}
              <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>
                {f.desc}
              </p>
            </div>
          ))}
        </div>

        {/* Mobile View: Carousel slider */}
        <div className="sm:hidden flex flex-col items-center">
          {/* Active Card */}
          <div
            className="w-full rounded-xl p-6 mb-6"
            style={{
              background: 'var(--background-primary)',
              border: '1px solid var(--border-default)',
              minHeight: '180px',
            }}
          >
            {/* Icon */}
            <div
              className="mb-4 inline-flex items-center justify-center rounded-xl p-2.5"
              style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}
            >
              {features[activeIdx].icon}
            </div>

            {/* Title */}
            <h3 className="mb-1.5 text-base font-semibold" style={{ color: 'var(--primary-text)' }}>
              {features[activeIdx].title}
            </h3>

            {/* Description */}
            <p className="text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>
              {features[activeIdx].desc}
            </p>
          </div>

          {/* Dots Indicator */}
          <div className="flex gap-2">
            {features.map((_, idx) => (
              <button
                key={idx}
                onClick={() => setActiveIdx(idx)}
                className="w-2.5 h-2.5 rounded-full transition-all duration-300"
                style={{
                  background: activeIdx === idx ? 'var(--accent-blue)' : 'var(--border-default)',
                  border: 'none',
                  cursor: 'pointer',
                  padding: 0,
                }}
                aria-label={`Slide ${idx + 1}`}
              />
            ))}
          </div>
        </div>

      </div>
    </section>
  );
}
