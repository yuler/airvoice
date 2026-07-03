import { useState } from 'react';
import type { DocsMessages } from '../../i18n/messages';

interface SidebarProps {
  m: DocsMessages;
  currentPath: string;
  paths: {
    background: string;
    quickStart: string;
    development: string;
    architecture: string;
    platformDeps: string;
  };
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

export default function DocsSidebar({ m, currentPath, paths }: SidebarProps) {
  const [open, setOpen] = useState(false);

  const sections = [
    {
      title: m.overview,
      items: [
        { href: paths.background, label: m.background },
        { href: paths.quickStart, label: m.quickStart },
      ],
    },
    {
      title: m.guide,
      items: [
        { href: paths.development, label: m.development },
        { href: paths.architecture, label: m.architecture },
        { href: paths.platformDeps, label: m.platformDeps },
      ],
    },
  ];

  return (
    <>
      <button
        className="fixed left-4 top-3.5 z-50 md:hidden flex items-center justify-center rounded-md p-1.5 transition-colors"
        style={{ color: 'var(--secondary-text)', background: 'transparent', border: 'none', cursor: 'pointer' }}
        onClick={() => setOpen(!open)}
        aria-label={m.sidebarToggle}
      >
        {open ? <CloseIcon /> : <HamburgerIcon />}
      </button>
      {open && (
        <div
          className="fixed inset-0 z-30 md:hidden"
          style={{ background: 'rgba(0,0,0,0.5)' }}
          onClick={() => setOpen(false)}
        />
      )}
      <aside
        className={`fixed top-0 left-0 z-40 h-full w-64 overflow-auto transition-transform duration-300 ${
          open ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
        }`}
        style={{
          background: 'var(--background-secondary)',
          borderRight: '1px solid var(--border-default)',
        }}
      >
        <nav className="p-4 pt-16 md:pt-[68px]">
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
                    currentPath.startsWith(item.href.endsWith('/') ? item.href : `${item.href}/`);
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
