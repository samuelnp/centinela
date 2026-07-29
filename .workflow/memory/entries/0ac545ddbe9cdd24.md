---
id: 0ac545ddbe9cdd24
feature: merge-truthful-delivery
step: validate
type: verdict
title: ### Adversarial Verifier Report: merge-truthful-delivery
tags: gatekeeper, verdict
sourceArtifact: .workflow/merge-truthful-delivery-gatekeeper.md
createdAt: 2026-07-29T08:49:05Z
---

### Adversarial Verifier Report: merge-truthful-delivery
**Date:** 2026-07-29
**Status:** WARNING

#### Inputs Read
- `git diff origin/main...HEAD` (36 files, +2442/−13) and the uncommitted tree
  (3 commits: plan/code/tests; untracked new source + test files listed by
  `git status`).
- `docs/features/merge-truthful-delivery.md` — the contract.
- `specs/merge-truthful-delivery.feature` — **21** scenarios.
- `docs/plans/merge-truthful-delivery.md` — the plan.
- Source read to attack, not as narrative: `cmd/centinela/{merge,merge_report,
  merge_continue,merge_dispatch,merge_validate,hook_merge,deliver,validate}.go`,
  `internal/worktree/{primary,merge_verify,merger,merger_git,registry,remove,
  merge_finish,merge_pending,merge_pending_directive,finalize,
  finalize_escalation}.go`, `internal/gates/spec_traceability*.go`, plus
  `git show origin/main:...` for the three files whose behaviour I compared
  against the pre-fix baseline.
- Output of every command I executed myself (see Commands Run), including 14
  timed end-to-end runs of a scratch binary in throwaway git repositories with
  no network and no remotes.

Explicitly **not** read as evidence: the `senior-engineer`, `qa-senior`,
`validation-specialist`, `big-thinker` and `feature-specialist` narratives. The
prior gatekeeper report was skimmed for one purpose only — to enumerate which
claims a previous pass had refuted, so I could re-attack them on the fixed tree.
Nothing in it was accepted as evidence; every verdict below comes from a command
I ran in this session. The evidence JSON was re-initialised before this file was
written, so no prior verdict survives in it.

**Prompt-narrative flag:** the task prompt asserts that fixes have been applied
and lists the specific defects to re-attack. I treated that list as a set of
claims to falsify, not as a statement that they are fixed. The `origin/main`
baseline for `dispatchSteward`/`runMergeContinue` was read directly from git to
confirm what actually changed.

#### Refutation Attempts

**Claim attacked:** "The original defect is fixed — `deliver --via merge` from
inside a worktree verifiably advances main and removes the worktree."
**How:** Throwaway repo, `.worktrees/high-score` with a commit main lacks, a
hand-written `.workflow/high-score.json` carrying `worktreePath`; ran
`deliver high-score --via merge` **from inside the worktree**, then asserted on
the ref, on `merge-base --is-ancestor`, and on `git worktree list`.
**Result:** could not refute. Exit 0; main `076dbdf → 9ad4454`; branch is an
ancestor; `.worktrees/high-score` gone from disk **and** from the registry. The
same shape via `merge` (not `deliver`) also passes, including with a repo path
containing a space and with the primary branch renamed to `develop` (wording
follows `TargetBranch`: `Merged "high-score" into develop`).

**Claim attacked:** "`merge --continue` never prints a success line while the
target ref is unmoved" (the previously-refuted defect #1).
**How:** Induced a **real** text conflict, stalled the merge from
`.worktrees/high-score`, `git merge --abort`ed so the primary tree was clean and
the branch was NOT an ancestor, then wrote schema-valid merge-steward evidence
with `handoffTo: complete` (an APPLY verdict) and ran `merge --continue` from the
worktree CWD.
**Result:** could not refute. Exit 1,
`merge of "high-score" was not completed … "high-score" is not an ancestor of
main`; no success line; main byte-identical; worktree still present. The APPLY
verdict is treated as a claim, ancestry as the proof.

**Claim attacked:** "A worktree-initiated stall is resumable from BOTH CWDs."
**How:** Two independent repos. In each: real text conflict, stall started from
`.worktrees/high-score`, steward resolution committed in the primary tree, valid
APPLY evidence. Resumed once from the worktree CWD, once from the primary CWD.
**Result:** could not refute. Both exit 0 with `Merged "high-score" into main and
removed its worktree.`, ancestry true, worktree gone from disk and registry,
pending marker cleared. `centinela hook merge` also finds the marker from either
CWD and goes silent once valid evidence exists.

