import React from 'react';
import { AbsoluteFill, interpolate, spring, useCurrentFrame, useVideoConfig } from 'remotion';
import { color, sans } from './theme';

export const clamp = { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' } as const;

export const progress = (frame: number, start: number, length: number) =>
  interpolate(frame, [start, start + length], [0, 1], clamp);

export const fadeOut = (frame: number, duration: number, length = 12) =>
  interpolate(frame, [duration - length, duration], [1, 0], clamp);

export const Background: React.FC = () => {
  const frame = useCurrentFrame();
  const drift = Math.sin(frame / 240) * 80;
  return (
    <AbsoluteFill style={{ background: color.bg2, overflow: 'hidden' }}>
      <div
        style={{
          position: 'absolute', width: 1700, height: 1700, left: 120 + drift, top: -1050,
          background: 'radial-gradient(closest-side, rgba(0,110,254,0.16), rgba(0,110,254,0) 70%)'
        }}
      />
      <div
        style={{
          position: 'absolute', width: 1400, height: 1400, right: -500 - drift, bottom: -950,
          background: 'radial-gradient(closest-side, rgba(40,169,72,0.08), rgba(40,169,72,0) 70%)'
        }}
      />
      <AbsoluteFill
        style={{
          backgroundImage:
            'linear-gradient(rgba(0,0,0,0.035) 1px, transparent 1px), linear-gradient(90deg, rgba(0,0,0,0.035) 1px, transparent 1px)',
          backgroundSize: '64px 64px',
          maskImage: 'radial-gradient(ellipse at 50% 30%, black 20%, transparent 75%)'
        }}
      />
    </AbsoluteFill>
  );
};

type WordsProps = {
  text: string;
  delay?: number;
  stagger?: number;
  size?: number;
  weight?: number;
  accent?: string[];
  style?: React.CSSProperties;
};

export const Words: React.FC<WordsProps> = ({ text, delay = 0, stagger = 4, size = 72, weight = 700, accent = [], style }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const words = text.split(' ');
  return (
    <div style={{ fontFamily: sans, fontSize: size, fontWeight: weight, letterSpacing: '-0.035em', color: color.text, lineHeight: 1.1, ...style }}>
      {words.map((word, i) => {
        const p = spring({ frame: frame - delay - i * stagger, fps, config: { damping: 18, stiffness: 140 } });
        const highlighted = accent.includes(word.replace(/[.,]/g, ''));
        return (
          <span
            key={i}
            style={{
              display: 'inline-block', marginRight: i < words.length - 1 ? '0.24em' : 0, opacity: p,
              transform: `translateY(${(1 - p) * 0.5}em)`, filter: `blur(${(1 - p) * 10}px)`,
              color: highlighted ? color.blue : undefined
            }}
          >
            {word}
          </span>
        );
      })}
    </div>
  );
};

export const Soundwave: React.FC<{ size: number; stroke?: string; strokeWidth?: number }> = ({ size, stroke = 'currentColor', strokeWidth = 2 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={stroke} strokeWidth={strokeWidth} strokeLinecap="round" strokeLinejoin="round">
    <path d="M2 10v4" />
    <path d="M6 6v12" />
    <path d="M10 3v18" />
    <path d="M14 6v12" />
    <path d="M18 10v4" />
    <path d="M22 12v0" />
  </svg>
);

export const StatusDot: React.FC<{ on: boolean; size?: number }> = ({ on, size = 10 }) => (
  <div
    style={{
      width: size, height: size, borderRadius: size / 2, flexShrink: 0, background: on ? color.green : color.yellow,
      boxShadow: on ? `0 0 ${size}px rgba(40,169,72,0.55)` : undefined
    }}
  />
);
