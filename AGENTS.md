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
