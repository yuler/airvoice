import { AbsoluteFill, spring, useCurrentFrame, useVideoConfig } from 'remotion';
import { Words, fadeOut } from '../components';
import { AndroidIcon, AppleIcon, FinderIcon, LinuxIcon, WindowsIcon } from '../icons';
import { Sfx } from '../sound';
import { copy } from '../strings';
import { color, sans, sec } from '../theme';

export const PLATFORMS_FRAMES = sec(2.5);

const PLATFORMS = [
  ['iOS', AppleIcon],
  ['Android', AndroidIcon],
  ['macOS', FinderIcon],
  ['Windows', WindowsIcon],
  ['Linux', LinuxIcon]
] as const;
const CARD_AT = (i: number) => 10 + i * 5;

export const Platforms = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  return (
    <AbsoluteFill style={{ alignItems: 'center', justifyContent: 'center', gap: 80, opacity: fadeOut(frame, PLATFORMS_FRAMES) }}>
      {PLATFORMS.map(([label], i) => (
        <Sfx key={label} name="click" at={CARD_AT(i)} volume={0.35} />
      ))}
      <Words text={copy.platforms} size={88} delay={2} accent={['everywhere']} />
      <div style={{ display: 'flex', gap: 28 }}>
        {PLATFORMS.map(([label, Icon], i) => {
          const p = spring({ frame: frame - CARD_AT(i), fps, config: { damping: 14, stiffness: 150 } });
          return (
            <div
              key={label}
              style={{
                width: 220, height: 200, borderRadius: 24, background: color.bg, border: `1px solid ${color.border}`,
                boxShadow: '0 20px 50px rgba(15,23,42,0.08)', display: 'flex', flexDirection: 'column',
                alignItems: 'center', justifyContent: 'center', gap: 22, opacity: Math.min(1, p),
                transform: `translateY(${(1 - p) * 40}px) scale(${0.9 + 0.1 * p})`
              }}
            >
              <Icon size={72} />
              <div style={{ fontFamily: sans, fontSize: 28, fontWeight: 600, color: color.text }}>{label}</div>
            </div>
          );
        })}
      </div>
    </AbsoluteFill>
  );
};
