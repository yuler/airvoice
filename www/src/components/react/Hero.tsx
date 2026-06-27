import { ArrowRight } from '@phosphor-icons/react';

interface HeroProps {
  lang: 'en' | 'zh';
  base: string;
}

export default function Hero({ lang, base }: HeroProps) {
  const isZh = lang === 'zh';
  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  return (
    <section className="py-24 md:py-32">
      <div className="mx-auto max-w-3xl px-6 text-center">
        <h1 className="text-4xl font-bold tracking-tight text-kumo-default md:text-5xl">
          {isZh ? '手机说话，电脑打字' : 'Speak on your phone, type on your PC'}
        </h1>
        <p className="mt-6 text-lg text-kumo-subtle">
          {isZh
            ? '局域网语音桥接工具。无云端、无账号、数据不出局域网。'
            : 'A LAN voice bridge. No cloud, no accounts, no data leaves your network.'}
        </p>

        <div className="mt-10 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <a
            href="#get-started"
            className="inline-flex items-center gap-2 rounded-full px-6 py-2.5 text-sm font-medium text-white transition-opacity hover:opacity-90"
            style={{ backgroundColor: '#006efe' }}
          >
            {isZh ? '快速开始' : 'Get Started'}
            <ArrowRight size={14} />
          </a>
          <a
            href={loc('docs/background/')}
            className="inline-flex items-center gap-2 rounded-full border border-kumo-hairline px-6 py-2.5 text-sm font-medium text-kumo-default transition-colors hover:bg-kumo-control"
          >
            {isZh ? '了解更多' : 'Learn more'}
          </a>
        </div>
      </div>
    </section>
  );
}
