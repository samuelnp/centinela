# merge-truthful-delivery — edge-case-tester

### Edge-Case Report: merge-truthful-delivery
**Date:** 2026-07-09

Grounding: `internal/worktree/primary.go` (`PrimaryTree`, `parseFirstWorktreeBlock`),
`internal/worktree/merge_verify.go` (`headSHA`, `currentBranch`, `isAncestor`,
`verifyRemoved`, `verifyAdvance`), `internal/worktree/merger.go` (`Merge`),
`cmd/centinela/merge.go` (`runMerge`, `runValidateForMerge`, `docsPortalRegen`),
`cmd/centinela/merge_report.go` (`reportMergeSuccess`),
`cmd/centinela/merge_dispatch.go` (`dispatchSteward`),
`internal/worktree/merge_pending.go` (`WritePending`/`ClearPending`),
`internal/worktree/remove.go` (`Remove`), `internal/worktree/path.go`
(`DetectFeatureFromCwd`, `Exists`, `Path`), `internal/worktree/finalize.go`
(`ResolveMerge`), `cmd/centinela/merge_continue.go` (`runMergeContinue`).
Dispositions: **[spec]** covered-by-spec-scenario · **[test]** needs-test ·
**[risk]** accepted-risk-with-reason.

#### Risk Matrix

- **Case:** EC-01 Hardcoded "main" in success wording when the primary branch is renamed
  - **Impact:** High · **Likelihood:** Low · **[test] — HIGH, do not miss**
  - **Why:** `Merge` (merger.go:45) already computes `currentBranch(repo)` and the
    verification is branch-agnostic (correct), but `reportMergeSuccess`
    (merge_report.go:20,23) prints "…into main" for a primary tree on `develop`
    or `master`, and `runMergeContinue` (merge_continue.go:35) does the same.
    The feature's whole point is truthful messaging; the plan/big-thinker
    (outstanding question 1) asked for the actual target branch to be named.
    `MergeOutcome` has no `TargetBranch` field, so the message cannot be fixed
    without a small code change. Spec Background pins "main", so NO scenario
    exercises a renamed primary — this is a truthfulness hole inside the fix
    itself. Fix now (carry the branch on the outcome) or explicitly accept.

- **Case:** EC-02 `--continue` finalize path bypasses ALL new verification
  - **Impact:** High · **Likelihood:** Medium · **[risk] (deferred `merge-continue-primary-tree`) — flag to qa**
  - **Why:** `runMergeContinue` → `ResolveMerge(".", …)` (merge_continue.go:25,
    finalize.go:30) uses the invoking CWD and has no `verifyAdvance`/
    `verifyRemoved`; merge_continue.go:35 prints the unconditional
    "Merged %q into main and removed its worktree." — the exact fabricated
    line this feature kills, still alive on the steward-resume path. Deferred
    by plan, but interaction is real: EC-13/EC-14 misroute non-conflict git
    errors INTO the steward flow, i.e. into the unverified path. Tests must at
    minimum pin that the happy `runMerge` path never routes through it.

- **Case:** EC-03 Pending-marker write/clear CWD asymmetry
  - **Impact:** Medium · **Likelihood:** Medium · **[risk] (deferred `merge-pending-marker-primary-tree`)**
  - **Why:** New inconsistency introduced by this change: `dispatchSteward`
    still calls `worktree.WritePending(".", o)` (merge_dispatch.go:15) —
    marker + its `WorktreePath: Path(".", feature)` are CWD-relative — while
    `runMerge` clears at the resolved primary (`ClearPending(repo, …)`,
    merge.go:66). Stall from inside the worktree → marker lands in
    `<worktree>/.workflow/` and only disappears because worktree removal
    deletes it; stall from a nested subdir strands it forever (hook re-emit,
    `merge --continue` LoadPending miss). One-line fix candidate: pass `repo`
    into `dispatchSteward`. Recorded as deferred; qa should decide fix-now vs
    keep deferred.

