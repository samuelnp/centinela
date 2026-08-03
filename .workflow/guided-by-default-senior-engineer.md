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

---

## Post-verification fixes (adversarial verifier round 1: WARNING, 0 CRITICAL)

The verifier could not refute proof parity — including a hand-built differential
through the tail-default path the acceptance suite never took, and a live
mutation proving the AST guard fails. Three MEDIUM findings were real and are
fixed in-branch. **None of the three fixes touches a proof path**; the AST guard
was re-mutated after the fixes and still fails (exit 1) and passes when restored.

### F2 — `hook autostart` failed OPEN to guided on an unparseable config

`cmd/centinela/hook_autostart.go` discarded `config.Load`'s error, so a nil cfg
reached `ResolveStart`'s guided tail and `NewWithOrder` pinned
`profileContract` **for life** — even for a `centinela.toml` that explicitly said
`enforcement_profile = "strict"`, and fixing the typo afterwards could not undo
it. It now fails CLOSED: it declines to autostart, prints the parse failure, and
writes no state at all — matching `setupRequiresGrading` and every other error
path in that function. `centinela start` already propagated the same error.

### F1 — `EffectiveProfile` bypassed the "declared-but-unmapped driver ⇒ strict" rule

Tier 3 returned only on a capability HIT and otherwise fell through to the
guided tail, while `ResolveStart`, `ProjectDefaultProfile` and
`ProfileProvenance` all returned strict there. Result: `centinela status` printed
`Profile strict (driver: … → no capability, default strict)` while
`centinela verdict` — the machine payload, also served over MCP — reported
`"profile":"guided"` for the identical state file. Tier 3 is now TERMINAL when a
driver is declared. The existing tests could not see this: their tier-3-miss case
used an *unpinned* workflow, which masks the disagreement exactly.

### F3 — the advertised guided cold start dead-ended

In the exact tree AC6 blesses, `start` on a Backlog finding or a draft names
`roadmap promote`, which was unconditionally gated on the grading artifacts
guided no longer produces. The cascade slimming reached `start_guard_cascade.go`
and `hook_setup_cascade.go` but not `promote`. Now `promote` seeds an empty,
role-correct artifact pair when the resolved profile does not require grading
(`roadmap.SeedArtifactsIfAbsent`, never touching an existing file), and its
post-write coverage check is advisory under guided/outcome. Under strict nothing
changed: preflight still refuses a missing artifact before any write.

### Files changed in this round

| File | Lines | Change |
|---|---|---|
| `cmd/centinela/hook_autostart.go` | 70 | F2 — fail closed, name the parse failure |
| `internal/workflow/profile.go` | 50 | F1 — tier 3 terminal on a declared driver |
| `internal/roadmap/promote_seed.go` | **new** 50 | F3 — `SeedArtifactsIfAbsent` |
| `cmd/centinela/roadmap_promote_cascade.go` | **new** 58 | F3 — profile-scoped promote grading |
| `cmd/centinela/roadmap_promote.go` | 96 | F3 — call the two seams above |
| `internal/workflow/profile_agreement_test.go` | **new** 74 | F1 unit — status/verdict agreement table |
| `cmd/centinela/hook_autostart_failclosed_test.go` | **new** 85 | F2 colocated |
| `cmd/centinela/roadmap_promote_cascade_test.go` | **new** 87 | F3 colocated |
| `cmd/centinela/roadmap_promote_advisory_test.go` | **new** 68 | F3 advisory + strict happy path |
| `internal/roadmap/promote_seed_test.go` | **new** 97 | F3 seed unit |
| `tests/acceptance/guided_by_default_cold_walk_test.go` | **new** 93 | F3 — the FULL walk |
| `tests/acceptance/guided_by_default_surface_agreement_test.go` | **new** 92 | F1 — status vs verdict |
| `tests/acceptance/guided_by_default_autostart_test.go` | **new** 98 | F2 — binary-driving |
| `cmd/centinela/roadmap_promote_{errors,final}_test.go` | 60 / 49 | pinned strict — they test the strict post-write check |

Every new/edited function in all of these is at 100% statement coverage. The
three deferral records the verifier filed were removed with
`centinela roadmap remove` and `ROADMAP.md` regenerated, because they are no
longer true.

---

## Post-verification fixes, round 2 (verifier: WARNING, 0 CRITICAL, 37 attacks)

Round 1's three fixes were verified real by direct probe; proof parity held and
the AST guard was still mutation-live. Four more findings, all fixed. **None
touches a proof path**; the guard was re-mutated afterwards (mutated exit 1,
restored exit 0).

### F2 — the SAME fail-open class, one file away: centralized instead of patched

