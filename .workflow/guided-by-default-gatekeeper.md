### Adversarial Verifier Report: guided-by-default (round 4, fresh context)

**Status:** WARNING

I re-derived the verdict from scratch, from the tree and from commands I ran myself.

**The central claim holds.** Profiles scale process, never proof. `centinela validate`
passes, the full suite passes, the AST guard on `complete_validate_gates.go` fails when
mutated and passes when restored, the proof packages carry no profile branch, all 30
scenarios of `specs/guided-by-default.feature` are covered under the real matcher
(including the four markers that were line-wrapped in round 3), and the provisional mark
still keeps a guided-seeded artifact from satisfying the strict `start` rung.

**The round-4 structural fix works, for the vector it targets.** `config.resolvedByLoad`
is unexported, written only by `Load`'s two success returns, unreachable by TOML/JSON
decode, preserved across copies, and mutation-proven load-bearing in *both* resolver
packages. Across 9 profile-reporting/acting surfaces × 8 config shapes I could not make
any surface answer guided on an unreadable, unparseable, permission-denied or
invalid-value `centinela.toml`. Round 3's F1 (`hook context` cadence), F2 (doctor), F3
(guard blind to the rebind idiom, and blind to `internal/doctor`) and F4 (MCP `read_rules`
scope) are all closed and re-verified by mutation. **The unreadable-config fail-open class
is closed, and closed structurally rather than per surface.**

**What is not closed is the wider class the fix was named for: surfaces disagreeing about
the same tree.** Two new instances survive, on input axes the provenance marker does not
reach. Neither touches proof, so WARNING, not CRITICAL.

#### Inputs Read
- `git diff origin/main...HEAD` (103 files) + the uncommitted round-4 tree (32 files), at `1f5cc0d`
- `docs/features/guided-by-default.md`
- `specs/guided-by-default.feature` (30 scenarios: 23 `Scenario:` + 7 `Scenario Outline:`)
- `docs/plans/guided-by-default.md`
- `internal/config/`: `config.go`, `defaults.go`, `profile_load.go`, `project_profile.go`, `profile_defaults.go`, `driver_model.go`, `profile_load_test.go`, `project_profile_test.go`
- `internal/workflow/`: `profile.go`, `profile_provenance.go`, `profile_contract.go`, `start_resolve.go`, `order.go`
- `cmd/centinela/`: `status_model.go`, `mcp_rules.go`, `hook_context.go`, `hook_context_review_mode.go`, `hook_prewrite.go`, `hook_plan_advisor.go`, `hook_autostart.go`, `hook_statusline_view.go`, `hook_setup_cascade.go`, `roadmap_promote_cascade.go`, `start.go`, `start_guard.go`, `start_guard_cascade.go`, `verdict.go`, `complete.go`, `mcp.go`
- `cmd/centinela/profile_resolver_{guard,scan,match,allowlist}_test.go`, `complete_validate_gates_invariant_test.go`, `mcp_rules_{profile,scope}_test.go`
- `internal/doctor/context.go`, `internal/doctor/check_profile_default.go`
- `internal/roadmap/`: `provisional.go`, `promote_seed.go`, `promote.go`, `promote_inplace.go`, `analysis.go`, `quality.go`
- `internal/hookpolicy/prewrite.go`, `internal/verdict/assemble.go`, `internal/gates/spec_traceability_match.go`, `internal/ui/render_status.go`
- `tests/acceptance/guided_by_default_*.go`, `enforcement_profiles_invariant_*.go`
- `centinela.toml`, `.workflow/roadmap.json`
- The prior round's report was read ONLY to re-attack its claims. No narrative was accepted as evidence.

#### Refutation Attempts

1. **Built the full surface × config-shape matrix myself** (below), driving the real binary
   in scratch projects. Every surface fails closed on every unreadable-config shape.
   NOT REFUTED.
2. **Attacked the provenance marker directly.** `resolvedByLoad` is unexported; `markResolved`
   is called from exactly two places, both `Load` success returns; the field is absent from
   every struct literal outside `defaultConfig`/`strictFallbackConfig`; TOML and JSON decode
   cannot set it (`RawEnforcementProfile` is `toml:"-"`, `resolvedByLoad` is unexported);
   a struct copy preserves it, which is the correct direction; a zero value reads false.
   No surface bypasses it — no non-resolver code anywhere reads
   `cfg.Workflow.EnforcementProfile`. NOT REFUTED.