- **Case:** EC-04 Porcelain attribute placement/annotations (`bare`, `locked <reason>`, `prunable <reason>`)
  - **Impact:** Medium · **Likelihood:** Low · **[test]**
  - **Why:** `parseFirstWorktreeBlock` (primary.go:35) matches `line == "bare"`
    anywhere inside the first block (placement-independent, per git's
    documented contract the main tree is the first block) and ignores
    `locked …`/`prunable …` by exact-match, so `locked bare` cannot false-
    positive. Needs unit pins: bare-after-HEAD ordering, locked-with-reason
    line in the first block, annotations only on later blocks. Ordering of
    worktree *blocks* (main first) is documented and stable across git
    versions — no known version emits linked trees first. Locally verified
    shape on git 2.50.1.

- **Case:** EC-05 Porcelain output with no trailing blank line / CRLF endings
  - **Impact:** Low · **Likelihood:** Low · **[test]**
  - **Why:** Real git always terminates the last block with `\n\n` (verified
    by byte-dump locally), but the parser's loop-exit return (primary.go:55)
    handles a trimmed/CRLF stream from exotic wrappers. That return path and
    the `\r` TrimRight (primary.go:37) need unit coverage or they will show
    as uncovered branches against the 95% gate.

- **Case:** EC-06 Empty/garbage porcelain vs runner failure
  - **Impact:** Medium · **Likelihood:** Low · **[spec]+[test]**
  - **Why:** Runner failure (not a git repo) is spec-covered ("cannot resolve
    primary working tree", merge.go:47→primary.go:17). Exit-0-but-garbage
    (no `worktree ` line, primary.go:22) is only reachable via a stubbed
    `gitRunner` — needs the planned unit test; both refuse, never guess.

- **Case:** EC-07 Primary path containing spaces
  - **Impact:** Medium · **Likelihood:** Low · **[test]**
  - **Why:** `parseFirstWorktreeBlock` takes everything after the first space
    (primary.go:46) — correct because porcelain does not escape paths. Needs
    both the stubbed unit AND one real-git case (repo under a dir with a
    space) so `gitRunner(cmd.Dir=repo)` and `Remove`'s relative-arg call
    (remove.go:16-20) are proven against a spacey base.

- **Case:** EC-08 Symlinked paths (macOS `/tmp`→`/private/tmp`, `/var/folders`)
  - **Impact:** Low · **Likelihood:** Medium (in tests, not production) · **[test hazard]**
  - **Why:** Porcelain prints git's stored (realpath-resolved-at-add) path;
    `DetectFeatureFromCwd` EvalSymlinks its input (path.go:27) — so the
    notice check is consistent, but tests MUST NOT string-compare
    `os.Getwd()`/`t.TempDir()` to `PrimaryTree` output without resolving
    symlinks on both sides, or they will be macOS-flaky. All disk checks
    (`Exists`, `Remove`) stat through symlinks — functionally safe.

- **Case:** EC-09 Invocation from a nested subdirectory of the worktree (or of the primary)
  - **Impact:** Medium · **Likelihood:** Medium · **[test]**
  - **Why:** `PrimaryTree(".")` is safe (git walks up); `IsInsideWorktree(".")`
    matches the `.worktrees/<f>` path segment from any depth (path.go:30-37),
    so the notice still prints; `DetectSpecConflicts`/`Merge`/`ClearPending`
    all get the resolved `repo`. The residue is EC-03 (marker into
    `<subdir>/.workflow`) and the deferred portal regen (EC-12). One cheap
    variant of the cmd-level regression test should chdir into
    `<worktree>/internal/` instead of the worktree root.

- **Case:** EC-10 Single-worktree repo (no linked trees)
  - **Impact:** Low · **Likelihood:** Medium · **[spec]**
  - **Why:** One porcelain block → `PrimaryTree` returns it; no notice
    (`IsInsideWorktree` false). Covered by the "run from the primary tree"
    scenario; add the single-block string to the parser unit table. Note
    `deliver --via merge` additionally requires `wf.WorktreePath != ""`
    (deliver.go:46) so no-worktree delivery refuses earlier — plain
    `centinela merge` for a never-created worktree hits EC-13 instead.

- **Case:** EC-11 Bare primary / detached HEAD / feature branch checked out in primary / dirty primary
  - **Impact:** High · **Likelihood:** Low · **[spec]**
  - **Why:** All four refusals implemented exactly where planned: bare
    (primary.go:24), detached (merge_verify.go:21-27 — note the guard also
    makes the unborn-HEAD repo fail fast at `headSHA`, merger.go:52),
    self-merge (merger.go:49 — kills the trivial-self-ancestor
    `AlreadyMerged` fabrication), dirty (merger.go:40 — `isDirty` uses
    `status --porcelain`, so an UNTRACKED file in the primary also refuses;
    stricter than the spec's "uncommitted changes" wording, honest, worth a
    test assertion so the behavior is pinned deliberately).

- **Case:** EC-12 Post-removal deleted process CWD → portal regen always fails from inside a worktree
  - **Impact:** Low · **Likelihood:** High · **[risk] (deferred `merge-portal-regen-primary-tree`)**
  - **Why:** After `Remove` succeeds, the process CWD (inside the removed
    worktree) is a deleted directory; `docsPortalRegen` (merge.go:20-22)
    writes the RELATIVE `docs/project-docs/index.html` → ENOENT → the
    "notice: portal regen skipped" branch (merge.go:72-74) fires on EVERY
    from-inside-worktree merge; when run from a non-primary CWD that still
    exists, the portal is written to the WRONG tree. Best-effort by design so
    success stays truthful; needs one test asserting notice-not-crash and
    that the success line still follows.

- **Case:** EC-13 Any git-merge error conflated with TextConflict (nonexistent branch, deleted branch on re-run, index.lock)
  - **Impact:** Medium · **Likelihood:** Medium · **[risk] (deferred `merge-git-error-conflated-conflict`)**
  - **Why:** merger.go:58-62 maps EVERY `git merge` error to
    `TextConflict=true` → `dispatchSteward` + pending marker, with empty
    `ConflictedPaths`. Misroutes: merging a feature whose branch never
    existed, an already-merged re-run whose branch was manually deleted, and
    the loser of two concurrent merges (index.lock). Truthfulness holds — no
    success is ever claimed — but the operator is sent to the steward (and
    thence the unverified EC-02 path) for a non-conflict. Pre-existing
    behavior, now recorded on the roadmap.

- **Case:** EC-14 Concurrent merge attempts
  - **Impact:** Medium · **Likelihood:** Low · **[risk]**
  - **Why:** Two `centinela merge` racing on the same primary: git's
    index.lock serializes the actual merges; the loser lands in EC-13.
    `WritePending` is atomic (temp+rename, merge_pending.go:47-56) so the
    marker never tears. Residual TOCTOU: both pass `isDirty`/`headSHA`
    before one merges — the second's `before` is stale, but its own merge
    either errors (lock) or re-reads reality; `verifyAdvance` re-reads HEAD
    after, so no false success is constructible. Accepted: no cross-process
    lock exists in centinela, git's own lock plus ref-verification bound the
    damage to a misrouted steward dispatch.

- **Case:** EC-15 External commit races between `headSHA(before)` and the merge
  - **Impact:** Low · **Likelihood:** Low · **[risk]**
  - **Why:** If another process commits to the primary in the window
    (merger.go:52-56), `before != after` even when the merge itself printed
    "Already up to date" → labeled `RefAdvanced` instead of `AlreadyMerged`.
    Since `git merge` exit 0 guarantees the branch IS an ancestor of the new
    HEAD, both messages remain factually true — the mislabel is wording-only.
    Accepted; not economically testable.

- **Case:** EC-16 `git worktree remove` refuses on untracked/modified files in the feature worktree
  - **Impact:** Medium · **Likelihood:** Medium · **[spec]+[test]**
  - **Why:** `Remove(repo, feature, false)` (merger.go:74, remove.go:16) has
    no `--force`; git refuses on modified OR untracked files (gitignored
    files — the `.workflow` machine `.json`/`.lock` — do NOT block;
    submodules would always block). Sequence hazard: main has ALREADY
    advanced when the removal error aborts (merger.go:71 ran first), so the
    command exits non-zero with the merge landed — the honest-failure
    behavior the spec's "removal is only claimed…" scenario wants, but the
    re-run path must be pinned: after cleaning the worktree, re-run yields
    the truthful `AlreadyMerged` exit 0. Also assert the error surfaces
    git's own "use --force" hint (big-thinker outstanding question 3:
    message actionability).

