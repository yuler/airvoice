import { ArrowSquareOut } from '@phosphor-icons/react';

interface GetStartedProps {
  lang: 'en' | 'zh';
  base: string;
}

// OS icon components
function WindowsIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      <path d="M0 3.449L9.75 2.1v9.451H0m10.949-9.602L24 0v11.549H10.949M0 12.6h9.75v9.451L0 20.699M10.949 12.6H24V24l-12.9-1.801" />
    </svg>
  );
}

function AppleIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      <path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701" />
    </svg>
  );
}

function LinuxIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      {/* Tux penguin — simplified clean version */}
      <path d="M12 2C9.243 2 7 4.243 7 7v2c0 .738.203 1.43.556 2.022C6.613 12.482 6 14.207 6 16c0 2.21 1.343 3 3 3h6c1.657 0 3-.79 3-3 0-1.793-.613-3.518-1.556-4.978C16.797 10.43 17 9.738 17 9V7c0-2.757-2.243-5-5-5zm-1.5 6a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5zm3 0a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5zM12 18c-1.5 0-2-.5-2-2s.895-3.5 2-3.5 2 1.5 2 3.5-.5 2-2 2z" />
    </svg>
  );
}

function AndroidIcon({ size = 16 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor">
      {/* Material Design Android robot icon */}
      <path d="M17.6 9.48l1.84-3.18c.16-.31.04-.69-.26-.85a.637.637 0 0 0-.83.22l-1.88 3.24C14.79 8.33 13.44 8 12 8s-2.79.33-4.47.91L5.65 5.67a.637.637 0 0 0-.83-.22c-.3.16-.42.54-.26.85L6.4 9.48C3.3 11.25 1.28 14.44 1 18h22c-.28-3.56-2.3-6.75-5.4-8.52zM7 15.25a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5zm10 0a1.25 1.25 0 1 1 0-2.5 1.25 1.25 0 0 1 0 2.5z" />
    </svg>
  );
}

export default function GetStarted({ lang, base }: GetStartedProps) {
  const isZh = lang === 'zh';
  const loc = (path: string) => lang === 'en' ? `${base}${path}` : `${base}zh/${path}`;

  return (
    <section id="get-started" className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="mb-16 text-center">
          <h2 className="text-2xl font-bold text-kumo-default" style={{ letterSpacing: '-0.02em' }}>
            {isZh ? '快速开始' : 'Get Started'}
          </h2>
          <p className="mt-3 text-sm text-kumo-subtle">
            {isZh
              ? '需要安装两个客户端：桌面端 CLI + 手机端 App'
              : 'Two clients needed: Desktop CLI + Mobile App'}
          </p>
        </div>

        <div className="mx-auto grid max-w-4xl grid-cols-1 gap-6 md:grid-cols-2">
          {/* Desktop CLI Card */}
          <div
            className="rounded-xl border border-kumo-hairline p-6"
            style={{ backgroundColor: 'var(--color-kumo-canvas, #fff)' }}
          >
            <div className="mb-5">
              <div className="mb-3 flex items-center gap-2">
                <div
                  className="inline-flex items-center justify-center rounded-lg p-1.5"
                  style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)', color: 'var(--accent-blue, #006efe)' }}
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" />
                  </svg>
                </div>
                <h3 className="text-sm font-semibold text-kumo-default">
                  {isZh ? '桌面端 CLI' : 'Desktop CLI'}
                </h3>
              </div>

              {/* OS badges */}
              <div className="mb-4 flex flex-wrap items-center gap-1.5">
                {[
                  { icon: <WindowsIcon />, label: 'Windows' },
                  { icon: <AppleIcon />, label: 'macOS' },
                  { icon: <LinuxIcon />, label: 'Linux' },
                ].map((os) => (
                  <span
                    key={os.label}
                    className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2 py-1 text-xs text-kumo-subtle"
                    style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)' }}
                  >
                    <span style={{ opacity: 0.7 }}>{os.icon}</span>
                    {os.label}
                  </span>
                ))}
              </div>
            </div>

            {/* Install commands */}
            <div className="mb-5 space-y-2">
              <p className="text-xs font-medium text-kumo-subtle uppercase tracking-wider">
                {isZh ? '安装方式' : 'Install'}
              </p>
              <div
                className="rounded-lg p-3"
                style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)', border: '1px solid var(--color-kumo-hairline, #eaeaea)' }}
              >
                <code className="font-mono text-xs text-kumo-default">
                  brew tap yuler/airvoice && brew install airvoice
                </code>
              </div>
              <div
                className="rounded-lg p-3"
                style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)', border: '1px solid var(--color-kumo-hairline, #eaeaea)' }}
              >
                <code className="font-mono text-xs text-kumo-default">
                  go install github.com/yuler/airvoice/cli@latest
                </code>
              </div>
            </div>

            <a
              href="https://github.com/yuler/airvoice/releases"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-xs font-medium text-white transition-opacity hover:opacity-90"
              style={{ backgroundColor: 'var(--accent-blue, #006efe)' }}
            >
              {isZh ? '下载 Release' : 'Download Release'}
              <ArrowSquareOut size={11} />
            </a>
          </div>

          {/* Mobile App Card */}
          <div
            className="rounded-xl border border-kumo-hairline p-6"
            style={{ backgroundColor: 'var(--color-kumo-canvas, #fff)' }}
          >
            <div className="mb-5">
              <div className="mb-3 flex items-center gap-2">
                <div
                  className="inline-flex items-center justify-center rounded-lg p-1.5"
                  style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)', color: 'var(--accent-blue, #006efe)' }}
                >
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <rect x="5" y="2" width="14" height="20" rx="2" ry="2" />
                    <line x1="12" y1="18" x2="12.01" y2="18" />
                  </svg>
                </div>
                <h3 className="text-sm font-semibold text-kumo-default">
                  {isZh ? '手机 App' : 'Mobile App'}
                </h3>
              </div>

              {/* Mobile OS badges */}
              <div className="mb-4 flex flex-wrap items-center gap-1.5">
                {[
                  { icon: <AndroidIcon />, label: 'Android' },
                  { icon: <AppleIcon />, label: 'iOS' },
                ].map((os) => (
                  <span
                    key={os.label}
                    className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2 py-1 text-xs text-kumo-subtle"
                    style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)' }}
                  >
                    <span style={{ opacity: 0.7 }}>{os.icon}</span>
                    {os.label}
                  </span>
                ))}
              </div>
            </div>

            <p className="mb-5 text-sm leading-relaxed text-kumo-subtle">
              {isZh
                ? '使用手机自带的系统语音识别，扫码连接桌面端，无需额外账号或云服务。'
                : 'Uses your phone\'s native speech recognition. Scan to connect — no account or cloud service needed.'}
            </p>

            <div className="flex flex-wrap items-center gap-2">
              <a
                href={loc('docs/quick-start/')}
                className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-xs font-medium text-white transition-opacity hover:opacity-90"
                style={{ backgroundColor: 'var(--accent-blue, #006efe)' }}
              >
                {isZh ? '安装文档' : 'Install Docs'}
                <ArrowSquareOut size={11} />
              </a>
              <a
                href="https://github.com/yuler/airvoice/releases"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 rounded-full border border-kumo-hairline px-4 py-2 text-xs font-medium text-kumo-default transition-colors hover:bg-kumo-control"
              >
                {isZh ? '从 GitHub 下载' : 'GitHub Releases'}
                <ArrowSquareOut size={11} />
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
