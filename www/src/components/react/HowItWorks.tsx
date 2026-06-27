import { LayerCard, Badge } from '@cloudflare/kumo';

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
        ? '运行 mise install 获取 airvoice 二进制文件。'
        : 'Run mise install to get the airvoice binary.',
    },
    {
      num: '02',
      title: isZh ? '扫码连接' : 'Scan QR',
      desc: isZh
        ? '运行 airvoice serve，用 iPhone 扫描二维码。'
        : 'Run airvoice serve, scan the QR code with your iPhone.',
    },
    {
      num: '03',
      title: isZh ? '说话打字' : 'Speak & Type',
      desc: isZh
        ? '手机语音口述，文字出现在电脑光标处。'
        : 'Dictate on your phone, text appears at your cursor.',
    },
  ];

  return (
    <section className="border-t border-kumo-hairline py-20">
      <div className="mx-auto max-w-6xl px-6">
        <h2 className="mb-16 text-center text-2xl font-bold text-kumo-default">
          {isZh ? '工作原理' : 'How It Works'}
        </h2>
        <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
          {steps.map((s) => (
            <LayerCard key={s.num} className="rounded-xl p-6">
              <Badge className="mb-4">{s.num}</Badge>
              <h3 className="mb-2 text-sm font-semibold text-kumo-default">{s.title}</h3>
              <p className="text-sm text-kumo-subtle">{s.desc}</p>
            </LayerCard>
          ))}
        </div>
      </div>
    </section>
  );
}
