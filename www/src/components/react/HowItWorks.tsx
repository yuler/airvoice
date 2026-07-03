import HowItWorksDiagram from './HowItWorksDiagram';
import type { HowItWorksMessages } from '../../i18n/messages';

interface HowItWorksProps {
  m: HowItWorksMessages;
}

export default function HowItWorks({ m }: HowItWorksProps) {
  return (
    <section
      className="border-t py-16 md:py-20"
      style={{ borderColor: 'var(--border-default)', background: 'var(--background-secondary)' }}
    >
      <div className="mx-auto max-w-6xl px-4 md:px-6">
        <div className="mb-10 text-center max-w-2xl mx-auto">
          <h2
            className="text-2xl font-bold"
            style={{ color: 'var(--primary-text)', letterSpacing: '-0.02em' }}
          >
            {m.title}
          </h2>
          <p className="mt-4 text-sm leading-relaxed" style={{ color: 'var(--secondary-text)' }}>
            {m.subtitle}
          </p>
        </div>
        <div
          className="rounded-2xl px-4 py-8 md:px-8 md:py-10"
          style={{
            background: 'var(--background-primary)',
            border: '1px solid var(--border-default)',
          }}
        >
          <HowItWorksDiagram labels={m.diagram} />
        </div>
      </div>
    </section>
  );
}