`status_model.go` did `cfg, _ := config.Load()`, so an unreadable centinela.toml
made status print `guided` where start, autostart and verdict all failed closed.
Round 1 fixed one call site; that was whack-a-mole. There is now ONE seam —
`config.LoadForProfile()` — which never returns nil and, on a read failure,
returns a config that pins strict **explicitly** (`RawEnforcementProfile` set, so
it reads as a global pin and outranks the capability tier). Every existing
consumer — `ProfileProvenance`, `EffectiveProfile`, `ProjectDefaultProfile`,
`RenderStatusWithConfig` — is unchanged and simply inherits the direction.

**Surfaces routed through it:** `status_model.go` (status), `mcp_rules.go` (MCP
read_rules), `hook_setup_cascade.go` (setup cascade weight),
`roadmap_promote_cascade.go` (promote's grading requirement), `hook_autostart.go`
(which goes further and declines outright, because it pins a contract for life).

Two guards, not one: a pinned-surface test asserting each of those five uses
`LoadForProfile` and calls no `config.Load`; and a **sweep** over every non-test
file in `cmd/centinela` failing on any discarded-error `config.Load()`, with a
one-entry documented allowlist (`hook_migrate.go`, provider sync, no profile
decision) that is itself checked for honesty. A new fail-open surface now fails
a test instead of waiting for a verifier.

### F1 — MCP read_rules disagreed with every resolver

It reported `cfg.Workflow.EnforcementProfile`, i.e. applyDefaults' normalized
`strict` for any project whose centinela.toml merely omits the knob — which is
exactly what the new scaffold ships — and `""` with no config at all. It now
reports `config.ProjectDefaultProfile(LoadForProfile())`, so read_rules, status,
verdict and workflow_state agree by construction. The other fields stay
best-effort; a broken config reports strict and invents no rule surface.

### F3 — a guided seed could satisfy the STRICT cascade

Seeded artifacts carry the evaluator role (promote must be able to append to
them), which made the role a forgeable claim: one guided promote, then pin
strict, and `start` sailed through a grading rung no evaluator ever ran —
reproduced end to end. Seeded artifacts now carry `"provisional": true`;
`ValidateAnalysis` and `ValidateQuality` refuse a provisional artifact, naming
the file and the exact edit that clears it. `writeArtifact` preserves
non-`features` keys, so the mark survives every append and can only be removed
deliberately. Seeding is NOT removed — round 1's guided cold walk is still
traversable end to end, asserted in the same acceptance file — and clearing the
mark genuinely unblocks strict, so the refusal is a redirect, not a dead end.

### F4 — a failed guided promote left four unrequested files

`prepareGradingArtifacts` ran before `roadmap.Promote` validated the slug and
phase. Seeding moved INSIDE `Promote`, into a `seedThenPreflight` seam placed
after every slug/phase check in both branches, driven by a new
`PromoteRequest.SeedArtifacts` field so `internal/roadmap` stays config-free.
A bad slug or phase now writes nothing at all, matching strict and main.

### Files changed in round 2

| File | Lines | Change |
|---|---|---|
| `internal/config/profile_load.go` | **new** 37 | `LoadForProfile` + strict fallback |
| `cmd/centinela/status_model.go` | 73 | F2 |
| `cmd/centinela/mcp_rules.go` | 48 | F1 |
| `cmd/centinela/hook_setup_cascade.go` | 92 | routed through the shared seam |
| `cmd/centinela/roadmap_promote_cascade.go` | 43 | routed; `prepareGradingArtifacts` removed |
| `cmd/centinela/hook_autostart.go` | 69 | routed; still declines outright |
| `internal/roadmap/provisional.go` | **new** 28 | the mark + its refusal |
| `internal/roadmap/analysis.go` | 72 | refuse provisional; 2 docstrings |
| `internal/roadmap/quality.go` | 91 | refuse provisional |
| `internal/roadmap/promote_seed.go` | 69 | mark seeds; `seedThenPreflight` |
| `internal/roadmap/promote.go` | 100 | `SeedArtifacts` field; F4 ordering |
| `internal/roadmap/promote_inplace.go` | 65 | F4 ordering |
| `cmd/centinela/roadmap_promote.go` | 96 | pass `SeedArtifacts` |
| `internal/config/profile_load_test.go` | **new** 72 | fail-closed, both directions |
| `cmd/centinela/profile_resolver_guard_test.go` | **new** 87 | surface + sweep + allowlist guards |
| `cmd/centinela/mcp_rules_profile_test.go` | **new** 72 | read_rules vs resolver table |
| `cmd/centinela/roadmap_promote_nowrite_test.go` | **new** 38 | F4 |
| `internal/roadmap/provisional_test.go` | **new** 82 | provisional refuse + clear + seam |
| `internal/roadmap/promote_seed_test.go` | 89 | seed pair, idempotence, partial |
| `tests/acceptance/guided_by_default_failclosed_test.go` | **new** 68 | F2 across status/verdict/start |
| `tests/acceptance/guided_by_default_seed_isolation_test.go` | **new** 74 | F3 cross-profile + the clear path |

Every new or edited function is at 100% statement coverage. The four deferral
records the round-2 verifier filed were removed with `centinela roadmap remove`
and `ROADMAP.md` regenerated; the two it marked pre-existing
(`provenance-local-default-ignores-capability-profile-override`,
`deferred-findings-spec-scenario-uncovered`) were deliberately left, as they
reproduce on `main` and are out of this feature's scope.

---

## Post-verification fixes, round 4

Round 3's structural fix was confirmed working (marker unforgeable, all 9
surfaces × 8 config shapes fail closed). Two MEDIUM + one LOW remained, on axes
the config marker cannot reach — they are about *which driver source* and *which
artifact quality*, not about whether the config loaded.