**Claim attacked:** "`verifyRemoved` consults git's registry, not the
`.worktrees/<feature>` path convention" (previously-refuted defect #2).
**How:** `git worktree move .worktrees/high-score /tmp/mtd-r1-elsewhere`, then
merged from the primary tree. Repeated with an untracked file inside the
relocated worktree so removal had to fail.
**Result:** could not refute for the supported shape — the registered path is
what gets removed (`/private/tmp/mtd-r1-elsewhere` deleted, registry empty), and
when it is busy the failure names the real path and refuses to claim removal.
**But see Finding 2:** a worktree moved *out-of-band* (`mv`, not
`git worktree move`) leaves a `prunable` registry record that
`findRegisteredWorktree` deliberately skips — and removal is then claimed while
`git worktree list` still lists it.

**Claim attacked:** "Removal is refused when the registry cannot be read."
**How:** Code path audit plus behaviour: `removeTarget` swallows the registry
error (`err == nil && ok`) and falls back to the convention or a no-op, but
`verifyRemoved` re-reads the registry and returns
`cannot verify worktree removal for %q` on any read error, so the composite is
fail-closed.
**Result:** could not refute — no ordering makes a read failure produce success.

**Claim attacked:** "The busy-worktree half-success is truthful, idempotent and
recoverable" (previously-refuted defect #3).
**How:** Untracked file in `.worktrees/high-score`; ran `merge` (run 1), plain
re-run (run 2), then `--force-remove` (run 3).
**Result:** could not refute. Run 1 exit 1, verbatim: `"high-score" merged into
main — verified; worktree removal failed: … ; re-run `centinela merge high-score
--force-remove` to retry removal`. Run 2 exit 1 with the same two-part shape,
correctly re-worded to `already merged into main — verified`. Run 3 exit 0,
worktree gone from disk and registry.

**Claim attacked:** "The post-merge validate gates the merged result rather than
being a vacuous diff-aware no-op."
**How:** Put a 150-line `src/big.go` on the feature branch, merged from the
worktree CWD.
**Result:** could not refute. Header reads `Built-in Gates (full scan)` (not
`diff-aware: 0 files changed`), `✗ G1: File Size` fires, the merge stops on the
validate-fail outcome, the worktree is kept and no success line is printed. The
portal regen likewise writes `docs/project-docs/index.html` into the **primary**
tree (9.2 KB) when its inputs exist there. **But see Finding 4:** the merge
commit created before validate ran is left on main and nothing says so.

**Claim attacked:** "Every refusal path refuses truthfully and moves nothing."
**How:** Ran `merge high-score` from a non-repo directory; from a linked worktree
of a **bare** repository; with the primary tree in **detached HEAD**; with the
primary tree holding the **feature branch itself**; with a **dirty** primary
tree.
**Result:** could not refute. `cannot resolve primary working tree: git worktree
list failed: fatal: not a git repository`; `cannot resolve primary working tree:
primary working tree is bare`; `primary working tree … has a detached HEAD`;
`… has "high-score" checked out — cannot merge a branch into itself`; `main
working tree is dirty …`. All exit 1, main unchanged, worktree kept, no success
line in any of them.

**Claim attacked:** "Re-running an already-delivered feature stays honest."
**How:** Pre-merged the branch, then ran `merge` with the worktree present, then
again with it already gone.
**Result:** could not refute. Both exit 0 with `Branch "high-score" was already
merged into main — worktree cleaned up, nothing new delivered.` — never the
fabricated `Merged … into main` — and main is byte-identical.

**Claim attacked:** "Every scenario in the spec has a real executing acceptance
marker."
**How:** Extracted all 21 `Scenario:` names from the spec and every
`// Scenario:` comment sitting under an
`// Acceptance: specs/merge-truthful-delivery.feature` header in
`tests/acceptance/*.go`, then normalised and diffed both sets — the same keying
`internal/gates/spec_traceability_match.go` uses. Cross-checked against the
`spec-traceability-gate` line in my own `centinela validate` run.
**Result:** **REFUTED.** 15 of 21 covered; 6 uncovered. `centinela validate`
independently reports `⚠ spec-traceability-gate  Scenarios without acceptance
coverage:` and only passes because the gate is configured `severity = "warn"`.
See Finding 1.

