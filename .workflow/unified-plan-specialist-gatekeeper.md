# Adversarial Verifier Report — unified-plan-specialist (round 5, fresh context)

**Status:** WARNING

Round 5 is the closing round. Fresh context; the verdict is re-derived from
scratch against `origin/main...HEAD` plus the uncommitted tree, the brief, the
spec, the plan, and the output of commands run here. Rounds 1–4 were treated as
claims to refute, not facts to accept — no prior narrative was read as input.

I could not refute the completion claim. Every refutation attempt below either
reproduced correct behavior or landed on an item already recorded as deferred.
Three LOW findings remain; one (F1) is actionable before merge, which is why
this is WARNING rather than SAFE.

---

## Inputs Read

- `git diff origin/main...HEAD` (122 files, +3953/−672) and `git status --porcelain`
- `docs/features/unified-plan-specialist.md`
- `specs/unified-plan-specialist.feature` (264 lines, 20 scenarios)
- `docs/plans/unified-plan-specialist.md` (447 lines)
- `.workflow/roadmap.json`, `ROADMAP.md`, `.workflow/unified-plan-specialist.json`
- `internal/workflow/{contract,validate_orchestration,order,steps,state}.go`
- `internal/orchestration/{policy,models,evidence,output_rules,plan_snapshot,validate}.go`
- `internal/evidence/{roles_retired,invalidation_targets}.go` + its test
- `internal/config/orchestration_plan_alias.go`
- `internal/setup/opencode_agent_config.go`, `internal/setup/testdata/golden/opencode/opencode.json`, `opencode.json`
- `internal/planadvisor/{advisor,questions}.go`
- `internal/scaffold/planner_role_asset_test.go`; all 27 `internal/scaffold/assets/docs/architecture/*` vs `docs/architecture/*`
- `cmd/centinela/{verify_deps,evidence_init,hook_statusline_view,verify,verdict,mcp,complete_verify}.go`
- `tests/acceptance/unified_plan_specialist_*.go`
- `docs/guides/configuration.md`, `docs/guides/configuration-reference.md`
- `centinela.toml`, `scripts/check-coverage.sh`

Never read: any orchestrator summary or another role's report. The prior
gatekeeper JSON was opened only to append to it.

---

## Refutation Attempts

