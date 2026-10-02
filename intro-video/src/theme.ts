import { loadFont as loadInter } from '@remotion/google-fonts/Inter';
import { loadFont as loadMono } from '@remotion/google-fonts/JetBrainsMono';

export const { fontFamily: sans } = loadInter('normal', { weights: ['400', '500', '600', '700', '800'], subsets: ['latin'] });
export const { fontFamily: mono } = loadMono('normal', { weights: ['400', '600'], subsets: ['latin'] });

export const FPS = 60;
export const WIDTH = 1920;
export const HEIGHT = 1080;

export const color = {
  text: '#171717',
  secondary: '#666666',
  muted: '#888888',
  bg: '#ffffff',
  bg2: '#fafafa',
  border: '#eaeaea',
  blue: '#006efe',
  green: '#28a948',
  yellow: '#ffae00',
  red: '#fc0035',
  terminal: '#0d0e15',
  terminalBar: '#15161f',
  terminalText: '#ededed',
  terminalDim: '#666666',
  terminalBorder: '#2e2e2e'
};

export const sec = (s: number) => Math.round(s * FPS);