**Claim attacked:** "A stall started from a worktree is always resumable in a
real Centinela-governed repo."
**How:** Rebuilt the fixture with a `.gitignore` mirroring **Centinela's own**
policy (`git check-ignore` proves `.workflow/*-merge-pending.json` and
`.workflow/*-merge-steward.md` are NOT ignored here, only
`.workflow/*-merge-steward.json` is), stalled a merge, wrote steward evidence,
and ran `merge --continue` from both CWDs.
**Result:** **REFUTED.** Both refuse with `main working tree is dirty — commit or
stash before continuing "high-score"`, because the marker and the steward report
this feature now writes into the **primary** tree are untracked there. The two
"resumable from … CWD" scenarios pass in CI only because
`tests/acceptance/merge_steward_auto_dispatch_helper_test.go:46` gitignores the
whole `.workflow/` directory. See Finding 3.

**Claim attacked:** "The truthful two-part outcome always names a retry command
that works."
**How:** `git worktree lock .worktrees/high-score`, merged, then followed the
tool's own advice and re-ran with `--force-remove`.
**Result:** **REFUTED (low severity).** Both runs exit 1; `--force-remove` maps
to `git worktree remove --force`, which still refuses a locked worktree
(`use 'remove -f -f' to override or unlock first`). The message is still truthful
about both halves and echoes git's own guidance, but the command Centinela tells
the operator to run cannot succeed. See Finding 5.

**Claim attacked:** "The gates are green because they were run on this tree."
**How:** Ran `centinela validate` once myself (exit 0, 602 s) and `go test ./...`
myself (exit 0, 3606 tests in 45 packages) at revision
`0ba4381ef3174f0ae7effb15bc52bc815f5bf1e6`, and re-derived the two `⚠` gate lines
from the gate source rather than trusting the summary line.
**Result:** could not refute the pass, but `All gates passed.` coexists with two
warn-severity gate warnings, one of which is this feature's own traceability gap.

#### Commands Run
| # | argv | exit | duration |
|---|------|------|----------|
| 1 | `centinela validate` | 0 | 602 000 ms |
| 2 | `go test ./...` | 0 | 304 000 ms |
| 3 | `go build -o /tmp/centinela-verify-mtd2 ./cmd/centinela` | 0 | ~4 s |
| 4 | `centinela-verify-mtd2 merge high-score` (from `.worktrees/high-score`) | 0 | 196 ms |
| 5 | `centinela-verify-mtd2 merge high-score` (real text conflict, from worktree) | 1 | 109 ms |
| 6 | `centinela-verify-mtd2 merge --continue high-score` (from `.worktrees/high-score`) | 0 | 129 ms |
| 7 | `centinela-verify-mtd2 merge --continue high-score` (from primary CWD) | 0 | 122 ms |
| 8 | `centinela-verify-mtd2 merge --continue high-score` (APPLY evidence, branch NOT ancestor) | 1 | 67 ms |
| 9 | `centinela-verify-mtd2 merge high-score` (busy worktree, half-success) | 1 | 148 ms |
| 10 | `centinela-verify-mtd2 merge high-score` (plain re-run, idempotent two-part outcome) | 1 | 137 ms |
| 11 | `centinela-verify-mtd2 merge high-score --force-remove` (recovery) | 0 | 144 ms |
| 12 | `centinela-verify-mtd2 merge high-score` (worktree registered outside `.worktrees/<f>`) | 0 | 148 ms |
| 13 | `centinela-verify-mtd2 merge high-score` (worktree moved out-of-band → prunable record) | 0 | 132 ms |
| 14 | `centinela-verify-mtd2 merge high-score` (detached HEAD in primary) | 1 | 56 ms |
| 15 | `centinela-verify-mtd2 merge high-score` (bare primary) | 1 | 29 ms |
| 16 | `centinela-verify-mtd2 merge high-score` (not a git repository) | 1 | 28 ms |
| 17 | `centinela-verify-mtd2 merge --continue high-score` (repo tracks `.workflow/`) | 1 | 46 ms |

Additional untimed probes run in the same session (exit codes observed):
`deliver high-score --via merge` from a worktree (0); `merge` with the primary on
`develop` (0); already-merged re-run with the worktree present (0) and already
gone (0); dirty primary from a worktree CWD (1); feature branch checked out in
primary (1); validate-failure merge from a worktree CWD (1); 150-line-file
merge-time G1 failure (1); portal regen into the primary tree (0); locked
worktree with and without `--force-remove` (1, 1); repo path containing a space
(0); `hook merge` from both CWDs and outside any repo (0, 0, 0);
`evidence validate merge-truthful-delivery` (0);
`git check-ignore -v .workflow/x-merge-pending.json` (1 = not ignored),
`.workflow/x-merge-steward.md` (1 = not ignored),
`.workflow/x-merge-steward.json` (0 = ignored);
`git show origin/main:cmd/centinela/{merge_dispatch,merge_continue}.go`,
`git show origin/main:internal/worktree/finalize.go` (0).

