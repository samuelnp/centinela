### Adversarial Verifier Report: hook-context-panel-diet
**Date:** 2026-08-04
**Status:** SAFE

#### Inputs Read

- `git diff main...HEAD` (23 files). Merge-base == `main` == `7513283`, so no rebase drift.
- Working tree: clean before and after this review (`git status --porcelain=v1 -uall` empty). `coverage.out`
  and `.workflow/.roadmap-digest` produced by my runs are both gitignored — confirmed with `git check-ignore`.
- `specs/hook-context-panel-diet.feature` (11 scenarios), `docs/features/hook-context-panel-diet.md`,
  `docs/plans/hook-context-panel-diet.md`
- Source under test: `internal/ui/panel.go`, `render.go`, `render_brief.go`, `render_review.go` — the only
  four non-test files changed
- Tests under test: `internal/ui/panel_budget_test.go`, `panel_trim_test.go`, `render_trailing_ws_test.go`,
  `render_cli_padding_test.go`; `tests/unit/hook_context_panel_diet_unit_test.go`;
  `tests/integration/hook_context_panel_diet_integration_test.go`;
  `tests/acceptance/hook_context_panel_diet{,_guard,_helper}_test.go`
- Call sites re-derived from source, not from any narrative: `cmd/centinela/hook_context.go`,
  `hook_context_roadmap.go`, `hook_context_review_mode.go`, `hook_workflows.go`, `internal/gates/file_size.go`,
  and the `UserPromptSubmit` wiring in `.claude/settings.json`
- The output of every command listed under Commands Run

**Flag (required by the Input Contract):** this verifier prompt DID contain a narrative summary of the
implementation — "six render functions padded every line… a `trimTrailingWS` helper now wraps those six
returns… 2004B → 1195B with content byte-identical". That narrative was treated strictly as a set of claims
to falsify. Every number in it was re-derived from source below. No role's `.workflow/*.md` was accepted as
proof; the planner / senior-engineer / qa-senior / edge-cases artifacts were read only to locate claims.

#### Analyzed Specs

`specs/hook-context-panel-diet.feature`, 11 scenarios across four groups: padding removal (3), governance
signal survives (3), human TTY rendering unchanged (2), size guard (3). Traceability re-derived
independently in refutation attempt A9 below — not taken from the gate's own report.

#### Refutation Attempts

**A1 — "The saving is real and no content changed." → could not refute; independently reproduced.**
Built two binaries from source: `/tmp/centinela-main` from `main` (`7513283`, verified clean Go tree) and
`/tmp/centinela-v14` from this worktree. Drove `hook context` on both across **17 distinct project states**
from an isolated sandbox (never the real `.workflow/`), comparing content, exit code, stderr and
padded-line count. Result: v14 emitted **0 trailing-whitespace lines in all 17 states**; exit codes and
stderr identical everywhere; content **byte-identical after stripping trailing whitespace in all 17** —
i.e. the only characters removed are trailing spaces/tabs.

| state | main B | v14 B | main padded | v14 padded | content identical |
|---|---|---|---|---|---|
| 0 workflows | 113 | 113 | 0 | 0 | yes |
| 1 canonical @plan (brief nudge fires) | 1512 | 883 | 16 | 0 | yes |
| 1 canonical @plan (brief present) | 147 | 114 | 1 | 0 | yes |
| 5 mixed archetypes | 3333 | 2232 | 38 | 0 | yes |
| 7 workflows (capped, "+2 more active") | 3727 | 2525 | 42 | 0 | yes |
| canonical / hotfix / refactor / spike | 549/493/525/475 | 349/321/344/310 | 6/6/5/6 | 0/0/0/0 | yes |
| done workflow present | 555 | 352 | 6 | 0 | yes |
| docs step, internal feature (changelog nudge) | 452 | 345 | 5 | 0 | yes |
| docs step, user-facing feature (docs nudge) | 472 | 369 | 5 | 0 | yes |
| tests step, edge-case nudge | 547 | 367 | 5 | 0 | yes |
| degraded state file (schemaVersion 99, unmodellable) | 1092 | 688 | 12 | 0 | yes |
| corrupt state file (invalid JSON) | 555 | 352 | 6 | 0 | yes |
| roadmap-summary line emitted | 1562 | 933 | 16 | 0 | yes |
| review-required panel armed | 612 | 399 | 6 | 0 | yes |

