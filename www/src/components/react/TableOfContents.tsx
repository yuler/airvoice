import { useState, useEffect, useMemo } from 'react';
import type { DocsMessages } from '../../i18n/messages';

interface TocProps {
  headings: { depth: number; slug: string; text: string }[];
  m: Pick<DocsMessages, 'tocTitle' | 'backToTop'>;
}

export default function TableOfContents({ headings, m }: TocProps) {
  const filtered = useMemo(() => headings.filter((h) => h.depth <= 3), [headings]);
  const [activeId, setActiveId] = useState<string>('');

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) setActiveId(entry.target.id);
        }
      },
      { rootMargin: '-80px 0px -70% 0px', threshold: 0 }
    );
    for (const h of filtered) {
      const el = document.getElementById(h.slug);
      if (el) observer.observe(el);
    }
    return () => observer.disconnect();
  }, [filtered]);

  if (filtered.length === 0) return null;

  return (
    <nav>
      <p className="mb-3 text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--muted-text)', letterSpacing: '0.08em' }}>
        {m.tocTitle}
      </p>
      <ul className="space-y-2">
        {filtered.map((h) => (
          <li key={h.slug} style={{ paddingLeft: `${(h.depth - 1) * 12}px` }}>
            <a
              href={`#${h.slug}`}
              className="block text-sm transition-colors"
              style={{
                color: activeId === h.slug ? '#006efe' : 'var(--secondary-text)',
                fontWeight: activeId === h.slug ? 500 : 400,
                textDecoration: 'none',
              }}
            >
              {h.text}
            </a>
          </li>
        ))}
      </ul>
      <a
        href="#"
        className="mt-6 inline-block text-xs transition-colors hover:opacity-80"
        style={{ color: 'var(--muted-text)', textDecoration: 'none' }}
        onClick={(e) => { e.preventDefault(); window.scrollTo({ top: 0, behavior: 'smooth' }); }}
      >
        ↑ {m.backToTop}
      </a>
    </nav>
  );
}
