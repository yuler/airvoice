---
name: intro
description: Render and publish the Airvoice README intro video (Remotion, rebuilt device UIs, synthesized audio, cloned voice line). Use when the user asks to update the intro video, README hero video, launch video, materials/intro-poster.jpg, or mise intro:voice / intro:studio / intro:render.
---

# Intro

The intro is a ~25s launch video: hook copy, an install card, one persistent `Devices` scene (macOS window + phone) that pairs, dictates, and pastes, then platforms and an outro. Everything lives in `intro-video/`. Spec: `docs/specs/21-intro-video-design.md`.

Every device UI is rebuilt in React. Do not add screen capture: the phone app is native, and iOS Simulator cannot dictate. Text lands on the computer as a paste, because that is what Airvoice does. Do not show third-party keyboard branding.

## Preconditions

- `npm`, `uv` (audio synthesis), `ffmpeg`.
- Only for `intro:voice`: a local `~/Explore/explore-tts` checkout (`TTS_DIR` to override). The first run downloads about 4.5GB of models (OmniVoice and Whisper) into `~/.cache/huggingface`.

## Commands

From the repo root:

```bash
mise run intro:voice    # re-clone the spoken line → intro-video/public/voice.wav (committed)
mise run intro:studio   # preview in Remotion Studio
mise run intro:render   # audio + intro-video/out/intro.mp4 + materials/intro-poster.jpg
```

Publish: `gh release upload intro-video intro-video/out/intro.mp4 --clobber`. The first time, create the release with `gh release create intro-video intro-video/out/intro.mp4 --title "Intro video" --notes "README intro video asset." --latest=false`. `README.md` links the poster to `releases/download/intro-video/intro.mp4`, so the URL never changes. The MP4 is not committed.

## Layout

| Path                                | Role                                                               |
| ----------------------------------- | ------------------------------------------------------------------ |
| `intro-video/src/strings.ts`        | Every on-screen string and the spoken line                         |
| `intro-video/src/theme.ts`          | `DESIGN.md` colors, fonts, 1920×1080 @ 60fps, `sec()`              |
| `intro-video/src/scenes/Devices.tsx`| Pair → dictate → paste → LAN only timeline; `VOICE_SECONDS`        |
| `intro-video/src/devices/`          | `MacWindow`, `TerminalPane`, `EditorPane`, `Phone`                 |
| `intro-video/src/sound.tsx`         | `Music` (ducks under the voice), `Sfx`, `Voice`                    |
| `intro-video/audio/generate.py`     | Synthesizes `music.wav` + SFX into `public/audio/` (gitignored)    |
| `intro-video/audio/voice.sh`        | Clones `copy.spoken`, trims silence, prints the duration           |

## Fragile facts

| Pitfall                        | Required approach                                                                 |
| ------------------------------ | --------------------------------------------------------------------------------- |
| Changed the spoken line        | `mise run intro:voice`, then set `VOICE_SECONDS` to the printed duration (≤ 5.5s) |
| Poster frame                   | `--frame 950` in `package.json` must land after the paste highlight fades         |
| Release marked latest          | Always `--latest=false`; README and www rely on `releases/latest` for `v*` builds |
| `voice.sh` finds no line       | `spoken:` must stay a single-quoted one-liner in `strings.ts`                     |
| Tuning a frame                 | `./node_modules/.bin/remotion still src/index.ts Intro /tmp/f.jpg --frame N` in `intro-video/` |
