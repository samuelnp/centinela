### Adversarial Verifier Report: evidence-schema-skeleton-legacy-handoff
**Date:** 2026-08-03
**Status:** WARNING

#### Inputs Read

- `git diff main...HEAD` (36 files) plus the uncommitted working tree
  (`specs/evidence-schema-skeleton-legacy-handoff.feature` + 3 files under
  `tests/acceptance/` — re-read; all four are scenario/comment additions only,
  no assertion or production-code change).
- `docs/features/evidence-schema-skeleton-legacy-handoff.md`
- `specs/evidence-schema-skeleton-legacy-handoff.feature`
- `docs/plans/evidence-schema-skeleton-legacy-handoff.md`
- Source read directly, not through any role narrative:
  `internal/worktreepath/path.go`, `internal/worktree/path.go`,
  `internal/evidence/schema_active.go`, `internal/evidence/repair.go`,
  `internal/evidence/schema_init.go`, `internal/evidence/validate.go`,
  `internal/workflow/handoff.go`, `internal/workflow/handoff_chain.go`,
  `internal/workflow/validate_orchestration.go`, `internal/workflow/active.go`,
  `internal/orchestration/policy.go`, `internal/orchestration/feature_surface.go`,
  `internal/gates/file_size.go`, `internal/gates/file_size_scan.go`,
  `cmd/centinela/evidence_schema.go`, `cmd/centinela/merge_validate.go`,
  `cmd/centinela/start.go`, `internal/worktree/start_hook.go`, `centinela.toml`.
- The output of every command listed under Commands Run, all executed by me.

The orchestrator's prompt did NOT contain a narrative summary of the
implementation — only the claim under test and a probe list. No `.workflow/*.md`
role narrative (planner / senior-engineer / qa-senior / edge-cases) was accepted
as evidence; every behavior below was re-derived by executing a binary built from
this tree.

#### Refutation Attempts

**Claim attacked:** the printed `handoffTo` is depth-independent inside a worktree.
**How:** built `/tmp/centinela-v6` from this tree and ran `evidence schema` from
the worktree root, `internal/evidence`, `.workflow`, `docs/architecture`, a
synthetic `a/b/c/d/e/f`, a symlink pointing at a worktree root, and a symlink
pointing at a nested package dir — in this real worktree and in a synthetic
canonical-internal fixture.
**Result:** COULD NOT REFUTE. All 4 depths in this worktree print
`feature=evidence-schema-skeleton-legacy-handoff, step=validate, handoffTo=complete`,
identical apart from timestamps. The synthetic fixture prints
`demo-internal / complete` from all 6 locations including both symlink routes.
The legacy value (`documentation-specialist`) never appeared.

