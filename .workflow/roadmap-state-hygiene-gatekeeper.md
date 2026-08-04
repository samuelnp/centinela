# roadmap-state-hygiene — gatekeeper (round 4)

## Status

**Status:** WARNING

Both round-3 findings are genuinely fixed, and I confirmed each by execution
through a real `git merge` conflict driven by the real binary, not by reading
the diff. The Backlog phase shell now three-ways against base in both
directions, refuses a both-sides divergence with the index and markers intact,
and honours a one-sided `note` DELETION either way. The summary now reconciles:
I recounted the attribution independently from the captured `:1:`/`:2:`/`:3:`
stages of the project's own 130-finding merge and got the published numbers
exactly.

I did find one NEW defect the rewrite did not close, in the same code it
touched: when one side's `roadmap.json` lacks the Backlog phase that the merge
base had, `resolve` emits a phase object with **no `name` key**, and the
surviving deferred findings are silently reclassified as schedulable — behind a
green `✓ Resolved roadmap state`. It is MEDIUM rather than CRITICAL because no
Centinela command can produce that input shape (both `Backlog`/`backlog`
spellings are reserved; `phase remove` and `phase rename` refuse them), and
because no finding's bytes are destroyed — they are misfiled, not lost.

## Inputs Read

Paths only; every claim below comes from a command I ran myself.

- `git diff origin/main...HEAD` (120 files, +6821/-478) and the uncommitted tree
- `docs/features/roadmap-state-hygiene.md`, `docs/plans/roadmap-state-hygiene.md`
- `specs/roadmap-state-hygiene.feature` (43 scenarios)
- `internal/roadmap/resolve.go`, `resolve_shell.go`, `resolve_survivor.go`,
  `resolve_backlog.go`, `resolve_threeway.go`, `resolve_side.go`,
  `resolve_index.go`, `resolve_order.go`, `resolve_errors.go`, `resolve_stages.go`
- `internal/roadmap/mutate_lock.go`, `mutate_lock_path.go`, `mutate_lock_canon.go`,
  `statesync.go`, `defer.go`, `regenerate.go`
- `internal/filelock/filelock.go`, `flock_unix.go`
- `internal/gitutil/commit_guard.go`, `internal/treestate/digest.go`,
  `internal/ui/render_resolve.go`, `cmd/centinela/roadmap_resolve.go`,
  `cmd/centinela/roadmap_sync.go`
- `centinela.toml` / `PROJECT.md` diffs (checked for gate weakening — additive only)

I was handed a narrative of what changed since round 3. I did not rely on it:
every claim in it was re-derived by execution below.

## Analyzed Specs

- `specs/roadmap-state-hygiene.feature` — all 43 scenarios. Each has an
  unwrapped `// Scenario:` marker in `tests/acceptance/`; none is inside a block
  comment or a string. The `spec-traceability-gate` independently reports 45/45
  scenarios covered in scope.

## Refutation Attempts

Each line is a claim I tried to break, and what actually happened. Every probe
ran in a THROWAWAY repo through a real `git merge` conflict and the real binary.

- **"A one-sided incoming `note` edit now survives."** Ours edited a finding,
  theirs edited the Backlog `note`, real conflict. Refuted my doubt: output kept
  `THEIR NOTE`; the round-3 MEDIUM is gone.
- **"…and a one-sided LOCAL note edit still wins."** Mirror direction: kept
  `OUR NOTE`. Holds.
- **"Both-sides note divergence refuses."** Refused with the phase named, exit 1,
  3 unmerged index stages still present, conflict markers still in the worktree.
  Nothing written. Holds.
- **"Both sides editing the note IDENTICALLY resolves."** Resolved cleanly. Holds.
- **"A pure key REORDER of the shell is not a conflict."** Ours reordered
  `note`/`name`, theirs added a finding: merged cleanly, `note` preserved, keys
  re-emitted canonically. Holds.
- **"Deletion of `note` is an edit, not a value to restore."** Tried BOTH
  directions (ours deletes / theirs deletes). The key is absent from the output
  both times — deletion honoured, never silently restored. Holds.
- **"Modify/delete on a SHELL key refuses."** Ours deleted `note`, theirs
  modified it: refused. Holds.
