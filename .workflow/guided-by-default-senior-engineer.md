# guided-by-default — senior-engineer

Slices 1, 2, 3, 5 implemented in full; slice 4's colocated (source-guard) half
implemented, its acceptance-tier half deferred to `qa-senior` (see Handoff).

## Files Touched

### New source (10)
| File | Lines | Purpose |
|---|---|---|
| `internal/workflow/profile_contract.go` | 22 | `ProfileContractGuidedDefault = "guided-default-v1"` + `(*Workflow).UsesGuidedDefault()` |
| `internal/config/project_profile.go` | 31 | `ProjectDefaultProfile(cfg)` for workflow-less surfaces |
| `internal/roadmap/quality_advisory.go` | 32 | `QualityAdvisories()` — low scores as advice, never as a check |
| `cmd/centinela/start_guard_cascade.go` | 52 | `roadmapGradingGuard` — profile-scoped grading rungs |
| `cmd/centinela/hook_setup_cascade.go` | 95 | rung table + `setupRequiresGrading` + advisory emitter |
| `internal/doctor/check_profile_default.go` | 51 | inherited-default advisory (Warn, no Repair) |
| `cmd/centinela/complete_validate_gates_invariant_test.go` | 58 | AST source guard + its own meta-test |
| `internal/workflow/profile_contract_test.go` | 92 | tail/tier/provenance table |
| `internal/config/project_profile_test.go` | 54 | tier table + `RequireRoadmapGrading` |
| `internal/roadmap/quality_advisory_test.go` | 84 | threshold-ignored / range-still-fires / not-rewritten |
| `internal/doctor/check_profile_default_test.go` | 75 | advisory both directions, never fatal, never fixed |
| `cmd/centinela/start_guard_cascade_test.go` | 48 | guided cold start ✅ / strict refusal ❌ |
| `cmd/centinela/start_guard_guided_refusals_test.go` | 72 | AC8 refusal table under guided |
| `cmd/centinela/hook_setup_profile_test.go` | 85 | guided advises, roadmap.json still required, knob table |
| `cmd/centinela/roadmap_promote_range_test.go` | 30 | out-of-range still refuses with zero writes |

### Modified source
`internal/workflow/{profile,profile_provenance,state,order,start_resolve}.go`,
`internal/config/profile_defaults.go`,
`internal/roadmap/{quality,quality_features,promote_scores,types}.go`,
`cmd/centinela/{start,start_guard,start_guard_draft,hook_setup,hook_autostart,roadmap_validate,roadmap_promote}.go`,
`internal/doctor/registry.go`,
`internal/ui/{render_promote,render_roadmap,render_roadmap_analysis,render_roadmap_quality}.go`,
`centinela.toml`, `internal/scaffold/assets/centinela.toml`,
`docs/architecture/workflow-enforcement.md` (+ scaffold mirror, byte-identical),
`docs/guides/{configuration-reference,configuration,getting-started}.md`,
`specs/deferred-findings-roadmap-capture.feature`.

Every file, including every `_test.go`, is ≤100 lines. `state.go` lands exactly
at 100; `start_guard.go` at 99 via the planned `*_cascade.go` extraction.

## Architecture Compliance

**The invariant held.** `cmd/centinela/complete_validate_gates.go` is byte-for-byte
unchanged and now carries an AST-level test asserting it contains no identifier
matching `profile` — with a companion meta-test proving that scan fails on a
source that *does* branch on a profile, so the guard cannot rot into a silent
no-op. No profile condition was added to any gate, test run, freshness check,
claim verifier, gatekeeper check or production-readiness check.

Per-behavior ledger, as shipped:

