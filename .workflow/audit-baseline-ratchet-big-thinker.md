# audit-baseline-ratchet — big-thinker

## Problem

Centinela's mechanical gates are all-or-nothing: a legacy repo with any
pre-existing violation fails `validate` until every one is fixed, forcing a
big-bang cleanup before any enforcement can be adopted. A baseline+ratchet
snapshots today's violations as tolerated debt and fails only on NEW ones.

## Scope

New `internal/audit` package + `centinela audit` command group: record a
committed baseline (`.workflow/audit-baseline.json`) of current violations
(full-repo scan), then ratchet — new→fail, baselined→tolerate, resolved→prune.
Optional `[gates.audit_baseline]` gate folded into `validate`. Reuses
`gates.RunWithFilter(cfg, nil)`; `internal/gates` is untouched.

## Dependencies & Assumptions

- Builds on the existing gate suite. Gates expose violations only as
  `Result.Details []string`, so the baseline parses those strings (no
  structured-`Finding` refactor in v1 — deferred).
- Unblocks `precommit-and-pr-gate` (needs a fast machine-readable ratchet
  verdict → `centinela audit --json`).

## Risks

- **Fingerprint stability (central).** Resolved by a per-gate identity extractor
  reducing each Detail to a stable key (file-size→path, import-graph→edge,
  spec→feature+scenario, secrets→path+rule) + a generic fallback (strip trailing
  `(…)`/digits). Hash = `sha256(scheme+gate+key)`, versioned in the file, so an
  oversized file growing by lines keeps its identity (AC-5). Verified the
  file-size detail format is `path (N lines)` (incl. "justified" variants) — the
  "text before ` (`" extractor is stable across all three.
- **Import-graph (G2).** `internal/audit` joins the `aggregator` layer (allows
  domain+leaf — verified in centinela.toml); its edges audit→`gates` (domain) +
  audit→`config` (leaf) add no failing edge. **No cycle:** `gates` must not
  import `audit`, so the gate is wired from `cmd/centinela/validate.go`, not
  inside `gates.RunWithFilter`. One toml line adds `internal/audit/**`.
- **Scaffold/toml parity + dogfooding** (known traps): mirror the toml
  import-graph change into `internal/scaffold/assets`; build a `/tmp` binary to
  dry-run `audit`/`audit baseline` before the validate gate.

## Rollout

Safe-adoption defaults: severity `warn`, missing baseline → non-blocking Skip.
Existing repos see no behavior change until they run `centinela audit baseline`.

## Handoff

→ feature-specialist. Plan authored at `docs/plans/audit-baseline-ratchet.md`
with per-file budgets (all ≤95 lines), the baseline JSON schema, and the
cycle-avoidance seam. v1 parses `Details`; structured-findings refactor is
explicitly out of scope.