Also verified live in this worktree with fresh session ids: main 234 B / 1 padded line → v14 214 B / 0
padded, content identical. (A first live comparison appeared to differ; the cause was the roadmap-summary
line being session-digest-gated — the main run consumed the digest for that session id. Re-run with
distinct session ids: identical. Recorded here rather than reported as a finding.)

**A2 — "0 trailing whitespace across all states / a missed emitter." → could not refute.**
Enumerated every emitter inside `runHookContext` from source: `emitRoadmapSummary`→`ui.RenderRoadmapSummary`,
`ui.RenderSuccess`, the bare `fmt.Println` directive, plus the six trimmed renderers. The two untrimmed
ones go through `renderSystemLine`, which is a **single line with no `lipgloss.JoinVertical`** —
structurally incapable of padding. Confirmed empirically in the 0-workflow and roadmap-summary states
above. No missed emitter exists in `hook context`.

**A3 — "Padding survives when colour is on." → could not refute (the most promising attack).**
Hypothesis: `lipgloss` might append padding *inside* the styled span, so a padded line would end in the
ANSI reset (`\x1b[0m`) rather than a space, and `TrimRight(l, " \t")` would silently miss it — invisible in
the non-TTY environment every test runs in. Forced colour with `CLICOLOR_FORCE=1` on a 3-workflow fixture:
output carried **96 ANSI escape sequences** (2058 raw bytes, 1624 after stripping escapes) and still **0
padded lines, both raw and after ANSI-stripping**. `JoinVertical` pads outside the reset; the trim is
correct in colour mode too.

**A4 — "No content was cut." → could not refute.** Read all six trimmed renderers: no body line in
`render.go`, `render_review.go` or `render_brief.go` ends in a meaningful space or tab — no tab-separated
columns, no NBSP, no literal `\r`. `panel_trim_test.go` pins those three latent hazards explicitly.
Combined with the 17-state content-identity diff in A1, nothing but padding is removed.

**A5 — "Claimed 2004B → 1195B." → could not refute; BOTH numbers reproduced exactly.**
- *1195 (after)*: binary-searched the shipped `CENTINELA_PANEL_DIET_PAD` knob against the 1400-byte budget.
  `PAD=205` → pass; `PAD=206` → `measured 1401 bytes, budget 1400`. Therefore measured = 1401 − 206 =
  **1195**. Exact match. Cross-checked at `PAD=233` (`1428`) and `PAD=234` (`1429`).
- *2004 (before)*: `rsync`ed the worktree to an out-of-tree scratch copy, rewrote `trimTrailingWS` to the
  identity function **there** (never in the real tree), and ran the guard: `measured 2004 bytes, budget
  1400`. Exact match. The before/after pair is not a self-report — it is reproducible from source.

**A6 — "The size guard is honest / not vacuous / actually fails." → could not refute.**
- *Fails when output grows*: proved twice, two different ways. Synthetic growth (`PAD=206`) → FAIL. Real
  regression (fix reverted out-of-tree) → FAIL at 2004 B, **and** `TestPanelBudgetNoTrailingWhitespace`,
  `TestTrimTrailingWSTable`, `TestTrimTrailingWSLongLine` and `TestHookPanelsCarryNoTrailingWhitespace` all
  went red simultaneously. This is a guard I watched fail, not one I trusted.
- *Not vacuous*: 205 B headroom over a 1195 B measurement (17%); a reverted padding pass lands 604 B past
  budget. Tight enough to catch the regression it exists for.
- *No repo-state dependency*: `panelDietRender` builds a literal in-memory `&workflow.Workflow{}` — no
  `.workflow` read, no `ActiveWorkflows`, no cwd dependency. `TestPanelBudgetIndependentOfRepoState`
  re-measures from a temp dir seeded with 7 unrelated state files and requires byte-identity;
  `TestPDSizeGuardIgnoresLiveWorkflowCount` additionally greps the guard source for `ActiveWorkflows`,
  `workflow.Load`, `os.ReadDir`. All pass. The `t.Chdir` in that test is safe: nothing in `internal/ui`
  calls `t.Parallel()`, so package tests are sequential and `t.Chdir` restores on cleanup.