- **Case:** EC-17 Already-merged re-run with worktree already gone (idempotency)
  - **Impact:** Low · **Likelihood:** Medium · **[spec]**
  - **Why:** `Remove` no-ops on `!Exists` (remove.go:12), `verifyRemoved`
    passes, `verifyAdvance` sets `AlreadyMerged` only after an unmoved HEAD
    (merge_verify.go:56-57 — ancestor is never consulted pre-merge, killing
    the false-AlreadyMerged risk from the big-thinker table). Spec-covered.
    Variant with the BRANCH also gone falls into EC-13.

- **Case:** EC-18 Already-merged re-run while validate is red
  - **Impact:** Low · **Likelihood:** Low · **[risk]**
  - **Why:** The no-op merge still runs validate (merger.go:64); a failure
    yields `ValidateFail` → steward dispatch for an already-delivered branch
    instead of the honest exit-0. No false success; pre-existing sequencing.
    Accepted — fixing it means moving verification before validate, which
    would change steward semantics out of scope.

- **Case:** EC-19 Merge-time validate runs in the invoking CWD, not the merged tree
  - **Impact:** Medium · **Likelihood:** High · **[risk] (deferred `merge-validate-primary-tree`)**
  - **Why:** `runValidateForMerge` (merge.go:78) discards its repo argument;
    from inside the feature worktree the "post-merge validate" actually
    validates the pre-merge WORKTREE. Interaction with this feature: the
    validate gate that stands between a clean text merge and the truthful
    success line is measured on the wrong tree, so a merge that breaks the
    merged primary can still reach the (ref-verified, honestly-worded)
    success. The ref claims stay true; the quality claim does not. Deferred,
    but qa must not let a test encode the wrong-tree behavior as desirable.