| Behavior | Process / Proof | Where |
|---|---|---|
| Tail default for **pinned** workflows: strict → guided | PROCESS | `profile.go`, `start_resolve.go` |
| Confirmation cadence → `after_plan` by default | PROCESS (falls out of the flip) | none — already wired |
| Orchestration evidence bundle not required under guided | PROCESS | none — already wired via `order.go` |
| Greenfield grading rungs advisory under guided/outcome | PROCESS | `start_guard_cascade.go`, `hook_setup_cascade.go` |
| Self-graded `overall >= 9` deleted, **all profiles** | theater removal — neither side | `quality.go`, `quality_features.go`, `promote_scores.go` |
| Gatekeeper report + grounded verdict | **PROOF — untouched** | verified by dogfood (b) |
| Gate set, suite, ×2 freshness, claim verification | **PROOF — untouched** | `complete_validate_gates.go` unchanged |
| Production-readiness gate | **PROOF — untouched, config-driven** | not read by any new code |
| Prewrite step-gating | PROCESS, **unchanged** for strict *and* guided | untouched |

Back-compat is **state-dated**: `NewWithOrder` pins `profileContract`, and the
tail returns guided only for a pinned workflow. `EffectiveProfile(nil, cfg)` and
every legacy/in-flight state file still resolve to strict. `DisplayProfile` was
left alone (pinned-only, no config dependency).

Layering respected: `internal/config` gained no dependency; `cmd/centinela`
consumes `config.ProfileDefaults`/`ProjectDefaultProfile`; `internal/doctor`
reuses `workflow.ActiveWorkflows`.

## Type-Safety Notes

- `RequireRoadmapGrading` is a typed `ProfileKnobs` field, not a string compare
  at each call site — the three enforcement points read the knob, never a
  profile literal.
- The hook-setup cascade is a `[]setupRung` value table (predicate + directive +
  panel), so strict's ordering and exact output are structurally preserved
  rather than re-typed per rung.
- `qualityEnvelope.Threshold` still decodes and is ignored, so pre-deletion
  artifacts parse unchanged; `advisoryScore` is deliberately isolated in the
  advisory reporter and can never produce an `error`.
- The AST guard uses `go/parser` + `ast.Ident`, so the target file's own prose
  (which says "profile" repeatedly) cannot trip it — only real code can.

## Trade-Offs

1. **Doctor severity is `Warn`, not a new `INFO`.** The plan says "INFO". The
   doctor `Status` enum has exactly OK / Warn / Error, and Warn is the one
   documented as "an advisory problem that does not fail the command". Adding a
   fourth status would have touched `Counts`, `RenderDoctorSummary` and
   `glyphLine` for a cosmetic distinction. The contract the spec actually asserts
   — advisory, never fails the run, never auto-fixed — is met and tested
   (`ExitError == false`, `Repair == nil`).
2. **`hook setup` fails safe.** An unparseable `centinela.toml` makes the profile
   unknowable; `setupRequiresGrading` returns `true` (full cascade) rather than
   inheriting guided. Fail-safe direction is *more* scaffolding.
3. **A declared-but-unmapped driver model still lands on strict** in both
   `ResolveStart` and `ProjectDefaultProfile`. Naming a model Centinela has no
   class for is read as a request for maximum scaffolding, not as zero-config.
4. **Three roadmap UI panels were rewritten** (`render_roadmap`,
   `render_roadmap_analysis`, `render_roadmap_quality`) plus
   `render_promote`. They instructed agents to "iterate until every feature is
   overall >= 9" — an instruction to inflate a number that now gates nothing.
   Not in the plan's file table; leaving them would have shipped a documented lie.
5. **`specs/deferred-findings-roadmap-capture.feature` scenario updated.** Its
   "overall below 9 is rejected before any write" scenario asserted the deleted
   gate. Rewritten (with a note naming this feature as the supersession) to
   "a low overall score is recorded rather than rejected", and its acceptance
   test now asserts the score is written verbatim. The sibling out-of-range
   scenario is untouched and still refuses with zero writes.
6. **Existing strict-behavior tests were pinned, not deleted.** Every hook-setup
   and start-guard test that asserted the full cascade now writes
   `enforcement_profile = "strict"` (or passes `config.ProfileStrict`), turning
   them into ❌-direction regression guards that strict is byte-identical.

## Verification

