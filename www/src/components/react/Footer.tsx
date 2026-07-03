import { ANDROID_APK_URL } from '../../lib/downloads';
import type { FooterMessages } from '../../i18n/messages';

interface FooterProps {
  m: FooterMessages;
  paths: {
    docs: string;
    quickStart: string;
    architecture: string;
  };
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

export default function Footer({ m, paths }: FooterProps) {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="border-t" style={{ backgroundColor: 'var(--background-secondary)', borderColor: 'var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6 py-12">
        <div className="grid grid-cols-1 gap-10 sm:grid-cols-3">
          <div className="order-2 sm:order-1">
            <div className="flex items-center gap-2 mb-3" style={{ color: 'var(--primary-text)' }}>
              <SoundwaveIcon />
              <span className="font-semibold text-sm">Airvoice</span>
            </div>
            <p className="text-xs leading-relaxed" style={{ color: 'var(--muted-text)', maxWidth: '200px' }}>
              {m.tagline}
            </p>
            <p className="mt-4 text-xs" style={{ color: 'var(--muted-text)' }}>
              © {currentYear} Airvoice. {m.rights}
            </p>
          </div>
          {/* Links Columns: 2 columns on mobile, spanning 2 columns on desktop */}
          <div className="order-1 sm:order-2 grid grid-cols-2 gap-8 sm:col-span-2">
            {/* Product Column */}
            <div>
              <h4 className="mb-4 text-[11px] font-semibold uppercase tracking-wider" style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}>
                {m.product}
              </h4>
              <ul className="space-y-3">
                {[
                  { href: 'https://github.com/yuler/airvoice/releases/latest', label: m.downloadCli },
                  { href: 'https://github.com/yuler/airvoice/releases/latest', label: m.downloadDesktop },
                  { href: ANDROID_APK_URL, label: m.downloadApk },
                  { href: 'https://github.com/yuler/airvoice/releases', label: m.releases },
                ].map((l) => (
                  <li key={l.label}>
                    <a href={l.href} target="_blank" rel="noopener noreferrer" className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>
                      {l.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>

            {/* Resources Column */}
            <div>
              <h4 className="mb-4 text-[11px] font-semibold uppercase tracking-wider" style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}>
                {m.resources}
              </h4>
              <ul className="space-y-3">
                <li>
                  <a href={paths.docs} className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>{m.documentation}</a>
                </li>
                <li>
                  <a href={paths.quickStart} className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>{m.quickStart}</a>
                </li>
                <li>
                  <a href={paths.architecture} className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>{m.architecture}</a>
                </li>
                <li>
                  <a href="https://github.com/yuler/airvoice" target="_blank" rel="noopener noreferrer" className="text-sm hover:underline transition-all" style={{ color: 'var(--secondary-text)' }}>GitHub</a>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