### F1 — the resolvers read only `wf.DriverModel`, ignoring the project declaration

`[orchestration] driver_model` and `$CENTINELA_MODEL` feed every ACTING path
(via `DriverModelFrom` inside `ProjectDefaultProfile`/`ResolveStart`) but neither
reporting resolver consulted them. For any workflow predating the declaration —
which has no pin — `status`, `verdict` and MCP `read_rules` said guided while
`start`, `hook setup`, `roadmap promote` and `doctor` acted strict on the same
tree. Both resolvers now go through `workflow.TierDriverModel`: the workflow pin
wins when present (it is the state-dated decision), otherwise it falls through to
the same `DriverModelFrom` chain the acting paths use.

**The fixture was fixed first, deliberately.** `mcp_rules_profile_test.go`
injected `DriverModel: config.DriverModelFrom("", cfg)`, resolving the driver
*for* the resolver and hiding whether it consults the source itself. Removing the
injection made the test fail on the real bug before any production change; the
new `profile_driver_source_test.go` builds real `centinela.toml` files and lets
the resolver find the driver, and asserts reporting == acting on every shape.

### F2 — the strict cascade's rungs tested existence, not evaluation

`postRoadmapRungs()` checked only that the four grading files exist, so
guided-seeded `"provisional": true` artifacts satisfied `hook setup` on the exact
tree where `start` refused them. Round 3's mark reached
`ValidateAnalysis`/`ValidateQuality` but not this path. Both surfaces now ask one
question, `roadmap.ArtifactIsProvisional`, through `evaluatedArtifact`. An absent
or unreadable file is explicitly NOT provisional — turning "missing" into
"seeded" would make the cascade report the wrong reason.

### F3 — the fallback faked a pin, so provenance lied and a branch was dead

`strictFallbackConfig` set `RawEnforcementProfile = "strict"` to force tier 2.
That worked but attributed the profile to a global setting the operator may never
have written (`Profile strict (global)` on a chmod-000 file) and left the honest
branch unreachable. It now sets an unexported `loadFailed` marker instead,
checked by all three resolvers ahead of every tier below the per-feature pin, so
the direction is unchanged (a declared frontier model still cannot lift an
unreadable config to outcome) and status renders
`Profile strict (unreadable centinela.toml, default strict)`. A pinned workflow
under a merely *unloaded* config keeps a separate, equally honest note,
`default (strict, unresolved config)` — it is not a legacy workflow and must not
be labelled one.

### Files changed in round 4

| File | Lines | Change |
|---|---|---|
| `internal/workflow/profile.go` | 71 | F1 `TierDriverModel`; F3 `LoadFailed` tier |
| `internal/workflow/profile_provenance.go` | 54 | F1 driver source; F3 honest notes |
| `internal/config/profile_load.go` | 65 | F3 `loadFailed` + `LoadFailed()` |
| `internal/config/config.go` | 93 | F3 field; `defaultConfig` moved out for budget |
| `internal/config/defaults.go` | 48 | received `defaultConfig` |
| `internal/config/project_profile.go` | 45 | F3 `LoadFailed` tier |
| `internal/roadmap/provisional.go` | 55 | F2 `ArtifactIsProvisional` |
| `cmd/centinela/hook_setup_cascade.go` | 95 | F2 rungs use the shared predicate |
| `cmd/centinela/hook_setup_artifacts.go` | **new** 14 | F2 `evaluatedArtifact` |
| `internal/workflow/profile_driver_source_test.go` | **new** 95 | F1, real-resolution fixtures |
| `cmd/centinela/hook_setup_provisional_test.go` | **new** 85 | F2 both directions + guided |
| `internal/roadmap/provisional_predicate_test.go` | **new** 46 | F2 predicate incl. absent/malformed |
| `internal/workflow/profile_provenance_notes_test.go` | **new** 57 | F3 attribution |
| `cmd/centinela/mcp_rules_profile_test.go` | 75 | fixture no longer injects the driver |
| `internal/config/profile_load_test.go` | 78 | fallback marks, does not fake a pin |
| `internal/workflow/profile_provenance_test.go` | 78 | split for the 100-line budget |

Every round-4 function is at 100% statement coverage. Five mutations were run and
all five fail the suite: the two config markers, the F1 driver source, the F2
provisional rung and the F3 provenance attribution. No new deferrals; the three
pre-existing ones are untouched.
