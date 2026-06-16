# audit-baseline-ratchet — documentation-specialist

Internal-surface (right-sized) docs step.

## KB Pages

No standalone KB page — this adds a new `centinela audit` command surface
documented via the feature brief, plan, and regenerated project docs.

## project-docs Entries

- `.workflow/audit-baseline-ratchet-changelog.md` — one-line `feat` changelog.
- Regenerated `docs/project-docs/index.html` (picks up the brief, plan, and
  changelog for the feature).

## User-facing note

New `centinela audit` command group for adopting Centinela's mechanical gates on
legacy codebases without a big-bang cleanup:

- `centinela audit baseline` — full-scans the participating gates and records the
  current violations as a committed, deterministic baseline at
  `.workflow/audit-baseline.json`.
- `centinela audit` — re-scans and **ratchets**: NEW violations block (exit 1);
  pre-existing baselined violations are tolerated; resolved ones are reported and
  pruned on the next `audit baseline` (the ratchet only tightens). `--json`
  emits the verdict for tooling.

Fingerprints are stable across cosmetic churn (e.g. a baselined oversized file
growing by lines stays the same tolerated violation). Configurable via
`[gates.audit_baseline]` (`enabled`, `severity`, `baseline_path`,
`target_gates`); defaults are safe (off/warn). When enabled it also runs as a
gate inside `centinela validate`.

## Outcome

Docs generated and validated. Handoff → complete.
