import { AbsoluteFill, spring, useCurrentFrame, useVideoConfig } from 'remotion';
import { Soundwave, Words } from '../components';
import { Sfx } from '../sound';
import { copy } from '../strings';
import { color, mono, sans, sec } from '../theme';

export const OUTRO_FRAMES = sec(3);

const Pill: React.FC<{ children: string }> = ({ children }) => (
  <div
    style={{
      fontFamily: mono, fontSize: 28, color: color.text, padding: '14px 26px', borderRadius: 14,
      background: color.bg, border: `1px solid ${color.border}`, boxShadow: '0 10px 30px rgba(15,23,42,0.06)'
    }}
  >
    {children}
  </div>
);

export const Outro = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const pop = spring({ frame, fps, config: { damping: 12, stiffness: 150 } });
  const line = spring({ frame: frame - 20, fps, config: { damping: 18 } });
  const links = spring({ frame: frame - 32, fps, config: { damping: 18 } });

  return (
    <AbsoluteFill style={{ alignItems: 'center', justifyContent: 'center', gap: 40 }}>
      <Sfx name="pop" at={2} />
      <div style={{ display: 'flex', alignItems: 'center', gap: 36 }}>
        <div
          style={{
            width: 150, height: 150, borderRadius: 36, background: color.blue, display: 'flex',
            alignItems: 'center', justifyContent: 'center', transform: `scale(${pop}) rotate(${(1 - pop) * -20}deg)`,
            boxShadow: `0 30px 80px rgba(0,110,254,${0.4 * pop})`
          }}
        >
          <Soundwave size={92} stroke="#ffffff" strokeWidth={2.2} />
        </div>
        <Words text={copy.name} size={128} weight={800} delay={6} />
      </div>
      <div style={{ fontFamily: sans, fontSize: 42, color: color.secondary, letterSpacing: '-0.02em', opacity: line, transform: `translateY(${(1 - line) * 20}px)` }}>
        {copy.tagline}
      </div>
      <div style={{ display: 'flex', gap: 20, opacity: links, transform: `translateY(${(1 - links) * 20}px)` }}>
        <Pill>{copy.url}</Pill>
        <Pill>{copy.repo}</Pill>
      </div>
    </AbsoluteFill>
  );
};
