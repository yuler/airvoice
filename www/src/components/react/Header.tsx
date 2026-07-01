import { useState, useEffect } from 'react';

interface HeaderProps {
  lang: 'en' | 'zh';
  base: string;
  active?: 'home' | 'docs';
  currentPath?: string;
}

function SoundwaveIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M2 10v4" />
      <path d="M6 6v12" />
      <path d="M10 3v18" />
      <path d="M14 6v12" />
      <path d="M18 10v4" />
      <path d="M22 12v0" />
    </svg>
  );
}

function BookIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" className="mr-1.5">
      <path d="M4 19.5v-15A2.5 2.5 0 0 1 6.5 2H20v20H6.5a2.5 2.5 0 0 1-2.5-2.5Z" />
      <path d="M6 6h10" />
      <path d="M6 10h10" />
    </svg>
  );
}

function GithubIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className="mr-1.5">
      <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/>
    </svg>
  );
}

function HamburgerIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="3" y1="6" x2="21" y2="6" />
      <line x1="3" y1="12" x2="21" y2="12" />
      <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

function SunIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2" />
      <path d="M12 20v2" />
      <path d="M4.93 4.93l1.41 1.41" />
      <path d="M17.66 17.66l1.41 1.41" />
      <path d="M2 12h2" />
      <path d="M20 12h2" />
      <path d="M6.34 17.66l-1.41 1.41" />
      <path d="M19.07 4.93l-1.41 1.41" />
    </svg>
  );
}

function MoonIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
    </svg>
  );
}

export default function Header({ lang, base, active = 'home', currentPath }: HeaderProps) {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [theme, setTheme] = useState<'light' | 'dark'>('light');
  const isZh = lang === 'zh';

  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  useEffect(() => {
    const htmlMode = document.documentElement.getAttribute('data-mode') as 'light' | 'dark' | null;
    if (htmlMode) {
      setTheme(htmlMode);
    } else {
      const stored = localStorage.getItem('theme');
      if (stored) {
        setTheme(stored as 'light' | 'dark');
      } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        setTheme('dark');
      }
    }
  }, []);

  const toggleTheme = () => {
    const nextTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(nextTheme);
    document.documentElement.setAttribute('data-mode', nextTheme);
    localStorage.setItem('theme', nextTheme);
  };

  const getLangTogglePath = () => {
    const path = currentPath || '/';
    if (lang === 'en') {
      if (path.startsWith(base + 'zh/')) return path;
      const subPath = path.startsWith(base) ? path.slice(base.length) : path;
      return `${base}zh/${subPath}`.replace(/\/+/g, '/');
    } else {
      const subPath = path.startsWith(base + 'zh/') ? path.slice((base + 'zh/').length) : path;
      return `${base}${subPath}`.replace(/\/+/g, '/');
    }
  };

  // Close mobile menu on outside click
  useEffect(() => {
    if (!mobileOpen) return;
    const handle = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (!target.closest('[data-mobile-nav]')) setMobileOpen(false);
    };
    document.addEventListener('mousedown', handle);
    return () => document.removeEventListener('mousedown', handle);
  }, [mobileOpen]);

  return (
    <header className="sticky top-0 z-50 border-b backdrop-blur-sm" style={{ backgroundColor: 'var(--background-primary)', borderColor: 'var(--border-default)', opacity: 0.95 }} data-mobile-nav>
      <div className="mx-auto flex h-[52px] max-w-6xl items-center justify-between px-4 md:px-6">
        {/* Logo */}
        <a href={loc('')} className="flex items-center gap-2 text-kumo-default hover:opacity-80 transition-opacity">
          <SoundwaveIcon />
          <span className="text-base font-semibold">Airvoice</span>
        </a>

        {/* Desktop Nav + CTA grouped on the right */}
        <div className="hidden md:flex items-center gap-5">
          <a
            href={loc('docs/background/')}
            className={`flex items-center text-sm transition-colors ${active === 'docs' ? 'text-kumo-default font-medium' : 'text-kumo-subtle hover:text-kumo-default'}`}
            title={isZh ? '文档' : 'Docs'}
          >
            <BookIcon />
          </a>
          <a
            href="https://github.com/yuler/airvoice"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center text-sm text-kumo-subtle hover:text-kumo-default transition-colors"
            title="GitHub"
          >
            <GithubIcon />
          </a>

          {/* Lang Toggle */}
          <a
            href={getLangTogglePath()}
            className="flex items-center justify-center text-base text-kumo-subtle hover:text-kumo-default transition-colors w-6 h-6 cursor-pointer"
            title={isZh ? 'Switch to English' : '切换至中文'}
          >
            {isZh ? '🇺🇸' : '🇨🇳'}
          </a>

          {/* Theme Toggle */}
          <button
            onClick={toggleTheme}
            className="flex items-center justify-center text-kumo-subtle hover:text-kumo-default transition-colors w-6 h-6 cursor-pointer bg-transparent border-none outline-none p-0"
            title={isZh ? '切换主题' : 'Toggle Theme'}
          >
            {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
          </button>

        </div>

        {/* Mobile hamburger */}
        <button
          className="flex md:hidden items-center justify-center p-1.5 rounded-md text-kumo-subtle hover:text-kumo-default transition-colors"
          onClick={() => setMobileOpen(!mobileOpen)}
          aria-label={isZh ? '菜单' : 'Menu'}
        >
          {mobileOpen ? <CloseIcon /> : <HamburgerIcon />}
        </button>
      </div>

      {/* Mobile dropdown */}
      {mobileOpen && (
        <div className="md:hidden border-t px-4 py-4 space-y-3" style={{ backgroundColor: 'var(--background-primary)', borderColor: 'var(--border-default)' }}>
          <a
            href={loc('docs/background/')}
            className="flex items-center text-sm text-kumo-subtle hover:text-kumo-default transition-colors py-1"
            onClick={() => setMobileOpen(false)}
          >
            <BookIcon />
            {isZh ? '文档' : 'Docs'}
          </a>
          <a
            href="https://github.com/yuler/airvoice"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center text-sm text-kumo-subtle hover:text-kumo-default transition-colors py-1"
            onClick={() => setMobileOpen(false)}
          >
            <GithubIcon />
            GitHub
          </a>

          {/* Lang/Theme switch for mobile */}
          <div className="flex items-center gap-4 py-2 border-t" style={{ borderColor: 'var(--border-default)' }}>
            <span className="text-xs text-kumo-inactive" style={{ color: 'var(--muted-text)' }}>{isZh ? '语言与主题:' : 'Lang & Theme:'}</span>
            <a
              href={getLangTogglePath()}
              className="flex items-center justify-center text-base text-kumo-subtle hover:text-kumo-default transition-colors w-6 h-6 cursor-pointer"
            >
              {isZh ? '🇺🇸' : '🇨🇳'}
            </a>
            <button
              onClick={toggleTheme}
              className="flex items-center justify-center text-kumo-subtle hover:text-kumo-default transition-colors w-6 h-6 cursor-pointer bg-transparent border-none outline-none p-0"
            >
              {theme === 'dark' ? <SunIcon /> : <MoonIcon />}
            </button>
          </div>

        </div>
      )}
    </header>
  );
}
