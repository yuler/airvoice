# Agent Instructions

Rules for AI agents working in this repository.

---

## Planning documents

**All implementation plans must live under `docs/plans/`.**

| Rule | Detail |
|------|--------|
| Location | `docs/plans/` only — not repo root, not `docs/` root |
| Naming | Zero-padded prefix + kebab-case, e.g. `00-mvp-plan.md`, `01-ios-polish-plan.md` |
| Before writing | Read existing plans in `docs/plans/` to avoid duplication or contradiction |
| Superseding | If a new plan replaces an old one, say so at the top of the new file |

**Examples**

- `docs/plans/00-mvp-plan.md` — oneshot MVP build spec
- `docs/plans/01-…` — follow-up milestones

**Do not** create implementation plans outside `docs/plans/`.

---

## Superpowers workflow

When using the **Superpowers** skill chain (`brainstorming` → `writing-plans` → `executing-plans`), this repo's paths override the skill defaults.

| Artifact | Location | Naming |
|----------|----------|--------|
| Design spec | `docs/specs/` | Zero-padded prefix + kebab-case + `-design` suffix, matching the paired plan number when one exists (e.g. `06-android-client-design.md`). For exploratory work without a plan yet, use `YYYY-MM-DD-<topic>-design.md`. |
| Implementation plan | `docs/plans/` | Same rules as above |

**Workflow**

1. **brainstorming** — explore requirements, present design, get approval, then write the design doc to `docs/specs/`.
2. **writing-plans** — read the approved spec from `docs/specs/`, then write the implementation plan to `docs/plans/` (not `docs/superpowers/plans/`).
3. **executing-plans** or **subagent-driven-development** — implement the plan task-by-task.

**Do not** write design specs under `docs/plans/` or `docs/superpowers/specs/`.
