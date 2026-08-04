# Adversarial Verifier Report — guided-by-default (round 6, final)

**Status:** SAFE

Fresh context, refutation stance, verdict re-derived from scratch. I attacked
the newest and least-reviewed change first — the `dasWithoutTimestamps`
normalizer that closed round 5's blocking flake — with a 14-case mutation table
plus an over-broad positive control, then re-swept the feature itself across
proof parity, the unforgeable provenance markers, the provisional rungs, the
driver-source matrix and the spec markers.

The round-5 blocker is gone: the mandated `centinela validate` now exits **0**
with **All gates passed**, and `go test ./...` passes 4400 tests across 47
packages. Nothing I could construct made the completion claim false, and no
attack produced a new finding.

---

## Inputs Read

1. `git diff --stat origin/main...HEAD` (148 files, +6392/−269) and the
   uncommitted tree (the flake fix, `.workflow/roadmap.json`, `ROADMAP.md`)
2. `git diff` of `tests/acceptance/deterministic_artifact_scaffolds_helper_test.go`
   and `..._validate_test.go` — the round-6 delta, read line by line
3. `internal/evidence/{schema.go,schema_marshal.go,schema_init.go,io_write.go,plan_inputs.go}`
   and `internal/orchestration/plan_snapshot.go` — to enumerate which fields of
   the written blob are non-deterministic and which are construction-derived
4. `internal/workflow/{profile.go,profile_provenance.go,profile_contract.go,order.go,validate_orchestration.go,stamp.go}`
5. `internal/config/{profile_load.go,project_profile.go,profile_defaults.go,config.go,driver_model.go}`
6. `internal/roadmap/{provisional.go,promote_seed.go,analysis.go,quality.go}`
7. `cmd/centinela/{start_guard_cascade.go,hook_setup_cascade.go,hook_setup_artifacts.go,roadmap_promote_cascade.go,roadmap_validate.go,artifact_stamp.go,status_model.go}`
8. `internal/doctor/check_profile_default.go`
9. `internal/gates/spec_traceability.go` and its parse/match helpers
10. `cmd/centinela/profile_resolver_{guard,scan,match,allowlist}_test.go` and
    `cmd/centinela/complete_validate_gates_invariant_test.go`
11. `tests/acceptance/enforcement_profiles_invariant_*.go` and
    `tests/acceptance/guided_by_default_*.go`
12. `specs/guided-by-default.feature`, `centinela.toml`,
    `internal/scaffold/assets/centinela.toml`
13. `.workflow/roadmap.json` — the nine deferral records
14. Output of every command in **Commands Run**, all executed by me

---

## Refutation Attempts

### A. The flake fix — priority attack

1. **Does `dasWithoutTimestamps` mask real content drift?** REFUTED. I built a
   14-case mutation table in a byte-identical scratch copy of the tree. Each case
   writes the evidence blob twice, mutating exactly one thing on the second write
   — `outputs`, an extra `inputs` entry, `status`, `handoffTo`, `feature`, `step`,
   `role`, `edgeCases`, `_meta.cli_version`, `_meta` dropped entirely, `coverage`,
   `mobileFirst`, an `Extra` key named `zz_generatedAt`, and an `inputs` entry
   whose text embeds a literal `"generatedAt": "x"` — and asserts the NORMALIZED
   strings still differ. All 14 detected the drift.
2. **Is that harness vacuous?** REFUTED by positive control. I replaced the
   normalizer with an over-broad one (`"[^"]*":\s*"[^"]*"` → `"":""`) and re-ran:
   7 of the 14 cases immediately reported `MASKED`. The harness is live; the
   shipped regex is narrow enough to survive it.
3. **Is the regex scoped correctly — absent fields, multiple occurrences, nested
   objects, over-reach?** REFUTED. `"(generatedAt|written_at)":\s*"[^"]*"` blanks
   both occurrences inside a nested object, leaves `{"x":1}` byte-identical when
   neither field is present, does NOT touch a *value* that equals the field name
   (`{"a":"generatedAt"}`), does not touch `"c":"written_at: z"`, and cannot match
   a JSON-escaped occurrence (`\"generatedAt\":` has a backslash where the pattern
   needs a quote).