| Check | Result |
|---|---|
| `go build ./...` | pass |
| `go vet ./...` | pass |
| `go test ./... -run xxxNONE` | pass (no test-compile breakage) |
| `go test ./...` | pass (whole repo) |
| `./scripts/check-fmt.sh` | exit 0 |
| `centinela validate` | **pass** — "All gates passed" (G1 ✓, cross-compile ✓, roadmap_drift ✓, docstring-gate ✓ 44/44; import_graph and spec-traceability warn on their pre-existing backlogs) |
| Aggregate coverage | 97.3% (gate 95.0%) |
| Per-package coverage | workflow 97.1%, config 98.1%, doctor 96.1%, ui 99.8%, cmd/centinela 96.5%, roadmap 95.0% — **every new or edited function in every one of those files is at 100%**; the package numbers are pre-existing baselines |

Dogfooded with a scratch binary built from this tree:
(a) zero-config project → `guided (default (guided))`, `profileContract` pinned,
no `orchestrationMode`, plan step completes with no evidence bundle;
(b) same project's validate step still refuses a missing gatekeeper report AND an
ungrounded verdict ("no commands-run record — a narrated verdict is not evidence");
(c) a hand-written legacy state file with no `profileContract` → `strict
(default (strict, legacy workflow))`, while a new workflow in the same tree → guided;
(d) `roadmap promote --scores 3,3,3,3,3,3` exits 0 and records `overall: 3`;
`roadmap validate` exits 0 with an advisory line;
(e) `"scores": []` → JSON-array shape fault; missing `scores` → missing-field
fault; `"overall": "nine"` → type fault; `overall: 11` → range fault.

### One timing note the verifier should not misread

Under `go test ./... -coverprofile` the `tests/acceptance` package sits close to
Go's default 10-minute per-package timeout. Two of my `centinela validate` runs
tripped it (605s / 607s) while other work ran concurrently on the machine. I
measured this against a clean `main` worktree rather than assuming: isolated,
`tests/acceptance` is **411.9s on main vs 428.9s on this branch (+17s, ~4%)** —
contention, not a regression this feature introduced. A serial `centinela
validate` on this tree passes end to end. The tier's proximity to the cap is a
standing, pre-existing risk (it is already a recorded lesson), not a new one.

## Handoff

**Next role:** `qa-senior`.

Left deliberately for the tests step (acceptance tier, per the plan's Slice 4/5):

- `tests/acceptance/guided_by_default_parity_test.go` — the {strict, guided}
  parity table over (a) failing validate command, (b) missing gatekeeper report,
  (c) ungrounded verdict, (d) `**Status:** BLOCKING` readiness, plus the clean-tree
  ✅ case so parity is not claimed via everything-failing.
- `tests/acceptance/guided_by_default_parity_helpers_test.go` — shared fixture builder.
- Extension of `tests/acceptance/enforcement_profiles_invariant_test.go` with the
  gatekeeper-report cases.
- `tests/acceptance/guided_by_default_self_governance_test.go` — this repo's
  `centinela.toml` carries an explicit `enforcement_profile` and it is `strict`
  (the pin is already in place; the test that guards it is not).
- The one legitimate divergence to assert explicitly: guided skips the
  orchestration-evidence requirement while strict does not.
- End-to-end guided cold start driving the real binary (the unit-level guard
  exists in `cmd/centinela/start_guard_cascade_test.go`).
- `.workflow/guided-by-default-edge-cases.md`.

### Deferred finding raised

`profile-display-unused-and-now-misleading` (source
`guided-by-default/senior-engineer`): `workflow.DisplayProfile` has no production
caller, and because the guided default is state-dated on `ProfileContract` rather
than on `EnforcementProfile`, it now returns `strict` for every guided-by-default
workflow. Delete it or make it contract-aware before any surface adopts it. The
two `ProfileKnobs` inert-field findings were already deferred by the brief and
were not re-raised.

Manual dogfood transcripts for (a)-(e) are reproducible from the scripts in the
session scratchpad; qa-senior should re-derive them as executable acceptance tests
rather than trusting this report.