3. **Mutation-tested the marker's load-bearingness.** Deleting `&& cfg.ResolvedByLoad()` from
   `workflow.EffectiveProfile` → tests fail (exit 1). Deleting the `!cfg.ResolvedByLoad()`
   guard from `config.ProjectDefaultProfile` → tests fail (exit 1). Both restored, green.
   NOT REFUTED.
4. **Mutation-tested the swallow guard on BOTH idioms and BOTH trees.** Injecting
   `cfg, _ := config.Load()` → exit 1. Injecting
   `cfg, err := config.Load(); if err != nil { cfg = &config.Config{} }` (round 3's blind
   spot) → exit 1. Injecting the rebind idiom into `internal/doctor/context.go` (outside the
   old glob) → exit 1. All restored, green. NOT REFUTED.
5. **Swept every package for a profile resolver the guard cannot see.** Repo-wide,
   `config.Load` is called only in `cmd/centinela/*` and `internal/doctor/context.go`; the
   four resolvers live in `internal/config` and `internal/workflow` and are the only readers
   of the profile fields. The guard's own docstring admits it is defense in depth; that is
   accurate and the real defense holds. NOT REFUTED.
6. **Re-attacked round 3's F1 (hook context cadence).** Post-flip workflow at the `code` step:
   REVIEW panel absent on a readable guided config (correct), present on valid-strict,
   unparseable-pinning-strict, chmod-000 and invalid-value. Fails closed. CLOSED.
7. **Re-attacked round 3's F2 (doctor).** An unreadable config now reports "enforcement
   profile is unknown — centinela.toml could not be read" instead of claiming guided. CLOSED
   — but see F1 below for the case where doctor stays silent when it should speak.
8. **Re-attacked round 3's F4 (MCP scope).** Drove `mcp serve` over real stdio JSON-RPC across
   12 tree/config combinations. `read_rules` now equals `status` and `verdict` in every one,
   including the legacy pre-flip workflow that was the round-3 disagreement. `workflow_state`
   carries no profile field and its `mcpVerdict` returns the load error rather than guessing.
   CLOSED.
9. **Re-attacked proof parity.** `centinela validate` exit 0; `go test ./...` exit 0. The AST
   guard fires on the real file when a `zzProfileProbe` identifier is inserted (exit 1) and
   passes restored (exit 0); its synthetic positive control is present. `internal/gates`,
   `internal/verify` and `internal/verdict` contain exactly one profile mention, the reporting
   field `Packet.Profile`, which nothing branches on. `precommit`, `pr-gate`, `audit`, `route`
   and `validate` contain no profile identifier at all. NOT REFUTED.
10. **Re-attacked the provisional-mark hole.** A guided promote seeds; a second promote appends
    and the mark survives; `roadmap generate` does not launder it; with strict pinned, `start`
    and `roadmap validate` both refuse naming the file and the exact edit; a failed promote
    (unknown slug) leaves `.workflow/` containing only `roadmap.json`; the guided cold walk
    still traverses. NOT REFUTED at `start` — **REFUTED at `hook setup`, see F2.**
11. **Re-checked spec-marker honesty against the real matcher, reimplemented independently**
    (`^//\s*Scenario:\s*(.+?)\s*$`, counted only under an `// Acceptance:` header). 0 of 30
    guided-by-default scenarios uncovered. The four round-3 line-wrapped markers
    (`surface_agreement:68`, `autostart:82`, `cold_walk:54`, `precedence:54`) are single-line
    and matching. The only in-scope gap left is the recorded
    `deferred-findings-spec-scenario-uncovered` deferral. CLOSED.
12. **Attacked the driver-model tier across scopes.** A project that declares
    `driver_model` in `centinela.toml` is honoured by `ProjectDefaultProfile`/`ResolveStart`
    but ignored by `EffectiveProfile`/`ProfileProvenance`, which read only `wf.DriverModel`.
    REFUTED — see F1.
13. **Confirmed the three recorded deferrals are present in the roadmap Backlog** and did not
    re-litigate them.

