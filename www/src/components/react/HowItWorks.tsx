interface Step {
  num: string;
  title: string;
  desc: string;
}

interface HowItWorksProps {
  lang: 'en' | 'zh';
}

export default function HowItWorks({ lang }: HowItWorksProps) {
  const isZh = lang === 'zh';

  const steps: Step[] = [
    {
      num: '01',
      title: isZh ? '安装 CLI' : 'Install CLI',
      desc: isZh
        ? '在 Windows、macOS 或 Linux 上下载并运行 airvoice 二进制文件。'
        : 'Download and run the airvoice binary on Windows, macOS, or Linux.',
    },
    {
      num: '02',
      title: isZh ? '扫码连接' : 'Scan QR Code',
      desc: isZh
        ? '运行 airvoice serve，用 Android 或 iPhone 扫描终端二维码，连接局域网。'
        : 'Run airvoice serve, scan the QR code with your Android or iPhone over LAN.',
    },
    {
      num: '03',
      title: isZh ? '说话打字' : 'Speak & Type',
      desc: isZh
        ? '在手机上开口说话，文字立即出现在电脑光标处。无需互联网。'
        : 'Dictate on your phone — text appears at your cursor instantly. No internet.',
    },
  ];

  return (
    <section className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <div className="mb-16 text-center">
          <h2 className="text-2xl font-bold text-kumo-default" style={{ letterSpacing: '-0.02em' }}>
            {isZh ? '三步即用' : 'Three Steps'}
          </h2>
          <p className="mt-3 text-sm text-kumo-subtle">
            {isZh ? '全程在局域网内完成，无需任何云服务。' : 'Everything happens on your LAN. No cloud services involved.'}
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {steps.map((s, i) => (
            <div key={s.num} className="relative">
              {/* Connector line between steps */}
              {i < steps.length - 1 && (
                <div
                  className="absolute top-[42px] hidden md:block"
                  style={{
                    left: 'calc(100% - 12px)',
                    width: 'calc(100% - 48px + 24px)',
                    height: '1px',
                    background: 'var(--color-kumo-hairline, #eaeaea)',
                    zIndex: 0,
                  }}
                />
              )}
              <div
                className="relative rounded-xl border border-kumo-hairline p-6 transition-colors hover:border-kumo-subtle"
                style={{ backgroundColor: 'var(--color-kumo-canvas, #fff)' }}
              >
                {/* Step number badge */}
                <div
                  className="mb-5 inline-flex h-9 w-9 items-center justify-center rounded-lg text-xs font-bold"
                  style={{
                    backgroundColor: 'var(--color-kumo-control, #fafafa)',
                    color: 'var(--accent-blue, #006efe)',
                    border: '1px solid var(--color-kumo-hairline, #eaeaea)',
                    fontFamily: 'ui-monospace, monospace',
                  }}
                >
                  {s.num}
                </div>
                <h3 className="mb-2 text-sm font-semibold text-kumo-default">{s.title}</h3>
                <p className="text-sm leading-relaxed text-kumo-subtle">{s.desc}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
