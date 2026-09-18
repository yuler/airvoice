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

## Version

All product version strings come from the [`VERSION`](VERSION) file at the repo root. Inject it at build time (Go `-ldflags`, Android Gradle, Xcode `APP_VERSION` / `APP_BUILD`, Wails via `scripts/with-wails-version.sh`, www `VERSION?raw` import). Do not duplicate the version in `package.json`, `wails.json`, Info.plist, or source constants.

iOS `CFBundleVersion` and Android `versionCode` are `major * 10000 + minor * 100 + patch` from `VERSION`. A second store/TestFlight binary for the same marketing version needs a `VERSION` bump (or a formula change); there is no separate monotonic build counter.

Do not run `xcodegen` alone: `ios/project.yml` needs `APP_VERSION` / `APP_BUILD` from `load_app_version`. Use `./scripts/ios-xcodegen.sh` (or `mise setup` / `mise ios:dev`).

Bump with `mise bump` (`patch` | `minor` | `major` | `X.Y.Z`). That command updates only `VERSION`, then optionally commits and tags `vX.Y.Z`.
