import { useCurrentFrame } from 'remotion';
import { copy } from '../strings';
import { color, sans } from '../theme';

export const EditorPane: React.FC<{ text: string; flash: number }> = ({ text, flash }) => {
  const frame = useCurrentFrame();
  const caretOn = Math.floor(frame / 30) % 2 === 0;
  return (
    <div style={{ position: 'absolute', inset: 0, padding: '52px 68px', fontFamily: sans, color: color.text }}>
      <div style={{ fontSize: 44, fontWeight: 700, letterSpacing: '-0.03em' }}>{copy.editorTitle}</div>
      <div style={{ fontSize: 20, color: color.muted, marginTop: 8 }}>{copy.editorDate}</div>
      <div style={{ fontSize: 30, lineHeight: 1.55, marginTop: 36, color: color.secondary }}>{copy.editorLine}</div>
      <div style={{ fontSize: 30, lineHeight: 1.55, marginTop: 14 }}>
        <span style={{ background: `rgba(0,110,254,${0.18 * flash})`, borderRadius: 6 }}>{text}</span>
        <span
          style={{
            display: 'inline-block', width: 3, height: '1.15em', verticalAlign: '-0.22em', marginLeft: 2,
            background: color.blue, opacity: caretOn ? 1 : 0
          }}
        />
      </div>
    </div>
  );
};
