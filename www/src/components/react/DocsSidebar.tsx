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
      <path d="M2 10v4" />
      <path d="M6 6v12" />
      <path d="M10 3v18" />
      <path d="M14 6v12" />
      <path d="M18 10v4" />
      <path d="M22 12v0" />
    </svg>
  );
}

export default function DocsSidebar({ lang, base, currentPath }: SidebarProps) {
  const [open, setOpen] = useState(false);
  const isZh = lang === 'zh';
  const normalizedBase = base.endsWith('/') ? base : `${base}/`;
  const loc = (path: string) =>
    lang === 'en'
      ? `${normalizedBase}${path}`
      : `${normalizedBase}zh/${path}`;

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
      {/* Mobile toggle */}
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
        <div
          className="flex h-[52px] items-center border-b px-4"
          style={{ borderColor: 'var(--border-default)' }}
        >
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
                className="mb-2 uppercase"
                style={{
                  fontSize: '11px',
                  fontWeight: 600,
                  letterSpacing: '0.08em',
                  color: 'var(--muted-text)',
                }}
              >
                {section.title}
              </h3>
              <ul className="space-y-0.5">
                {section.items.map((item) => {
                  const isActive =
                    currentPath === item.href ||
                    currentPath === item.href.replace(/\/$/, '') ||
                    (item.href !== normalizedBase && currentPath.startsWith(item.href));
                  return (
                    <li key={item.href}>
                      <a
                        href={item.href}
                        onClick={() => setOpen(false)}
                        className="block py-2 text-sm transition-colors"
                        style={
                          isActive
                            ? {
                                borderLeft: '2px solid #006efe',
                                borderRadius: '0 6px 6px 0',
                                paddingLeft: '10px',
                                paddingRight: '12px',
                                color: 'var(--primary-text)',
                                fontWeight: 500,
                                background: 'transparent',
                                textDecoration: 'none',
                              }
                            : {
                                paddingLeft: '12px',
                                paddingRight: '12px',
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