#### Surface × config-shape matrix

Workflow fixture: post-flip state (`profileContract: guided-default-v1`), no `driverModel`
pin. `A`=absent, `E`=empty, `U`=unparseable, `0`=chmod-000, `V`=valid-no-profile,
`S`=valid-strict, `I`=invalid-profile-value, `D`=declared-unmapped-driver (`[orchestration]
driver_model`). **Truth** = the profile the operator's configuration asks for.

| Surface | A | E | U | 0 | V | S | I | D |
|---|---|---|---|---|---|---|---|---|
| **truth** | guided | guided | strict | strict | guided | strict | strict | **strict** |
| `status <f>` | guided | guided | strict | strict | guided | strict | strict | **guided** ⚠ |
| `verdict <f>` | guided | guided | refuse | refuse | guided | strict | refuse | **guided** ⚠ |
| MCP `read_rules` | guided | guided | strict | strict | guided | strict | strict | **guided** ⚠ |
| MCP `workflow_state` | ok | ok | refuse | refuse | ok | ok | refuse | ok |
| `doctor` profile-default | warn:guided | warn:guided | "unknown" | "unknown" | warn:guided | OK | "unknown" | **silent** ⚠ |
| `hook context` cadence | after_plan | after_plan | every_step | every_step | after_plan | every_step | every_step | after_plan ⚠ |
| `hook statusline` | ok | ok | ok | ok | ok | ok | ok | ok |
| `hook prewrite` | block | block | block | block | block | block | block | block |
| `hook plan-advisor` | advise | advise | warn+advise | warn+advise | advise | advise | warn+advise | advise |
| `hook autostart` | guided | guided | **decline** | **decline** | guided | strict | decline | strict |
| `hook setup` cascade | advisory | advisory | **halt** | **halt** | advisory | halt | halt | **halt** |
| `start <f>` | proceed | proceed | **refuse** | **refuse** | proceed | refuse | refuse | **refuse** |
| `roadmap promote` | seed | seed | require | require | seed | require | require | require |
| `precommit` / `pr-gate` / `audit` / `validate` | profile-blind in every shape |

Columns A–I: every surface agrees, and every unreadable shape lands strict or refuses.
Column D is the disagreement: three reporting surfaces say guided while four
acting surfaces enforce strict on the identical tree (F1).

#### Commands Run

- `go build -o /tmp/centinela-verify-gbd4 ./cmd/centinela` → exit 0
- `/tmp/centinela-verify-gbd4 validate` → **exit 0**, 493s — "All gates passed";
  `import_graph` and `spec-traceability-gate` warn, both pre-existing and warn-severity.
  (An earlier invocation of the same command was captured through a filtering shell wrapper
  that truncated its log and produced an unreliable exit record; it is disregarded, and the
  run recorded here is the one whose output I have in full.)
- `go test ./...` → **exit 0**, 477s, zero `FAIL` lines
- `go test ./cmd/centinela/ -run 'TestNoSwallowedConfigLoadOutsideTheAllowlist|TestProfileSurfacesUseTheSharedResolver|TestLoadEscapeAllowlistStaysHonest'` → exit 0 baseline; exit 1 under each of three injected mutants (discard idiom, rebind idiom, `internal/doctor` rebind); exit 0 restored
- `go test ./cmd/centinela/ -run TestValidateGatesHasNoProfileIdentifier` → exit 0 baseline; exit 1 with a profile identifier injected into `complete_validate_gates.go`; exit 0 restored
- `go test ./internal/workflow/ ./internal/config/ ./internal/doctor/` → exit 0 baseline; exit 1 with `ResolvedByLoad` removed from `EffectiveProfile`; exit 1 with it removed from `ProjectDefaultProfile`; exit 0 restored
- Binary-driven surface matrix across the eight config shapes (`status`, `verdict`, `doctor`, `hook context|statusline|prewrite|plan-advisor|setup|autostart`, `start`, `roadmap promote`)
- Driver-tier probe (`driver_model` = `haiku` / `opus` / unmapped, workflow without a pin)
- Confirmation-cadence probe (post-flip workflow at `code`, five config shapes)
- Provisional-mark lifecycle probe (seed → append → `roadmap generate` → pin strict → `start` / `hook setup`)
- MCP stdio JSON-RPC probe of `read_rules` across 12 tree/config combinations
- Independent reimplementation of the spec-traceability matcher over `specs/` × `tests/acceptance/`
- All mutations reverted; the post-restore working diff is byte-identical to the pre-mutation one.

