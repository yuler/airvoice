import { useState, useEffect, useMemo } from 'react';

interface TocProps {
  headings: { depth: number; slug: string; text: string }[];
  lang: 'en' | 'zh';
}

export default function TableOfContents({ headings, lang }: TocProps) {
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
      <h4
        className="mb-3 uppercase"
        style={{ fontSize: '11px', fontWeight: 600, letterSpacing: '0.08em', color: 'var(--muted-text)' }}
      >
        {lang === 'en' ? 'On this page' : '本页目录'}
      </h4>
      <ul className="space-y-1">
        {filtered.map((h) => {
          const isActive = activeId === h.slug;
          return (
            <li key={h.slug}>
              <a
                href={`#${h.slug}`}
                className="block text-sm whitespace-nowrap transition-colors"
                style={{
                  paddingLeft: h.depth === 3 ? '16px' : undefined,
                  color: isActive ? '#006efe' : 'var(--muted-text)',
                  textDecoration: 'none',
                }}
              >
                {h.text}
              </a>
            </li>
          );
        })}
      </ul>
      <div className="mt-6 pt-4" style={{ borderTop: '1px solid var(--border-default)' }}>
        <button
          className="text-xs transition-opacity hover:opacity-70 bg-transparent border-none cursor-pointer p-0"
          style={{ color: 'var(--muted-text)' }}
          onClick={() => window.scrollTo({ top: 0, behavior: 'smooth' })}
        >
          ↑ {lang === 'en' ? 'Back to top' : '回到顶部'}
        </button>
      </div>
    </nav>
  );
}
