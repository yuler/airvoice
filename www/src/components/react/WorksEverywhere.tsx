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
    <svg width="24" height="24" viewBox="0 0 256 256" aria-label="Linux">
      <path fill="#DD4814" d="M255.637 127.683c0 70.514-57.165 127.68-127.683 127.68C57.434 255.363.27 198.197.27 127.683C.27 57.165 57.436 0 127.954 0c70.519 0 127.683 57.165 127.683 127.683"/>
      <path fill="#FFF" d="M41.133 110.633c-9.419 0-17.05 7.631-17.05 17.05c0 9.414 7.631 17.046 17.05 17.046c9.415 0 17.046-7.632 17.046-17.046c0-9.419-7.631-17.05-17.046-17.05m121.715 77.478c-8.153 4.71-10.95 15.13-6.24 23.279c4.705 8.154 15.125 10.949 23.279 6.24c8.153-4.705 10.949-15.125 6.24-23.28c-4.705-8.148-15.131-10.943-23.279-6.239m-84.686-60.428c0-16.846 8.368-31.73 21.171-40.742L86.87 66.067c-14.914 9.97-26.012 25.204-30.624 43.047c5.382 4.39 8.826 11.075 8.826 18.568c0 7.489-3.444 14.174-8.826 18.565C60.852 164.094 71.95 179.33 86.87 189.3l12.463-20.88c-12.803-9.007-21.171-23.89-21.171-40.737m49.792-49.797c26.013 0 47.355 19.944 49.595 45.38l24.29-.358c-1.194-18.778-9.398-35.636-22.002-48.032c-6.482 2.449-13.97 2.074-20.44-1.656c-6.483-3.741-10.548-10.052-11.659-16.902a74.3 74.3 0 0 0-19.785-2.69a73.8 73.8 0 0 0-32.819 7.663l11.845 21.227a49.6 49.6 0 0 1 20.975-4.632m0 99.59a49.6 49.6 0 0 1-20.974-4.632l-11.845 21.225a73.7 73.7 0 0 0 32.82 7.671a74 74 0 0 0 19.784-2.697c1.111-6.85 5.177-13.155 11.658-16.902c6.476-3.737 13.959-4.105 20.44-1.656c12.605-12.396 20.808-29.254 22.004-48.032l-24.297-.358c-2.235 25.443-23.576 45.38-49.59 45.38m34.888-110.231c8.154 4.708 18.575 1.92 23.279-6.234c4.71-8.154 1.92-18.575-6.234-23.285c-8.154-4.704-18.574-1.91-23.285 6.244c-4.703 8.15-1.908 18.57 6.24 23.275"/>
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