#### Findings

**Finding 1 — 6 of the feature's own 21 scenarios have no acceptance coverage, and the spec asserts otherwise (WARNING)**
- **Affected spec:** `specs/merge-truthful-delivery.feature` (header comment,
  line 15: "Scenario titles map 1:1 to Go acceptance tests (`// Scenario: <name>`)").
- **Affected scenario:** `a text conflict still dispatches the Merge Steward with
  no success claim`; `a validate failure after a clean text merge still
  dispatches the steward`; `merge refuses when the primary tree is in detached
  HEAD state`; `merge refuses when the primary working tree is bare`; `removal is
  only claimed when the worktree directory is actually gone`; `the success
  message is never printed when the ref did not advance`.
- **Risk:** the spec's own 1:1 claim is false, and `centinela validate` reports
  `⚠ spec-traceability-gate  Scenarios without acceptance coverage:` on this very
  tree. It passes only because `centinela.toml` sets
  `[gates.spec_traceability] severity = "warn"` — the gate is doing its job and
  is being ignored. The underlying behaviours are not unverified (I reproduced
  four of the six end-to-end myself, and all six have colocated unit or
  integration coverage in `internal/worktree/{merger_guards,merge_verify,
  merge_verify_more,merger_verify_fail,finalize_detached,merge_realgit_more}_test.go`
  and `tests/integration/`), so this is a traceability defect, not a correctness
  one — but the artifact that ships claims a mapping it does not have, which is
  the same class of untruth the feature exists to eliminate.
- **Suggestion:** add the six `// Scenario:` markers to the acceptance files that
  already drive these paths (`merge_truthful_delivery_refusals_test.go` covers
  the bare/detached shapes trivially with the binary; the two steward-dispatch
  scenarios are already driven by `mtdcStall`), or delete the scenarios from the
  spec if they are genuinely lower-tier concerns. Do not close this by relaxing
  the gate.

