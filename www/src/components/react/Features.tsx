interface Feature {
  icon: React.ReactNode;
  title: string;
  desc: string;
}

interface FeaturesProps {
  lang: 'en' | 'zh';
}

function WifiIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <path d="M5 12.55a11 11 0 0 1 14.08 0" /><path d="M1.42 9a16 16 0 0 1 21.16 0" />
      <path d="M8.53 16.11a6 6 0 0 1 6.95 0" /><circle cx="12" cy="20" r="1" fill="currentColor" stroke="none" />
    </svg>
  );
}

function ShieldIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    </svg>
  );
}

function PlatformsIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="3" width="20" height="14" rx="2" /><path d="M8 21h8M12 17v4" />
    </svg>
  );
}

function CodeIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="16 18 22 12 16 6" /><polyline points="8 6 2 12 8 18" />
    </svg>
  );
}

export default function Features({ lang }: FeaturesProps) {
  const isZh = lang === 'zh';

  const features: Feature[] = [
    {
      icon: <WifiIcon />,
      title: isZh ? '局域网专属' : 'LAN Only',
      desc: isZh
        ? '无需连接互联网。数据在局域网内点对点传输，永远不经过任何服务器。'
        : 'No internet connection needed. Data flows peer-to-peer within your LAN, never touching any server.',
    },
    {
      icon: <ShieldIcon />,
      title: isZh ? '零云端' : 'Zero Cloud',
      desc: isZh ? '无账号、无订阅、无追踪。完全离线工作。' : 'No accounts, no subscriptions, no tracking. Works completely offline.',
    },
    {
      icon: <PlatformsIcon />,
      title: isZh ? '全平台支持' : 'All Platforms',
      desc: isZh
        ? '桌面端支持 Windows、macOS、Linux；手机端支持 Android 和 iOS。'
        : 'Desktop: Windows, macOS, Linux. Mobile: Android & iOS.',
    },
    {
      icon: <CodeIcon />,
      title: isZh ? '开源' : 'Open Source',
      desc: isZh ? 'MIT 许可证，完全透明，无任何隐藏逻辑。' : 'MIT License, fully transparent. No hidden logic.',
    },
  ];

  return (
    <section className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {features.map((f) => (
            <div
              key={f.title}
              className="rounded-xl border border-kumo-hairline p-5 transition-colors hover:border-kumo-subtle"
              style={{ backgroundColor: 'var(--color-kumo-canvas, #fff)' }}
            >
              <div
                className="mb-4 inline-flex items-center justify-center rounded-lg p-2"
                style={{
                  backgroundColor: 'var(--color-kumo-control, #fafafa)',
                  color: 'var(--accent-blue, #006efe)',
                }}
              >
                {f.icon}
              </div>
              <h3 className="mb-1.5 text-sm font-semibold text-kumo-default">{f.title}</h3>
              <p className="text-sm leading-relaxed text-kumo-subtle">{f.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