- **"Backlog is not specially privileged."** One-sided edits to a NON-Backlog
  phase (`Phase 0`) survive from either side; both-sides divergence refuses
  naming `Phase 0`. Non-Backlog phases three-way at whole-phase granularity,
  which is coarser but strictly fail-safe.
- **"`Kept == FromBase + FromOurs + FromTheirs` in every shape."** Ran eight
  shapes: one-sided ours, one-sided theirs, both-identical, base-unchanged, an
  entry deleted on one side, a slug both sides deferred with DIFFERENT capture
  bytes in both orderings, and both modify/delete directions. Reconciled in
  every case that produced a summary. **Broke it once** — see Finding 2.
- **"The rendered line never names a side that did not contribute the bytes."**
  The double-defer case is the one that used to lie: with theirs' capture
  earlier, the surviving bytes are theirs and the line now says `1 from theirs`.
  Reversed the ordering and it flipped correctly. The round-3 LOW is gone. One
  residual ambiguity, not a lie: when both sides make the BYTE-IDENTICAL edit it
  is credited to "our side" — the bytes genuinely are ours (and also theirs).
- **"The counts are not just self-consistent but TRUE."** Cloned the repo,
  merged `origin/main` into the branch for real (base 106 / ours 120 / theirs 116
  Backlog findings), captured all three index stages, ran `resolve`, then
  recounted attribution independently in Python from the stages. Published
  `130 kept: 105 unchanged, 14 ours, 11 theirs`; my independent recount:
  `base 105, ours 14, theirs 11, sum 130`. Exact match. One slug
  (`hook-context-panel-diet`) was dropped — base+ours had it byte-identically,
  theirs deleted it: the one-sided-deletion rule, correctly applied.
- **"The symlink lock race is closed at the PROCESS tier."** Did not trust the
  new in-process test. Ran 12 CONCURRENT real `centinela roadmap defer`
  processes, half entering through `/private/tmp/.../racereal` and half through
  a symlink to it. All 12 exited 0, all 12 records survived, exactly ONE lock
  file was created in `.git`, and `git status --porcelain` was empty. Holds.
- **"The lock is invisible to `git status`."** Confirmed in the same run — the
  lock lives in `.git/`, and status was clean.
- **"Concurrent auto-commits never report a commit that did not happen."** Six
  concurrent defers produced 1 commit and 5 honest `⚠ … was not committed`
  lines naming `index.lock`; all 6 records were in HEAD, tree clean. Pessimistic
  but never false. This is the already-deferred index-lock race, correctly
  characterised — no data loss.
- **"Mutations stay inside their declared pathspec."** Dirtied a tracked file and
  added an untracked one, then deferred: the commit contained exactly
  `.workflow/roadmap.json` + `ROADMAP.md`; both foreign changes were left alone.
- **"`disable_auto_commit` skips the commit but never the regeneration."**
  Confirmed: no new commit, both files regenerated and left modified in the tree,
  honest message.
- **"`roadmap generate` never commits."** Confirmed, including WHILE a merge was
  in progress (`MERGE_HEAD` survived untouched).
- **"In-progress and detached guards block the commit without losing the record."**
  Merge-in-progress → `⚠ … merge in progress`, record on disk. Detached HEAD →
  `⚠ … detached HEAD`, record on disk, and crucially NO orphan commit created.
- **"The freshness exemption is symmetric and not over-broad."** Drove
  `treestate.Digest` directly with a throwaway probe test: roadmap-state-only
  churn does not move the digest; a non-roadmap change DOES; a mixed change does;
  a rename OUT of roadmap state does. Probe file removed afterwards.
- **"Spec markers are honest."** All 43 scenarios in the `.feature` have an
  unwrapped `// Scenario:` marker in `tests/acceptance/`; none sits inside a
  block comment. The traceability gate agrees (45/45 in scope).
- **"No proof path was weakened."** No `t.Skip`/`testing.Short` guard was added
  beyond environment preconditions (`git not installed`); the `centinela.toml`
  and `PROJECT.md` diffs are purely additive (two new leaf packages mapped into
  the import-graph layer), no threshold lowered, no assertion relaxed. The
  apparent deletion of `durable_workflow_state_*_test.go` is a two-dot diff
  artifact — the branch is 8 commits behind `origin/main` and deletes nothing.
- **NEW ANGLE — "what if a side does not have the Backlog phase at all?"**
  BROKE IT. See Finding 1.
