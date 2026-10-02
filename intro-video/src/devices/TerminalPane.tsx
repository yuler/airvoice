import React from 'react';
import { Img } from 'remotion';
import qr from '../../../www/public/qrcode.svg';
import { color, mono } from '../theme';

const Row: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div>
    <span style={{ color: color.terminalDim }}>{label.padEnd(16)}</span>
    {children}
  </div>
);

export const TerminalPane: React.FC<{ connected: boolean }> = ({ connected }) => (
  <div
    style={{
      position: 'absolute', inset: 0, padding: '44px 48px', display: 'flex', gap: 48,
      fontFamily: mono, fontSize: 22, lineHeight: 1.75, color: color.terminalText, whiteSpace: 'pre'
    }}
  >
    <div style={{ width: 300, height: 300, padding: 14, borderRadius: 16, background: '#ffffff', flexShrink: 0 }}>
      <Img src={qr} style={{ width: '100%', height: '100%' }} />
    </div>
    <div>
      <Row label="Token:">277129e4-35ea-40af-a122</Row>
      <Row label="WebSocket URL:">ws://192.168.1.24:7654/ws</Row>
      <div>{' '}</div>
      {connected ? (
        <>
          <Row label="Status:">
            <span style={{ color: color.green }}>Connected</span>
          </Row>
          <Row label="Device:">iPhone</Row>
        </>
      ) : (
        <div style={{ color: color.terminalDim }}>[airvoice] waiting for phone connection...</div>
      )}
    </div>
  </div>
);
