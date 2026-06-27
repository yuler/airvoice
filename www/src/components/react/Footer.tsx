interface FooterProps {
  lang: 'en' | 'zh';
  base: string;
}

export default function Footer({ lang, base }: FooterProps) {
  const isZh = lang === 'zh';
  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  return (
    <footer className="border-t border-kumo-hairline py-8">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-6 md:flex-row">
        <p className="text-xs text-kumo-subtle">
          © 2026 Airvoice. MIT License.
        </p>
        <div className="flex items-center gap-6">
          <a
            href="https://github.com/yuler/airvoice"
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-kumo-subtle hover:text-kumo-default transition-colors"
          >
            GitHub
          </a>
          <a
            href={loc('docs/background/')}
            className="text-xs text-kumo-subtle hover:text-kumo-default transition-colors"
          >
            {isZh ? '文档' : 'Docs'}
          </a>
        </div>
      </div>
    </footer>
  );
}