4. **Is the diagnosis right — was the flake really the timestamps, and not a
   genuine content nondeterminism the normalizer now hides?** REFUTED twice over.
   Statically: `Skeleton()` derives every field but `GeneratedAt` and
   `_meta.written_at` from `(feature, role, cliVersion)`; `Extra` is empty; and
   `PlanInputs` → `orchestration.RequiredPlanInputs` builds two paths from the
   feature name and sorts them, with no filesystem glob. Empirically: a
   600-iteration probe of the exact two-write sequence measured **raw-differs
   6/600 (1.0%), normalized-differs 0/600**. The normalizer removes exactly the
   nondeterminism and nothing else.
5. **Is the fixed test stable?** CONFIRMED. `go test ./tests/acceptance/ -run
   'TestDAS_' -count=30` → 630/630 pass. The sibling DAS tests
   (`PreFillPassesSnapshotWithZeroAppends`, `PreFillIgnoresLateSiblingBrief`,
   `SkeletonStaysEmpty`) are untouched and green.
6. **Did the fix silently weaken another assertion in that file?** REFUTED. The
   diff is +8/−0 in the helper and +4/−1 in the test; the only replaced line is
   the comparison itself. The duplicate-`inputs` check below it is intact, and
   `TestDAS_PreFillIgnoresLateSiblingBrief` still uses `reflect.DeepEqual` on the
   decoded inputs, untouched by any normalization.
7. **Does anything else depend on the raw-string comparison?** REFUTED —
   `dasWithoutTimestamps` has exactly one call site repo-wide.
8. **Hygiene.** Helper file 82 lines (cap 100), `gofmt -l tests/acceptance/`
   empty, `go vet ./tests/acceptance/` clean.

### B. Proof parity — guided vs strict

9. **Can a guided tree complete on weaker proof?** REFUTED at five cells, each
   with a live acceptance test I read: missing gatekeeper report blocks both
   (`TestEP_MissingGatekeeperReportBlocksEveryProfile`), ungrounded verdict blocks
   both (`TestEP_UngroundedVerdictBlocksEveryProfile`), revision skew blocks both
   (`TestEP_StaleVerificationBlocksEveryProfile`), BLOCKING production readiness
   blocks both (`TestEP_BlockingProductionReadinessBlocksEveryProfile`), and a
   failing `validate.commands` entry blocks all three profiles
   (`TestEP_GatesRunUnderEveryProfile`, `TestGBD_GateFailureBlocksBothProfiles`).
10. **Is parity achieved by everything failing?** REFUTED by the counterweight:
    `TestEP_CleanTreeCompletesUnderEveryProfile` drives a fully-satisfied tree to
    exit 0 under both profiles.
11. **Is there a profile branch in a proof path?** REFUTED. Grep over
    `internal/{gates,verify,acceptance,docstring,importgraph}` finds zero profile
    references; `internal/verdict` mentions the profile only as a reported field.
    The one divergence is `RequireSubagentEvidence` → `OrchestrationMode`, whose
    single consumer is `strictOrchestrationEnabled` in
    `internal/workflow/validate_orchestration.go` — who authored what, not what is
    true of the tree — and it is asserted in both directions by
    `TestEP_GuidedSkipsOrchestrationEvidenceStrictRequires`.
12. **Is the AST guard on the validate-step gate runner mutation-live?**
    CONFIRMED. I appended `func zzProfileProbe() string { return "guided" }` to
    `complete_validate_gates.go` in the scratch copy;
    `TestValidateGatesHasNoProfileIdentifier` failed. It also ships its own
    positive control (`...GuardCatchesAProfileBranch`).

### C. Fail-closed resolution

