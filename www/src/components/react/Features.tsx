import { LayerCard } from '@cloudflare/kumo';

interface Feature {
  icon: string;
  title: string;
  desc: string;
}

interface FeaturesProps {
  lang: 'en' | 'zh';
}

export default function Features({ lang }: FeaturesProps) {
  const isZh = lang === 'zh';

  const features: Feature[] = [
    {
      icon: '📡',
      title: isZh ? '仅限局域网' : 'LAN Only',
      desc: isZh ? '数据从不离开本地网络' : 'Data never leaves your local network',
    },
    {
      icon: '🔒',
      title: isZh ? '零云端' : 'Zero Cloud',
      desc: isZh ? '无账号、无订阅、无追踪' : 'No accounts, no subscriptions, no tracking',
    },
    {
      icon: '💻',
      title: isZh ? '跨平台' : 'Cross-Platform',
      desc: isZh ? 'macOS + Linux (X11/Wayland)' : 'macOS + Linux (X11/Wayland)',
    },
    {
      icon: '📖',
      title: isZh ? '开源' : 'Open Source',
      desc: isZh ? 'MIT 许可证，完全透明' : 'MIT License, fully transparent',
    },
  ];

  return (
    <section className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {features.map((f) => (
            <LayerCard key={f.title} className="rounded-xl p-5">
              <div className="mb-3 text-2xl">{f.icon}</div>
              <h3 className="mb-1 text-sm font-semibold text-kumo-default">{f.title}</h3>
              <p className="text-sm text-kumo-subtle">{f.desc}</p>
            </LayerCard>
          ))}
        </div>
      </div>
    </section>
  );
}