1. **"The round-4 mirror sync is cosmetic — the shipped bytes still name the retired pair."** Refuted. `cmp` over all 27 embedded arch assets: `workflow-enforcement.md` and `planner-prompt.md` are byte-identical source↔mirror. The three drifting pairs (`gatekeepers.md`, `new-project-guide.md`, `testing-strategy.md`) drift *identically on `origin/main`* — local "Preserved Custom Sections" additions with no plan-role text. `production-readiness-prompt.md.template` is a pre-existing `.template`-suffixed pair, not a new gap.
2. **"The new embedded-asset guard is decorative and would not catch a regression."** Refuted by mutation. Restoring the mirror to its `origin/main` bytes fails **both** guards (`TestEmbeddedArchAssets_NamePlannerNotRetiredPlanRoles` twice, `TestEmbeddedWorkflowEnforcement_NamesPlanner` — "never names the planner role"). File restored and re-verified byte-identical. The guard asserts the *embedded* bytes, which is the right surface: `workflow-enforcement.md` sits on `mirrorParityAllowlist`, so byte-parity never covered it.
3. **"The config guides were not really moved to the planner world."** Refuted. `grep` over `docs/guides/` returns **zero** occurrences of `big-thinker`/`feature-specialist`/`big_thinker`/`feature_specialist`. The reference now lists `planner` and documents the retired keys, the `big-thinker` > `feature-specialist` alias precedence, and the one-time doctor/start notice — matching `internal/config/orchestration_plan_alias.go` exactly.
4. **"A recorded deferral was quietly dropped to make the list look clean."** Refuted. `verdict-mcp-contract-blind-roles` is gone *because it was fixed in-branch* (`cmd/centinela/verify_deps.go` + 3 colocated tests). All **12** slugs handed to me are present exactly once in `.workflow/roadmap.json`, and `roadmap_drift` is green so `ROADMAP.md` matches.
5. **"The retired-role refusal is a unit-test fiction; the CLI still lets you author legacy evidence."** Refuted end-to-end in a scratch repo driven by a binary built from this branch: pinned `planner-v1` + `evidence init X big-thinker` → exit 1 with the retired-role message; unpinned workflow + `big-thinker` → allowed.
6. **"D2's transition hole is open — a fresh feature can pass plan with hand-authored legacy files."** Refuted. A pinned workflow carrying a **forged complete legacy pair** is refused by `complete`, naming `.workflow/<f>-planner.{md,json}`. Conversely an unpinned workflow with a **partial** legacy set is refused with the contract annotation, and an unpinned workflow holding **planner-only** evidence is refused too — back-compat cannot be dodged in either direction. Positive path confirmed: planner evidence (outputs plan+spec, non-empty edgeCases, status done) advances plan → code.
7. **"The round-4 fixes introduced a regression."** Refuted. `go test ./...` green (3632 tests, 45 packages). The removed `InvalidationTargets` plan branch is genuinely unreachable — `plan` is index 0 of both `DefaultStepOrder` and `BootstrapStepOrder` (the only two orders in the tree), and `TestPlanIsFirstStepSoInvalidationNeverSeesIt` fails the moment that stops being true. `verify.Deps` is now built in exactly one production site (`verifyDepsFor`); `orchestration.ValidateStep` retains no production caller.
8. **"The spec-traceability warn hides uncovered scenarios in this feature's own spec."** Refuted by re-implementing the check: 20 scenarios in `specs/unified-plan-specialist.feature` map 1:1 onto 20 `// Scenario:` markers, including both `Scenario Outline`s. The warn's only uncovered scenarios — "Defer auto-resolves --source from worktree CWD…" and "Runtime configuration is unchanged" — exist verbatim on `origin/main`.
9. **"Legacy role strings are orphaned across the tree."** Partially true, and every instance is already recorded: `internal/brownmap/{baseline,gaps}.go` still stamp `Source.Role="big-thinker"` (`brownmap-roadmap-source-role-retired`); the repo's own `opencode.json` still carries the two retired agents because `mergeOpenCodeAgents` only adds keys (`managed-agent-retirement-sweep`); other specs still describe legacy behavior (`stale-spec-text-after-role-retirement`). `cmd/centinela/hook_statusline_view.go` correctly covers `-planner` alongside the legacy suffixes.
10. **"`start` never pins the contract, so the planner is required by nobody."** Refuted — `internal/workflow/order.go:44` sets `PlanContract: PlanContractUnified` on every new workflow, and the scratch-repo runs above depend on it.
11. **"Brief criteria 3/5/6 are unmet."** Refuted: `RolePlanner: TierReasoning` and `AllowedRunnerKeys` include planner; plan-advisor questions are re-lensed to `strategy`/`spec` and the directive reads "One planner agent, two lenses"; the OpenCode generator emits `planner`, stops emitting the two legacy plan agents, and the golden moved in the same change.
12. **"`doctor` reports drift this feature introduced."** Not sustainable — see F3; unreproducible across four subsequent runs on two binaries, with `opencode.json` already carrying the `planner` agent and its `build` task permission.
13. **Dogfood check.** `.workflow/unified-plan-specialist.json` carries **no** `planContract`, so this feature's own plan step was gated on the complete legacy pair and holds exactly `big-thinker` + `feature-specialist` evidence. The state-dated rule is proven on the very workflow that shipped it.

---

## Commands Run

| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `go build -o /tmp/centinela-verify-ups5 ./cmd/centinela` | 0 | ~9s |
| 2 | `centinela validate` | **0** | **602s** |
| 3 | `go test ./...` | **0** | **301s** (3632 tests, 45 packages) |
| 4 | `git diff origin/main...HEAD --stat` / `git status --porcelain` / `git log --oneline origin/main..HEAD` | 0 | <1s |
| 5 | `cmp` loop over all 27 `internal/scaffold/assets/docs/architecture/*` vs `docs/architecture/*` | 0 | <1s |
| 6 | `diff <(git show origin/main:docs/architecture/<f>) <(git show origin/main:internal/scaffold/assets/…/<f>)` ×3 | 0 | <1s |
| 7 | `go test ./internal/scaffold/... -run TestEmbedded` | 0 | 1s |
| 8 | *mutation*: mirror ← `origin/main` bytes, then `go test ./internal/scaffold/... -run TestEmbedded` | **1 (2 FAIL)** | 1s |
| 9 | restore mirror + `cmp` re-verify | 0 | <1s |
| 10 | `git diff -- docs/guides/…` + `rg 'big-thinker\|feature-specialist' docs/guides/` | 1 (no match) | <1s |
| 11 | `grep -c '"name":"<slug>"' .workflow/roadmap.json` ×12 | 0 (all 1) | <1s |
| 12 | scratch repo: `evidence init freshfeat big-thinker` (pinned `planner-v1`) | **1** — retired-role error | <1s |
| 13 | scratch repo: `evidence init legacyfeat big-thinker` (unpinned) | 0 — allowed | <1s |
| 14 | scratch repo: `complete freshfeat` with forged complete legacy pair | **1** — "missing: …-planner.md, …-planner.json" | <1s |
| 15 | scratch repo: `complete legacyfeat` with partial legacy set | **1** — missing + invalid + contract annotation | <1s |
| 16 | scratch repo: `complete legacyfeat` with planner-only evidence | **1** — names the legacy pair | <1s |
| 17 | scratch repo: `evidence init/append/set` + `complete freshfeat` (positive path) | 0 — plan → code | <1s |
| 18 | scratch repo: `verdict freshfeat` / `verdict legacyfeat` | 0 | <1s |
| 19 | scenario↔marker cross-check (`specs/unified-plan-specialist.feature` vs `tests/acceptance/unified_plan_specialist_*.go`) | 0 — 20/20 | <1s |
| 20 | uncovered-scenario check on the two other modified specs + `git show origin/main:` confirmation | 0 | <1s |
| 21 | `centinela doctor` ×4 (branch binary ×3, installed 0.49.1 ×1) + one `doctor --fix` | 0 | <1s each |
| 22 | `go tool cover -func=coverage.out \| tail -1` | 0 — **total 97.1%** | 2s |
| 23 | `centinela evidence append/set unified-plan-specialist gatekeeper …` | 0 | <1s |
| 24 | `centinela artifact stamp unified-plan-specialist` | 0 | <1s |

`centinela validate` was run exactly once, as required.

### Gate detail from run #2

```
✓ G1: File Size          All files under 100 lines.
✓ G-Build: Cross-Compile All 6 release targets compile.
⚠ import_graph           Packages match no configured layer:
⚠ spec-traceability-gate Scenarios without acceptance coverage:
✓ roadmap_drift          ROADMAP.md is in sync.
✓ go test ./...   ✓ ./scripts/check-coverage.sh   ✓ ./scripts/check-fmt.sh
All gates passed.
```

Both warns are pre-existing categories; the traceability warn's two uncovered
scenarios exist unchanged on `origin/main`. Coverage 97.1% vs the 95.0% gate.

### Reading the `centinela:verification` record

The machine-readable block below carries only entries that are a *single
executed command*, with argv verbatim. Two shapes in it need a word of prose:

- **`go test ./internal/scaffold/... -run TestEmbedded` appears three times** —
  exit 0 (baseline), exit **1** (run against the mutated mirror, failing both
  guards), exit 0 (after restoring the mirror). The mutation itself was a file
  copy, not a command, so it is described here and in Refutation Attempt #2
  rather than recorded as a pseudo-argv.
- **`complete legacyfeat` appears twice at exit 1** — two distinct probes
  against the same unpinned workflow: once with a *partial* legacy set, once
  with *planner-only* evidence. Both are refusals, for different reasons.

The `cmp` sweep over the 27 scaffold assets, the 12-slug roadmap grep, and the
scenario↔marker cross-check were shell loops rather than single commands; they
are reported in the table above and in Refutation Attempts #1, #4 and #8.

All exit codes in the block are measured, not inferred: the four refusal probes
and the mutated scaffold run were re-executed with `$?` captured directly after
the first pass had masked it behind a pipe.

Argv[0] is recorded as it was actually invoked — `centinela` for the installed
0.49.1 binary, `/tmp/centinela-verify-ups5` for the scratch binary built from
this branch, which is what drove every scratch-repo probe.

