import React from 'react';

interface FooterProps {
  lang: 'en' | 'zh';
}

export default function Footer({ lang }: FooterProps) {
  const isZh = lang === 'zh';
  const currentYear = new Date().getFullYear();

  return (
    <footer className="border-t py-12" style={{ backgroundColor: 'var(--background-secondary)', borderColor: 'var(--border-default)' }}>
      <div className="mx-auto max-w-6xl px-4 md:px-6 flex flex-col md:flex-row justify-between items-center gap-6">
        {/* Left side: branding & copyright */}
        <div className="flex flex-col items-center md:items-start gap-1">
          <span className="text-sm font-semibold" style={{ color: 'var(--primary-text)' }}>Airvoice</span>
          <span className="text-xs" style={{ color: 'var(--muted-text)' }}>
            © {currentYear} Airvoice. {isZh ? '保留所有权利。' : 'All rights reserved.'}
          </span>
        </div>

        {/* Right side: links */}
        <div className="flex flex-wrap justify-center gap-x-6 gap-y-2 text-sm">
          <a
            href={`${import.meta.env.BASE_URL || '/'}${isZh ? 'zh/docs/background/' : 'docs/background/'}`.replace(/\/+/g, '/')}
            className="hover:underline transition-all"
            style={{ color: 'var(--secondary-text)' }}
          >
            {isZh ? '指南与文档' : 'Documentation'}
          </a>
          <a
            href="https://github.com/yuler/airvoice"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline transition-all"
            style={{ color: 'var(--secondary-text)' }}
          >
            GitHub
          </a>
          <a
            href="https://github.com/yuler/airvoice/releases"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline transition-all"
            style={{ color: 'var(--secondary-text)' }}
          >
            {isZh ? '版本发布' : 'Releases'}
          </a>
          <a
            href="https://github.com/yuler/airvoice/blob/main/LICENSE"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:underline transition-all"
            style={{ color: 'var(--secondary-text)' }}
          >
            {isZh ? '许可证 (MIT)' : 'License'}
          </a>
        </div>
      </div>
    </footer>
  );
}
