import React from 'react';
import type { FeaturesMessages } from '../../i18n/messages';

interface FeaturesProps {
  m: FeaturesMessages;
}

// Icons matching the design exactly (24x24)
function QRIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
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

function SyncIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" aria-hidden="true">
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
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="5" y="11" width="14" height="11" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
      <circle cx="12" cy="16" r="1.5" fill="currentColor" stroke="none" />
    </svg>
  );
}

function TerminalIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="2" y="4" width="20" height="16" rx="2" />
      <path d="M7 9l3 3-3 3" />
      <path d="M13 15h4" />
    </svg>
  );
}

function DesktopIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
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

export default function Features({ m }: FeaturesProps) {
  const features: FeatureCard[] = [
    { icon: <QRIcon />, title: m.pairing.title, desc: m.pairing.desc },
    { icon: <SyncIcon />, title: m.sync.title, desc: m.sync.desc },
    { icon: <LockIcon />, title: m.secure.title, desc: m.secure.desc },
    { icon: <TerminalIcon />, title: m.cli.title, desc: m.cli.desc },
    { icon: <DesktopIcon />, title: m.desktop.title, desc: m.desktop.desc },
  ];

  const steps = [
    { step: '01', text: m.step1 },
    { step: '02', text: m.step2 },
    { step: '03', text: m.step3 },
  ];

  return (
    <section className="border-t py-16 md:py-20" style={{ borderColor: 'var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <h2
          className="mb-10 text-center text-2xl font-bold"
          style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}
        >
          {m.title}
        </h2>

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
          {/* Easy Pairing — tall (rowspan 2) */}
          <div
            className="rounded-xl p-4 flex flex-col"
            style={{ gridArea: 'pairing', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
          >
            <div className="mb-2.5 inline-flex items-center justify-center rounded-lg p-2 self-start"
              style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
              {features[0].icon}
            </div>
            <h3 className="mb-1 text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>{features[0].title}</h3>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[0].desc}</p>
            
            {/* 3-step connection guide */}
            <div className="mt-auto pt-3 border-t space-y-1.5" style={{ borderColor: 'var(--border-default)' }}>
              {steps.map((item) => (
                <div key={item.step} className="flex gap-1.5 items-center whitespace-nowrap">
                  <span className="text-[10px] font-mono font-bold" style={{ color: 'var(--secondary-text)' }}>
                    {item.step}.
                  </span>
                  <p className="text-[11px]" style={{ color: 'var(--secondary-text)' }}>
                    {item.text}
                  </p>
                </div>
              ))}
            </div>
          </div>

          {/* Real-Time Sync — featured */}
          <div
            className="rounded-xl p-4 flex flex-col justify-between"
            style={{ gridArea: 'sync', background: 'var(--background-secondary)', border: '1px solid var(--border-default)' }}
          >
            <div className="flex justify-between items-start mb-2">
              <h3 className="text-base font-semibold" style={{ color: 'var(--primary-text)' }}>{features[1].title}</h3>
              <div className="inline-flex items-center justify-center rounded-lg p-2 shrink-0 ml-3"
                style={{ background: 'var(--background-primary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
                {features[1].icon}
              </div>
            </div>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[1].desc}</p>
          </div>

          {/* Secure & Local */}
          <div
            className="rounded-xl p-4 flex flex-col justify-between"
            style={{ gridArea: 'secure', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
          >
            <div className="flex justify-between items-start mb-2">
              <h3 className="text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>{features[2].title}</h3>
              <div className="inline-flex items-center justify-center rounded-lg p-2 shrink-0 ml-3"
                style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
                {features[2].icon}
              </div>
            </div>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[2].desc}</p>
          </div>

          {/* CLI Power */}
          <div
            className="rounded-xl p-4 flex flex-col justify-between"
            style={{ gridArea: 'cli', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
          >
            <div className="flex justify-between items-start mb-2">
              <h3 className="text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>{features[3].title}</h3>
              <div className="inline-flex items-center justify-center rounded-lg p-2 shrink-0 ml-3"
                style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
                {features[3].icon}
              </div>
            </div>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{features[3].desc}</p>
          </div>

          {/* Desktop App */}
          <div
            className="rounded-xl p-4 flex flex-col justify-between"
            style={{ gridArea: 'desktop', background: 'var(--background-primary)', border: '1px solid var(--border-default)' }}
          >
            <div className="flex justify-between items-start mb-2">
              <h3 className="text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>{features[4].title}</h3>
              <div className="inline-flex items-center justify-center rounded-lg p-2 shrink-0 ml-3"
                style={{ background: 'var(--background-secondary)', color: 'var(--primary-text)', border: '1px solid var(--border-default)' }}>
                {features[4].icon}
              </div>
            </div>
            <p className="text-xs leading-relaxed sm:whitespace-nowrap" style={{ color: 'var(--secondary-text)' }}>{features[4].desc}</p>
          </div>
        </div>

        {/* Mobile View: Stacked Vertical Layout */}
        <div className="sm:hidden flex flex-col gap-4">
          {features.map((f, idx) => (
            <div
              key={f.title}
              className="w-full rounded-xl p-4 flex flex-col justify-between"
              style={{
                background: idx === 1 ? 'var(--background-secondary)' : 'var(--background-primary)',
                border: '1px solid var(--border-default)',
                minHeight: '110px',
              }}
            >
              <div>
                <div className="flex justify-between items-start mb-2">
                  <h3 className="text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>{f.title}</h3>
                  <div className="inline-flex items-center justify-center rounded-lg p-2 shrink-0 ml-3"
                    style={{
                      background: idx === 1 ? 'var(--background-primary)' : 'var(--background-secondary)',
                      color: 'var(--primary-text)',
                      border: '1px solid var(--border-default)'
                    }}
                  >
                    {f.icon}
                  </div>
                </div>
                <p className="text-xs leading-relaxed" style={{ color: 'var(--secondary-text)' }}>{f.desc}</p>
              </div>

              {/* If it is the first card (pairing), show the 3 steps below it */}
              {idx === 0 && (
                <div className="mt-4 pt-3 border-t space-y-1.5" style={{ borderColor: 'var(--border-default)' }}>
                  {steps.map((item) => (
                    <div key={item.step} className="flex gap-1.5 items-center whitespace-nowrap">
                      <span className="text-[10px] font-mono font-bold" style={{ color: 'var(--secondary-text)' }}>
                        {item.step}.
                      </span>
                      <p className="text-[11px]" style={{ color: 'var(--secondary-text)' }}>
                        {item.text}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
