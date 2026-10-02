# Intro Video Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a ~25s Remotion-rendered README intro: speak on the phone, the text lands at the cursor on the computer.

**Architecture:** A standalone Remotion project in `intro-video/`. Five scenes in a `Series`; scenes 3–5 of the storyboard are one persistent `Devices` scene (rebuilt macOS window + rebuilt phone). Music and SFX are synthesized by a Python script; one spoken line is cloned locally with OmniVoice and committed. `mise intro:*` tasks drive it; the MP4 is a GitHub Release asset.

**Tech Stack:** Remotion 4.0.532, React 19, TypeScript, npm, `uv` + numpy/scipy, OmniVoice (`~/Explore/explore-tts`), ffmpeg.

## Global Constraints

- Spec: `docs/specs/21-intro-video-design.md`
- 1920×1080 @ 60fps, total 1500 frames (25s), H.264, < 6MB
- Light theme; only accent `#006efe`; colors from `DESIGN.md`
- All on-screen copy in `intro-video/src/strings.ts`, English only
- No screen capture, no third-party keyboard branding
- Text lands as a paste (with highlight), not per-character typing
- MP4 and synthesized audio are gitignored; `intro-video/public/voice.wav` and `materials/intro-poster.jpg` are committed
- Release `intro-video` is created with `--latest=false` (README/www rely on `releases/latest`)
- Plans live only under `docs/plans/`

## File Structure

| File | Responsibility |
| --- | --- |
| `intro-video/package.json`, `tsconfig.json` | Remotion project, scripts |
| `intro-video/src/index.ts`, `Root.tsx`, `Intro.tsx` | Composition + scene series + music |
| `intro-video/src/theme.ts` | Fonts, size, fps, colors, `sec()` |
| `intro-video/src/strings.ts` | Every on-screen string + spoken line |
| `intro-video/src/components.tsx` | Background, `Words`, `Soundwave`, `StatusDot`, timing helpers |
| `intro-video/src/icons.tsx` | Platform icons (copied from `www/.../WorksEverywhere.tsx`) |
| `intro-video/src/sound.tsx` | `Music` (with ducking), `Sfx`, `Voice` |
| `intro-video/src/devices/*.tsx` | `MacWindow`, `TerminalPane`, `EditorPane`, `Phone` |
| `intro-video/src/scenes/*.tsx` | `Hook`, `Install`, `Devices`, `Platforms`, `Outro` |
| `intro-video/audio/generate.py` | Music + SFX → `public/audio/` |
| `intro-video/audio/voice.sh` | Cloned voice → trimmed `public/voice.wav` |
| `mise.toml`, `.gitignore`, `README.md` | Tasks, ignores, poster link |
| `.agents/skills/intro/SKILL.md` | How to update and publish the intro |

Verification is visual: `npm run typecheck`, then `npx remotion still` at fixed frames and inspect the JPEGs.

---

### Task 1: Project scaffold, shared pieces, and the text-only scenes

**Files:** Create `intro-video/{package.json,tsconfig.json}`, `intro-video/src/{index.ts,Root.tsx,Intro.tsx,assets.d.ts,theme.ts,strings.ts,components.tsx,icons.tsx,sound.tsx}`, `intro-video/src/scenes/{Hook,Install,Platforms,Outro}.tsx`, `intro-video/audio/generate.py`.

**Interfaces produced:** `sec(s)`, `color`, `sans`, `mono`, `copy`, `clamp`, `progress(frame, start, length)`, `fadeOut(frame, duration, length?)`, `Words`, `Soundwave`, `StatusDot`, `Sfx`, `Music({ frames, duck })`, `Voice({ at })`, `HOOK_FRAMES`, `INSTALL_FRAMES`, `PLATFORMS_FRAMES`, `OUTRO_FRAMES`.

- [ ] **Step 1:** Write the files. `package.json`:

```json
{
  "name": "airvoice-intro",
  "private": true,
  "type": "module",
  "scripts": {
    "audio": "uv run --quiet --script audio/generate.py",
    "voice": "bash audio/voice.sh",
    "typecheck": "tsc",
    "studio": "remotion studio src/index.ts",
    "render": "remotion render src/index.ts Intro out/intro.mp4 --codec h264 --crf 23 --audio-bitrate 128k",
    "poster": "remotion still src/index.ts Intro ../materials/intro-poster.jpg --frame 950 --jpeg-quality 90"
  },
  "dependencies": {
    "@remotion/cli": "4.0.532",
    "@remotion/google-fonts": "4.0.532",
    "react": "^19.3.0",
    "react-dom": "^19.3.0",
    "remotion": "4.0.532"
  },
  "devDependencies": {
    "@types/react": "^19.0.0",
    "typescript": "^5.6.0"
  }
}
```

