import React from 'react';

interface WorksEverywhereProps {
  lang: 'en' | 'zh';
}

function AppleIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor" aria-label="Apple">
      <path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701" />
    </svg>
  );
}

function AndroidIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="#3ddc84" aria-label="Android">
      <path d="M17.6 9.48l1.84-3.18c.16-.31.04-.69-.26-.85a.637.637 0 0 0-.83.22l-1.88 3.24C14.79 8.33 13.44 8 12 8s-2.79.33-4.47.91L5.65 5.67a.637.637 0 0 0-.83-.22c-.3.16-.42.54-.26.85L6.4 9.48C3.3 11.25 1.28 14.44 1 18h22c-.28-3.56-2.3-6.75-5.4-8.52zM7 15.25a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5zm10 0a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5z" />
    </svg>
  );
}

function FinderIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" aria-label="macOS">
      <mask id="finder-mask-we">
        <rect x="2" y="2" width="20" height="20" rx="4" fill="white" />
      </mask>
      <g mask="url(#finder-mask-we)">
        <rect x="2" y="2" width="20" height="20" fill="#E0F2FE" />
        <rect x="2" y="2" width="10" height="20" fill="#38BDF8" />
        <path d="M12 2v10c0 1.1-.9 2-2 2H8" stroke="#0369A1" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        <circle cx="7.5" cy="8.5" r="1.5" fill="#0369A1" />
        <circle cx="16.5" cy="8.5" r="1.5" fill="#0369A1" />
        <path d="M7 17c1.5 2 5.5 2 7 0" stroke="#0369A1" strokeWidth="2" strokeLinecap="round" />
      </g>
      <rect x="2" y="2" width="20" height="20" rx="4" stroke="#0284C7" strokeWidth="2" />
    </svg>
  );
}

function WindowsIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="#0078d4" aria-label="Windows">
      <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.549H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801" />
    </svg>
  );
}

function TuxIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" aria-label="Linux">
      {/* Feet */}
      <path d="M5 19.5c0-.8 1-1.5 2.5-1.5h1c.8 0 1.5.3 1.8.8.4-.5 1.1-.8 1.8-.8h1c1.5 0 2.5.7 2.5 1.5 0 1-.8 1.5-2.5 1.5h-1.3c-.3 0-.6-.2-.7-.5-.2.3-.5.5-.9.5H7.5C5.8 21 5 20.5 5 19.5z" fill="#FFA000"/>
      {/* Flippers (back) */}
      <path d="M3.5 14.5c0-1.8.8-3.5 2.2-4.5.3.5.7 1 1.2 1.3-1 1-1.6 2-1.6 3.2 0 1.5.8 2.8 2 3.5-.5.2-1 .5-1.5.9C4.6 18 3.5 16.4 3.5 14.5z" fill="#2D3748"/>
      <path d="M20.5 14.5c0 1.9-1.1 3.5-2.3 4.4-.5-.4-1-.7-1.5-.9 1.2-.7 2-2 2-3.5 0-1.2-.6-2.2-1.6-3.2.5-.3.9-.8 1.2-1.3 1.4 1 2.2 2.7 2.2 4.5z" fill="#2D3748"/>
      {/* Body/Head (Dark Grey/Black) */}
      <path d="M12 2c-3 0-5.5 2.5-5.5 6v4.5c0 1.5.5 2.8 1.5 3.8C7 17.5 6 19.5 6 21c0 .6.4 1 1 1h10c.6 0 1-.4 1-1 0-1.5-1-3.5-2-4.7 1-1 1.5-2.3 1.5-3.8V8c0-3.5-2.5-6-5-6z" fill="#1A202C"/>
      {/* Belly/Face background (White) */}
      <path d="M12 4.5c-2.2 0-4 1.8-4 4.5v3.5c0 2 1.5 3.5 4 4.5 2.5-1 4-2.5 4-4.5V9c0-2.7-1.8-4.5-4-4.5z" fill="#FFFFFF"/>
      {/* Eyes */}
      <circle cx="10" cy="8.5" r="1.5" fill="#FFFFFF"/>
      <circle cx="10" cy="8.5" r="0.75" fill="#1A202C"/>
      <circle cx="14" cy="8.5" r="1.5" fill="#FFFFFF"/>
      <circle cx="14" cy="8.5" r="0.75" fill="#1A202C"/>
      {/* Beak */}
      <path d="M10 10.5c0-1.1.9-2 2-2s2 .9 2 2H10z" fill="#FFB300"/>
      <path d="M10 10.5c0 1.1.9 2 2 2s2-.9 2-2H10z" fill="#FFA000"/>
    </svg>
  );
}

export default function WorksEverywhere({ lang }: WorksEverywhereProps) {
  const isZh = lang === 'zh';
  
  const platforms = [
    { icon: <AppleIcon />, name: 'iOS' },
    { icon: <AndroidIcon />, name: 'Android' },
    { icon: <FinderIcon />, name: 'macOS' },
    { icon: <WindowsIcon />, name: 'Windows' },
    { icon: <TuxIcon />, name: 'Linux' },
  ];

  return (
    <section className="py-12 md:py-16" style={{ background: 'var(--background-secondary)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <h2 className="text-center text-2xl font-bold mb-10 text-kumo-default" style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}>
          {isZh ? '多平台运行' : 'Works everywhere'}
        </h2>
        
        <div className="flex flex-wrap justify-center gap-4">
          {platforms.map((p) => (
            <div
              key={p.name}
              className="flex items-center gap-3 rounded-xl px-6 py-4 min-w-[160px] shadow-sm hover:shadow-md transition-shadow"
              style={{
                background: 'var(--background-primary)',
                border: '1px solid var(--border-default)',
              }}
            >
              <div className="flex-shrink-0" style={{ color: 'var(--primary-text)' }}>
                {p.icon}
              </div>
              <span className="text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>{p.name}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
