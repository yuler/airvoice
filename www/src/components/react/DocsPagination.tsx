import type { DocsMessages } from '../../i18n/messages';

interface PaginationProps {
  m: DocsMessages;
  currentSlug: string;
  pages: Array<{ slug: string; label: string; href: string }>;
}

export default function DocsPagination({ m, currentSlug, pages }: PaginationProps) {
  const idx = pages.findIndex((p) => p.slug === currentSlug);
  if (idx === -1) return null;

  const prev = idx > 0 ? pages[idx - 1] : null;
  const next = idx < pages.length - 1 ? pages[idx + 1] : null;

  if (!prev && !next) return null;

  return (
    <nav
      className="mt-12 grid grid-cols-1 gap-3 sm:grid-cols-2"
      style={{ borderTop: '1px solid var(--border-default)', paddingTop: '24px' }}
    >
      {prev ? (
        <a
          href={prev.href}
          className="flex flex-col rounded-xl p-4 transition-opacity hover:opacity-80"
          style={{
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            textDecoration: 'none',
          }}
        >
          <span className="text-xs mb-1" style={{ color: 'var(--muted-text)' }}>
            ← {m.prev}
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
          href={next.href}
          className="flex flex-col rounded-xl p-4 transition-opacity hover:opacity-80 sm:text-right"
          style={{
            border: '1px solid var(--border-default)',
            borderRadius: '12px',
            textDecoration: 'none',
          }}
        >
          <span className="text-xs mb-1" style={{ color: 'var(--muted-text)' }}>
            {m.next} →
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