**Finding 2 — "removed its worktree" is claimed while `git worktree list` still lists it (WARNING)**
- **Affected spec:** `specs/merge-truthful-delivery.feature`
- **Affected scenario:** `removal is verified against git's worktree registry,
  not a path convention` ("removal is never claimed while `git worktree list`
  still lists it") and, by contradiction, `removal is only claimed when the
  worktree directory is actually gone`.
- **Risk:** `findRegisteredWorktree` (`internal/worktree/registry.go:37`) skips
  any porcelain block carrying `prunable`, on the stated premise that such a
  record is "not a worktree on disk". That premise is falsifiable: move the
  worktree directory out-of-band (`mv .worktrees /tmp/elsewhere`) and git emits
  `prunable gitdir file points to non-existent location` for a worktree whose
  files are very much still on disk. Reproduced end-to-end: `merge high-score`
  exits 0 printing `Merged "high-score" into main and removed its worktree.`
  while `git worktree list` still shows
  `/private/tmp/mtd-pd/.worktrees/high-score … [high-score] prunable` and
  `/tmp/mtd-pd-wt/high-score` still contains the checkout. `Remove` also no-ops
  here (`removeTarget` returns `""`), so nothing was removed at all. This is a
  fail-**open** false success — exactly the defect class the feature targets —
  though it requires a hand-corrupted repo state, not a supported operation
  (`git worktree move` is handled correctly and was verified).
- **Suggestion:** treat a `prunable` block for the feature branch as "still
  registered" and refuse the removal claim, or at minimum surface it
  (`git worktree prune` first, then re-read the registry, and require the
  post-prune registry to be clean before claiming removal).

**Finding 3 — in repos that track `.workflow/`, `merge --continue` is refused from BOTH CWDs (WARNING)**
- **Affected spec:** `specs/merge-truthful-delivery.feature`
- **Affected scenario:** `a merge stalled from a worktree CWD is resumable from
  that same worktree CWD`; `a merge stalled from a worktree CWD is resumable from
  the primary CWD`.
- **Risk:** this feature deliberately moved `WritePending`/steward-evidence
  lookup from the invoking CWD into the **primary** tree. In Centinela's own
  repository `git check-ignore` shows `.workflow/<f>-merge-pending.json` and
  `.workflow/<f>-merge-steward.md` are **not** ignored (only
  `.workflow/*-merge-steward.json` is, `.gitignore:48`). So the act of stalling a
  merge now dirties the primary tree with untracked files, and `ResolveMerge`'s
  `isDirty(repo)` guard then refuses the resume from both CWDs with
  `main working tree is dirty — commit or stash before continuing`. Reproduced
  end-to-end on a validate-fail stall (where the tree would otherwise be clean):
  exit 1 from the worktree CWD and from the primary CWD alike. The acceptance
  tests do not catch this because their fixture
  (`tests/acceptance/merge_steward_auto_dispatch_helper_test.go:46`) writes
  `.gitignore` = `.worktrees/` + `.workflow/`, ignoring the entire directory — a
  fixture strictly more permissive than the framework's own policy. The message
  is actionable and the failure is fail-closed, so this is friction, not a false
  success; but the two scenarios are green on a fixture that does not represent
  the shipping repo shape.
- **Suggestion:** either exclude the merge bookkeeping paths from the dirty check
  (`git status --porcelain -- . ':(exclude).workflow/*-merge-pending.json'
  ':(exclude).workflow/*-merge-steward.*'`), or add
  `.workflow/*-merge-pending.json` to the managed ignore set, and change the
  acceptance fixture to mirror Centinela's real `.gitignore` so the scenario
  proves something.

**Finding 4 — a validate failure leaves the merge commit on main, unannounced, and `--continue` ships it (WARNING)**
- **Affected spec:** `specs/merge-truthful-delivery.feature`
- **Affected scenario:** `a validate failure after a clean text merge still
  dispatches the steward` (the stall is correct; what it leaves behind is not
  reported) and the Feature preamble's "success is asserted on the ref".
- **Risk:** `Merge` runs `git merge --no-ff` **before** `run(repo)`, so by the
  time validate fails the merge commit already exists on main in the primary
  tree. Reproduced: main `61d4990 → 1641da9 Merge branch 'high-score'` while the
  CLI printed only `Merge Steward required — post-merge-validate-failed`. An
  operator or agent reading "the merge stalled" will reasonably believe main is
  unmoved; it is not, and it now fails `centinela validate`. Worse, writing an
  APPLY verdict and running `merge --continue` closes it: exit 0,
  `Merged "high-score" into main and removed its worktree.`, **no revalidation at
  all** (`ResolveMerge` never calls the validate runner), and the worktree is
  destroyed — leaving a main that still fails `centinela validate` (verified:
  exit 1) with no worktree to recover from. `ResolveMerge` did not revalidate on
  `origin/main` either, so the "no revalidation" half is pre-existing and the
  brief explicitly scopes steward dispatch logic Out; but this feature is what
  makes the whole path reachable from the documented worktree CWD, and the
  silence about the landed merge commit is a truthfulness gap inside a
  truthfulness feature.
- **Suggestion:** state the landed merge in the stall message ("main already
  advanced to <sha>; `git reset --hard <base>` to unwind"), and re-run the
  validate runner inside `ResolveMerge` before finalising a stall whose recorded
  `Reason` is `post-merge-validate-failed`.

**Finding 5 — the `--force-remove` retry advice cannot succeed on a locked worktree (low)**
- **Affected spec:** `specs/merge-truthful-delivery.feature`
- **Affected scenario:** `a merge that lands but cannot remove the worktree
  reports both halves and stays recoverable` ("stderr names the command to re-run
  to retry removal").
- **Risk:** `WithForceRemove` appends a single `--force`. For a worktree under
  `git worktree lock`, `git worktree remove --force` still refuses
  (`use 'remove -f -f' to override or unlock first`), so the retry the tool names
  loops forever. Verified: both the plain run and the `--force-remove` run exit 1
  with identical two-part messages. The scenario's `Given` is an untracked file,
  which does work, so this is outside the letter of the spec; and git's own
  guidance is echoed verbatim in the error, so the operator is not left blind.
- **Suggestion:** detect `cannot remove a locked working tree` and name
  `git worktree unlock <path>` in the retry hint instead of `--force-remove`.

#### Deferred Findings
Executed:
- `centinela roadmap defer merge-prunable-registry-false-removal --summary "merge claims 'removed its worktree' while git worktree list still lists the branch's worktree as prunable and the directory survives elsewhere" --source merge-truthful-delivery/gatekeeper` → Backlog
- `centinela roadmap defer merge-pending-marker-dirties-primary-tree --summary "the pending marker and steward .md are untracked in repos that track .workflow/, so isDirty(primary) refuses merge --continue from both CWDs" --source merge-truthful-delivery/gatekeeper` → Backlog
- `centinela roadmap defer merge-validate-fail-leaves-main-advanced --summary "a post-merge validate failure leaves the merge commit on main unannounced and merge --continue closes it without re-validating" --source merge-truthful-delivery/gatekeeper` → Backlog
- `centinela roadmap defer merge-force-remove-locked-worktree --summary "the --force-remove retry advice does not work for a locked worktree (git needs -f -f or unlock)" --source merge-truthful-delivery/gatekeeper` → Backlog

Finding 1 is **not** deferred: it is a gap in this feature's own deliverable and
should be closed before the docs step.

#### Recommendation
**WARNING — the completion claim is substantially true but not fully true.**

The defect this feature exists to kill is dead, and I could not resurrect it.
Fourteen timed end-to-end runs of a scratch binary against real git repositories
confirm: `deliver --via merge` and `merge` from inside a worktree advance the
primary tree's ref, prove ancestry, and remove the worktree from disk and from
git's registry before any success line; every refusal path (non-repo, bare,
detached, self-merge, dirty) refuses truthfully and moves nothing; the
already-merged path is honest and idempotent; the busy-worktree half-success is
truthful, idempotent and recoverable; `merge --continue` resumes a
worktree-initiated stall from either CWD and refuses an unbacked APPLY verdict
with ancestry as the proof; the post-merge validate is a genuine full scan of the
merged primary tree that really fails on branch content. `centinela validate`
exits 0 and `go test ./...` exits 0 (3606 tests) at
`0ba4381ef3174f0ae7effb15bc52bc815f5bf1e6`. All three previously-refuted claims
survived re-attack.

What blocks a SAFE verdict is that the deliverable overstates itself in one place
and under-tests itself in another. The spec asserts a 1:1 scenario→acceptance
mapping that holds for only 15 of 21 scenarios, and the project's own
traceability gate says so out loud on this tree — it passes only because it is
configured to warn (Finding 1). Two of the scenarios that *are* marked are green
against a fixture whose `.gitignore` is more permissive than Centinela's own, and
under the real policy both fail (Finding 3). And two fail-open shapes remain: a
prunable registry record yields a false "removed its worktree" (Finding 2), and a
validate-failed merge is left sitting on main unannounced and can be shipped by
`--continue` without revalidation (Finding 4).

Close Finding 1 (six markers, or six honest deletions) before the docs step.
Findings 2–5 are deferred to the Backlog. Nothing here warrants CRITICAL: no
supported operation produces a false success, and every failure I found is either
fail-closed or requires a hand-corrupted repository.

```json centinela:verification
{
  "revision": "0ba4381ef3174f0ae7effb15bc52bc815f5bf1e6",
  "treeDigest": "sha256:8aabe130a4696c69e45df04f55f371441d3981efc93edc9e747930aa87483e82",
  "commands": [
    {"argv": ["centinela", "validate"], "exitCode": 0, "durationMs": 602000},
    {"argv": ["go", "test", "./..."], "exitCode": 0, "durationMs": 304000},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 0, "durationMs": 196},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 1, "durationMs": 109},
    {"argv": ["centinela-verify-mtd2", "merge", "--continue", "high-score"], "exitCode": 0, "durationMs": 129},
    {"argv": ["centinela-verify-mtd2", "merge", "--continue", "high-score"], "exitCode": 0, "durationMs": 122},
    {"argv": ["centinela-verify-mtd2", "merge", "--continue", "high-score"], "exitCode": 1, "durationMs": 67},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 1, "durationMs": 148},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 1, "durationMs": 137},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score", "--force-remove"], "exitCode": 0, "durationMs": 144},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 0, "durationMs": 148},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 0, "durationMs": 132},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 1, "durationMs": 56},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 1, "durationMs": 29},
    {"argv": ["centinela-verify-mtd2", "merge", "high-score"], "exitCode": 1, "durationMs": 28},
    {"argv": ["centinela-verify-mtd2", "merge", "--continue", "high-score"], "exitCode": 1, "durationMs": 46}
  ]
}
```