---

## Findings

**3 findings, all LOW. No CRITICAL, no HIGH, no MEDIUM.**

### F1 — LOW — the entire validate-step remediation is uncommitted, junk included

Carried from round 4 and still open at verdict time. `git status` shows the
`.gitignore` hardening unstaged and the deletions of `docs/.DS_Store` and
`docs/plans/.unified-plan-specialist.md.swp` staged but uncommitted, alongside
every code/spec/test fix from rounds 3–5. Merging before `centinela complete`
auto-commits the validate step lands the editor/OS junk on `main`; a tracked
`.swp` makes every later editor session prompt for swap recovery on the doc it
shadows. Self-resolving via the normal `complete` path — but only if that path
runs before the merge.

### F2 — LOW — warn-severity gates print their header with an empty detail list

Pre-existing and out of this feature's scope (`internal/ui/render_gates.go` is
untouched by this branch). `ui.RenderGateResult` renders `Result.Details` only
for `Fail`, so my own `centinela validate` printed
`⚠ spec-traceability-gate  Scenarios without acceptance coverage:` followed by
nothing at all. I had to re-implement the gate's matcher by hand to learn which
two scenarios it meant — exactly the work the gate exists to save. Not on the
deferred list; recommend recording it (`.workflow/roadmap.json` **and**
`ROADMAP.md` together, or `roadmap_drift` goes red).

### F3 — LOW — OBSERVATION, unreproducible, recorded for honesty not as a defect

The first `centinela doctor` of this session reported an **error**:
`prewrite-hook needs update at .claude/settings.json` and
`opencode-config needs update at opencode.json`. `doctor --fix` cleared it while
writing **nothing** — `opencode.json` and `.claude/settings.json` are
byte-identical to backups taken immediately before, with mtimes still 12:38 and
10:54 — and four subsequent runs (branch binary ×3, installed 0.49.1 ×1) report
clean with zero errors. I cannot ground it, and it is not plan-role-shaped:
`opencode.json` already carries the `planner` agent and its `build` task
permission, matching both the generator and the golden.

---

## Deferred Findings

All **12** slugs handed to this round are recorded in `.workflow/roadmap.json`
(exactly one record each) and mirrored in `ROADMAP.md` (`roadmap_drift` green).
Not re-litigated; several were independently reproduced in passing and are noted
as such.

| slug | recorded | reproduced this round |
|------|----------|-----------------------|
| `codex-claude-role-agent-registry` | ✓ | — |
| `managed-agent-retirement-sweep` | ✓ | ✓ `mergeOpenCodeAgents` only adds keys; the repo's own `opencode.json` still carries both retired agents |
| `prompt-doc-budget-ratchet` | ✓ | — (`planner-prompt.md` is 129 lines, inside the 130 budget) |
| `statusline-plan-role-display` | ✓ | — |
| `revise-to-plan-sheds-no-evidence` | ✓ | ✓ no `plan` branch in `InvalidationTargets`; unreachable by construction |
| `legacy-plan-key-mixed-form-precedence` | ✓ | — |
| `planner-skeleton-prompt-header-drift` | ✓ | — |
| `stale-spec-text-after-role-retirement` | ✓ | ✓ ~8 other specs still describe legacy `evidence init … big-thinker` behavior the retired-role guard now refuses |
| `brownmap-roadmap-source-role-retired` | ✓ | ✓ `internal/brownmap/{baseline,gaps}.go` still stamp `Role: "big-thinker"` |
| `handoff-chain-unvalidated` | ✓ | — |
| `planner-stub-on-legacy-workflow-unguarded` | ✓ | ✓ `evidence init legacyfeat planner` succeeds on an unpinned workflow whose gate can never accept it |
| `mirror-workflow-enforcement-doc` | ✓ | still legitimately open — the file remains on `mirrorParityAllowlist`, so byte-parity still does not cover it; the new embedded guard covers only the planner property |

New deferrals proposed this round: **none**. F2 is recommended for recording.

---

## Recommendation

