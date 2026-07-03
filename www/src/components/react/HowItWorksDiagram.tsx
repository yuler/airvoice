import { type CSSProperties } from 'react';
import type { HowItWorksMessages } from '../../i18n/messages';

interface HowItWorksDiagramProps {
  labels: HowItWorksMessages['diagram'];
}

export default function HowItWorksDiagram({ labels }: HowItWorksDiagramProps) {
  const cycle = 5; // 5-second complete cycle
  const isZh = labels.sampleText.includes('你好');
  const rectWidth = isZh ? 68 : 84;
  const foWidth = isZh ? 58 : 74;
  const typedMaxWidth = isZh ? '54px' : '70px';

  return (
    <div
      className="relative w-full select-none"
      aria-hidden="true"
      style={{ '--av-typed-max-width': typedMaxWidth } as CSSProperties}
    >
      <style>{`
        /* ─── iOS-style voice waveform flow ─── */
        @keyframes av-wave-flow-1 {
          from { transform: translateX(0); }
          to { transform: translateX(-40px); }
        }
        @keyframes av-wave-flow-2 {
          from { transform: translateX(0); }
          to { transform: translateX(-30px); }
        }
        @keyframes av-wave-flow-3 {
          from { transform: translateX(0); }
          to { transform: translateX(-50px); }
        }
        .av-wave-1 { animation: av-wave-flow-1 2s linear infinite; }
        .av-wave-2 { animation: av-wave-flow-2 1.5s linear infinite; }
        .av-wave-3 { animation: av-wave-flow-3 2.8s linear infinite; }

        /* ─── Voice amplitude breathing & status dimming ─── */
        @keyframes av-wave-breathe {
          0%, 100% { transform: scaleY(0.5); }
          20% { transform: scaleY(1.2); }
          40% { transform: scaleY(0.7); }
          65% { transform: scaleY(1.3); }
          85% { transform: scaleY(0.9); }
        }
        @keyframes av-wave-opacity {
          0%, 75% { opacity: 1; }
          80%, 95% { opacity: 0.15; }
          100% { opacity: 1; }
        }
        .av-wave-breathe {
          transform-origin: 44px 104px;
          animation:
            av-wave-breathe 3.5s ease-in-out infinite,
            av-wave-opacity ${cycle}s ease-in-out infinite;
        }

        /* ─── Data flow (Send & ACK marching) ─── */
        @keyframes av-flow-send {
          0%, 10% { stroke-dashoffset: 0; opacity: 0.2; }
          15%, 45% { stroke-dashoffset: -24; opacity: 0.8; }
          50%, 100% { stroke-dashoffset: 0; opacity: 0.2; }
        }
        .av-flow-line-send {
          stroke-dasharray: 6 6;
          animation: av-flow-send ${cycle}s linear infinite;
        }

        @keyframes av-flow-ack {
          0%, 60% { stroke-dashoffset: 0; opacity: 0.2; }
          65%, 85% { stroke-dashoffset: 24; opacity: 0.8; }
          90%, 100% { stroke-dashoffset: 0; opacity: 0.2; }
        }
        .av-flow-line-ack {
          stroke-dasharray: 6 6;
          animation: av-flow-ack ${cycle}s linear infinite;
        }

        /* ─── Horizontal packets ─── */
        @keyframes av-packet-1 {
          0%, 15% { cx: 168; opacity: 0; }
          17% { opacity: 1; }
          28% { opacity: 1; }
          30%, 100% { cx: 320; opacity: 0; }
        }
        @keyframes av-packet-2 {
          0%, 30% { cx: 472; opacity: 0; }
          32% { opacity: 1; }
          43% { opacity: 1; }
          45%, 100% { cx: 624; opacity: 0; }
        }
        .av-pkt-1 { animation: av-packet-1 ${cycle}s ease-in-out infinite; }
        .av-pkt-2 { animation: av-packet-2 ${cycle}s ease-in-out infinite; }

        @keyframes av-packet-ack-1 {
          0%, 65% { cx: 624; opacity: 0; }
          67% { opacity: 1; }
          73% { opacity: 1; }
          75%, 100% { cx: 472; opacity: 0; }
        }
        @keyframes av-packet-ack-2 {
          0%, 75% { cx: 320; opacity: 0; }
          77% { opacity: 1; }
          83% { opacity: 1; }
          85%, 100% { cx: 168; opacity: 0; }
        }
        .av-pkt-ack-1 { animation: av-packet-ack-1 ${cycle}s ease-in-out infinite; }
        .av-pkt-ack-2 { animation: av-packet-ack-2 ${cycle}s ease-in-out infinite; }

        /* ─── Vertical packets (Mobile) ─── */
        @keyframes av-packet-v1 {
          0%, 15% { cy: 176; opacity: 0; }
          17% { opacity: 1; }
          28% { opacity: 1; }
          30%, 100% { cy: 210; opacity: 0; }
        }
        @keyframes av-packet-v2 {
          0%, 30% { cy: 310; opacity: 0; }
          32% { opacity: 1; }
          43% { opacity: 1; }
          45%, 100% { cy: 380; opacity: 0; }
        }
        .av-pkt-v1 { animation: av-packet-v1 ${cycle}s ease-in-out infinite; }
        .av-pkt-v2 { animation: av-packet-v2 ${cycle}s ease-in-out infinite; }

        @keyframes av-packet-ack-v1 {
          0%, 65% { cy: 380; opacity: 0; }
          67% { opacity: 1; }
          73% { opacity: 1; }
          75%, 100% { cy: 310; opacity: 0; }
        }
        @keyframes av-packet-ack-v2 {
          0%, 75% { cy: 210; opacity: 0; }
          77% { opacity: 1; }
          83% { opacity: 1; }
          85%, 100% { cy: 176; opacity: 0; }
        }
        .av-pkt-ack-v1 { animation: av-packet-ack-v1 ${cycle}s ease-in-out infinite; }
        .av-pkt-ack-v2 { animation: av-packet-ack-v2 ${cycle}s ease-in-out infinite; }

        /* ─── Speech bubble ─── */
        @keyframes av-bubble-in {
          0%, 80% { opacity: 1; transform: scale(1); }
          85%, 95% { opacity: 0; transform: scale(0.8); }
          100% { opacity: 1; transform: scale(1); }
        }
        .av-bubble {
          transform-origin: 38px 24px;
          animation: av-bubble-in ${cycle}s ease-in-out infinite;
        }

        /* ─── Bubble text clear (ACK trigger) ─── */
        @keyframes av-bubble-text {
          0%, 75% { opacity: 1; }
          76%, 95% { opacity: 0; }
          100% { opacity: 1; }
        }
        .av-bubble-text {
          animation: av-bubble-text ${cycle}s step-end infinite;
        }

        /* ─── Typing reveal ─── */
        @keyframes av-type {
          0%, 45% { width: 0; }
          60%, 85% { width: var(--av-typed-max-width, 70px); }
          90%, 100% { width: 0; }
        }
        .av-typed {
          display: inline-block;
          overflow: hidden;
          white-space: nowrap;
          vertical-align: bottom;
          animation: av-type ${cycle}s steps(14) infinite;
        }

        /* ─── Cursor blink ─── */
        @keyframes av-cursor-blink {
          0%, 100% { opacity: 1; }
          50% { opacity: 0; }
        }
        @keyframes av-cursor-appear {
          0%, 43% { opacity: 0; }
          45%, 88% { opacity: 1; }
          90%, 100% { opacity: 0; }
        }
        .av-cursor {
          animation:
            av-cursor-appear ${cycle}s step-end infinite,
            av-cursor-blink 0.6s step-end infinite;
        }

        /* ─── PC text highlight ─── */
        @keyframes av-text-hl {
          0%, 60% { opacity: 0; }
          62%, 68% { opacity: 1; }
          70%, 100% { opacity: 0; }
        }
        .av-text-hl {
          animation: av-text-hl ${cycle}s ease-in-out infinite;
        }

        /* ─── Bridge ─── */
        @keyframes av-ring-pulse {
          0%   { r: 18; stroke-opacity: 0.6; }
          50%  { r: 26; stroke-opacity: 0; }
          100% { r: 26; stroke-opacity: 0; }
        }
        .av-ring-pulse {
          animation: av-ring-pulse 2.4s ease-out infinite;
        }

        @keyframes av-bridge-glow {
          0%, 30%, 100% { stroke-opacity: 0.25; }
          50%, 70% { stroke-opacity: 0.8; }
        }
        .av-bridge-ring {
          animation: av-bridge-glow ${cycle}s ease-in-out infinite;
        }

        /* ─── Phone screen shimmer ─── */
        @keyframes av-screen-shimmer {
          0%, 100% { opacity: 0.04; }
          50% { opacity: 0.08; }
        }
        .av-screen-shimmer {
          animation: av-screen-shimmer 3s ease-in-out infinite;
        }
      `}</style>

      {/* Defs to share between both SVGs */}
      <svg style={{ position: 'absolute', width: 0, height: 0 }} aria-hidden="true">
        <defs>
          <filter id="av-glow-filter" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <filter id="av-ring-glow" x="-100%" y="-100%" width="300%" height="300%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="2" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <filter id="av-bubble-glow" x="-20%" y="-20%" width="140%" height="140%">
            <feDropShadow dx="0" dy="1" stdDeviation="3" floodColor="#006efe" floodOpacity="0.15" />
          </filter>
          <clipPath id="av-wave-clip">
            <rect x="12" y="88" width="64" height="32" rx="4" />
          </clipPath>
        </defs>
      </svg>

      {/* Desktop View: Horizontal Layout (>= sm) */}
      <div className="hidden sm:block w-full max-w-4xl mx-auto">
        <svg
          viewBox="0 0 800 260"
          className="w-full h-auto"
          role="img"
          aria-label={labels.ariaLabel}
        >
          {/* Connection paths: Top (Send - Blue) */}
          <path d="M 168 112 L 320 112" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 472 112 L 624 112" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 168 112 L 320 112" fill="none" stroke="#006efe" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-send" />
          <path d="M 472 112 L 624 112" fill="none" stroke="#006efe" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-send" />

          {/* Connection paths: Bottom (ACK - Green) */}
          <path d="M 624 124 L 472 124" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 320 124 L 168 124" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 624 124 L 472 124" fill="none" stroke="#22c55e" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-ack" />
          <path d="M 320 124 L 168 124" fill="none" stroke="#22c55e" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-ack" />

          {/* Animated Send packets (Blue, moving left to right) */}
          <circle cx="-100" cy="112" r="4" fill="#006efe" opacity="0" className="av-pkt-1" filter="url(#av-glow-filter)" />
          <circle cx="-100" cy="112" r="4" fill="#006efe" opacity="0" className="av-pkt-2" filter="url(#av-glow-filter)" />

          {/* Animated ACK packets (Green, moving right to left) */}
          <circle cx="-100" cy="124" r="4" fill="#22c55e" opacity="0" className="av-pkt-ack-1" filter="url(#av-glow-filter)" />
          <circle cx="-100" cy="124" r="4" fill="#22c55e" opacity="0" className="av-pkt-ack-2" filter="url(#av-glow-filter)" />

          {/* Phone */}
          <g transform="translate(56, 40)">
            <rect x="0" y="0" width="88" height="156" rx="14" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="2" />
            <rect x="10" y="14" width="68" height="108" rx="6" fill="var(--background-secondary)" />
            <rect x="10" y="14" width="68" height="108" rx="6" fill="#006efe" className="av-screen-shimmer" />

            <g transform="translate(6, 52)">
              <g className="av-bubble" filter="url(#av-bubble-glow)">
                <path
                  d="M 8,0 H 68 Q 76,0 76,8 V 16 Q 76,24 68,24 H 46 L 38,31 L 30,24 H 8 Q 0,24 0,16 V 8 Q 0,0 8,0 Z"
                  fill="var(--background-primary)"
                  stroke="#006efe"
                  strokeWidth="1.5"
                />
                <text x="38" y="16" textAnchor="middle" fontSize="10" fontWeight="600" fill="var(--primary-text)" className="av-bubble-text">
                  {labels.sampleText}
                </text>
              </g>
            </g>

            <line x1="14" y1="104" x2="74" y2="104" stroke="#006efe" strokeWidth="0.5" opacity="0.1" />

            <g clipPath="url(#av-wave-clip)">
              <g className="av-wave-breathe">
                <g className="av-wave-1">
                  <path
                    d="M -40,104 Q -30,96 -20,104 Q -10,112 0,104 Q 10,96 20,104 Q 30,112 40,104 Q 50,96 60,104 Q 70,112 80,104 Q 90,96 100,104 Q 110,112 120,104"
                    fill="none" stroke="#006efe" strokeWidth="2.5" strokeLinecap="round" opacity="0.7"
                  />
                </g>
                <g className="av-wave-2">
                  <path
                    d="M -30,104 Q -22.5,99 -15,104 Q -7.5,109 0,104 Q 7.5,99 15,104 Q 22.5,109 30,104 Q 37.5,99 45,104 Q 52.5,109 60,104 Q 67.5,99 75,104 Q 82.5,109 90,104 Q 97.5,99 105,104"
                    fill="none" stroke="#006efe" strokeWidth="1.5" strokeLinecap="round" opacity="0.35"
                  />
                </g>
                <g className="av-wave-3">
                  <path
                    d="M -50,104 Q -37.5,92 -25,104 Q -12.5,116 0,104 Q 12.5,92 25,104 Q 37.5,116 50,104 Q 62.5,92 75,104 Q 87.5,116 100,104 Q 112.5,92 125,104 Q 137.5,116 150,104"
                    fill="none" stroke="#006efe" strokeWidth="1" strokeLinecap="round" opacity="0.15"
                  />
                </g>
              </g>
            </g>

            <circle cx="44" cy="104" r="11" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="1.5" />
            <circle cx="44" cy="104" r="11" fill="none" stroke="#006efe" strokeWidth="1" strokeOpacity="0.15" />
            <rect x="41.5" y="98" width="5" height="8" rx="2.5" fill="none" stroke="var(--secondary-text)" strokeWidth="1.5" />
            <path d="M39 105 C39 107.8 41.2 110 44 110 C46.8 110 49 107.8 49 105" fill="none" stroke="var(--secondary-text)" strokeWidth="1.5" strokeLinecap="round" />
            <line x1="44" y1="110" x2="44" y2="112" stroke="var(--secondary-text)" strokeWidth="1.5" strokeLinecap="round" />

            <rect x="30" y="132" width="28" height="4" rx="2" fill="var(--border-default)" />

            <text x="44" y="178" textAnchor="middle" fontSize="11" fontWeight="600" fill="var(--primary-text)">
              {labels.phone}
            </text>
            <text x="44" y="194" textAnchor="middle" fontSize="9" fill="var(--muted-text)">
              {labels.phoneDesc}
            </text>
          </g>

          {/* Bridge */}
          <g transform="translate(320, 68)">
            <rect x="0" y="0" width="152" height="100" rx="16" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="2" />
            <rect x="8" y="8" width="136" height="84" rx="12" fill="var(--background-secondary)" opacity="0.38" />
            <rect x="8" y="8" width="136" height="84" rx="12" fill="none" stroke="#006efe" strokeWidth="1.25" strokeOpacity="0.28" className="av-bridge-ring" />

            <g transform="translate(76, 42)">
              <circle cx="0" cy="0" r="22" fill="none" stroke="#006efe" strokeWidth="1" className="av-ring-pulse" filter="url(#av-ring-glow)" />
              <circle cx="0" cy="0" r="22" fill="none" stroke="#006efe" strokeWidth="1" className="av-ring-pulse" style={{ animationDelay: '1.2s' }} filter="url(#av-ring-glow)" />

              <rect x="-26" y="-17" width="52" height="34" rx="9" fill="var(--background-primary)" stroke="#006efe" strokeWidth="1.5" />
              <rect x="-20" y="-11" width="40" height="22" rx="6" fill="#006efe" opacity="0.08" />
              <text x="0" y="4" textAnchor="middle" fontSize="11" fontWeight="700" fill="#006efe" fontFamily="ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace">
                WS
              </text>

              <line x1="-48" y1="0" x2="-26" y2="0" stroke="#006efe" strokeWidth="1.5" strokeLinecap="round" strokeOpacity="0.75" />
              <line x1="26" y1="0" x2="48" y2="0" stroke="#006efe" strokeWidth="1.5" strokeLinecap="round" strokeOpacity="0.75" />
              <circle cx="-52" cy="0" r="4" fill="var(--background-primary)" stroke="#006efe" strokeWidth="1.5" />
              <circle cx="52" cy="0" r="4" fill="var(--background-primary)" stroke="#006efe" strokeWidth="1.5" />
              <circle cx="-52" cy="0" r="1.8" fill="#006efe" />
              <circle cx="52" cy="0" r="1.8" fill="#006efe" />
            </g>

            <text x="76" y="116" textAnchor="middle" fontSize="11" fontWeight="600" fill="var(--primary-text)">
              {labels.bridge}
            </text>
            <text x="76" y="132" textAnchor="middle" fontSize="9" fill="var(--muted-text)">
              {labels.lanOnly}
            </text>
          </g>

          {/* PC */}
          <g transform="translate(624, 76)">
            <rect x="0" y="0" width="120" height="84" rx="8" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="2" />
            <rect x="8" y="10" width="104" height="60" rx="4" fill="var(--background-secondary)" />

            <circle cx="18" cy="20" r="2.5" fill="var(--border-default)" />
            <circle cx="27" cy="20" r="2.5" fill="var(--border-default)" />
            <circle cx="36" cy="20" r="2.5" fill="var(--border-default)" />

            <rect x="16" y="30" width="40" height="3" rx="1.5" fill="var(--border-default)" opacity="0.5" />
            <rect x="16" y="37" width="60" height="3" rx="1.5" fill="var(--border-default)" opacity="0.35" />

            <rect x="12" y="44" width={rectWidth} height="16" rx="3" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="0.5" className="av-text-hl" />
            <text x="15" y="55" fontSize="10" fontWeight="700" fill="#006efe" opacity="0.6" fontFamily="ui-monospace, monospace">›</text>

            <foreignObject x="22" y="44" width={foWidth} height="20">
              <div
                xmlns="http://www.w3.org/1999/xhtml"
                style={{
                  fontSize: '10px',
                  fontWeight: 600,
                  color: 'var(--primary-text)',
                  fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <span className="av-typed">{labels.sampleText}</span>
                <span className="av-cursor" style={{ color: '#006efe', marginLeft: '1px', fontWeight: 400 }}>|</span>
              </div>
            </foreignObject>

            <rect x="44" y="88" width="32" height="8" rx="2" fill="var(--border-default)" />
            <rect x="28" y="96" width="64" height="4" rx="2" fill="var(--border-default)" />

            <text x="60" y="120" textAnchor="middle" fontSize="11" fontWeight="600" fill="var(--primary-text)">
              {labels.pc}
            </text>
            <text x="60" y="136" textAnchor="middle" fontSize="9" fill="var(--muted-text)">
              {labels.pcDesc}
            </text>
          </g>
        </svg>
      </div>

      {/* Mobile View: Vertical Layout (< sm) */}
      <div className="sm:hidden w-full max-w-[280px] mx-auto">
        <svg
          viewBox="0 0 260 520"
          className="w-full h-auto"
          role="img"
          aria-label={labels.ariaLabel}
        >
          {/* Vertical Connection Paths: Left (Send - Blue) */}
          <path d="M 68 176 L 68 210" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 68 310 L 68 380" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 68 176 L 68 210" fill="none" stroke="#006efe" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-send" />
          <path d="M 68 310 L 68 380" fill="none" stroke="#006efe" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-send" />

          {/* Vertical Connection Paths: Right (ACK - Green) */}
          <path d="M 84 380 L 84 310" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 84 210 L 84 176" fill="none" stroke="var(--border-default)" strokeWidth="1.5" strokeOpacity="0.5" />
          <path d="M 84 380 L 84 310" fill="none" stroke="#22c55e" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-ack" />
          <path d="M 84 210 L 84 176" fill="none" stroke="#22c55e" strokeWidth="1.5" strokeOpacity="0.35" className="av-flow-line-ack" />

          {/* Animated Send packets (Blue, moving down) */}
          <circle cx="68" cy="-100" r="4" fill="#006efe" opacity="0" className="av-pkt-v1" filter="url(#av-glow-filter)" />
          <circle cx="68" cy="-100" r="4" fill="#006efe" opacity="0" className="av-pkt-v2" filter="url(#av-glow-filter)" />

          {/* Animated ACK packets (Green, moving up) */}
          <circle cx="84" cy="-100" r="4" fill="#22c55e" opacity="0" className="av-pkt-ack-v1" filter="url(#av-glow-filter)" />
          <circle cx="84" cy="-100" r="4" fill="#22c55e" opacity="0" className="av-pkt-ack-v2" filter="url(#av-glow-filter)" />

          {/* Phone (left-aligned, text on right) */}
          <g transform="translate(32, 20)">
            <rect x="0" y="0" width="88" height="156" rx="14" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="2" />
            <rect x="10" y="14" width="68" height="108" rx="6" fill="var(--background-secondary)" />
            <rect x="10" y="14" width="68" height="108" rx="6" fill="#006efe" className="av-screen-shimmer" />

            <g transform="translate(6, 52)">
              <g className="av-bubble" filter="url(#av-bubble-glow)">
                <path
                  d="M 8,0 H 68 Q 76,0 76,8 V 16 Q 76,24 68,24 H 46 L 38,31 L 30,24 H 8 Q 0,24 0,16 V 8 Q 0,0 8,0 Z"
                  fill="var(--background-primary)"
                  stroke="#006efe"
                  strokeWidth="1.5"
                />
                <text x="38" y="16" textAnchor="middle" fontSize="10" fontWeight="600" fill="var(--primary-text)" className="av-bubble-text">
                  {labels.sampleText}
                </text>
              </g>
            </g>

            <line x1="14" y1="104" x2="74" y2="104" stroke="#006efe" strokeWidth="0.5" opacity="0.1" />

            <g clipPath="url(#av-wave-clip)">
              <g className="av-wave-breathe">
                <g className="av-wave-1">
                  <path
                    d="M -40,104 Q -30,96 -20,104 Q -10,112 0,104 Q 10,96 20,104 Q 30,112 40,104 Q 50,96 60,104 Q 70,112 80,104 Q 90,96 100,104 Q 110,112 120,104"
                    fill="none" stroke="#006efe" strokeWidth="2.5" strokeLinecap="round" opacity="0.7"
                  />
                </g>
                <g className="av-wave-2">
                  <path
                    d="M -30,104 Q -22.5,99 -15,104 Q -7.5,109 0,104 Q 7.5,99 15,104 Q 22.5,109 30,104 Q 37.5,99 45,104 Q 52.5,109 60,104 Q 67.5,99 75,104 Q 82.5,109 90,104 Q 97.5,99 105,104"
                    fill="none" stroke="#006efe" strokeWidth="1.5" strokeLinecap="round" opacity="0.35"
                  />
                </g>
                <g className="av-wave-3">
                  <path
                    d="M -50,104 Q -37.5,92 -25,104 Q -12.5,116 0,104 Q 12.5,92 25,104 Q 37.5,116 50,104 Q 62.5,92 75,104 Q 87.5,116 100,104 Q 112.5,92 125,104 Q 137.5,116 150,104"
                    fill="none" stroke="#006efe" strokeWidth="1" strokeLinecap="round" opacity="0.15"
                  />
                </g>
              </g>
            </g>

            <circle cx="44" cy="104" r="11" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="1.5" />
            <circle cx="44" cy="104" r="11" fill="none" stroke="#006efe" strokeWidth="1" strokeOpacity="0.15" />
            <rect x="41.5" y="98" width="5" height="8" rx="2.5" fill="none" stroke="var(--secondary-text)" strokeWidth="1.5" />
            <path d="M39 105 C39 107.8 41.2 110 44 110 C46.8 110 49 107.8 49 105" fill="none" stroke="var(--secondary-text)" strokeWidth="1.5" strokeLinecap="round" />
            <line x1="44" y1="110" x2="44" y2="112" stroke="var(--secondary-text)" strokeWidth="1.5" strokeLinecap="round" />

            <rect x="30" y="132" width="28" height="4" rx="2" fill="var(--border-default)" />
          </g>

          {/* Phone Labels on the right */}
          <text x="136" y="88" textAnchor="start" fontSize="12" fontWeight="600" fill="var(--primary-text)">
            {labels.phone}
          </text>
          <text x="136" y="106" textAnchor="start" fontSize="10" fill="var(--muted-text)">
            {labels.phoneDesc}
          </text>

          {/* Bridge (left-aligned, text on right) */}
          <g transform="translate(0, 210)">
            <rect x="0" y="0" width="152" height="100" rx="16" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="2" />
            <rect x="8" y="8" width="136" height="84" rx="12" fill="var(--background-secondary)" opacity="0.38" />
            <rect x="8" y="8" width="136" height="84" rx="12" fill="none" stroke="#006efe" strokeWidth="1.25" strokeOpacity="0.28" className="av-bridge-ring" />

            <g transform="translate(76, 42)">
              <circle cx="0" cy="0" r="22" fill="none" stroke="#006efe" strokeWidth="1" className="av-ring-pulse" filter="url(#av-ring-glow)" />
              <circle cx="0" cy="0" r="22" fill="none" stroke="#006efe" strokeWidth="1" className="av-ring-pulse" style={{ animationDelay: '1.2s' }} filter="url(#av-ring-glow)" />

              <rect x="-26" y="-17" width="52" height="34" rx="9" fill="var(--background-primary)" stroke="#006efe" strokeWidth="1.5" />
              <rect x="-20" y="-11" width="40" height="22" rx="6" fill="#006efe" opacity="0.08" />
              <text x="0" y="4" textAnchor="middle" fontSize="11" fontWeight="700" fill="#006efe" fontFamily="ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace">
                WS
              </text>

              <line x1="-48" y1="0" x2="-26" y2="0" stroke="#006efe" strokeWidth="1.5" strokeLinecap="round" strokeOpacity="0.75" />
              <line x1="26" y1="0" x2="48" y2="0" stroke="#006efe" strokeWidth="1.5" strokeLinecap="round" strokeOpacity="0.75" />
              <circle cx="-52" cy="0" r="4" fill="var(--background-primary)" stroke="#006efe" strokeWidth="1.5" />
              <circle cx="52" cy="0" r="4" fill="var(--background-primary)" stroke="#006efe" strokeWidth="1.5" />
              <circle cx="-52" cy="0" r="1.8" fill="#006efe" />
              <circle cx="52" cy="0" r="1.8" fill="#006efe" />
            </g>
          </g>

          {/* Bridge Labels on the right */}
          <text x="164" y="256" textAnchor="start" fontSize="12" fontWeight="600" fill="var(--primary-text)">
            {labels.bridge}
          </text>
          <text x="164" y="274" textAnchor="start" fontSize="10" fill="var(--muted-text)">
            {labels.lanOnly}
          </text>

          {/* PC (left-aligned, text on right) */}
          <g transform="translate(16, 380)">
            <rect x="0" y="0" width="120" height="84" rx="8" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="2" />
            <rect x="8" y="10" width="104" height="60" rx="4" fill="var(--background-secondary)" />

            <circle cx="18" cy="20" r="2.5" fill="var(--border-default)" />
            <circle cx="27" cy="20" r="2.5" fill="var(--border-default)" />
            <circle cx="36" cy="20" r="2.5" fill="var(--border-default)" />

            <rect x="16" y="30" width="40" height="3" rx="1.5" fill="var(--border-default)" opacity="0.5" />
            <rect x="16" y="37" width="60" height="3" rx="1.5" fill="var(--border-default)" opacity="0.35" />

            <rect x="12" y="44" width={rectWidth} height="16" rx="3" fill="var(--background-primary)" stroke="var(--border-default)" strokeWidth="0.5" className="av-text-hl" />
            <text x="15" y="55" fontSize="10" fontWeight="700" fill="#006efe" opacity="0.6" fontFamily="ui-monospace, monospace">›</text>

            <foreignObject x="22" y="44" width={foWidth} height="20">
              <div
                xmlns="http://www.w3.org/1999/xhtml"
                style={{
                  fontSize: '10px',
                  fontWeight: 600,
                  color: 'var(--primary-text)',
                  fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <span className="av-typed">{labels.sampleText}</span>
                <span className="av-cursor" style={{ color: '#006efe', marginLeft: '1px', fontWeight: 400 }}>|</span>
              </div>
            </foreignObject>

            <rect x="44" y="88" width="32" height="8" rx="2" fill="var(--border-default)" />
            <rect x="28" y="96" width="64" height="4" rx="2" fill="var(--border-default)" />
          </g>

          {/* PC Labels on the right */}
          <text x="148" y="422" textAnchor="start" fontSize="12" fontWeight="600" fill="var(--primary-text)">
            {labels.pc}
          </text>
          <text x="148" y="440" textAnchor="start" fontSize="10" fill="var(--muted-text)">
            {labels.pcDesc}
          </text>
        </svg>
      </div>
    </div>
  );
}