- **NEW ANGLE — "what if two phase names both match Backlog case-insensitively?"**
  BROKE IT. See Finding 2.
- **"A corrupted document degrades gracefully on the next pass."** It does not:
  re-conflicting the nameless phase yields `phase "" changed on both sides` — the
  Backlog union stops working for those findings entirely. Compounding, but
  downstream of Finding 1.

## Commands Run

Mandated, each exactly once, both uncached:

- `go build -o /tmp/centinela-verify-rsh4 ./cmd/centinela` → exit 0, 616 ms
- `/tmp/centinela-verify-rsh4 validate` → exit 0, 483780 ms. All built-in gates
  pass (G1 file size, G-Build cross-compile 6/6, spec-traceability 45/45,
  roadmap_drift in sync, docstring 80/80); `import_graph` emits its pre-existing
  non-failing "packages match no configured layer" warning. All three validate
  commands green: `go test ./... -coverprofile=coverage.out`,
  `check-coverage.sh`, `check-fmt.sh`. Final line: `All gates passed.`
- `go test ./... -count=1` → exit 0, 481042 ms. Every package `ok`, including
  `tests/acceptance` (463.7 s), `tests/integration`, `tests/unit`.

Defect reproducer, self-contained and timed:

- `/tmp/rsh4/repro.sh` → exit 0, 825 ms. Builds a throwaway repo, stages a real
  merge conflict, runs `roadmap resolve`, then `roadmap ready` and
  `roadmap backlog`. Output: `✓ Resolved roadmap state — kept 1 findings` /
  `🔓 c` / `No deferred findings in the Backlog.`

Untimed probes (all in throwaway repos under `/private/tmp/rsh4/`, all driving
the real binary through real `git merge` conflicts): shell three-way A1–A7,
non-Backlog phase B1–B4, attribution matrix C1–C8, hostile-git D1–D4,
generate/freshness E1–E2, the 12-process symlink lock race, the 6-process
index-lock race, the full `origin/main` dogfood merge with an independent
Python recount of the attribution, and the case-variant/absent-Backlog probes.

## Findings

**Finding 1 — MEDIUM — `resolve` emits a NAMELESS phase and silently promotes
deferred findings to schedulable.**
Affected: `internal/roadmap/resolve_shell.go` (`mergeShellKeys`),
`internal/roadmap/resolve_order.go` (`backlogPhase`). Spec: scenarios at
`specs/roadmap-state-hygiene.feature:313/321/327` (Backlog phase note),
`:280` (union).

When the merge base HAS the Backlog phase and one side's `roadmap.json` does
NOT, `phaseObject` returns an empty map for that side, so for every shell key
`threeWay(base=X, ours=nil, theirs=X)` sees "theirs equals base" and returns
ours — `nil` — deleting the key. That includes `name`. `backlogPhase` then
unconditionally rebuilds an object around the merged findings, so the output is
`{"features":[…]}` with no name. Reproduced end to end:

```
✓ Resolved roadmap state — kept 1 findings: 0 unchanged, 0 from our side, 1 from theirs.
$ centinela roadmap ready
  🔓 c                 ← a DEFERRED finding, now listed as ready to start
$ centinela roadmap backlog
No deferred findings in the Backlog.
```

The phase `note` is lost too, and `ROADMAP.md` regenerates with an empty `## `
heading — so `roadmap_drift` still passes and nothing flags it. Exit code 0,
green success line. This is the exact "silent reclassification behind a green ✓"
shape the feature exists to prevent, applied to the Backlog's validate-exemption
rather than to a finding's bytes.

Not CRITICAL: no Centinela command can create the triggering input. Both
`Backlog` and `backlog` are reserved — `roadmap phase remove Backlog` and
`roadmap phase rename Backlog backlog` are both refused. It needs an
out-of-band edit, revert, or restore of `roadmap.json` that drops the whole
Backlog phase on one branch. No finding's bytes are destroyed; they are misfiled.

Suggestion: in `backlogPhase`, when the merged shell has no `name` key, either
emit no phase at all (the whole-phase deletion the three-way actually decided) or
force the phase name back in — `name` is the phase's identity and can never
legitimately be deleted by a merge.