**PROCEED to the docs step.** `centinela validate` exit 0, `go test ./...` exit 0
(3632 tests across 45 packages), coverage 97.1% against a 95.0% gate, G1 and
cross-compile green, and the state-dated `planner-v1` contract holds in both
directions under direct end-to-end exercise — including the forged-legacy-pair
attack the brief named as its High risk. All three round-4 fixes are real:
mutation-proven for the mirror guard, grep-proven for the guides, and
record-proven for the removed false deferral.

Before merge:

1. Commit the working tree — especially `.gitignore` and the two junk-file
   deletions (F1). Let `centinela complete unified-plan-specialist` do it.
2. Record F2 as a deferred finding, updating `.workflow/roadmap.json` **and**
   `ROADMAP.md` together.
3. Re-run `centinela doctor` post-merge; if F3's hook/opencode drift reappears,
   it deserves a real investigation rather than a `--fix`.

Status is WARNING rather than SAFE solely because F1 is actionable *now*: the
tree that passed these gates is not yet the tree that would be merged.

Note on the evidence JSON: `evidence set` cannot clear a list field, so
`edgeCases` retains round-4's records beneath the `R5 …` entries. One R4 line —
the claim that `verdict-mcp-contract-blind-roles` is still recorded — is
explicitly superseded by an `R5 SUPERSEDES` entry rather than deleted.

---

```json centinela:verification
{
  "revision": "128dcdb0f6a7ac3ffab205e70a2ee272ee120610",
  "treeDigest": "sha256:db5abfaf07f0cb0c04f6b996eb5c5d0881a7fa6c4e3f594787ea80c517cceb28",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-verify-ups5", "./cmd/centinela"], "exitCode": 0},
    {"argv": ["centinela", "validate"], "exitCode": 0, "durationMs": 602000},
    {"argv": ["go", "test", "./..."], "exitCode": 0, "durationMs": 301000},
    {"argv": ["go", "test", "./internal/scaffold/...", "-run", "TestEmbedded"], "exitCode": 0, "durationMs": 1000},
    {"argv": ["go", "test", "./internal/scaffold/...", "-run", "TestEmbedded"], "exitCode": 1, "durationMs": 1000},
    {"argv": ["go", "test", "./internal/scaffold/...", "-run", "TestEmbedded"], "exitCode": 0, "durationMs": 1000},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "init", "freshfeat", "big-thinker"], "exitCode": 1},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "init", "legacyfeat", "big-thinker"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "init", "freshfeat", "planner"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "init", "legacyfeat", "planner"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "init", "legacyfeat", "feature-specialist"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "complete", "freshfeat"], "exitCode": 1},
    {"argv": ["/tmp/centinela-verify-ups5", "complete", "legacyfeat"], "exitCode": 1},
    {"argv": ["/tmp/centinela-verify-ups5", "complete", "legacyfeat"], "exitCode": 1},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "append", "freshfeat", "planner", "outputs", "docs/plans/freshfeat.md"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "append", "freshfeat", "planner", "outputs", "specs/freshfeat.feature"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "append", "freshfeat", "planner", "edgeCases", "empty input"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "evidence", "set", "freshfeat", "planner", "status", "done"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "complete", "freshfeat"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "verdict", "freshfeat"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "verdict", "legacyfeat"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "doctor"], "exitCode": 0},
    {"argv": ["/tmp/centinela-verify-ups5", "doctor", "--fix"], "exitCode": 0},
    {"argv": ["centinela", "doctor"], "exitCode": 0},
    {"argv": ["go", "tool", "cover", "-func=coverage.out"], "exitCode": 0, "durationMs": 2000},
    {"argv": ["centinela", "evidence", "set", "unified-plan-specialist", "gatekeeper", "extra.verdict", "WARNING"], "exitCode": 0},
    {"argv": ["centinela", "evidence", "set", "unified-plan-specialist", "gatekeeper", "status", "done"], "exitCode": 0},
    {"argv": ["centinela", "evidence", "set", "unified-plan-specialist", "gatekeeper", "handoffTo", "documentation-specialist"], "exitCode": 0},
    {"argv": ["centinela", "evidence", "append", "unified-plan-specialist", "gatekeeper", "outputs", ".workflow/unified-plan-specialist-gatekeeper.md"], "exitCode": 0},
    {"argv": ["centinela", "artifact", "stamp", "unified-plan-specialist"], "exitCode": 0}
  ]
}
```