13. **Can a surface reach guided through an unreadable config?** REFUTED at the
    decision, not merely at the call sites. `loadFailed` and `resolvedByLoad` are
    unexported and set only by `strictFallbackConfig` and `markResolved`, and
    `markResolved` is called from exactly the two success returns of
    `config.Load`. No package outside `internal/config` can forge either marker.
    All three resolvers — `EffectiveProfile`, `ProjectDefaultProfile`,
    `ProfileProvenance` — check `LoadFailed()` above the capability and tail tiers
    and gate the tail on `ResolvedByLoad()`.
14. **Is the fail-open guard mutation-live for a NEW surface?** CONFIRMED, both
    idioms and both guarded trees. I (a) swapped `LoadForProfile` for `Load` in
    `status_model.go` → both `TestProfileSurfacesUseTheSharedResolver` and
    `TestNoSwallowedConfigLoadOutsideTheAllowlist` failed; (b) added a brand-new
    `cmd/centinela/zz_newsurface.go` using the rebind-to-empty idiom and a new
    `internal/doctor/zz_newcheck.go` using the discard idiom → the sweep named
    both. All mutants removed; every restore confirmed with `cmp`.
15. **Is an absent `centinela.toml` conflated with an unreadable one?** REFUTED.
    `Load` returns `markResolved(defaultConfig())` only on `os.IsNotExist`; any
    other read error and any parse or validation error returns the error, which
    `LoadForProfile` converts into the `loadFailed` stand-in.

### D. Provisional rungs and the cross-profile hole

16. **Can guided-seeded grading artifacts satisfy strict?** REFUTED at all three
    surfaces. `roadmap promote` seeds with `"provisional": true`;
    `ValidateAnalysis`/`ValidateQuality` refuse it (used by `start` and by
    `roadmap validate`), and `hook setup`'s rungs ask the same question through
    `evaluatedArtifact`. `TestGBD_GuidedSeedCannotSatisfyTheStrictCascade` walks
    the exact promote → pin-strict → `start` sequence and requires the refusal to
    name both "provisional" and `roadmap-analysis.json`.
17. **Is that refusal a dead end?** REFUTED by
    `TestGBD_ClearingProvisionalUnblocksStrict`.
18. **Does seeding poison a genuine evaluator artifact?** REFUTED.
    `SeedArtifactsIfAbsent` skips any path that already exists — never read,
    rewritten or touched — and `writeArtifact` preserves every non-`features`
    key, so an append neither adds nor clears the mark.
19. **Does the guided cold walk still traverse?** CONFIRMED by
    `TestGBD_GuidedColdStartBacklogWalkIsTraversable` and its draft variant: the
    refusal names a command, that command runs on that very tree, and the feature
    then starts on `plan`. The ❌ direction
    (`TestGBD_StrictColdWalkStillRefusesPromote`) asserts a refused promote leaves
    `roadmap.json` byte-identical.

### E. Surface agreement

20. **Can any two surfaces disagree about the same tree?** REFUTED across the
    driver-source matrix. `TierDriverModel` falls through to
    `config.DriverModelFrom`, whose precedence is flag > `$CENTINELA_MODEL` >
    `[orchestration] driver_model` > local block.
    `TestDeclaredDriverAppliesToWorkflowsWithoutThePin` checks five driver classes
    against `EffectiveProfile`, `ProfileProvenance` AND `ProjectDefaultProfile` at
    once; `TestEnvDriverModelReachesTheResolvers` and
    `TestWorkflowPinOutranksTheProjectDeclaration` close the env and pin sources.
    At the binary level, `TestGBD_StatusAndVerdictAgreeOnUnmappedDriver`,
    `...OnCapableDriver`, `TestGBD_StatusFailsClosedOnUnreadableConfig` and
    `TestGBD_VerdictAndStartAgreeOnUnreadableConfig` compare real CLI output.

### F. Spec honesty and self-governance

