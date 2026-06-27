import { LayerCard, Badge } from '@cloudflare/kumo';
import { DesktopIcon, DeviceMobileIcon, ArrowSquareOutIcon } from '@phosphor-icons/react';

interface GetStartedProps {
  lang: 'en' | 'zh';
  base: string;
}

export default function GetStarted({ lang, base }: GetStartedProps) {
  const isZh = lang === 'zh';
  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  return (
    <section id="get-started" className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <h2 className="mb-4 text-center text-2xl font-bold text-kumo-default">
          {isZh ? '快速开始' : 'Get Started'}
        </h2>
        <p className="mb-12 text-center text-sm text-kumo-subtle">
          {isZh
            ? '需要安装两个客户端：桌面端 CLI 和手机端 App'
            : 'Two clients required: Desktop CLI and Mobile App'}
        </p>

        <div className="mx-auto grid max-w-3xl grid-cols-1 gap-6 md:grid-cols-2">
          <LayerCard className="rounded-xl p-6">
            <div className="mb-4 flex items-center gap-2">
              <DesktopIcon size={20} className="text-kumo-subtle" />
              <h3 className="text-sm font-semibold text-kumo-default">
                {isZh ? '桌面端 CLI' : 'Desktop CLI'}
              </h3>
            </div>

            <div className="mb-4 space-y-2">
              <div className="rounded-lg bg-kumo-control p-3">
                <code className="font-mono text-xs text-kumo-default">mise trust && mise install</code>
              </div>
              <div className="rounded-lg bg-kumo-control p-3">
                <code className="font-mono text-xs text-kumo-default">go install github.com/yuler/airvoice/cli@latest</code>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <a
                href="https://github.com/yuler/airvoice/releases"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-xs font-medium text-white transition-opacity hover:opacity-90"
                style={{ backgroundColor: '#006efe' }}
              >
                {isZh ? '下载' : 'Download'}
                <ArrowSquareOutIcon size={12} />
              </a>
              <Badge variant="neutral">
                <kbd className="font-mono">⌘S</kbd>
              </Badge>
            </div>
          </LayerCard>

          <LayerCard className="rounded-xl p-6">
            <div className="mb-4 flex items-center gap-2">
              <DeviceMobileIcon size={20} className="text-kumo-subtle" />
              <h3 className="text-sm font-semibold text-kumo-default">
                {isZh ? 'iOS 客户端' : 'iOS Client'}
              </h3>
            </div>

            <p className="mb-4 text-sm text-kumo-subtle">
              {isZh
                ? '从源码构建，免费 Apple ID 即可。'
                : 'Build from source with a free Apple ID.'}
            </p>

            <div className="flex items-center gap-2">
              <a
                href={loc('docs/quick-start/')}
                className="inline-flex items-center gap-1.5 rounded-full px-4 py-1.5 text-xs font-medium text-white transition-opacity hover:opacity-90"
                style={{ backgroundColor: '#006efe' }}
              >
                {isZh ? '安装文档' : 'Install Docs'}
                <ArrowSquareOutIcon size={12} />
              </a>
              <Badge variant="neutral">
                <kbd className="font-mono">⌘R</kbd>
              </Badge>
            </div>
          </LayerCard>
        </div>
      </div>
    </section>
  );
}
