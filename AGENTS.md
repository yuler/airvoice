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
