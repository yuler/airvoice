import { SunIcon, MoonIcon, GithubLogoIcon, CaretDownIcon } from '@phosphor-icons/react';
import { Button } from '@cloudflare/kumo';
import { useState, useEffect, useRef } from 'react';


interface HeaderProps {
  lang: 'en' | 'zh';
  base: string;
  active?: 'home' | 'docs';
}

export default function Header({ lang, base, active = 'home' }: HeaderProps) {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');
  const [langOpen, setLangOpen] = useState(false);
  const langRef = useRef<HTMLDivElement>(null);
  const isZh = lang === 'zh';

  useEffect(() => {
    const stored = localStorage.getItem('theme') as 'light' | 'dark' | null;
    if (stored) {
      setTheme(stored);
    } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      setTheme('dark');
    }
  }, []);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (langRef.current && !langRef.current.contains(e.target as Node)) {
        setLangOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  const toggleTheme = () => {
    const next = theme === 'dark' ? 'light' : 'dark';
    setTheme(next);
    localStorage.setItem('theme', next);
    document.documentElement.setAttribute('data-mode', next);
  };

  return (
    <header className="sticky top-0 z-50 h-12 border-b border-kumo-hairline bg-kumo-canvas/80 backdrop-blur-sm">
      <div className="mx-auto flex h-full max-w-6xl items-center justify-between px-6">
        <div className="flex items-center gap-8">
          <a href={loc('')} className="text-lg font-bold text-kumo-default hover:text-kumo-subtle transition-colors">
            Airvoice
          </a>
          <nav className="hidden items-center gap-6 md:flex">
            <a
              href={loc('docs/background/')}
              className={`text-sm transition-colors ${
                active === 'docs'
                  ? 'text-kumo-default font-medium'
                  : 'text-kumo-subtle hover:text-kumo-default'
              }`}
            >
              {isZh ? '文档' : 'Docs'}
            </a>
          </nav>
        </div>

        <div className="flex items-center gap-1">
          <a
            href="https://github.com/yuler/airvoice"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center rounded-md p-1.5 text-kumo-subtle hover:text-kumo-default transition-colors"
          >
            <GithubLogoIcon size={16} />
          </a>
          <Button
            variant="ghost"
            shape="square"
            size="sm"
            icon={theme === 'dark' ? SunIcon : MoonIcon}
            onClick={toggleTheme}
            aria-label={isZh ? '切换主题' : 'Toggle theme'}
          />
          <div ref={langRef} className="relative">
            <button
              onClick={() => setLangOpen(!langOpen)}
              className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-kumo-subtle hover:text-kumo-default transition-colors"
            >
              <span>{lang === 'en' ? '🇺🇸' : '🇨🇳'}</span>
              <span className="hidden sm:inline text-xs">{lang === 'en' ? 'EN' : '中文'}</span>
              <CaretDownIcon size={12} />
            </button>
            {langOpen && (
              <div className="absolute right-0 top-full mt-1 w-36 rounded-lg border border-kumo-hairline bg-kumo-canvas shadow-lg py-1">
                <a
                  href={`${base}${active === 'docs' ? 'docs/background/' : ''}`}
                  onClick={() => setLangOpen(false)}
                  className={`flex items-center gap-2.5 px-3 py-2 text-sm transition-colors ${
                    lang === 'en' ? 'text-kumo-default bg-kumo-control' : 'text-kumo-subtle hover:text-kumo-default hover:bg-kumo-control'
                  }`}
                >
                  🇺🇸 English
                </a>
                <a
                  href={`${base}zh/${active === 'docs' ? 'docs/background/' : ''}`}
                  onClick={() => setLangOpen(false)}
                  className={`flex items-center gap-2.5 px-3 py-2 text-sm transition-colors ${
                    lang === 'zh' ? 'text-kumo-default bg-kumo-control' : 'text-kumo-subtle hover:text-kumo-default hover:bg-kumo-control'
                  }`}
                >
                  🇨🇳 中文
                </a>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