21. **Are the spec markers honest?** CONFIRMED — and recomputed rather than taken
    from the gate's summary line. Driving `parseScenarios` / `coveredScenarios` /
    `uncovered` directly over the tree: the two changed specs contribute 55
    scenarios (`guided-by-default` 23, `deferred-findings-roadmap-capture` 32)
    with exactly **1** uncovered — `"Defer auto-resolves --source from worktree
    CWD when flag is omitted"`, which is the already-recorded deferral
    `deferred-findings-spec-scenario-uncovered`. `specs/guided-by-default.feature`
    is 23/23 covered, and every `// Scenario:` marker is single-line and unwrapped
    (the matcher found them all).
    *Correction to the round-5 hand-off:* it described this as "30/30"; the true
    count for this spec is 23 scenarios, all covered. The substance holds, the
    number did not.
22. **Does this repo's own strict pin actually take effect?** REFUTED as a
    finding. The added `enforcement_profile = "strict"` sits on line 7, under the
    `[workflow]` header on line 1, and
    `TestGBD_RepoPinsEnforcementProfileExplicitly` loads the real `centinela.toml`
    and asserts `RawEnforcementProfile == "strict"`.

### G. Residual, checked and dismissed

23. **`import_graph` and `spec-traceability-gate` both report ⚠ with empty detail
    lines.** Both are `warn`-severity by configuration and neither is caused by
    this branch: `import_graph` warns that Go packages match no configured layer
    (repo-wide, pre-existing), and the traceability warning is the single deferred
    scenario above. `centinela validate` still exits 0.
24. **A hand-edited `enforcementProfile: "guided"` in `.workflow/<feature>.json`
    outranks the unreadable-config refusal.** Deliberate and documented — the pin
    is the state-dated decision for that workflow — and it can only relax PROCESS,
    which the parity table in §B bounds. Anyone able to edit that file can already
    edit `currentStep`. Not a finding.

---

## Commands Run

Every command below was executed by me in this session. The two mandated runs
executed against the real worktree; every mutation experiment ran in a
byte-identical `rsync` copy under the scratchpad, so the verified tree was never
modified, and each restore was confirmed with `cmp`.

| # | Command | Exit |
|---|---------|------|
| 1 | `go build -o /tmp/centinela-verify-gbd6 ./cmd/centinela` | 0 |
| 2 | `/tmp/centinela-verify-gbd6 validate` (mandated, run exactly once) | **0 — All gates passed** |
| 3 | `go test ./...` (mandated) | **0 — 4400 passed, 47 packages** |
| 4 | `go test ./tests/acceptance/ -run TestMUT_ -v` — 14-case masking table + normalizer purity (scratch copy) | 0 — 17/17 |
| 5 | `go test ./tests/acceptance/ -run TestMUT_RealDriftStillDetected` with an over-broad normalizer (positive control) | 1 — 7 MASKED, as designed |
| 6 | `go test ./tests/acceptance/ -run TestPROBE2_ -v` — 600-iteration flake-rate probe | 0 — raw 6/600, normalized 0/600 |
| 7 | `go test ./tests/acceptance/ -run 'TestDAS_' -count=30` | 0 — 630/630 |
| 8 | `gofmt -l tests/acceptance/` and `go vet ./tests/acceptance/` | 0 — clean |
| 9 | `go test ./cmd/centinela/ -run 'TestProfileSurfacesUseTheSharedResolver\|TestNoSwallowedConfigLoadOutsideTheAllowlist'` with `config.Load()` swapped into `status_model.go` | 1 — guard fired |
| 10 | the same guard with a new `zz_newsurface.go` (rebind idiom) and `internal/doctor/zz_newcheck.go` (discard idiom) | 1 — both named |
| 11 | `go test ./cmd/centinela/ -run TestValidateGatesHasNoProfileIdentifier` with a synthetic profile identifier appended | 1 — guard fired |
| 12 | `go test ./internal/gates/ -run TestZZProbeTraceGaps -v` — recompute traceability over the changed specs | 0 — 55 scenarios, 1 known gap |
| 13 | `/tmp/centinela-verify-gbd6 roadmap remove das-prefill-idempotency-timestamp-flake` | 0 |
| 14 | `/tmp/centinela-verify-gbd6 roadmap generate` | 0 |
| 15 | `/tmp/centinela-verify-gbd6 artifact stamp guided-by-default` | 0 |

