# Agent Instructions

Rules for AI agents working in this repository.

---

## Design

When modifying UI or UX — especially theme, colors, and typography — follow [`DESIGN.md`](DESIGN.md), the repo's design system. The same applies when working from design tools (Paper, Figma, etc.): treat `DESIGN.md` as the source of truth and keep outputs aligned with it.

## Planning documents & Superpowers

Implementation plans live only in `docs/plans/`, named with a zero-padded prefix and kebab-case (e.g. `00-mvp-plan.md`). Read existing plans before writing; if a new plan replaces an old one, say so at the top.

With the Superpowers workflow (`brainstorming` → `writing-plans` → `executing-plans`), write design specs to `docs/specs/` (`NN-topic-design.md`, or `YYYY-MM-DD-topic-design.md` for exploratory work), then the plan to `docs/plans/`, then implement task-by-task.

Do not put plans outside `docs/plans/`, or design specs under `docs/plans/` or `docs/superpowers/specs/`.

## Git commit messages

When committing, use the [git-commit](https://github.com/yuler/skills/tree/main/skills/git-commit) skill to generate the message from the staged diff.

## Version & Release

Git tags (`vX.Y.Z`) are the single source of truth for product versions — do not keep a static `VERSION` file or duplicate versions in `package.json`, `wails.json`, Info.plist, or source constants. At build time, all platforms resolve the version via `AIRVOICE_VERSION` env override → highest `vX.Y.Z` Git tag → `0.0.0` fallback (`scripts/resolve-version.sh`, `scripts/lib.sh`, Android Gradle, Wails via `scripts/with-wails-version.sh`, www Vite define in `www/astro.config.mjs`).

iOS `CFBundleVersion` and Android `versionCode` are `major * 1000000 + minor * 1000 + patch` (minimum `1`).

Do not run `xcodegen` alone: `ios/project.yml` needs `APP_VERSION` / `APP_BUILD` from `load_app_version`. Use `./scripts/ios-xcodegen.sh` (or `mise setup` / `mise ios:dev`).

Releases are automated via `.github/workflows/release.yml` after `CI` succeeds on `main`:
- Default bump is `patch`. Include `[minor]` or `[major]` in a commit message / PR title to bump minor or major, or `[skip release]` in the `HEAD` commit subject to skip.
- Manual releases can be triggered via `workflow_dispatch` on `Release` (`bump` / explicit `version` / `ios` mode), or locally with `mise bump` (`patch` | `minor` | `major` | `X.Y.Z`) to create and push a `vX.Y.Z` tag.