- *Can the env knob weaken the guard in a real run?* **No.** `grep` over `internal/` and `cmd/` for
  non-`_test.go` references returns nothing — it is read only in `panel_budget_test.go`. It is **monotone
  in one direction**: `strconv.Atoi` + `n > 0` means it can only *append* bytes, so any value (hostile,
  leaked, or in CI) can only push the measurement past budget, never under it. It feeds `panelDietMeasured`
  only, never `panelDietRender`, so it cannot mask the trailing-whitespace assertion. Negative, zero and
  garbage values are a no-op. Fail-closed by construction.

**A7 — "No CLI surface regressed." → could not refute; verified byte-for-byte against a main build.**
Ran main vs v14 and compared stdout, stderr and exit code with `cmp`:

| command | result | bytes | padded lines preserved |
|---|---|---|---|
| `status sample-feature` | BYTE-IDENTICAL, rc 0 | 368 | 11 |
| `hook prewrite` (blocked write) | BYTE-IDENTICAL, rc 2 | 1216 | 7 |
| `roadmap` (real repo, 181 features) | BYTE-IDENTICAL, rc 0 | 25936 | 0 |
| `doctor` | BYTE-IDENTICAL, rc 1 | 794 | 0 |
| `workflows`, `roadmap --json`, `--help` | BYTE-IDENTICAL | — | — |

`RenderRoadmap` genuinely has **zero** padding to preserve (it composes with `strings.Join`, not
`JoinVertical`), which my own measurement confirms. The shipped `TestRenderRoadmapUnaffected` says exactly
that rather than falsely asserting "still padded". The prompt's narrative that all three CLI renderers
"were deliberately left padded" is imprecise for `RenderRoadmap`; the code and its test are honest.

**A8 — "Tests assert the shipped contract, not its opposite." → could not refute.**
Read every new test. Names match behaviour in both directions: the padding-removal tests
(`TestHookPanelsCarryNoTrailingWhitespace`, `TestHookOnlyPanelsAreUnpadded`, `TestPD*HasNoTrailingWhitespace`)
assert absence; the four scope-boundary tests (`TestCLIRenderersKeepTheirPadding`,
`TestCLIRenderStatusStillPadded`, `TestPDStatusOutputKeepsItsPadding`, `TestPDBlockedWritePanelKeepsItsPadding`)
assert padding **survives** on `RenderStatus`/`RenderBlocked` — and I confirmed independently that it does
(11 and 7 padded lines, byte-identical to main). No inverted test found.
`TestDietFixturesActuallyPadWhenUntrimmed` is a genuine anti-tautology check: it composes a fixture the way
`RenderContextCapped` does *without* the trim and asserts that raw form IS padded, so the "no padding"
assertions have real padding to remove rather than passing by fixture accident.

**A9 — "11 scenarios, exact literal `// Scenario:` markers." → could not refute.**
Extracted all 11 `Scenario:` lines from the spec and all `// Scenario:` markers from every `.go` file in the
repo, then required **exact string equality** — not the gate's matcher, not substring. **11/11 matched
literally**, including the apostrophe in "this repository's live workflow count". Markers live at
`tests/acceptance/hook_context_panel_diet_test.go:{10,22,34,52,65,100,109,126}` and
`hook_context_panel_diet_guard_test.go:{18,28,69}`. The gate independently agrees.

**A10 — "A gate was weakened to make this pass." → could not refute.**
`git diff main...HEAD --name-only` touches nothing outside `.workflow/`, `docs/`, `specs/`, `internal/ui/`,
`tests/`. No change to `centinela.toml`, `.github/`, `scripts/`, or `internal/gates/`. No threshold moved.
Exactly four non-test production files changed.

**A11 — "File-size gate, FULL SCAN."** Ran the scan over the gate's real roots
(`src internal cmd lib app pkg`, skipping `.git node_modules vendor dist .next target build`, per
`internal/gates/file_size.go`): **1626 Go files scanned, 0 violations >100 lines.** 15 files sit at exactly
100, including `internal/ui/render.go` (99 → 100 in this diff). `centinela audit` was attempted for a
native full scan but exits with `no baseline`, so the equivalent scan was performed directly. Files over
100 lines exist only under `tests/`, which the gate does not scan.

**A12 — "The nested `go test` subprocesses worsen the known acceptance 10m timeout."**
Timed the 11 new acceptance tests: **4.7 s total**, of which three nested `go test ./internal/ui/` runs cost
~1 s each. Negligible. The full `validate` completed in 513 s with no acceptance panic, so the known flake
(`acceptance-tier-exceeds-default-test-timeout`) did not occur and one suite run sufficed.

