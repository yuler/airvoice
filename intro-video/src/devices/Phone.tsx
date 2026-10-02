import { interpolateColors, useCurrentFrame } from 'remotion';
import { StatusDot } from '../components';
import { copy } from '../strings';
import { color, sans } from '../theme';

export const PHONE_WIDTH = 360;
export const PHONE_HEIGHT = 740;

type PhoneProps = {
  connected: boolean;
  text: string;
  listening: number;
  pressed: number;
  scanning: number;
};

const PaperPlane: React.FC<{ fill: string }> = ({ fill }) => (
  <svg width={20} height={20} viewBox="0 0 24 24" fill={fill}>
    <path d="M2.5 20.5 22 12 2.5 3.5 2.5 10l13 2-13 2z" />
  </svg>
);

const Waveform: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <div style={{ display: 'flex', gap: 6, alignItems: 'center', height: 64 }}>
      {Array.from({ length: 15 }, (_, i) => {
        const envelope = 1 - Math.abs(i - 7) / 9;
        const h = 8 + 52 * envelope * Math.abs(Math.sin(frame / 5 + i * 0.9) * Math.cos(frame / 11 + i * 0.4));
        return <div key={i} style={{ width: 6, height: h, borderRadius: 3, background: color.blue }} />;
      })}
    </div>
  );
};

const Viewfinder: React.FC<{ opacity: number }> = ({ opacity }) => {
  const frame = useCurrentFrame();
  const corner = (rotate: number) => (
    <div
      style={{
        position: 'absolute', width: 44, height: 44, borderTop: '5px solid #fff', borderLeft: '5px solid #fff',
        borderTopLeftRadius: 18, transform: `rotate(${rotate}deg)`,
        ...(rotate === 0 && { left: 0, top: 0 }),
        ...(rotate === 90 && { right: 0, top: 0 }),
        ...(rotate === 180 && { right: 0, bottom: 0 }),
        ...(rotate === 270 && { left: 0, bottom: 0 })
      }}
    />
  );
  return (
    <div
      style={{
        position: 'absolute', inset: 0, background: 'rgba(12,12,16,0.94)', opacity, display: 'flex',
        flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 28
      }}
    >
      <div style={{ position: 'relative', width: 220, height: 220 }}>
        {[0, 90, 180, 270].map(corner)}
        <div style={{ position: 'absolute', left: 16, right: 16, top: 20 + ((frame * 4) % 180), height: 3, background: color.blue, boxShadow: `0 0 16px ${color.blue}` }} />
      </div>
      <div style={{ fontFamily: sans, fontSize: 17, color: '#ededed' }}>Scan the QR code</div>
    </div>
  );
};

export const Phone: React.FC<PhoneProps> = ({ connected, text, listening, pressed, scanning }) => (
  <div
    style={{
      width: PHONE_WIDTH, height: PHONE_HEIGHT, borderRadius: 58, padding: 10, background: '#111114',
      boxShadow: '0 50px 120px rgba(15,23,42,0.28), 0 10px 30px rgba(15,23,42,0.12)'
    }}
  >
    <div style={{ position: 'relative', width: '100%', height: '100%', borderRadius: 48, overflow: 'hidden', background: color.bg, fontFamily: sans }}>
      <div style={{ height: 54, display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 30px', fontSize: 17, fontWeight: 600, color: color.text }}>
        <span>9:41</span>
        <div style={{ width: 26, height: 12, borderRadius: 4, border: `1.5px solid ${color.text}`, padding: 1.5 }}>
          <div style={{ width: '75%', height: '100%', borderRadius: 2, background: color.text }} />
        </div>
      </div>
      <div style={{ position: 'absolute', top: 12, left: '50%', marginLeft: -56, width: 112, height: 32, borderRadius: 16, background: '#000' }} />

      <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '10px 22px 0', fontSize: 15, color: color.secondary }}>
        <StatusDot on={connected} size={9} />
        <span>{connected ? `Connected: ${copy.host}` : copy.connecting}</span>
      </div>

      <div
        style={{
          margin: '16px 20px', height: 300, padding: '14px 16px', borderRadius: 16, background: color.bg2,
          border: `1px solid ${color.border}`, fontSize: 21, lineHeight: 1.45, color: text ? color.text : color.muted
        }}
      >
        {text || copy.placeholder}
      </div>

      <div
        style={{
          margin: '0 20px', height: 54, borderRadius: 27, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10,
          background: interpolateColors(pressed, [0, 1], ['#ececec', color.blue]),
          color: interpolateColors(pressed, [0, 1], [color.text, '#ffffff']),
          fontSize: 17, fontWeight: 600, transform: `scale(${1 - 0.04 * pressed})`
        }}
      >
        <PaperPlane fill={interpolateColors(pressed, [0, 1], [color.text, '#ffffff'])} />
        {copy.send}
      </div>

      <div
        style={{
          position: 'absolute', left: 0, right: 0, bottom: 0, height: 250, borderTopLeftRadius: 32, borderTopRightRadius: 32,
          background: 'linear-gradient(180deg, #eef5ff, #d9e8ff)', display: 'flex', flexDirection: 'column',
          alignItems: 'center', justifyContent: 'center', gap: 18, transform: `translateY(${(1 - listening) * 270}px)`
        }}
      >
        <Waveform />
        <div style={{ fontSize: 17, color: color.secondary }}>{copy.listening}</div>
      </div>

      <Viewfinder opacity={scanning} />
    </div>
  </div>
);
