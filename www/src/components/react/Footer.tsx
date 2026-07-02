interface FooterProps {
  lang: 'en' | 'zh';
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

export default function Footer({ lang }: FooterProps) {
  const isZh = lang === 'zh';
  const currentYear = new Date().getFullYear();
  const base = import.meta.env.BASE_URL || '/';
  const loc = (path: string) =>
    isZh
      ? `${base}zh/${path}`.replace(/\/+/g, '/')
      : `${base}${path}`.replace(/\/+/g, '/');

  return (
    <footer className="border-t" style={{ backgroundColor: 'var(--background-secondary)', borderColor: 'var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6 py-12">
        <div className="grid grid-cols-1 gap-10 sm:grid-cols-3">

          {/* Branding */}
          <div>
            <div className="flex items-center gap-2 mb-3" style={{ color: 'var(--primary-text)' }}>
              <SoundwaveIcon />
              <span className="font-semibold text-sm">Airvoice</span>
            </div>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--muted-text)', maxWidth: '200px' }}>
              {isZh
                ? '跨设备通信的统一桥梁。'
                : 'A unified bridge for cross-device communication.'}
            </p>
            <p className="mt-4 text-xs" style={{ color: 'var(--muted-text)' }}>
              © {currentYear} Airvoice.{' '}
              {isZh ? '保留所有权利。' : 'All rights reserved.'}
            </p>
          </div>

          {/* Product */}
          <div>
            <h4
              className="mb-4 text-[11px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}
            >
              {isZh ? '产品' : 'Product'}
            </h4>
            <ul className="space-y-3">
              {[
                { href: 'https://github.com/yuler/airvoice/releases/latest', label: isZh ? '下载 CLI' : 'Download CLI', external: true },
                { href: 'https://github.com/yuler/airvoice/releases/latest', label: isZh ? '下载桌面版' : 'Download Desktop', external: true },
                { href: 'https://github.com/yuler/airvoice/releases/latest', label: isZh ? '移动端应用' : 'Mobile App', external: true },
                { href: 'https://github.com/yuler/airvoice/releases', label: isZh ? '版本发布' : 'Releases', external: true },
              ].map((l) => (
                <li key={l.label}>
                  <a
                    href={l.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm hover:underline transition-all"
                    style={{ color: 'var(--secondary-text)' }}
                  >
                    {l.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          {/* Resources */}
          <div>
            <h4
              className="mb-4 text-[11px] font-semibold uppercase tracking-wider"
              style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}
            >
              {isZh ? '资源' : 'Resources'}
            </h4>
            <ul className="space-y-3">
              <li>
                <a href={loc('docs/background/')} className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>
                  {isZh ? '文档' : 'Documentation'}
                </a>
              </li>
              <li>
                <a href={loc('docs/quick-start/')} className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>
                  {isZh ? '快速开始' : 'Quick Start'}
                </a>
              </li>
              <li>
                <a href={loc('docs/architecture/')} className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>
                  {isZh ? '架构' : 'Architecture'}
                </a>
              </li>
              <li>
                <a href="https://github.com/yuler/airvoice" target="_blank" rel="noopener noreferrer" className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>
                  GitHub
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </footer>
  );
}
