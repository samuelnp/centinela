### Big-Thinker Report: spec-traceability-gate
**Date:** 2026-06-10

#### Problem
Centinela's spec-first gates require a `.feature` file to exist before code, but
nothing mechanically verifies that each individual `Scenario:` is exercised by
an acceptance test. Scenarios drift into the spec and silently never get
implemented. This is the last "requested but not enforced" hole in the
mechanical-verification phase. The hard part is dogfooding: this gate must run
on Centinela's own repo without (a) being disabled or (b) being gamed with stub
tests — yet a strict whole-repo scan would fail on the large legacy backlog of
uncovered scenarios the day it merges. The crux is what the gate does in CI,
where `ResolveMode` forces a full scan (`CI=true` → `ModeFull` → `filter=nil`).

#### Scope
- **In:** A config-gated, diff-aware built-in gate `[gates.spec_traceability]`
  that maps every in-scope `Scenario:`/`Scenario Outline:` to a covering
  acceptance test via the `// Acceptance: specs/<slug>.feature` header +
  `// Scenario: <name>` comment convention; `severity = fail|warn`; registered
  in `RunWithFilter`; enabled on Centinela in **warn** mode for CI safety.
- **Out:** Per-scenario runtime pass/fail correlation (`go test -json`),
  step-level (Given/When/Then) traceability, Scenario-Outline example expansion,
  a whole-repo ratchet/baseline (that is `audit-baseline-ratchet`'s job),
  auto-generating missing tests, and non-Go acceptance runners.

#### Dependencies & Assumptions
- Reuses `internal/gitdiff.Set` filter contract exactly as G1 does in
  `file_size.go` (include only files in the diff set; nil filter = full scan).
- Lives in `internal/gates` (may import `internal/config`, `internal/gitdiff`
  only) per G2 layer matrix; config leaf in `internal/config`.
- Assumes slug = `.feature` basename without extension on both sides
  (spec path and the `specs/<slug>.feature` header reference).
- **Measured reality** (this worktree, today): 75 spec files, **406** scenarios
  (3 are Scenario Outline). The feature brief's "~208" is stale — the repo grew.
  Under the gate's exact (slug, normalized-name) matching, only **9 scenarios in
  4 spec files** are actually covered; **397 are uncovered**. ~32 acceptance
  files carry the canonical header, but their `// Scenario:` comments are
  paraphrased/grouped and seldom equal the spec's exact scenario text.
- The branch adds `specs/spec-traceability-gate.feature` with **9** scenarios,
  currently **0** covered — qa-senior's acceptance test must cover all 9.

#### Risks
| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| CI full-scan (`CI=true`→`ModeFull`→nil filter) fails on 397 legacy uncovered scenarios | High | Certain if `severity=fail` | Ship default-off; enable on Centinela with **`severity="warn"`** so CI reports gaps but never blocks until a backfill/ratchet lands |
| Convention is loosely followed today (paraphrased comments, `(AC4,AC5)` suffixes, `spec/` typo, `// Acceptance: Scenario:` freeform) → false "uncovered" | Medium | High | Gate defines a strict canonical convention; normalize aggressively (trim, collapse ws, strip one trailing period, case-insensitive); Details name the exact spec+scenario so the fix is obvious; document the convention in the gate message |
| Scenario Outline counted once but has N examples | Low | Low | v1 treats the outline as a single scenario (matches docgen counting); note in edge-cases |
| New gate pushes a file >100 lines (G1) | Medium | Medium | Pre-split config / parse / match / entry as in the plan (~200-250 lines total) |
| Pre-existing specs forced to be edited to pass | Medium | Low | Diff-aware scoping + warn severity means unchanged specs are never gated and never blocked — confirmed: no pre-existing spec needs touching |

#### Rollout
- **Step 1:** Config struct + `NormalizeSpecTraceability` + `validateSpecTraceability` (reject unknown severity); no behavior yet.
- **Step 2:** `parseScenarios` (diff-filtered) + `coveredScenarios` matcher with colocated unit tests in `internal/gates`/`internal/config`.
- **Step 3:** `checkSpecTraceability(cfg, filter)` gate entry honoring the SAME filter as G1; register in `RunWithFilter`; integration test over a temp repo tree (Pass + Fail).
- **Step 4:** Enable on Centinela in `centinela.toml` with `enabled=true`, **`severity="warn"`**; qa-senior writes `tests/acceptance/spec_traceability_gate_test.go` covering all 9 of this feature's own scenarios (closing the honest dogfood) so locally (diff-aware) and in CI (full-scan, warn) the branch is green.
- **Step 5 (later, not this branch):** once `audit-baseline-ratchet` or a backfill exists, ratchet `severity` to `fail`.

#### Handoff
- **Next role:** feature-specialist
- **Outstanding questions / CI-behavior recommendation (explicit):**
  Recommend **Option (b) layered on Option (a)'s mechanism** — the gate honors
  the same diff-aware/full-scan switch as G1/G11, AND on Centinela it ships
  `severity="warn"`. Rationale: in CI, `CI=true` forces `ModeFull` (nil filter),
  so a `severity="fail"` gate would deterministically fail on 397 legacy
  uncovered scenarios — exactly the "turn on a gate, drown in legacy, disable
  it" trap. `warn` makes CI surface every uncovered scenario (visible,
  actionable, ratchet-ready) without blocking merge, which is honest: it neither
  disables the gate nor games it with stubs. Locally (no `CI`), validate is
  diff-aware, so only `spec-traceability-gate.feature` is in scope and its 9
  scenarios must genuinely pass — a bounded real dogfood. Reject Option (a)-strict
  (would require backfilling 397 scenarios now or weakening the documented
  `CI`→full-scan invariant) and Option (c) (always-diff-aware diverges from how
  every other gate honors full-scan and masks CI regressions). The full-scan
  code path must still be implemented correctly (filter==nil walks all specs);
  `warn` is the adoption knob, not a code shortcut.

  **Convention hardening — YES, needed.** Measured data shows the convention is
  real but loosely honored: only 4 slugs match exactly, and live headers include
  `(AC4, AC5)` annotations, a `spec/` (singular) typo, and freeform
  `// Acceptance: Scenario: ...` lines. The gate should DEFINE the canonical
  form and parse defensively: header `^//\s*Acceptance:\s*specs/<slug>\.feature`
  (ignore any trailing annotation after the filename), scenario comment
  `^//\s*Scenario:\s*<name>` with aggressive normalization. This makes the
  loosely-followed convention enforced going forward without retroactively
  breaking unchanged specs.

  **No pre-existing specs need editing** — confirmed. Diff-aware local scope +
  `severity="warn"` in CI means the 397 legacy uncovered scenarios are reported
  but never block, and unchanged `.feature` files are out of the local gate
  scope entirely.
