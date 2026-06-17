# Feature Brief — audit-baseline-ratchet

> Phase 8: Continuous Governance. A whole-repo gate scan that records a baseline
> of existing violations and **ratchets**: never lets new violations in, lets
> teams pay down old ones over time. Enables adoption on legacy codebases
> without a big-bang cleanup. Unblocks `precommit-and-pr-gate` and (with
> `deep-codebase-analysis`) `adoption-baseline`.

## Problem

Centinela's mechanical gates (`centinela validate`) are all-or-nothing: a repo
with any pre-existing violation fails until every one is fixed. On a legacy
codebase that means a big-bang cleanup before the team can adopt *any*
enforcement — so they adopt nothing. The industry-standard answer is a
**baseline + ratchet**: snapshot today's violations as tolerated debt, then fail
only on *new* violations introduced after the snapshot, and let the baseline
shrink (never grow) as debt is paid down.

## How it works (mechanism)

1. **Record** — `centinela audit baseline` runs every participating gate in
   **full-repo scan** (bypassing diff-aware mode), collects each gate's
   `Result.Details` violation entries, fingerprints them, and writes a committed
   baseline file (`.workflow/audit-baseline.json`).
2. **Ratchet check** — a new audit run (and/or a gate wired into `validate`)
   re-scans full-repo, fingerprints current violations, and partitions them:
   - **new** (not in baseline) → **fail** (blocking).
   - **baselined** (in baseline, still present) → tolerated (reported, non-blocking).
   - **resolved** (in baseline, now gone) → baseline auto-prunes them on the next
     record/update so the ratchet only tightens.
3. Report: `N new (blocking), M baselined (tolerated), K resolved (prune)`.

## Key decisions to resolve in the plan

- **Violation identity / fingerprint stability.** Gates emit
  `Result.Details []string` (e.g. `"src/a.go (150 lines)"`,
  `"internal/ui → internal/orchestration (forbidden)"`). A naive
  `(gate, rawDetail)` hash is brittle: a file-size detail's line count changes
  whenever the file grows, making an old violation look "new". The plan must
  pick a normalization strategy (e.g. strip volatile numerics / fingerprint on
  the stable identity `(gate, path, rule)`) — per-gate extractor vs a generic
  normalizer. This is the central design call.
- **Command/gate surface.** A new `centinela audit` command group
  (`baseline` to record, a default/`check` to ratchet) vs a gate integrated
  into `centinela validate` vs both. Decide the minimal surface that unblocks
  `precommit-and-pr-gate`.
- **Which gates participate.** All gates that emit `Details`, or a configurable
  `target_gates` subset. Gates without per-violation details (pass/fail only)
  can't be baselined meaningfully.
- **Full-scan enforcement.** The audit scan must always be whole-repo
  regardless of `[validate] diff_mode` (note: import_graph already ignores the
  diff filter; file-size/secrets honor it).

## Acceptance Criteria

1. `centinela audit baseline` records a baseline file capturing every current
   violation across participating gates, scanning the full repo.
2. After a baseline exists, an audit/ratchet run with **no code change** reports
   all violations as baselined and exits 0 (non-blocking).
3. Introducing a **new** violation makes the ratchet run **fail** (non-zero),
   naming the new violation; pre-existing baselined violations stay tolerated.
4. **Fixing** a baselined violation never fails; the next baseline update prunes
   it so it can never be re-tolerated (ratchet only tightens).
5. Fingerprints are **stable** across cosmetic churn that doesn't change a
   violation's identity (e.g. a baselined oversized file growing by more lines
   is still the same tolerated violation, not a new one).
6. Behavior is configurable via `[gates.audit_baseline]` (enabled, severity,
   baseline path, participating gates) and defaults are safe (off or warn until
   a baseline is recorded).
7. Deterministic baseline file (stable ordering) so it diffs cleanly in git.
8. All new source files ≤100 lines; no cross-layer import violations.

## Edge Cases

- No baseline file yet → ratchet check reports "no baseline; run `audit
  baseline`" and does not block (or treats nothing as new), per chosen default.
- Empty repo / zero violations → baseline is empty; everything clean.
- A participating gate is newly enabled after the baseline → its violations are
  unbaselined; decide whether they count as new (likely yes — record again).
- A gate emits no `Details` (pass/fail only) → excluded from baselining.
- Violation detail string format changes between Centinela versions → fingerprint
  scheme must be versioned in the baseline file.
- Baselined file deleted/renamed → its violation resolves (prune).
- Diff-aware mode on → audit still scans full repo.
- Concurrent/duplicate identical details within one gate → deduplicated by
  fingerprint with a count, or kept stable.

## Data Model

New committed artifact `.workflow/audit-baseline.json`: a versioned schema with,
per participating gate, the set of violation fingerprints (and enough
human-readable context to be reviewable in a PR). New `config.AuditBaselineConfig`
under `GatesConfig` following the `RoadmapDriftConfig` pattern
(enabled/severity + baseline path + target gates), normalized in `applyDefaults`
and checked in `validateConfig`.

## Integration Points

- **Read gate results**: reuse `gates.RunWithFilter(cfg, nil)` (nil filter =
  full scan) to collect `Result.Details` across gates.
- **Fingerprint + compare**: new `internal/audit/` package (baseline record,
  load, diff/ratchet) — kept out of `internal/gates` to avoid bloating it and to
  respect layering.
- **Command**: new `cmd/centinela/audit*.go`.
- **Render**: reuse `internal/ui` gate-result style for the new/baselined/resolved
  summary.

## Risks

- **Fingerprint stability** (above) — the make-or-break correctness risk.
- **Layering / import-graph**: `internal/audit` will import `internal/gates` +
  `config`; confirm against the G2 matrix in the plan (gates is `domain`).
- **Scope creep**: structured per-violation findings would be cleaner than
  parsing `Details` strings, but extending every gate's `Result` is a large
  change — the plan should decide v1 (parse Details) vs deferring a structured
  findings refactor.