- **Case:** EC-20 `verifyRemoved`/`Exists` shape gaps
  - **Impact:** Low · **Likelihood:** Low · **[risk]**
  - **Why:** `Exists` requires `IsDir()` (path.go:49-50): a stray plain FILE
    left at `.worktrees/<f>` after removal would pass `verifyRemoved`.
    Pathological (git never leaves one); accepted. `Path(repo, feature)` is
    injection-safe because `ValidateFeatureSlug` (slug.go:12) rejects `/`,
    `..`, spaces and metacharacters before any path join or git argv.

- **Case:** EC-21 Primary tree moved/deleted after porcelain read
  - **Impact:** Low · **Likelihood:** Low · **[risk]**
  - **Why:** If the porcelain-reported primary path no longer exists,
    `gitRunner`'s `cmd.Dir=repo` fails with a chdir error at the first
    `isDirty` call → honest hard error, no guessing. Accepted without a test
    (exec-level failure, uniform across all call sites).

#### Missing or Weak Scenarios

1. **Renamed primary branch (EC-01)** — no spec scenario runs with a primary
   on anything but `main`; the hardcoded wording bug is invisible to the
   entire suite as planned.
2. Nested-subdirectory invocation (EC-09) — spec only covers worktree ROOT
   as CWD.
3. Post-advance removal-failure RE-RUN (EC-16 second half) — spec covers the
   failing run, not the recovery run's honest `AlreadyMerged`.
4. Portal-regen notice on deleted CWD (EC-12) — no scenario asserts the
   best-effort notice instead of a crash after from-inside-worktree success.
