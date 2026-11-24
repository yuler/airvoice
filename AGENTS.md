# AGENTS.md

Guidelines for coding agents working in this repository.

## Rules

### Git & PR Workflow

- **PR-first development**: All development must go through Pull Requests (PRs). Work on dedicated feature/bugfix branches; do not push directly to `main`.
- **Worktree workflow**: When writing code or experimenting with features, always create a new branch in a `git worktree` under `.worktree/` (gitignored) at the repo root. Once the work is done and the PR is opened, remove the worktree immediately.
- **Commit & PR title format**: Use the [git-commit](https://github.com/yuler/skills/tree/main/skills/git-commit) skill to generate commit messages from the staged diff, formatted as `emoji [scope] The main change` (e.g. `✨ [cli] Persist pairing token across restarts`). PR titles follow the exact same format.
- **No agent trailers**: Do not add agent trailers to commits (`Made-with:`, `Co-Authored-By: Claude`, Cursor, etc.). Include a message body only when it adds real context.
- **Auto-PR on completion**: Once a feature or bugfix is implemented and verified (checks/tests pass), commit, push the branch, create a PR, and remove the worktree automatically unless the user explicitly says not to.

### Version & Release

- **Single source of truth (Git tags)**: `vX.Y.Z` Git tags are the sole version source — do not keep a static `VERSION` file or duplicate versions in `package.json`, `wails.json`, `Info.plist`, or source constants. At build time, all platforms resolve the version via `AIRVOICE_VERSION` env override → highest `vX.Y.Z` Git tag → `0.0.0` fallback (`scripts/resolve-version.sh`, `scripts/lib.sh`, Android Gradle, Wails via `scripts/with-wails-version.sh`, www Vite define in `www/astro.config.mjs`).
- **Mobile build numbers**: iOS `CFBundleVersion` and Android `versionCode` are `major * 1000000 + minor * 1000 + patch` (minimum `1`).
- **XcodeGen**: Never run `xcodegen` directly — `ios/project.yml` requires `APP_VERSION` / `APP_BUILD` from `load_app_version`. Always use `./scripts/ios-xcodegen.sh` (or `mise setup` / `mise ios:dev`).
- **Automated trunk release**: Merging to `main` triggers `.github/workflows/release.yml` after `CI` succeeds to compute the next semver, tag `vX.Y.Z`, build/publish artifacts, and open the GitHub Release:
  - Default bump is `patch`. Include `[minor]` or `[major]` in a commit message or PR title to bump minor/major, or `[skip release]` in the `HEAD` commit subject to skip.
  - Manual releases can be triggered via `workflow_dispatch` on `Release` (`bump` / explicit `version` / `ios` mode), or locally with `mise bump` (`patch` | `minor` | `major` | `X.Y.Z`) to create and push a `vX.Y.Z` tag.

### Code, Design & Architecture

- **Simplicity first**: Keep everything from design to code as simple as possible. Avoid speculative over-engineering.
- **File size limit**: Keep each agent-authored file under 5,000 tokens (bytes ÷ 4). Lockfiles and generated code sit outside this limit. Split a file that would cross it by responsibility: one module per concern, helpers and types in a sibling file, and a one-line pointer in the original naming the new path.
- **Design system**: When modifying UI or UX — especially theme, colors, and typography — follow [`DESIGN.md`](DESIGN.md) as the source of truth, including outputs from design tools (Paper, Figma, etc.).

### Planning Documents & Specs

- **Plans**: Implementation plans live only in `docs/plans/`, named with a zero-padded prefix and kebab-case (e.g. `00-mvp-plan.md`). Read existing plans before writing; if a new plan replaces an old one, say so at the top.
- **Design specs**: Write design specs to `docs/specs/` (`NN-topic-design.md`, or `YYYY-MM-DD-topic-design.md` for exploratory work). When explicitly using the Superpowers workflow (`brainstorming` → `writing-plans` → `executing-plans`), write the design spec to `docs/specs/`, then the plan to `docs/plans/`, then implement task-by-task.
- **Locations**: Do not put plans outside `docs/plans/`, or design specs under `docs/plans/` or `docs/superpowers/specs/`.

### Agent Behavior & Formatting

- **Clarify before assuming**: If something is unclear, ask questions.
- **Restricted skills**: Do not use superpower or other speculative-driven skills unless explicitly declared.
- **Table alignment**: Markdown tables must be auto-aligned (pad columns so pipes line up).