The remaining files are as committed in `intro-video/` (theme tokens from `DESIGN.md`; `Words`/`fadeOut`/`Sfx`/`Music`/`generate.py` ported from Rails Studio with `LENGTH = 25.0` and a light background; `Music` takes `duck: [from, to]` and dips to 0.3 there).

Scene timings: Hook 3s, Install 3.5s (types `brew tap yuler/airvoice https://github.com/yuler/airvoice`, `brew install airvoice`, `airvoice` at 1.1 chars/frame), Platforms 2.5s (five cards, staggered springs), Outro 3s (blue soundwave tile, "Airvoice", tagline, `airvoice.yuler.cc` + `github.com/yuler/airvoice` pills).

- [ ] **Step 2:** `cd intro-video && npm install && npm run audio && npm run typecheck` — expected: no errors, `public/audio/{music,click,whoosh,pop,typing}.wav` exist.
- [ ] **Step 3:** `npx remotion still src/index.ts Intro /tmp/av-hook.jpg --frame 120` (and Install at 300) — inspect: centered copy, light gradient, blue accent word.

### Task 2: Devices, voice, and the hero shot

**Files:** Create `intro-video/src/devices/{MacWindow,TerminalPane,EditorPane,Phone}.tsx`, `intro-video/src/scenes/Devices.tsx`, `intro-video/audio/voice.sh`, `intro-video/public/voice.wav`. Modify `intro-video/src/Intro.tsx` to insert `Devices` after `Install` and duck music under the voice.

**Interfaces:** Consumes Task 1 helpers. Produces `DEVICES_FRAMES = sec(13)`, `VOICE_AT`, `VOICE_SECONDS`.

`Devices` timeline (scene-relative): window + phone spring in at 0; phone scan overlay from 0.8s; Connected (both) at 1.9s and the dotted link fades in; at 3s the terminal window crossfades to Notes; voice at 3.5s for `VOICE_SECONDS`; transcript reveals word by word over the voice; listening sheet shows meanwhile; tap Send 0.6s after the voice; a dot travels phone → window; paste 0.4s later with a fading highlight; at 10s "LAN only" pill + crossed-out cloud on the link; scene ends at 13s.

- [ ] **Step 1:** `voice.sh` reads `spoken` from `src/strings.ts`, runs `omni-voice-cli.sh --ref-audio ./yuler.sample.wav` in `$TTS_DIR` (default `~/Explore/explore-tts`), trims leading/trailing silence with ffmpeg `silenceremove`, writes `public/voice.wav`, prints the duration.
- [ ] **Step 2:** `npm run voice`; set `VOICE_SECONDS` in `Devices.tsx` to the printed duration (rounded to 0.1s). Must be ≤ 5.5s so the paste lands before 10s.
- [ ] **Step 3:** Write the device components and `Devices.tsx`; `npm run typecheck`.
- [ ] **Step 4:** Stills at frames 540 (paired), 720 (listening), 950 (poster: text on both), 1090 (LAN only) — inspect layout, no overlap between headline and devices, phone fully on screen.

### Task 3: Wiring, render, docs

**Files:** Modify `mise.toml`, `.gitignore`, `README.md`. Create `.agents/skills/intro/SKILL.md`, `materials/intro-poster.jpg`.

- [ ] **Step 1:** `mise.toml` — `intro:voice`, `intro:studio` (install, audio, studio), `intro:render` (install, audio, render, poster), `dir = "./intro-video"`.
- [ ] **Step 2:** `.gitignore` — `intro-video/out/`, `intro-video/public/audio/`.
- [ ] **Step 3:** `mise run intro:render`; `ffprobe` → duration 25.0s, 1920×1080, 60fps, size < 6MB. Spot-check the poster.
- [ ] **Step 4:** README — under the tagline, a centered poster image linking to `https://github.com/yuler/airvoice/releases/download/intro-video/intro.mp4`.
- [ ] **Step 5:** Skill `.agents/skills/intro/SKILL.md` — commands, layout, publish (`gh release create intro-video intro-video/out/intro.mp4 --title "Intro video" --notes "README intro video asset." --latest=false` the first time, `gh release upload intro-video intro-video/out/intro.mp4 --clobber` after), fragile facts.
- [ ] **Step 6:** Publishing the release and committing (git-commit skill) need the user's go-ahead.
