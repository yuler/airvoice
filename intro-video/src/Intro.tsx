import { AbsoluteFill, Series } from 'remotion';
import { Background } from './components';
import { Devices, DEVICES_FRAMES, VOICE_AT, VOICE_SECONDS } from './scenes/Devices';
import { Hook, HOOK_FRAMES } from './scenes/Hook';
import { Install, INSTALL_FRAMES } from './scenes/Install';
import { Outro, OUTRO_FRAMES } from './scenes/Outro';
import { Platforms, PLATFORMS_FRAMES } from './scenes/Platforms';
import { Music } from './sound';
import { sec } from './theme';

const DEVICES_AT = HOOK_FRAMES + INSTALL_FRAMES;
const VOICE_START = DEVICES_AT + VOICE_AT;

export const INTRO_FRAMES = HOOK_FRAMES + INSTALL_FRAMES + DEVICES_FRAMES + PLATFORMS_FRAMES + OUTRO_FRAMES;

export const Intro = () => (
  <AbsoluteFill>
    <Background />
    <Music frames={INTRO_FRAMES} duck={[VOICE_START, VOICE_START + sec(VOICE_SECONDS)]} />
    <Series>
      <Series.Sequence durationInFrames={HOOK_FRAMES}>
        <Hook />
      </Series.Sequence>
      <Series.Sequence durationInFrames={INSTALL_FRAMES}>
        <Install />
      </Series.Sequence>
      <Series.Sequence durationInFrames={DEVICES_FRAMES}>
        <Devices />
      </Series.Sequence>
      <Series.Sequence durationInFrames={PLATFORMS_FRAMES}>
        <Platforms />
      </Series.Sequence>
      <Series.Sequence durationInFrames={OUTRO_FRAMES}>
        <Outro />
      </Series.Sequence>
    </Series>
  </AbsoluteFill>
);