#### Findings

1. **[MEDIUM] The workflow-scoped resolvers ignore the project's declared driver model, so
   `status`, `verdict` and MCP `read_rules` report guided on a tree where `start`,
   `hook setup`, `roadmap promote` and `doctor` all act strict.**
   `workflow.EffectiveProfile` and `workflow.ProfileProvenance` engage tier 3 only when
   `wf.DriverModel != ""`. `config.ProjectDefaultProfile` and `workflow.ResolveStart` engage it
   from `config.DriverModelFrom("", cfg)` — i.e. `[orchestration] driver_model`,
   `[orchestration.local] model`, and `$CENTINELA_MODEL`. A workflow started *before* the
   operator declared a driver has no pin in its state, so it skips tier 3 entirely and lands on
   the guided tail.
   Reproduced: `centinela.toml` = `[orchestration]\ndriver_model = "haiku"` (a mapped,
   *limited* class → strict), workflow state carrying `profileContract: guided-default-v1` and
   no `driverModel`. `status` prints `Profile guided (default (guided))`; `verdict` and MCP
   `read_rules` both report `guided`; `centinela start setup` refuses with the full strict
   greenfield cascade; `hook setup` halts on `roadmap required`. Identical result for an
   unmapped driver id. This is the spec's own AC5 — *"A driver model's capability class still
   outranks the new default"* — failing for every workflow that predates the declaration.
   Before this feature the tail was strict, so the same asymmetry could only ever resolve
   *tighter*; the guided tail is what turns it into a fail-open.
   Second-order: `doctor`'s `profile-default` check computes `ProjectDefaultProfile` (strict
   here) and therefore returns OK with "pinned, resolves elsewhere, or no workflows exist" —
   it goes silent in precisely the state where workflows *are* inheriting guided against a
   declared limited driver. The one surface built to warn about inherited guided cannot see
   this case.
   The agreement test does not catch it because it constructs its fixture as
   `wf := &workflow.Workflow{... DriverModel: config.DriverModelFrom("", cfg)}` — it injects
   the pin whose absence is the bug, and its own comment says that injection "is what makes
   the project-level and workflow-level answers comparable."
   Blast radius is PROCESS only: the reported profile, and `effectiveConfirmationMode`
   (`after_plan` instead of `every_step`). Orchestration evidence is pinned at creation and
   unaffected; prewrite gating is identical under strict and guided; no gate, suite, freshness
   check or claim verification reads this.
   Fix: give `EffectiveProfile`/`ProfileProvenance` the same tier-3 fallback the other two
   resolvers already use — `config.DriverModelFrom("", cfg)` when `wf.DriverModel` is empty —
   and re-point the agreement test at a workflow with no pin.
   *Affected scenario:* "A driver model's capability class still outranks the new default".

2. **[MEDIUM] The strict greenfield setup cascade accepts guided-seeded (`provisional`)
   grading artifacts that `start` refuses, so the senior-PM directive never fires on the
   exact tree the provisional mark exists to catch.**
   Round 3's fix wired the mark into `roadmap.ValidateAnalysis` / `ValidateQuality`, which
   `start_guard_cascade.go`, `roadmap validate` and strict `promote` all call. The setup
   cascade does not call them: `postRoadmapRungs()` tests file existence only —
   `exists(".workflow/roadmap-analysis.md") && exists(".workflow/roadmap-analysis.json")`.
   Reproduced: greenfield tree, no config (guided) → `roadmap promote finding --phase
   "Phase 0: Bootstrap" --scores 3,3,3,3,3,3` seeds all four artifacts with
   `"provisional": true` → write `enforcement_profile = "strict"` →
   `hook setup` skips "CENTINELA DIRECTIVE: roadmap analysis required. Delegate to senior
   product manager." and advances to the roadmap checkpoint, while `centinela start setup` on
   the same tree refuses with the provisional message. Baseline control: the same strict tree
   *without* seeded artifacts halts on the analysis rung as expected.
   The consequence is direction-specific and unpleasant: `hook setup` is the surface that
   *tells the agent what to do*, so under strict it now reports a rung satisfied that no
   evaluator has satisfied, and the operator only discovers it at `start`. Enforcement itself
   still holds, so this is PROCESS, not proof.
   Fix: have the two grading rungs' `satisfied` funcs call `roadmap.ValidateAnalysis` /
   `ValidateQuality` (or at minimum check the `provisional` key) instead of `exists`.
   *Affected scenario:* "A strict greenfield project still requires the full cascade".

