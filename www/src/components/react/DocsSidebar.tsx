import { useState } from 'react';
import { List, X } from '@phosphor-icons/react';
import { Button } from '@cloudflare/kumo';

interface SidebarProps {
  lang: 'en' | 'zh';
  base: string;
  currentPath: string;
}

export default function DocsSidebar({ lang, base, currentPath }: SidebarProps) {
  const [open, setOpen] = useState(false);
  const isZh = lang === 'zh';
  const normalizedBase = base.endsWith('/') ? base : `${base}/`;
  const loc = (path: string) => lang === 'en' ? `${normalizedBase}${path}` : `${normalizedBase}zh/${path}`;

  const sections = [
    {
      title: isZh ? '入门' : 'Overview',
      items: [
        { href: loc('docs/background/'), label: isZh ? '背景' : 'Background' },
        { href: loc('docs/quick-start/'), label: isZh ? '快速开始' : 'Quick Start' },
      ],
    },
    {
      title: isZh ? '指南' : 'Guide',
      items: [
        { href: loc('docs/development/'), label: isZh ? '开发指南' : 'Development' },
        { href: loc('docs/architecture/'), label: isZh ? '架构' : 'Architecture' },
        { href: loc('docs/platform-deps/'), label: isZh ? '平台依赖' : 'Platform Deps' },
      ],
    },
  ];

  return (
    <>
      <div className="fixed left-4 top-3 z-50 md:hidden">
        <Button
          variant="ghost"
          shape="square"
          icon={open ? X : List}
          onClick={() => setOpen(!open)}
          aria-label={isZh ? '切换菜单' : 'Toggle menu'}
        />
      </div>

      {open && (
        <div
          className="fixed inset-0 z-30 bg-black/50 md:hidden"
          onClick={() => setOpen(false)}
        />
      )}

      <aside
        className={`fixed top-0 left-0 z-40 h-full w-64 overflow-auto border-r border-kumo-hairline bg-kumo-canvas transition-transform duration-300 ${
          open ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
        }`}
      >
        <div className="flex h-12 items-center border-b border-kumo-hairline px-4">
          <a href={loc('')} className="text-lg font-bold text-kumo-default hover:text-kumo-subtle transition-colors">
            Airvoice
          </a>
        </div>

        <nav className="p-4">
          {sections.map((section) => (
            <div key={section.title} className="mb-6">
              <h3 className="mb-2 text-xs font-semibold uppercase tracking-wider text-kumo-subtle">
                {section.title}
              </h3>
              <ul className="space-y-1">
                {section.items.map((item) => {
                  const isActive = currentPath === item.href;
                  return (
                    <li key={item.href}>
                      <a
                        href={item.href}
                        onClick={() => setOpen(false)}
                        className={`block rounded-md px-3 py-2 text-sm transition-colors ${
                          isActive
                            ? 'bg-kumo-control text-kumo-default font-medium'
                            : 'text-kumo-subtle hover:text-kumo-default hover:bg-kumo-control'
                        }`}
                      >
                        {item.label}
                      </a>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>
      </aside>
    </>
  );
}
