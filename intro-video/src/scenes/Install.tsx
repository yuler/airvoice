import { AbsoluteFill, spring, useCurrentFrame, useVideoConfig } from 'remotion';
import { Words, fadeOut } from '../components';
import { Sfx } from '../sound';
import { copy } from '../strings';
import { color, mono, sec } from '../theme';

export const INSTALL_FRAMES = sec(3.5);

type Token = [string, string?];
const LINES: Token[][] = [
  [['$ ', color.terminalDim], ['brew tap '], ['yuler/airvoice ', '#93c5fd'], ['https://github.com/yuler/airvoice', '#86efac']],
  [['$ ', color.terminalDim], ['brew install '], ['airvoice', '#93c5fd']],
  [['$ ', color.terminalDim], ['airvoice', '#93c5fd']]
];
const CHARS_PER_FRAME = 1.1;
const TYPE_START = 14;
const TYPE_FRAMES = LINES.flat().reduce((n, [text]) => n + text.length, 0) / CHARS_PER_FRAME;

export const Install = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enter = spring({ frame, fps, config: { damping: 18 } });
  let budget = Math.max(0, (frame - TYPE_START) * CHARS_PER_FRAME);

  return (
    <AbsoluteFill style={{ alignItems: 'center', justifyContent: 'center', gap: 64, opacity: fadeOut(frame, INSTALL_FRAMES) }}>
      <Sfx name="whoosh" at={0} />
      <Sfx name="typing" at={TYPE_START} frames={TYPE_FRAMES} />
      <Words text={copy.install} size={88} delay={4} accent={['Run']} />
      <div
        style={{
          width: 1240, padding: '32px 44px 40px', borderRadius: 24, background: color.terminal,
          border: `1px solid ${color.terminalBorder}`, boxShadow: '0 40px 120px rgba(15,23,42,0.22), 0 0 0 1px rgba(0,110,254,0.12)',
          transform: `perspective(1600px) rotateX(${(1 - enter) * 30}deg) translateY(${(1 - enter) * 80}px)`, opacity: enter
        }}
      >
        <div style={{ display: 'flex', gap: 10, marginBottom: 26 }}>
          {['#ff5f57', '#febc2e', '#28c840'].map((c) => (
            <div key={c} style={{ width: 14, height: 14, borderRadius: 7, background: c }} />
          ))}
        </div>
        <div style={{ fontFamily: mono, fontSize: 30, lineHeight: 1.7, color: color.terminalText, whiteSpace: 'pre' }}>
          {LINES.map((line, i) => (
            <div key={i} style={{ minHeight: '1.7em' }}>
              {line.map(([text, c], j) => {
                const shown = text.slice(0, Math.floor(budget));
                budget = Math.max(0, budget - text.length);
                return (
                  <span key={j} style={{ color: c }}>
                    {shown}
                  </span>
                );
              })}
            </div>
          ))}
        </div>
      </div>
    </AbsoluteFill>
  );
};