**Claim attacked:** every archetype agrees with `evidence init` and with the gate.
**How:** built 5 fixtures — canonical user-facing (`surface: user-facing` brief),
canonical internal, hotfix (`code,tests,validate`), spike (`plan,code`), and a
legacy-pinned workflow (neither `validateContract` nor `planContract`) — and for
30 role/archetype pairs compared `evidence schema <role>` at the root against the
same at `sub/deep/deeper`, against the `handoffTo` that `evidence init` writes,
then ran `centinela evidence validate <feature>` over the populated evidence set.
**Result:** COULD NOT REFUTE. 30/30 depth-identical, 30/30 identical to
`evidence init`, and `evidence validate` reported ZERO `handoffTo` issues in all
five fixtures (residual non-zero exits are the outputs-content rules — "outputs
must include a real UI file" etc. — unrelated to this feature). Spot values that
would have been wrong under the legacy chain and are right here:
canonical-internal gatekeeper → `complete` (not `documentation-specialist`),
hotfix gatekeeper → `complete`, spike senior-engineer → `complete`,
user-facing senior-engineer → `ux-ui-specialist`, legacy-pinned qa-senior →
`validation-specialist`.

**Claim attacked:** anti-guess — it degrades to the placeholder, never to a
confident wrong answer, and never leaks a candidate feature name.
**How:** `.worktrees/not-a-real-feature/sub` with no workflow state; a worktree
whose `.workflow/<f>.json` is corrupt (`{not json`); one that is a zero-byte file;
zero active workflows; a `.workflow` dir with two active workflows; a directory
with no `.workflow` at all; the primary tree of this repo.
**Result:** COULD NOT REFUTE for these. Every one prints exactly
`"feature": "<feature-slug>"` and `"handoffTo": "<successor-role>"`, exit 0, and
the unverifiable slug is NOT echoed. Two degenerate-but-readable states (`{}`, and
a file whose inner `feature` disagrees with its filename) do derive a value — but
`workflow.Load` succeeds there, so `CheckHandoffTo` derives from the same bytes
and agrees; no divergence from the gate is possible.

**Claim attacked:** anti-guess holds for a path where `.worktrees` appears twice.
**How:** built a REAL git repo, `git worktree add .worktrees/outer-uf`, then from
inside that worktree `git worktree add .worktrees/inner-feat` — which is what
`centinela start` does, since `cmd/centinela/start.go:50` calls
`worktree.MaybeProvision(".", …)` with the invoking CWD as the repo and there is
no inside-a-worktree guard. Gave `outer-uf` a user-facing brief and `inner-feat`
an internal one, stood at the ROOT of `inner-feat`, ran
`evidence schema senior-engineer`.
**Result:** **REFUTED.** It printed `"feature": "outer-uf"` and
`"handoffTo": "ux-ui-specialist"`. The truth for the feature you are standing in
is `inner-feat` / `qa-senior` (confirmed by `evidence init inner-feat
senior-engineer` in the same directory). Pasting that value produced:
`evidence handoffTo for "senior-engineer" is "ux-ui-specialist", but this
workflow's contract makes "qa-senior" its successor` — the command emitted a
`handoffTo` the chain gate refuses, wearing a real feature name. See Findings.

**Claim attacked:** the placeholder is refused by the gate where it matters.
**How:** set `handoffTo` to the literal `<successor-role>` on a contract-REQUIRED
role (gatekeeper at validate, adversarial pin) and on a non-required role
(documentation-specialist at docs on an internal feature), then ran
`evidence validate`.
**Result:** COULD NOT REFUTE — behavior matches the documented contract and the
two new spec scenarios exactly: the required role is refused with
`fix with: centinela evidence set canon-int gatekeeper handoffTo complete`; the
non-required role passes silently (already deferred as
`handoff-gate-skips-nonrequired-role-evidence`).

**Claim attacked:** output contract — valid JSON on success, zero stdout bytes on
every error path.
**How:** all 11 parseable roles against a resolved feature, JSON-parsed;
`evidence schema bogus`, `evidence schema` (0 args), `evidence schema gatekeeper
extra` (2 args), `evidence schema --nope`, each with stdout and stderr captured
separately.
**Result:** COULD NOT REFUTE. 11/11 valid JSON with EMPTY stderr; all 4 error
paths exit 1 with **stdout = 0 bytes**. One pre-existing leak survives and is
already deferred (`workflow-warnings-break-json-stdout-capture`): a corrupt
workflow file in a non-worktree repo makes `ActiveWorkflows` write
`workflow warning: …` to stderr while stdout still holds valid placeholder JSON —
harmless unless an agent captures with `2>&1`.

**Claim attacked:** the `internal/worktree.DetectFeatureFromCwd` refactor is
behavior-preserving.
**How:** copied main's pre-refactor body and the new leaf body into one scratch
program and diffed their outputs over 23 inputs — empty string, `.`, `..`, `/`,
relative paths, trailing/duplicate separators, `..` climbing out, `.worktreesX`
and `x.worktrees` near-misses, doubled `.worktrees`, `.worktrees/.worktrees`, a
symlink and a symlinked deep path, and a 200-segment path.
**Result:** COULD NOT REFUTE — **0 divergences / 23**. The only textual change is
`abs, err := filepath.Abs` → `abs, _ :=`; on the error path `Abs` returns `""`,
whose scan finds no segment, so both return `"",""`. All 8 non-test callers
(hook_workflows, active_feature, roadmap_defer, hook_postwrite, verify, doctor
context, IsInsideWorktree, Path) are therefore unaffected, and their packages'
tests pass.

**Claim attacked:** `internal/worktreepath` is stdlib-only and `internal/evidence`
gained no domain edge.
**How:** `go list -deps ./internal/worktreepath` filtered for module packages;
`go list -f '{{range .Imports}}…'` on `./internal/evidence` compared against
main's `internal/evidence/repair.go` imports.
**Result:** COULD NOT REFUTE. `worktreepath`'s only module package is itself —
stdlib-only. `internal/evidence` imports `orchestration`, `workflow` (both
pre-existing) and the new `worktreepath` leaf; no new domain edge. The
`import_graph` gate's ⚠ is the pre-existing unmapped-package warning, unchanged.

**Claim attacked:** the CLI arity change (`ExactArgs(1)`) breaks existing prompt
invocations.
**How:** grepped every prompt in `docs/architecture/` and the
`internal/scaffold/assets/` mirror for `evidence schema`.
**Result:** COULD NOT REFUTE — every invocation line is already
`centinela evidence schema <role>` (role-only). No caller passes a feature
argument. (3 mirror files — gatekeepers.md, new-project-guide.md,
testing-strategy.md — differ from `docs/architecture/`, but they already differ
on `main` and this branch touches neither side.)

**Claim attacked:** the file-size gate passes only because validate ran diff-aware.
**How:** re-implemented the gate's own scan rules (roots `src internal cmd lib app
pkg`, its 16 source extensions, its dot/vendor dir skips, >100 lines) and walked
the whole tree.
**Result:** COULD NOT REFUTE — 1559 files scanned, ZERO over 100 lines. No
`[[gates.file_size_exceptions]]` are configured, so nothing is being excused.

#### Commands Run

All from the worktree root
`/Users/samuelnp/projects/personal/centinela/.worktrees/evidence-schema-skeleton-legacy-handoff`
unless a probe fixture path is named. Exit codes captured directly
(`; echo EXIT=$?`), never inferred from banner text.

| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `go build -o /tmp/centinela-v6 ./cmd/centinela` | 0 | 370 ms |
| 2 | `/tmp/centinela-v6 validate` (backgrounded, polled to completion) | **0** | **448 s** |
| 3 | `go test ./internal/worktreepath/... ./internal/evidence/... ./cmd/centinela/... -count=1` | 0 | 12.99 s |
| 4 | `/tmp/centinela-v6 evidence schema gatekeeper` (worktree root) | 0 | 57 ms |
| 5 | `/tmp/centinela-v6 evidence schema bogus` | 1 | 33 ms |
| 6 | `go list -deps ./internal/worktreepath` | 0 | 68 ms |
| 7 | full-scan file-size sweep (gate rules re-implemented, 1559 files) | 0 | ~1 s |
| 8 | `evidence schema <role>` × 6 CWD depths × 2 fixtures (incl. 2 symlink routes) | 0 each | <1 s each |
| 9 | `evidence schema` × 30 role/archetype pairs × 2 depths + `evidence init` + `evidence validate` per fixture | mixed (see Refutation) | ~25 s total |
| 10 | `evidence schema` × 4 error paths, stdout/stderr captured separately | 1 each | <1 s each |
| 11 | `evidence schema` × 11 roles, JSON-parsed | 0 each | <1 s each |
| 12 | real `git worktree add` nested repro + `evidence init`/`set`/`validate` | 1 (gate refusal, as reported) | ~3 s |
| 13 | old-vs-new `DetectFeature` differential, 23 inputs (`go run`) | 0 | ~1 s |
| 14 | `/tmp/centinela-v6 artifact new … gatekeeper` then `evidence init … gatekeeper` | 0 / 0 | <1 s |

The single `centinela validate` run (command 2) executed every
`[validate] commands` entry, including `go test ./... -coverprofile=coverage.out`,
which IS the full suite (`tests/acceptance` is under `./...`). Per centinela.toml
that one profiled run is the mandated suite run, so no second full-suite execution
was performed. **No `test timed out after 10m0s` panic occurred** — the run
completed in 448 s on the FIRST attempt; no re-run was needed.

Command 2 output (tail, verbatim): `✓ G1: File Size`, `✓ G-Build: Cross-Compile`,
`⚠ import_graph  Packages match no configured layer:` (pre-existing),
`✓ spec-traceability-gate  All 12 scenarios have acceptance coverage.`,
`✓ roadmap_drift`, `✓ docstring-gate`, then all three validate commands `✓`, then
`🛡️👁️ CLI  All gates passed.`

Tree changes made AFTER command 2, by me, while authoring this report: this
report's `.md` + its evidence `.json`, the new deferral in
`.workflow/roadmap.json`, and the matching `centinela roadmap generate`
regeneration of `ROADMAP.md` (run so the `roadmap_drift` gate stays in sync —
verified by byte-comparing the file before and after). No Go source, spec, or
test file was touched after the run, so command 2's green result remains
representative of the stamped tree.

Coverage of the changed units, read from the profile that same run wrote:
`worktreepath.DetectFeature` 100%, `evidence.ResolveActiveFeature` 100%,
`evidence.SchemaSkeleton` 100%, `evidence.hasWorkflowState` 100%,
`worktree.DetectFeatureFromCwd` 100%, `cmd.runEvidenceSchema` 89.5% (uncovered
arms are the `os.Getwd` / `inDir` chdir failures).

#### Findings

- **Affected spec:** `specs/evidence-schema-skeleton-legacy-handoff.feature`
  **Affected scenario:** "A resolved slug with no readable workflow state is not
  guessed at" (its anti-guess intent) and "Derivation is identical from a
  subdirectory of the worktree" (its depth-independence intent).
  **Risk:** With nested worktrees the scan takes the OUTERMOST `.worktrees/<x>`
  segment, so anywhere inside `.worktrees/A/.worktrees/B` the command prints
  feature `A` and A's derived `handoffTo`. Reproduced end to end in a real git
  repo: at the ROOT of the inner worktree it printed
  `feature=outer-uf, handoffTo=ux-ui-specialist`, and pasting that into
  `inner-feat`'s evidence is refused by the gate with
  `…contract makes "qa-senior" its successor`. That is precisely the failure mode
  this feature exists to remove, now carrying a plausible feature name instead of
  an obvious placeholder. The existing deferral
  `nested-worktrees-resolve-outermost` calls this "harmless today because the
  resolved slug must now have real workflow state" — that rationale is **false**:
  in a real nested worktree the outer feature HAS real workflow state, so
  `hasWorkflowState` does not catch it. Reachability is not hypothetical:
  `cmd/centinela/start.go:50` calls `worktree.MaybeProvision(".", feature, cfg)`
  with the invoking CWD as the repo root and has no inside-a-worktree guard, so
  `centinela start <next>` run from within a worktree (the documented way of
  working here) creates the nested layout. Severity WARNING, not CRITICAL: it
  needs that operator sequence, the wrong value is refused loudly rather than
  silently accepted, and the scan's outermost-first behavior is unchanged from
  `main` (0/23 divergences) — an inherited quirk newly made consequential, not a
  regression introduced by this diff.
  **Suggestion:** make `worktreepath.DetectFeature` scan from the RIGHT (innermost
  segment wins) — a loop-direction change that leaves all 23 differential inputs
  except the nested ones unchanged — and/or refuse to provision a worktree from
  inside one in `MaybeProvision`. Re-scope the `nested-worktrees-resolve-outermost`
  deferral, whose stated justification this report disproves.

- **Affected spec:** `specs/evidence-schema-skeleton-legacy-handoff.feature`
  **Affected scenario:** "Derivation is identical from a subdirectory of the
  worktree" (scope note, not a defect).
  **Risk:** Depth independence holds ONLY for the worktree signal. In a repo with
  `use_worktrees = false`, one active workflow resolves at the repo root but
  degrades to the placeholder from any subdirectory, because
  `workflow.ActiveWorkflows` is CWD-relative and signal 2 returns no root.
  Verified: `<fixture>/one` → `solo` / `qa-senior`; `<fixture>/one/pkg/deep` →
  `<feature-slug>` / `<successor-role>`. This is a SAFE degradation (never a wrong
  value) and `--help` claims depth independence only for the worktree branch, so
  it is documented behavior — but agents in a non-worktree project will meet the
  placeholder more often than the feature's framing suggests.
  **Suggestion:** either walk upward for `.workflow/` in signal 2, or state the
  non-worktree limitation explicitly in the command's `Long` help.

- **Affected spec:** n/a (process observation)
  **Affected scenario:** n/a
  **Risk:** At verification time the branch carries 4 uncommitted files — the spec
  and three acceptance tests. I re-read all four: they add scenarios and
  `// Scenario:` traceability comments only; no assertion, no production code.
  They are what makes `spec-traceability-gate` report 12/12, and they were present
  in the tree command 2 tested, so the green run covers them. They are not yet in
  `git diff main...HEAD`.
  **Suggestion:** none beyond letting `centinela complete` commit them; noted so
  the stamped tree digest is not mistaken for the committed diff.

#### Deferred Findings

Already recorded in `.workflow/roadmap.json` under this feature (verified by
reading the file, not by trusting a role report):

- `nested-worktrees-resolve-outermost` — present, but its "harmless today"
  rationale is disproved above; it should be re-scoped, not closed.
- `prewrite-hook-cwd-relative-workflow-lookup`
- `workflow-warnings-break-json-stdout-capture`
- `handoff-gate-skips-nonrequired-role-evidence`
- `evidence-active-workflow-corrupt-json-ambiguity`
- `acceptance-tier-exceeds-default-test-timeout`

Filed by this report:

- `start-inside-worktree-creates-nested-worktree` — `centinela start` provisions
  with `repo = "."`, so run from inside a worktree it creates
  `.worktrees/A/.worktrees/B`; that nested layout is what makes
  `nested-worktrees-resolve-outermost` reachable with real workflow state on both
  levels.

#### Recommendation

**WARNING — may proceed.** Every headline claim holds in every supported
topology: depth independence (6 CWDs × 2 fixtures plus this worktree's 4 depths,
symlinks included), gate agreement across 30 role/archetype pairs with zero
`handoffTo` issues, placeholder-not-guess on 7 unresolvable states, a clean output
contract (11/11 valid JSON, 4/4 error paths with 0 stdout bytes), a
behavior-preserving refactor (0/23 divergences), a genuinely stdlib-only leaf, and
`centinela validate` green in 448 s with an independent full-scan file-size sweep
also clean. One reachable topology — nested worktrees, created by running
`centinela start` from inside a worktree — still prints a confident wrong feature
and a `handoffTo` the gate refuses; it is disclosed and deferred, is not a
regression from `main`, and fails loudly rather than silently, so it does not
block completion. Fix the scan direction (innermost wins) before this feature's
guarantee is quoted as unconditional.

```json centinela:verification
{
  "revision": "c8f6f70db92b1c2a3460a2583cd99f01c76aaf31",
  "treeDigest": "sha256:a626746a373570b4b1bcd52208afd67db4d100de0793b0b17b92a52cc67e2643",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-v6", "./cmd/centinela"], "exitCode": 0, "durationMs": 370},
    {"argv": ["/tmp/centinela-v6", "validate"], "exitCode": 0, "durationMs": 448000},
    {"argv": ["go", "test", "./internal/worktreepath/...", "./internal/evidence/...", "./cmd/centinela/...", "-count=1"], "exitCode": 0, "durationMs": 12985},
    {"argv": ["/tmp/centinela-v6", "evidence", "schema", "gatekeeper"], "exitCode": 0, "durationMs": 57},
    {"argv": ["/tmp/centinela-v6", "evidence", "schema", "bogus"], "exitCode": 1, "durationMs": 33},
    {"argv": ["go", "list", "-deps", "./internal/worktreepath"], "exitCode": 0, "durationMs": 68}
  ]
}
```
