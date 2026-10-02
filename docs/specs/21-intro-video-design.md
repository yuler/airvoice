# 21-intro-video-design

## Description

A short launch-style intro video for the README and the www hero: speak on the phone, the text lands at the cursor on the computer. It follows the Rails Studio intro strategy (Remotion, code-rendered, synthesized music + SFX, `mise intro:*` tasks) with two differences: the MP4 is committed under `www/public/` so README and www share it, and Airvoice spans a native phone app and a desktop, which headless capture cannot record. Every device UI is rebuilt in React instead.

---

## Decisions

| Topic       | Decision                                                                                          |
| ----------- | ------------------------------------------------------------------------------------------------- |
| Deliverable | One 16:9 master, ~25s, 1920×1080 @ 60fps, H.264, < 6MB                                            |
| Placement   | README poster link + www hero "Watch intro" button that opens the video in a dialog              |
| Copy        | English only; all strings in `intro-video/src/strings.ts` so a zh variant is cheap later          |
| Theme       | Light UI on a soft light gradient; `#006efe` the only accent; tokens from `DESIGN.md`             |
| Type        | Inter (UI) + JetBrains Mono (terminal)                                                            |
| Phone       | React rebuild of the iOS home screen: status pill, text box, "Send to Desktop", listening sheet  |
| Dictation   | Generic listening waveform; no third-party keyboard branding ("a bridge, not a speech engine")    |
| Desktop     | Rebuilt macOS window: `airvoice` terminal with QR for pairing, then a Notes window for the payoff |
| Delivery    | Text appears at the caret as a **paste** (that is what Airvoice does), with a short highlight     |
| Audio       | Synthesized music + click/whoosh/pop/typing SFX, plus one cloned-voice line (local OmniVoice)     |
| Location    | `intro-video/` (Remotion, npm); `mise run intro:voice`, `intro:studio`, `intro:render`            |
| Publishing  | Render to `www/public/intro.mp4` and commit it; README and www both use that one file             |
| Poster      | `www/public/intro-poster.jpg`, frame from the hero shot with text on both devices                 |

## Storyboard

| # | Time | Visual                                                                                                | Copy                                        |
| - | ---- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| 1 | 3s   | Hook lines fade in                                                                                    | "Typing is slow. Talking isn't."            |
| 2 | 3.5s | Dark terminal card types `brew tap …`, `brew install airvoice`, `airvoice`                            | "Install. Run."                             |
| 3 | 3s   | Terminal window (QR) + phone slide in; phone scans; both show green Connected; a link joins them      | "Scan. Paired."                             |
| 4 | 7s   | Window becomes Notes. Voice speaks; phone transcript fills word by word; tap Send; dot travels; paste | "Speak on your phone. It types on your PC." |
| 5 | 3s   | The link gets a "LAN only" pill and a crossed-out cloud                                              | "Local network. No cloud. No account."      |
| 6 | 2.5s | iOS / Android / macOS / Windows / Linux                                                               | "Works everywhere."                         |
| 7 | 3s   | Logo, tagline, `airvoice.yuler.cc`, GitHub URL                                                        | "Airvoice"                                  |

Scenes 3–5 are one persistent `Devices` scene so the phone and window never re-enter.

Spoken line: "Hey team, the release build is ready for review." (cloned from `yuler.sample.wav`, committed as `intro-video/public/voice.wav` because it needs a local model to regenerate). Music ducks under it.

## Out of scope

- Inline autoplay in the www hero, 1:1 / 9:16 social cuts, Chinese variant.
- Real screen capture of iOS, Android, or the Wails desktop app.
