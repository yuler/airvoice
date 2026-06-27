import { useState, useEffect } from 'react';

interface TocProps {
  headings: { depth: number; slug: string; text: string }[];
  lang: 'en' | 'zh';
}

export default function TableOfContents({ headings, lang }: TocProps) {
  const filtered = headings.filter((h) => h.depth <= 3);
  const [activeId, setActiveId] = useState<string>('');

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id);
          }
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
      <h4 className="mb-3 text-xs font-semibold uppercase tracking-wider text-kumo-subtle">
        {lang === 'en' ? 'On this page' : '本页目录'}
      </h4>
      <ul className="space-y-1">
        {filtered.map((h) => {
          const isActive = activeId === h.slug;
          return (
            <li key={h.slug}>
              <a
                href={`#${h.slug}`}
                className={`block text-sm transition-colors ${
                  h.depth === 3 ? 'pl-4' : ''
                } ${
                  isActive
                    ? 'text-kumo-default font-medium'
                    : 'text-kumo-inactive hover:text-kumo-subtle'
                }`}
              >
                {h.text}
              </a>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
