import { useState, useEffect } from 'react';
import { ArrowSquareOut, CheckCircle, Copy } from '@phosphor-icons/react';
import {
  getDownloadUrls,
  getDesktopDownloadUrl,
  getDesktopOS,
  getMobileDownloadInfo,
  RELEASES,
  type DesktopOS,
  type DownloadUrls,
  type MobileDownloadInfo,
} from '../../lib/downloads';
import { AppleIcon, AndroidIcon, MacOSIcon, WindowsIcon, LinuxIcon, CliIcon } from '../../lib/platform-icons';
import type { GetStartedMessages } from '../../i18n/messages';

interface GetStartedProps {
  m: GetStartedMessages;
  iosDocsUrl: string;
}

export default function GetStarted({ m, iosDocsUrl }: GetStartedProps) {
  const cliCommands = [
    'brew tap yuler/airvoice https://github.com/yuler/airvoice && brew install airvoice',
    'go install github.com/yuler/airvoice/cli@latest',
  ];
  const [urls, setUrls] = useState<DownloadUrls>({
    cli: 'https://github.com/yuler/airvoice/releases/latest',
    desktop: 'https://github.com/yuler/airvoice/releases/latest',
    mobile: 'https://github.com/yuler/airvoice/releases/latest',
  });
  const [mobileInfo, setMobileInfo] = useState<MobileDownloadInfo>({
    platform: 'other',
    apkUrl: 'https://github.com/yuler/airvoice/releases/latest',
    iosDocsPath: iosDocsUrl,
  });
  const [desktopOS, setDesktopOS] = useState<DesktopOS>('windows');
  const [copiedCommand, setCopiedCommand] = useState<string | null>(null);

  const desktopDownloadLabel = {
    windows: m.downloadDesktopWindows,
    macos: m.downloadDesktopMacos,
    linux: m.downloadDesktopLinux,
  }[desktopOS];
  const desktopDownloadUrl = getDesktopDownloadUrl(desktopOS);

  const copyCommand = async (command: string) => {
    try {
      if (navigator.clipboard) {
        await navigator.clipboard.writeText(command);
        setCopiedCommand(command);
        window.setTimeout(() => setCopiedCommand((current) => current === command ? null : current), 1600);
      }
    } catch (err) {
      console.error('Failed to copy command:', err);
    }
  };

  useEffect(() => {
    setUrls(getDownloadUrls(iosDocsUrl));
    setMobileInfo(getMobileDownloadInfo(iosDocsUrl));
    setDesktopOS(getDesktopOS());
  }, [iosDocsUrl]);

  return (
    <section id="get-started" className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="mb-16 text-center">
          <h2 className="text-2xl font-bold text-kumo-default" style={{ letterSpacing: '-0.02em' }}>
            {m.title}
          </h2>
          <p className="mt-3 text-sm text-kumo-subtle">{m.subtitle}</p>
        </div>
        <div className="mx-auto grid max-w-4xl grid-cols-1 gap-6 md:grid-cols-2">
          {/* PC Card */}
          <div
            className="rounded-xl border border-kumo-hairline p-6 flex flex-col justify-between"
            style={{ backgroundColor: 'var(--color-kumo-canvas, #fff)' }}
          >
            <div>
              <div className="mb-4">
                <div className="mb-3 flex items-center gap-2">
                  <div
                    className="inline-flex items-center justify-center rounded-lg p-1.5"
                    style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)', color: 'var(--accent-blue, #006efe)' }}
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <rect x="2" y="3" width="20" height="14" rx="2" />
                      <line x1="8" y1="21" x2="16" y2="21" />
                      <line x1="12" y1="17" x2="12" y2="21" />
                    </svg>
                  </div>
                  <h3 className="text-sm font-semibold text-kumo-default">{m.pc}</h3>
                </div>
              </div>
              <div className="mb-6 space-y-3">
                <div className="flex flex-wrap items-center gap-1.5">
                  <span
                    className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2 py-1 text-xs text-kumo-subtle"
                    style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)' }}
                  >
                    <CliIcon size={16} />
                    CLI
                  </span>
                </div>
                <div className="space-y-2">
                  {cliCommands.map((command, index) => {
                    const copied = copiedCommand === command;
                    return (
                      <div key={command}>
                        {index > 0 && (
                          <div className="text-center text-xs text-kumo-subtle py-1">{m.or}</div>
                        )}
                        <div
                          className="flex items-center gap-2 rounded-lg p-2"
                          style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)', border: '1px solid var(--color-kumo-hairline, #eaeaea)' }}
                        >
                          <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap font-mono text-xs text-kumo-default" style={{ scrollbarWidth: 'none' }}>
                            {command}
                          </code>
                          <button
                            type="button"
                            onClick={() => copyCommand(command)}
                            aria-label={copied ? m.copied : m.copyCommand}
                            className="inline-flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md border border-kumo-hairline text-kumo-subtle transition-colors hover:bg-kumo-canvas hover:text-kumo-default cursor-pointer"
                          >
                            {copied ? <CheckCircle size={15} weight="fill" style={{ color: 'var(--status-success, #28a948)' }} /> : <Copy size={14} />}
                          </button>
                        </div>
                      </div>
                    );
                  })}
                </div>
                <div className="pt-1">
                  <a
                    href={urls.cli}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-xs font-medium text-white transition-opacity hover:opacity-90"
                    style={{ backgroundColor: 'var(--accent-blue, #006efe)' }}
                  >
                    {m.downloadCli}
                    <ArrowSquareOut size={11} />
                  </a>
                </div>
              </div>
            </div>
            <div className="border-t border-kumo-hairline pt-5 space-y-3">
              <div className="flex flex-wrap items-center gap-1.5" role="tablist" aria-label={m.desktopDesc}>
                {([
                  { os: 'windows' as const, icon: <WindowsIcon size={16} />, label: 'Windows' },
                  { os: 'macos' as const, icon: <MacOSIcon size={16} />, label: 'macOS' },
                  { os: 'linux' as const, icon: <LinuxIcon size={16} />, label: 'Linux' },
                ]).map((item) => {
                  const active = desktopOS === item.os;
                  return (
                    <button
                      key={item.label}
                      type="button"
                      role="tab"
                      aria-selected={active}
                      onClick={() => setDesktopOS(item.os)}
                      className="inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-xs transition-colors cursor-pointer"
                      style={{
                        backgroundColor: active ? 'rgba(0, 110, 254, 0.08)' : 'var(--color-kumo-control, #fafafa)',
                        borderColor: active ? 'var(--accent-blue, #006efe)' : 'var(--color-kumo-hairline, #eaeaea)',
                        color: active ? 'var(--accent-blue, #006efe)' : undefined,
                      }}
                    >
                      {item.icon}
                      <span className={active ? 'font-medium text-kumo-default' : 'text-kumo-subtle'}>{item.label}</span>
                    </button>
                  );
                })}
              </div>
              <p className="text-sm leading-relaxed text-kumo-subtle">{m.desktopDesc}</p>
              <div className="flex flex-wrap items-center gap-2">
                <a
                  href={desktopDownloadUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-xs font-medium text-white transition-opacity hover:opacity-90"
                  style={{ backgroundColor: 'var(--accent-blue, #006efe)' }}
                >
                  {desktopDownloadLabel}
                  <ArrowSquareOut size={11} />
                </a>
                <a
                  href={RELEASES}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 rounded-full border border-kumo-hairline px-4 py-2 text-xs font-medium text-kumo-default transition-colors hover:bg-kumo-control"
                >
                  {m.downloadOtherDesktop}
                  <ArrowSquareOut size={11} />
                </a>
              </div>
            </div>
          </div>

          {/* Mobile Card */}
          <div
            className="rounded-xl border border-kumo-hairline p-6 flex flex-col justify-between"
            style={{ backgroundColor: 'var(--color-kumo-canvas, #fff)' }}
          >
            <div>
              <div className="mb-4">
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
                  <h3 className="text-sm font-semibold text-kumo-default">{m.mobile}</h3>
                </div>
              </div>
              
              {/* Added top spacing & increased text margin to push content down for perfect alignment */}
              <p className="mt-2 mb-6 text-sm leading-relaxed text-kumo-subtle md:mb-10">{m.mobileDesc}</p>
              
              {mobileInfo.platform === 'ios' && (
                <div
                  className="mb-4 rounded-lg px-3 py-2.5 text-xs leading-relaxed"
                  style={{
                    background: 'var(--background-secondary)',
                    border: '1px solid var(--border-default)',
                    color: 'var(--secondary-text)',
                  }}
                >
                  {m.iosBanner}
                </div>
              )}
              
              <div className="mb-6 space-y-3">
                <div className="flex flex-wrap items-center gap-1.5">
                  <span
                    className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2 py-1 text-xs text-kumo-subtle"
                    style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)' }}
                  >
                    <AndroidIcon size={16} />
                    Android
                  </span>
                </div>
                <p className="text-xs leading-relaxed text-kumo-subtle">{m.androidDesc}</p>
                <div className="pt-1">
                  <a
                    href={mobileInfo.apkUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-xs font-medium text-white transition-opacity hover:opacity-90"
                    style={{ backgroundColor: 'var(--accent-blue, #006efe)' }}
                  >
                    {m.downloadApk}
                    <ArrowSquareOut size={11} />
                  </a>
                </div>
              </div>
            </div>

            {/* iOS Group sits perfectly at the bottom with matching layout structure */}
            <div className="border-t border-kumo-hairline pt-5 space-y-3">
              <div className="flex flex-wrap items-center gap-1.5">
                <span
                  className="inline-flex items-center gap-1.5 rounded-md border border-kumo-hairline px-2 py-1 text-xs text-kumo-subtle"
                  style={{ backgroundColor: 'var(--color-kumo-control, #fafafa)' }}
                >
                  <AppleIcon size={16} />
                  iOS
                </span>
              </div>
              <p className="text-sm leading-relaxed text-kumo-subtle">{m.iosDesc}</p>
              <div>
                <a
                  href={iosDocsUrl}
                  className="inline-flex items-center gap-1.5 rounded-full border border-kumo-hairline px-4 py-2 text-xs font-medium text-kumo-default transition-colors hover:bg-kumo-control"
                >
                  {m.iosGuide}
                  <ArrowSquareOut size={11} />
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