**A13 — "Sibling `UserPromptSubmit` hooks were quietly missed."** `.claude/settings.json` wires **seven**
commands to `UserPromptSubmit`. Measured all seven on one fixture with v14: `context` 934 B / **0** padded,
`setup` 5533 B / **27** padded, `migrate` 300 B / **4** padded, `plan-advisor` 630 B / 0, and
`autostart`/`orchestration`/`merge` silent. The residual padding is real but sits in hooks that are gated or
rare (silent in a fully configured repo), is explicitly out of scope per plan decision D4, and is already
named in the plan as a follow-up. Not a false claim — recorded as a deferred finding.

#### Commands Run

All commands run from the worktree root
`/Users/samuelnp/projects/personal/centinela/.worktrees/hook-context-panel-diet` unless noted. Exit codes
captured directly (`; echo EXIT=$?` / `$?`), never inferred from banner text.

| # | argv | exit | duration |
|---|---|---|---|
| 1 | `git rev-parse HEAD` → `c4a8067583ce0ee48d88acd6d9b7bcf86d50dd8b` | 0 | <1s |
| 2 | `git status --porcelain=v1 -uall` (empty) | 0 | <1s |
| 3 | `git diff main...HEAD --stat` | 0 | <1s |
| 4 | `git merge-base main hook-context-panel-diet` → `7513283…` (== main) | 0 | <1s |
| 5 | `go build -o /tmp/centinela-v14 ./cmd/centinela` | 0 | 0.8s |
| 6 | `go build -o /tmp/centinela-main ./cmd/centinela` (cwd: primary worktree, on `main`) | 0 | ~9s |
| 7 | **`/tmp/centinela-v14 validate`** | **0** | **513s** |
| 8 | `python3 <scratch>/drive.py` — 17-state main-vs-v14 `hook context` differential | 0 | 21s |
| 9 | `/tmp/centinela-main hook context` and `/tmp/centinela-v14 hook context` (review-armed fixture) | 0 / 0 | <1s |
| 10 | `CENTINELA_PANEL_DIET_PAD=0 go test ./internal/ui -run TestPanelBudget -count=1` | 0 | 0.4s |
| 11 | `CENTINELA_PANEL_DIET_PAD=205 go test ./internal/ui -run TestPanelBudget -count=1` | 0 | 0.2s |
| 12 | `CENTINELA_PANEL_DIET_PAD=206 go test ./internal/ui -run TestPanelBudget -count=1` → `measured 1401, budget 1400` | 1 | 0.2s |
| 13 | `CENTINELA_PANEL_DIET_PAD=233 …` → `measured 1428` | 1 | 0.2s |
| 14 | `CENTINELA_PANEL_DIET_PAD=234 …` → `measured 1429` | 1 | 0.2s |
| 15 | `rsync -a --exclude .git <worktree>/ <scratch>/` then revert `trimTrailingWS` to identity in the copy | 0 | 3s |
| 16 | `go test ./internal/ui -run '…' -count=1` in the reverted copy → `measured 2004 bytes` + 4 other guards red | 1 | 1s |
| 17 | main-vs-v14 `cmp` of `status sample-feature`, `roadmap`, `doctor`, `workflows`, `--help`, `hook prewrite` | 0 | 2s |
| 18 | `/tmp/centinela-{main,v14} roadmap` / `status` / `workflows` / `roadmap --json` in the real worktree | identical | 3s |
| 19 | `CLICOLOR_FORCE=1 /tmp/centinela-v14 hook context` + ANSI-strip padding analysis | 0 | <1s |
| 20 | `/tmp/centinela-v14 hook {context,setup,migrate,autostart,orchestration,plan-advisor,merge}` | 0 (all) | 2s |
| 21 | python literal `// Scenario:` ↔ spec matcher over every `.go` in the repo → 11/11 exact | 0 | 1s |
| 22 | python full-scan file-size gate over `src internal cmd lib app pkg` → 1626 files, 0 violations | 0 | 2s |
| 23 | `/tmp/centinela-v14 audit` → `no baseline` (full scan done directly instead, #22) | 0 | <1s |
| 24 | `go test ./tests/acceptance/ -run TestPD -count=1 -v` → 11/11 PASS | 0 | 5s |
| 25 | `grep -rln CENTINELA_PANEL_DIET_PAD internal/ cmd/ \| grep -v _test.go` → empty (test-only confirmed) | 1 | <1s |
| 26 | `git check-ignore -v coverage.out .workflow/.roadmap-digest` → both ignored | 0 | <1s |
| 27 | `/tmp/centinela-v14 artifact new hook-context-panel-diet gatekeeper` | 0 | <1s |
| 28 | `/tmp/centinela-v14 evidence init hook-context-panel-diet gatekeeper` | 0 | <1s |

`centinela validate` was run **exactly once** (#7) and passed on the first attempt — the known
acceptance-tier flake did not occur, so there is no second attempt to record. No competing `go test` or
`centinela validate` process was observed. Verbatim output:

```
Built-in Gates (diff-aware: 23 files changed since main)
✓ G1: File Size  All in-scope files are within the 100-line cap (per-file exceptions are configurable under [[gates.file_size_exceptions]]).
✓ G-Build: Cross-Compile  All 6 release targets compile.
⚠ import_graph  Packages match no configured layer:
✓ spec-traceability-gate  All 11 scenarios have acceptance coverage.
✓ roadmap_drift  ROADMAP.md is in sync.
✓ docstring-gate  All 9 exported identifiers across 4 changed Go file(s) are documented.

Validate Commands
✓  go test ./... -coverprofile=coverage.out  go test (non-verbose) carries no skip data — add -json or -v to make skips detectable
✓  COVERAGE_PROFILE=coverage.out ./scripts/check-coverage.sh
✓  ./scripts/check-fmt.sh

 🛡️👁️  CLI  All gates passed.
```

The `⚠ import_graph` warning is pre-existing and ambient: this diff changes no package boundaries and no
layer configuration, and the warning is non-blocking (`validate` exited 0).

All scratch material was created outside the repository and deleted; `git status --porcelain=v1 -uall` is
empty immediately before the stamp.

#### Findings

**No CRITICAL findings. No WARNING findings. No blocking defect found.**

The core claim held under every attack I could construct. The two headline numbers (2004 → 1195) are not
self-reported: I reproduced both from source — the "after" via the shipped budget knob, the "before" by
reverting the fix in an out-of-tree copy. Content identity was verified across 17 project states plus a
colour-forced render, not asserted.

**LOW-1 — dangling backlog reference in shipped source.**
`internal/ui/render_cli_padding_test.go:32` tells the next engineer that cleaning up the remaining emitters
is tracked as `backlog: hook-panel-diet-remaining-emitters`. That slug is **not** in
`.workflow/roadmap.json` (181 features; only `token-diet` and `hook-context-panel-diet` match
`panel|diet`). The same slug is referenced in `docs/plans/hook-context-panel-diet.md:310` and in four
`.workflow/*.md` role artifacts. The pointer is real advice pointing at nothing. Non-blocking: a
traceability gap, not a correctness defect, and fixable in the docs step by deferring the slug to the
Backlog phase. Per this prompt's instruction I did **not** run `centinela roadmap defer` — it would land
the record on this branch and stale my own stamp. See Deferred Findings.

**INFO-1 — `internal/ui/render.go` is now exactly 100 lines**, at the G1 ceiling (99 on `main`). The gate
passes (`>100` fails), but the next single-line addition to that file trips G1.

**INFO-2 — the size-guard fixture is a conservative superset of any real turn.** `panelDietRender`
concatenates `RenderContextCapped` + `RenderReviewReady` + `RenderFeatureBriefNeeded`, but in the live hook
the last two are **mutually exclusive** for one feature at the plan step (the brief nudge requires
`ValidateArtifacts(plan) != nil`, the review nudge requires `== nil`). So 1195 B over-states a real
single-feature plan turn — the largest I measured live was 883 B. The guard therefore measures more than it
must, never less: the safe direction. The comment calling it "the three hook panels that dominate a
plan-step turn" is slightly generous phrasing for a deliberately synthetic fixture. Not a defect.

**INFO-3 — two acceptance scenarios assert less than their spec text says.** The spec says `centinela
status` and the blocked-write panel "should be identical to the pre-diet render"; the shipped tests
(`TestPDStatusOutputKeepsItsPadding`, `TestPDBlockedWritePanelKeepsItsPadding`) only assert that *some*
padding survives — they cannot detect a change that keeps padding while altering bytes. I closed that gap
myself by `cmp`-ing both surfaces against a binary built from `main`: **byte-identical** (status 368 B,
blocked-write 1216 B, identical stderr and exit code). The spec's claim is true; it is the test that is
weaker than the sentence it traces to. Non-blocking — a stronger test needs a golden fixture.

#### Deferred Findings

Listed as text only. `centinela roadmap defer` was deliberately NOT run, per the verifier prompt — it would
commit a roadmap record onto this branch and invalidate this report's own stamp.

1. **`hook-panel-diet-remaining-emitters`** — the follow-up the plan (§8) and
   `internal/ui/render_cli_padding_test.go:32` both already name, but which does not exist in
   `.workflow/roadmap.json`. Apply the same `trimTrailingWS` wrap to the padded panels emitted by the other
   `UserPromptSubmit` hooks and rare-event renderers: `hook setup` (measured **27 padded lines / 5533 B** on
   a partially-configured project), `hook migrate` (**4 padded lines / 300 B**), `RenderBlocked` /
   `RenderBlockedStaleBinary`, `hook merge`. Each is gated or rare rather than per-turn, which is why this
   feature correctly excluded them (plan decision D4) — but the token argument is identical whenever they
   fire. Creating this entry also repairs LOW-1.

2. **`hook-panel-diet-cli-render-golden`** — replace the "some padding still exists" assertions for
   `centinela status` and the blocked-write refusal (INFO-3) with byte-level golden fixtures, so the
   "identical to the pre-diet render" scenario is proved by the suite rather than by a manual two-binary
   `cmp` at review time.

3. **`ui-render-go-at-line-ceiling`** — `internal/ui/render.go` sits at exactly 100 lines (INFO-1). Either
   split `RenderContextCapped`/`stepBar` out now, or accept that the next edit to that file must include a
   split. Purely preventive.

#### Recommendation

**PROCEED — Status SAFE.**

This is the strongest verdict I am willing to give, and I give it because the falsification attempts were
mechanical rather than narrative: two binaries built from two revisions, 17 project states, a colour-forced
render, a deliberately reverted fix in an out-of-tree copy, and a binary search on the guard's own knob.
The feature does exactly what it claims — 809 bytes (40.4%) removed from the measurement fixture and ~41%
removed from real plan-step turns, with content provably unchanged and every CLI surface byte-identical to
`main`. The size guard is honest: I watched it go red for a synthetic overrun **and** for a genuine reverted
regression, and its test-only env knob is monotone in the failing direction, so it cannot weaken the check
in a real run. All 11 spec scenarios trace to acceptance tests by exact literal marker, the four
padding-must-survive tests correctly assert the opposite of the main change, `validate` passed on its first
and only run (exit 0, 513 s), and a full-scan file-size gate over 1626 files is clean.

The three findings are all LOW/INFO and none of them block. The dangling
`hook-panel-diet-remaining-emitters` backlog slug (LOW-1) should be created during the docs step; the other
two are recorded above as deferred follow-ups.

```json centinela:verification
{
  "revision": "c4a8067583ce0ee48d88acd6d9b7bcf86d50dd8b",
  "treeDigest": "sha256:96a296d224f285c67bee93c30f8a309157f0daa35dc5b87e410b78630a09cfc7",
  "commands": [
    {"argv": ["/tmp/centinela-v14", "validate"], "exitCode": 0, "durationMs": 513000},
    {"argv": ["go", "build", "-o", "/tmp/centinela-v14", "./cmd/centinela"], "exitCode": 0, "durationMs": 800},
    {"argv": ["go", "build", "-o", "/tmp/centinela-main", "./cmd/centinela"], "exitCode": 0, "durationMs": 9000},
    {"argv": ["go", "test", "./internal/ui", "-run", "TestPanelBudget", "-count=1"], "exitCode": 0, "durationMs": 400},
    {"argv": ["go", "test", "./internal/ui", "-run", "TestPanelBudget", "-count=1"], "exitCode": 1, "durationMs": 200},
    {"argv": ["go", "test", "./tests/acceptance/", "-run", "TestPD", "-count=1", "-v"], "exitCode": 0, "durationMs": 5000},
    {"argv": ["/tmp/centinela-v14", "audit"], "exitCode": 0, "durationMs": 300},
    {"argv": ["/tmp/centinela-v14", "artifact", "new", "hook-context-panel-diet", "gatekeeper"], "exitCode": 0, "durationMs": 200},
    {"argv": ["/tmp/centinela-v14", "evidence", "init", "hook-context-panel-diet", "gatekeeper"], "exitCode": 0, "durationMs": 200}
  ]
}
```