5. Parser micro-variants (EC-04/05): bare-after-HEAD, `locked <reason>`,
   CRLF, missing trailing newline, single-block, garbage-with-exit-0.
6. Untracked-file-in-PRIMARY refusal (EC-11 note) — pins that `isDirty`
   includes untracked, so the message matches reality.
7. Real-git repo under a path containing a space (EC-07).

#### Proposed/Added Tests

(Per plan file layout; no test files written by this role.)

- **Unit** (`internal/worktree/primary_test.go`, stubbed `gitRunner`):
  two-block happy path; single block; bare primary (attribute after HEAD
  line); `locked with reason` line in first block ignored; path with spaces;
  CRLF + no trailing newline (loop-exit return primary.go:55); empty output;
  garbage exit-0 output; runner error propagated with trimmed git output.
- **Unit** (`internal/worktree/merge_verify_test.go`, stubbed): RefAdvanced;
  unmoved HEAD + ancestor → AlreadyMerged; unmoved + non-ancestor → exact
  "did not advance" error; detached-HEAD message; headSHA failure (unborn
  repo); verifyRemoved failure when dir survives.
- **Unit** (`internal/worktree/merger_*_test.go`, stubbed): self-merge
  refusal (merger.go:49); removal-failure AFTER advance leaves error and
  RefAdvanced=true on the returned outcome; untracked-file dirty refusal.
- **Integration** (real git, `merge_realgit_test.go` + cmd
  `merge_truthful_test.go`): THE regression (from inside worktree: main SHA
  advanced, worktree gone, notice printed — resolve symlinks before path
  comparisons, EC-08); nested-subdir variant (EC-09); already-merged re-run
  exit 0 wording; re-run with worktree pre-deleted; untracked file blocks
  removal then cleaned re-run succeeds honestly (EC-16); repo path with a
  space (EC-07); unresolvable primary refusal; conflict and validate-fail
  paths keep worktree + never print success + `reportMergeSuccess` never
  called; defense-in-depth guard of merge_report.go:16 (zero-flag outcome →
  refusal error).
- **Acceptance** (`tests/acceptance/…`, temp-built binary, t.TempDir, NO
  network/remotes — a prior network push hung the suite): one scenario per
  spec title (`// Scenario:` traceability), driving `deliver <f> --via merge`
  from inside the worktree and asserting on `git rev-parse` output, never on
  stdout alone.

#### Residual Risks

- EC-01 hardcoded "main" wording — fix in-scope (add `TargetBranch` to
  `MergeOutcome`, populated from the `currentBranch` already computed) or
  record an explicit accept; currently the fix contradicts its own spec's
  truthfulness promise on renamed-main repos.
- EC-02 `--continue` path prints unverified success (deferred
  `merge-continue-primary-tree`) — mitigation until then: steward evidence
  gate is the only guard; do not extend it with new success claims.
- EC-03 marker CWD asymmetry (deferred `merge-pending-marker-primary-tree`);
  cheap hardening if touched anyway: `dispatchSteward(repo, o)`.
- EC-13/EC-14 git-error→conflict conflation incl. concurrent merges
  (deferred `merge-git-error-conflated-conflict`).
- EC-12/EC-19 wrong-CWD validate + portal regen (already deferred:
  `merge-validate-primary-tree`, `merge-portal-regen-primary-tree`).
- EC-15, EC-18, EC-20, EC-21 accepted with reasons above; none can produce a
  false success line.

#### Deferred Findings

Recorded this session via `centinela roadmap defer` (source
`merge-truthful-delivery/edge-case-tester`):

- `merge-pending-marker-primary-tree` (EC-03)
- `merge-git-error-conflated-conflict` (EC-13/EC-14)

Pre-existing (recorded by big-thinker/plan roles, confirmed here):
`merge-validate-primary-tree` (EC-19), `merge-continue-primary-tree`
(EC-02), `merge-portal-regen-primary-tree` (EC-12).