3. **[LOW] `ProfileProvenance`'s `"default (strict, unreadable config)"` note is unreachable,
   and the note the operator actually sees is misleading.**
   Every production surface obtains its config from `config.LoadForProfile`, whose fallback
   sets `RawEnforcementProfile = strict`, so tier 2 answers first and status renders
   `Profile strict (global)`. On a `chmod 000` config the operator is told the profile came
   from a global pin — which may be a file they never wrote, or one that says something else.
   The tier-4 branch that would have said "unreadable config" can only be reached by a caller
   that hand-builds a config, which no surface does. Either route the unreadable case so the
   honest note reaches the screen, or delete the dead branch. Cosmetic, and strict-direction.

#### Deferred Findings

Confirmed present in the roadmap Backlog, not re-litigated:
- `provenance-local-default-ignores-capability-profile-override`
- `deferred-findings-spec-scenario-uncovered`
- `profile-display-unused-and-now-misleading`

#### Recommendation

**WARNING — safe to proceed on proof, with two process-level disagreements open.**

The load-bearing guarantee of this feature is intact and independently verified: proof is
identical under every profile, the gates and the full suite pass, and the recurring
*unreadable-config* fail-open is now closed by construction rather than by patching call
sites — the marker cannot be forged, is preserved across copies, and both resolvers fail
when it is removed. Round 3's four findings are all closed and each was re-verified by
mutation rather than by reading the fix.

The recurring class as a whole is **not** closed. It has moved: F1 and F2 are the same
shape of defect — one surface answering looser than another about the same tree — on two
input axes the provenance marker does not reach (a config-declared driver model, and a
seeded grading artifact). Both are small, local fixes in the resolvers and the cascade
rungs respectively, and neither weakens any gate. I would fix F1 and F2 before merge or
record them as deferrals; F3 is polish.

```json centinela:verification
{
  "revision": "1f5cc0d8de14c575b43f293187bab6ef7c1162d0",
  "treeDigest": "sha256:f4d7e28166a8af0bd7aa702e4364aecde88985ed528c8cc44c016c6a5663e620",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-verify-gbd4", "./cmd/centinela"], "exitCode": 0, "durationMs": 386},
    {"argv": ["/tmp/centinela-verify-gbd4", "validate"], "exitCode": 0, "durationMs": 493551},
    {"argv": ["go", "test", "./..."], "exitCode": 0, "durationMs": 477385},
    {"argv": ["go", "test", "./cmd/centinela/", "-run", "TestNoSwallowedConfigLoadOutsideTheAllowlist|TestProfileSurfacesUseTheSharedResolver|TestLoadEscapeAllowlistStaysHonest", "-count=1"], "exitCode": 0, "durationMs": 1107},
    {"argv": ["go", "test", "./cmd/centinela/", "-run", "TestValidateGatesHasNoProfileIdentifier", "-count=1"], "exitCode": 0, "durationMs": 793},
    {"argv": ["go", "test", "./internal/workflow/", "./internal/config/", "./internal/doctor/", "-count=1"], "exitCode": 0, "durationMs": 960},
    {"argv": ["zsh", "scratchpad/matrix.sh"], "exitCode": 0, "durationMs": 2605},
    {"argv": ["zsh", "scratchpad/probe5.sh"], "exitCode": 0, "durationMs": 806},
    {"argv": ["zsh", "scratchpad/probe6.sh"], "exitCode": 0, "durationMs": 1055},
    {"argv": ["zsh", "scratchpad/probe7.sh"], "exitCode": 0, "durationMs": 882},
    {"argv": ["zsh", "scratchpad/mcpprobe.sh"], "exitCode": 0, "durationMs": 49304}
  ]
}
```
