interface PaginationProps {
  lang: 'en' | 'zh';
  currentSlug: string;
  base: string;
}

const ORDER_EN = [
  { slug: 'background', label: 'Background' },
  { slug: 'quick-start', label: 'Quick Start' },
  { slug: 'development', label: 'Development' },
  { slug: 'architecture', label: 'Architecture' },
  { slug: 'platform-deps', label: 'Platform Deps' },
];

const ORDER_ZH = [
  { slug: 'background', label: '背景' },
  { slug: 'quick-start', label: '快速开始' },
  { slug: 'development', label: '开发指南' },
  { slug: 'architecture', label: '架构' },
  { slug: 'platform-deps', label: '平台依赖' },
];

export default function DocsPagination({ lang, currentSlug, base }: PaginationProps) {
  const order = lang === 'zh' ? ORDER_ZH : ORDER_EN;
  const normalizedBase = base.endsWith('/') ? base : `${base}/`;
  const loc = (slug: string) =>
    lang === 'en'
      ? `${normalizedBase}docs/${slug}/`.replace(/\/+/g, '/')
      : `${normalizedBase}zh/docs/${slug}/`.replace(/\/+/g, '/');

  const idx = order.findIndex((p) => p.slug === currentSlug);
  if (idx === -1) return null;

  const prev = idx > 0 ? order[idx - 1] : null;
  const next = idx < order.length - 1 ? order[idx + 1] : null;

  if (!prev && !next) return null;

  return (
    <nav
      className="mt-12 grid grid-cols-1 gap-3 sm:grid-cols-2"
      style={{ borderTop: '1px solid var(--border-default)', paddingTop: '24px' }}
    >
      {prev ? (
        <a
          href={loc(prev.slug)}
          className="flex flex-col rounded-xl p-4 transition-opacity hover:opacity-80"
          style={{
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            textDecoration: 'none',
          }}
        >
          <span className="text-xs mb-1" style={{ color: 'var(--muted-text)' }}>
            ← {lang === 'en' ? 'Previous' : '上一篇'}
          </span>
          <span className="text-sm font-medium" style={{ color: 'var(--primary-text)' }}>
            {prev.label}
          </span>
        </a>
      ) : (
        <div />
      )}

      {next ? (
        <a
          href={loc(next.slug)}
          className="flex flex-col rounded-xl p-4 transition-opacity hover:opacity-80 sm:text-right"
          style={{
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            textDecoration: 'none',
          }}
        >
          <span className="text-xs mb-1" style={{ color: 'var(--muted-text)' }}>
            {lang === 'en' ? 'Next' : '下一篇'} →
          </span>
          <span className="text-sm font-medium" style={{ color: 'var(--primary-text)' }}>
            {next.label}
          </span>
        </a>
      ) : (
        <div />
      )}
    </nav>
  );
}
