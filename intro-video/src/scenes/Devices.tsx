import { AbsoluteFill, interpolate, Sequence, spring, useCurrentFrame, useVideoConfig } from 'remotion';
import { clamp, fadeOut, progress, Words } from '../components';
import { EditorPane } from '../devices/EditorPane';
import { MacWindow } from '../devices/MacWindow';
import { Phone } from '../devices/Phone';
import { TerminalPane } from '../devices/TerminalPane';
import { Sfx, Voice } from '../sound';
import { copy } from '../strings';
import { color, sans, sec } from '../theme';

// Length of public/voice.wav, printed by `npm run voice`. Keep <= 5.5 so the paste lands before LOCAL.
export const VOICE_SECONDS = 2.5;

const SCAN = sec(0.8);
const LINKED = sec(1.9);
const DICTATE = sec(3);
export const VOICE_AT = DICTATE + sec(0.5);
const LISTEN_END = VOICE_AT + sec(VOICE_SECONDS + 0.2);
const TAP = LISTEN_END + sec(0.4);
const PASTE = TAP + sec(0.4);
const LOCAL = sec(10);
export const DEVICES_FRAMES = sec(13);

const WIN = { left: 150, top: 300, width: 1100, height: 640 };
const PHONE = { left: 1430, top: 262 };
const LINK = { from: WIN.left + WIN.width + 14, to: PHONE.left - 14, y: 620 };
const LINK_MID = (LINK.from + LINK.to) / 2;

const Headline: React.FC<{ text: string; frames: number; accent: string[] }> = ({ text, frames, accent }) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ alignItems: 'center', top: 120, opacity: fadeOut(frame, frames) }}>
      <Words text={text} size={68} delay={4} stagger={3} accent={accent} />
    </AbsoluteFill>
  );
};

const CloudOff: React.FC = () => (
  <svg width={64} height={64} viewBox="0 0 24 24" fill="none" strokeWidth={1.8} strokeLinecap="round" strokeLinejoin="round">
    <path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z" stroke={color.muted} fill={color.bg} />
    <path d="M3 3l18 18" stroke={color.red} strokeWidth={2.4} />
  </svg>
);

export const Devices = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const winIn = spring({ frame, fps, config: { damping: 18 } });
  const phoneIn = spring({ frame: frame - 8, fps, config: { damping: 18 } });
  const connected = frame >= LINKED;
  const scanning = Math.min(progress(frame, SCAN - 8, 8), 1 - progress(frame, LINKED - 6, 10));
  const toEditor = progress(frame, DICTATE, 14);

  const words = copy.spoken.split(' ');
  const heard = Math.floor(progress(frame, VOICE_AT, sec(VOICE_SECONDS)) * words.length);
  const listening = Math.min(progress(frame, VOICE_AT - 18, 18), 1 - progress(frame, LISTEN_END, 18));
  const pressed = interpolate(frame, [TAP - 4, TAP, TAP + 10, TAP + 22], [0, 1, 1, 0], clamp);
  const packet = progress(frame, TAP + 4, PASTE - TAP - 4);
  const pasted = frame >= PASTE;
  const flash = pasted ? 1 - progress(frame, PASTE + 20, 40) : 0;

  const linkIn = progress(frame, LINKED, 18);
  const local = spring({ frame: frame - LOCAL, fps, config: { damping: 15, stiffness: 140 } });

  return (
    <AbsoluteFill style={{ opacity: fadeOut(frame, DEVICES_FRAMES, 14) }}>
      <Sfx name="whoosh" at={0} />
      <Sfx name="click" at={SCAN} />
      <Sfx name="pop" at={LINKED} />
      <Voice at={VOICE_AT} />
      <Sfx name="click" at={TAP} />
      <Sfx name="pop" at={PASTE} />
      <Sfx name="whoosh" at={LOCAL - 12} />

      <Sequence durationInFrames={DICTATE} layout="none">
        <Headline text={copy.pair} frames={DICTATE} accent={['Paired']} />
      </Sequence>
      <Sequence from={DICTATE} durationInFrames={LOCAL - DICTATE} layout="none">
        <Headline text={copy.dictate} frames={LOCAL - DICTATE} accent={['PC']} />
      </Sequence>
      <Sequence from={LOCAL} layout="none">
        <Headline text={copy.local} frames={DEVICES_FRAMES} accent={['No']} />
      </Sequence>

      <div
        style={{
          position: 'absolute', left: WIN.left, top: WIN.top, width: WIN.width, height: WIN.height, opacity: winIn,
          transform: `perspective(1800px) translateX(${(1 - winIn) * -140}px) rotateY(${(1 - winIn) * 14}deg)`
        }}
      >
        <MacWindow title="airvoice" dark width={WIN.width} height={WIN.height} style={{ position: 'absolute', inset: 0, opacity: 1 - toEditor }}>
          <TerminalPane connected={connected} />
        </MacWindow>
        <MacWindow
          title="Notes"
          width={WIN.width}
          height={WIN.height}
          style={{ position: 'absolute', inset: 0, opacity: toEditor, transform: `scale(${0.97 + 0.03 * toEditor})` }}
        >
          <EditorPane text={pasted ? copy.spoken : ''} flash={flash} />
        </MacWindow>
      </div>

      <svg width={1920} height={1080} style={{ position: 'absolute', inset: 0, opacity: linkIn }}>
        <line
          x1={LINK.from} y1={LINK.y} x2={LINK.to} y2={LINK.y} stroke={color.blue} strokeWidth={4}
          strokeLinecap="round" strokeDasharray="2 14" strokeDashoffset={frame * 0.8}
        />
        {packet > 0 && packet < 1 && (
          <circle cx={LINK.to - (LINK.to - LINK.from) * packet} cy={LINK.y} r={11} fill={color.blue} />
        )}
      </svg>

      <div
        style={{
          position: 'absolute', left: LINK_MID, top: LINK.y - 112, transform: `translateX(-50%) scale(${local})`,
          opacity: Math.min(1, local)
        }}
      >
        <CloudOff />
      </div>
      <div
        style={{
          position: 'absolute', left: LINK_MID, top: LINK.y + 30, transform: `translateX(-50%) scale(${local})`,
          opacity: Math.min(1, local), whiteSpace: 'nowrap', fontFamily: sans, fontSize: 20, fontWeight: 600,
          color: color.blue, background: color.bg, border: `1px solid ${color.border}`, borderRadius: 999,
          padding: '8px 18px', boxShadow: '0 10px 30px rgba(15,23,42,0.08)'
        }}
      >
        {copy.lan}
      </div>

      <div style={{ position: 'absolute', left: PHONE.left, top: PHONE.top, opacity: phoneIn, transform: `translateY(${(1 - phoneIn) * 160}px)` }}>
        <Phone
          connected={connected}
          text={words.slice(0, heard).join(' ')}
          listening={listening}
          pressed={pressed}
          scanning={scanning}
        />
      </div>
    </AbsoluteFill>
  );
};