**Finding 2 — LOW — `Kept` is assigned, not accumulated, so two
case-variant Backlog phases break the reconciliation invariant.**
Affected: `internal/roadmap/resolve_backlog.go:44` (`out.Kept = len(kept)`)
versus `internal/roadmap/resolve_survivor.go:61-71` (`out.From*++`).

`mergedOrder` de-duplicates phase names by exact string, but
`isBacklogPhaseName` matches case-insensitively. A document where one side spells
the phase `backlog` and the other `Backlog` therefore calls `mergeBacklog`
TWICE; the contribution counters accumulate across both calls while `Kept` is
overwritten by the second. Observed:

```
✓ Resolved roadmap state — kept 1 findings: 0 unchanged, 2 from our side, 1 from theirs.
```

`1 != 0 + 2 + 1`. That is precisely the arithmetic this round's rewrite promised
a human could check. LOW because the trigger is the same unreachable-by-CLI
shape as Finding 1 (both spellings are reserved), and because the visible
consequence is a wrong number rather than a wrong document.

Suggestion: make it `out.Kept += len(kept)`, and/or fold case-insensitive
Backlog names together in `mergedOrder` so `mergeBacklog` can only run once.

**Test-coverage gap behind both:** `resolve_test.go:32` asserts the
reconciliation invariant, but only for a single fixture shape. No test drives a
side that LACKS the Backlog phase, and none drives two case-variant Backlog
names — which is why both survived a rewrite aimed squarely at this area.

## Deferred Findings

Recommend recording both of the above as Backlog deferrals rather than blocking
round 4:

- `roadmap-resolve-emits-nameless-backlog-phase` — Finding 1. Reproducer:
  `/tmp/rsh4/repro.sh` (base has Backlog, ours lacks it, theirs adds a finding).
- `roadmap-resolve-kept-count-overwritten-on-duplicate-backlog` — Finding 2.

Already recorded on the branch and re-confirmed by execution this round, not
re-litigated: `roadmap-mutation-commit-races-git-index-lock` (verified: warns
honestly, zero loss), `roadmap-resolve-keyorder-false-conflict` (verified:
refuses, never loses — probe B4), `roadmap-resolve-write-outside-mutation-lock`
(confirmed by reading `runRoadmapResolve`, which takes no mutation lock).
`roadmap-resolve-summary-arithmetic-unreconcilable` was correctly REMOVED: the
invariant now holds in every shape the CLI can produce, including the project's
own real 130-finding merge, which I recounted independently.

## Recommendation

**SHIP.** Status WARNING.

The two round-3 findings are closed, and I verified both by execution rather
than by inspection — including the parts a passing test suite would not have
told me: the process tier of the symlink lock race, the hostile-git guards, the
pathspec containment, and an independent recount of the attribution numbers
against the project's own real `origin/main` merge stages. `centinela validate`
passes and the full suite passes uncached, both run once by me.

I am not blocking on Finding 1 even though it is the feature's signature failure
shape, because Centinela's own CLI cannot produce the input that triggers it —
it reserves both Backlog spellings and refuses to remove or rename the phase —
and because no captured finding's bytes are destroyed, only misfiled. The fix is
small and should be taken soon; recording it as a deferral with a working
reproducer is the honest disposition for round 4.

One operational note, not a finding: this branch is 8 commits behind
`origin/main` and overlaps it on `.workflow/roadmap.json`, `ROADMAP.md` and
`.workflow/memory/index.json`. The first two conflict on merge and are exactly
what `roadmap resolve` handles — I dogfooded that merge in a clone and it
resolved cleanly, 130 findings, arithmetic verified. `.workflow/memory/index.json`
auto-merged.

```json centinela:verification
{
  "revision": "10a273d6d9f4750bff62bc60c1d483494412fc95",
  "treeDigest": "sha256:001cc862c90c7a6b92c8bac2d935a9e3f0905289d56c44c570b2943823992f1d",
  "commands": [
    {"argv": ["go", "build", "-o", "/tmp/centinela-verify-rsh4", "./cmd/centinela"], "exitCode": 0, "durationMs": 616},
    {"argv": ["/tmp/centinela-verify-rsh4", "validate"], "exitCode": 0, "durationMs": 483780},
    {"argv": ["go", "test", "./...", "-count=1"], "exitCode": 0, "durationMs": 481042},
    {"argv": ["/tmp/rsh4/repro.sh"], "exitCode": 0, "durationMs": 825}
  ]
}
```
