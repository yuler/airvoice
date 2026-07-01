import React from 'react';

interface OpenSourceBannerProps {
  lang: 'en' | 'zh';
}

function GithubLogo() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.9-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.9 1.52 2.34 1.07 2.91.83.1-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z" />
    </svg>
  );
}

export default function OpenSourceBanner({ lang }: OpenSourceBannerProps) {
  const isZh = lang === 'zh';

  return (
    <section className="pb-16 md:pb-24">
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <div
          className="rounded-2xl border p-8 md:p-10 flex flex-col md:flex-row justify-between items-start md:items-center gap-6"
          style={{ background: 'var(--background-secondary)', borderColor: 'var(--border-default)' }}
        >
          {/* Text content */}
          <div>
            <p className="text-lg md:text-xl font-semibold" style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}>
              {isZh ? '共同帮助 Airvoice 变得更好。' : 'Help make Airvoice better.'}
            </p>
          </div>

          {/* View on GitHub Button */}
          <a
            href="https://github.com/yuler/airvoice"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 rounded-xl border px-5 py-3 text-sm font-semibold shadow-sm transition-colors hover:bg-black/5 dark:hover:bg-white/10"
            style={{ background: 'var(--background-primary)', borderColor: 'var(--border-default)', color: 'var(--primary-text)' }}
          >
            <GithubLogo />
            {isZh ? '在 GitHub 上查看' : 'View on GitHub'}
          </a>
        </div>
      </div>
    </section>
  );
}