---

## Findings

**None.** No new finding at any severity. Every attack in §A–§G was refuted, and
each refutation is backed by a command above rather than by reading alone.

---

## Deferred Findings

All nine records from rounds 1–5 are present in `.workflow/roadmap.json`, and I
re-litigated none of them:

| Slug | State |
|------|-------|
| `provenance-local-default-ignores-capability-profile-override` | present, LOW, deferred |
| `deferred-findings-spec-scenario-uncovered` | present — independently re-confirmed as the single traceability gap |
| `profile-display-unused-and-now-misleading` | present, LOW, deferred |
| `harness-model-alias-unverified` | present, deferred |
| `profile-plan-advisor-knob-inert` | present, deferred |
| `profile-step-gating-knob-inert` | present, deferred |
| `resolvestart-lacks-failclosed-value-guard` | present, deferred — not live: `start.go` propagates the load error |
| `setup-rung-content-fault-diverges-from-start` | present, deferred — fails in the safe direction |
| `das-prefill-idempotency-timestamp-flake` | **REMOVED — verified fixed in-branch** |

The removal is earned, not assumed. That record's own recommended remedy was
"inject a clock or exclude the timestamps from the comparison"; the branch took
the second option, and §A.1–A.6 prove the exclusion is exact (600-iteration
measurement, 6/600 → 0/600), non-masking (14/14 mutations still caught, with a
positive control proving the table is live) and stable (30 consecutive runs).

---

## Recommendation

**SHIP.**

Round 5 withheld SAFE for exactly one reason: the mandated `centinela validate`
exited 1 on a pre-existing timestamp flake. That reason is gone — validate now
exits 0 with all gates passed — and the fix that got it there is the narrowest
one available: it blanks the two wall-clock fields and provably nothing else.
The feature itself survived a second, independent sweep with no new finding, both
source-level guards are mutation-live, and the eight remaining deferrals are all
LOW-severity and unreachable on the shipped paths.

```json centinela:verification
{
  "revision": "3d40ba8f7b821a281e2803803fa9ba68e47f8cdf",
  "treeDigest": "sha256:f13fb25181fa17ecd2bc69c736b6c4aca01956fbb10278c6012d9bacdf372012",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-verify-gbd6", "./cmd/centinela"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-gbd6", "validate"], "exitCode": 0},
    {"argv": ["go", "test", "./..."], "exitCode": 0},
    {"argv": ["go", "test", "./tests/acceptance/", "-run", "TestMUT_", "-v"], "exitCode": 0},
    {"argv": ["go", "test", "./tests/acceptance/", "-run", "TestMUT_RealDriftStillDetected"], "exitCode": 1},
    {"argv": ["go", "test", "./tests/acceptance/", "-run", "TestPROBE2_", "-v"], "exitCode": 0},
    {"argv": ["go", "test", "./tests/acceptance/", "-run", "TestDAS_", "-count=30"], "exitCode": 0},
    {"argv": ["gofmt", "-l", "tests/acceptance/"], "exitCode": 0},
    {"argv": ["go", "vet", "./tests/acceptance/"], "exitCode": 0},
    {"argv": ["go", "test", "./cmd/centinela/", "-run", "TestProfileSurfacesUseTheSharedResolver|TestNoSwallowedConfigLoadOutsideTheAllowlist"], "exitCode": 1},
    {"argv": ["go", "test", "./cmd/centinela/", "-run", "TestValidateGatesHasNoProfileIdentifier"], "exitCode": 1},
    {"argv": ["go", "test", "./internal/gates/", "-run", "TestZZProbeTraceGaps", "-v"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-gbd6", "roadmap", "remove", "das-prefill-idempotency-timestamp-flake"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-gbd6", "roadmap", "generate"], "exitCode": 0}
  ]
}
```
