# Feature Brief — roadmap-state-hygiene

- surface: internal
- source: retrospective.md WS1.4 + WS6 + observed failure #10 (state footguns,
  deferral one-way valve) — plus one week of field evidence delivering Phase 13.

## Problem — what pain, who

Roadmap state (`.workflow/roadmap.json` + its generated `ROADMAP.md`) is the only
shared mutable file every parallel session writes, and Centinela treats it as if
one person edited it once:

1. **Mutations leave the tree dirty.** `roadmap defer/remove/promote/...` end in a
   plain file write; the CLI then prints *"Remember to sync ROADMAP.md"*. In the
   2048 field test a dirty `roadmap.json` in the main checkout aborted an ff-merge
   mid-delivery and a reflexive `git branch -D` nearly orphaned a completed
   feature (retrospective §7.2).
2. **ROADMAP.md drifts.** Regeneration is a separate command a human must
   remember; the `roadmap_drift` gate is set to `warn`, so drift ships.
3. **Every delivery this week hit the same merge conflict** in
   `.workflow/roadmap.json` AND `ROADMAP.md`. The resolution was identical every
   time: take main's side, replay the branch's own deferrals with `roadmap defer`,
   run `roadmap generate`. It is a mechanical procedure a human performs by hand
   under time pressure — and it has already lost data once: commit `f3886dd`
   ("record roadmap-checkpoint-marker-truncation — *verifier finding lost in
   merge*").
4. **Regeneration fights the freshness gate.** Verifiers must run
   `roadmap generate` after recording deferrals, but `artifact stamp` must be
   their LAST action; `ROADMAP.md` lives outside the `.workflow/` digest
   exclusion, so regenerating after stamping voids their own verification. One
   verifier resolved this by skipping the regeneration — drift shipped.
5. **The Backlog is a one-way valve.** 106 deferred findings (was 12 at the time
   of the retrospective), no aging view, no closure prompt. Deferral without a
   closure path is socially acceptable debt.

Who hurts: every agent and human running a roadmap mutation, every verifier, and
every delivery that has to merge.

## Scope

**In**
- One post-mutation choke point shared by every roadmap.json mutation:
  regenerate `ROADMAP.md` in-process, then commit **exactly** the roadmap-state
  pathspec, respecting `workflow.disable_auto_commit`.
- A stdlib-only leaf owning that pathspec, so the SAME list drives the commit,
  the tree-digest exclusion, and the freshness revision-range exemption.
- Freshness: roadmap-state-only churn (committed or not) no longer stales a
  gatekeeper stamp; anything else still does.
- `centinela roadmap backlog [--stale] [--older-than N] [--json]` — aging on the
  `deferredAt` field records already carry.
- A completion nudge when the last schedulable feature completes and the Backlog
  is non-empty.
- `centinela roadmap resolve` — mechanizes the observed conflict procedure:
  semantic union of both sides' Backlog findings, regenerate ROADMAP.md, stage
  both; refuse (leaving markers) when a schedulable phase diverged on both sides.
- Canonical one-object-per-line rendering for **every** phase, not only the
  mutated one (kills reformat churn; closes Backlog item
  `rawio-reformat-diff-churn`).

**Out**
- A git merge driver / `.gitattributes merge=union` (analysed and rejected — see
  the plan's "Merge-friendliness" section; union on a JSON array yields invalid
  JSON and silently unions semantic edits).
- Splitting the Backlog into a separate JSONL sidecar file (a schema split
  touching every reader; the semantic `roadmap resolve` gets the same win with a
  fraction of the blast radius).
- Full semantic 3-way merge of schedulable phases (v1 refuses that case).
- Auto-committing anything other than roadmap state; changing `complete`'s
  `git add -A` step commit; a config knob for the staleness threshold.
- Any change to what the `roadmap_drift` gate checks or its severity.

## User Stories

- As an agent recording a deferral, one command leaves the tree clean, ROADMAP.md
  in sync, and a conventional commit on my own branch — so the next `deliver`
  cannot abort on a dirty tree.
- As a verifier, I record deferrals and regenerate without voiding the stamp I am
  about to write, and I never have to choose between drift and freshness.
- As a session merging to main, `centinela roadmap resolve` replays the procedure
  I used to do by hand, and never silently drops a finding.
- As a maintainer, `roadmap backlog --stale` shows me what has rotted, and
  finishing the roadmap tells me the Backlog is still full.

## Acceptance Criteria (→ specs/roadmap-state-hygiene.feature)

1. **Regeneration is folded into every mutation.** `roadmap defer | add | remove |
   edit | move | reorder | promote | phase add | phase rename | phase remove`
   rewrite `ROADMAP.md` from the just-written `roadmap.json` in the same process,
   before exiting. After any of them the `roadmap_drift` gate is clean with no
   manual `roadmap generate`.
2. **Auto-commit is pathspec-scoped.** The same command creates exactly one
   commit containing only `.workflow/roadmap.json` and `ROADMAP.md`, message
   `chore(roadmap): <verb> <subject>`, made with `--no-verify` (the commit
   carries only generated state; the only gate that governs it — `roadmap_drift`
   — is clean by construction).
3. **Unrelated work is untouched.** With other files staged and other files
   dirty, the mutation commit contains neither; the index and working tree for
   every other path are byte-identical afterwards.
4. **`disable_auto_commit` is respected.** With `workflow.disable_auto_commit =
   true`, ROADMAP.md is still regenerated (correctness, not VCS policy), no
   commit is made, and one line says the state was left uncommitted.
5. **The commit is best-effort and never fails the mutation.** No git, no repo,
   no HEAD, a merge/rebase in progress, or a failing commit → the mutation still
   exits 0 and prints a warning naming the reason. (A non-zero exit would invite
   an agent to re-run the mutation and double-record it.)
6. **Worktree-local.** Run inside `.worktrees/<feature>`, the commit lands on that
   worktree's HEAD/branch; the primary checkout's tree and index are untouched.
7. **No empty commits.** A mutation that leaves both state files byte-identical
   creates no commit.
8. **`roadmap generate` stays standalone and never commits.** It regenerates and
   reports only — it is the repair path used mid-conflict, where a partial commit
   is impossible.
9. **Freshness is roadmap-state-blind.** After `artifact stamp`, a roadmap-state
   mutation (committed via AC2 or left dirty via AC4) does NOT report the
   verification stale. A commit in the same range touching any other path still
   does, with today's message.
10. **`roadmap backlog`** lists every Backlog finding oldest-first with age,
    slug, `source` and summary; `--stale` filters to older than 14 days,
    `--older-than N` overrides; `--json` emits a machine shape; a missing or
    unparseable `deferredAt` renders `unknown` and counts as stale.
11. **Completion nudge.** When `centinela complete` takes a workflow to `done`
    and no schedulable roadmap feature is left incomplete and the Backlog is
    non-empty, it prints the counts and the two follow-up commands. It is silent
    when schedulable work remains or the Backlog is empty.
12. **`roadmap resolve` (Backlog-only divergence)** merges the conflicted
    `roadmap.json` from git stages 1/2/3: findings are unioned by slug, ordered
    by `deferredAt` then name, ROADMAP.md is regenerated from the result, both
    are staged, and the summary names how many findings came from each side.
13. **`roadmap resolve` refuses a real conflict.** If a schedulable phase changed
    on both sides, it exits non-zero, names the phase, and leaves the conflict
    markers in place. With nothing conflicted it exits 0 as a no-op; with only
    `ROADMAP.md` conflicted it regenerates from the already-merged roadmap.json
    and stages it.
14. **Canonical rendering.** Every write emits every phase's `features` array as
    one compact object per line; a mutation to one phase produces a diff confined
    to the changed line(s) — untouched phases do not reformat.

## Edge Cases

| # | Case | Expected |
|---|------|----------|
| E1 | Mutation succeeds, `git commit` fails (hook, signing, detached, no identity) | exit 0, warning names the cause, state on disk correct |
| E2 | Merge or rebase in progress | no commit attempt (git refuses partial commits); warning; exit 0 |
| E3 | Other files staged when a mutation runs | those stay staged, uncommitted |
| E4 | roadmap.json already dirty from a hand edit before the mutation | that content is part of the read-modify-write, so it is committed with the mutation; the commit body names both files |
| E5 | Not a git repository at all | regenerate, no commit, one-line notice |
| E6 | `disable_auto_commit` + drift gate at `fail` | regeneration still happens, so the gate is clean |
| E7 | Backlog empty | `roadmap backlog` prints an explicit empty state, exit 0; no nudge |
| E8 | Every Backlog finding is newer than the threshold | `--stale` prints "no findings older than Nd", exit 0 |
| E9 | Finding without `deferredAt` (hand-edited/legacy) | age `unknown`, sorted first, counted stale |
| E10 | Finding with an unparseable `deferredAt` | same as E9; never an error |
| E11 | `resolve` when both sides added the SAME slug | one entry survives (dedupe by slug, earliest `deferredAt` wins) |
| E12 | `resolve` when one side DELETED a finding the other kept (promote/remove) | deletion wins over an unchanged side; both-sides-changed → refuse |
| E13 | `resolve` outside a merge / no conflicted paths | no-op, exit 0 |
| E14 | `resolve` with roadmap.json conflicted but unparseable on one side | refuse, exit non-zero, name the side |
| E15 | Stamp taken, then a deferral commits, then `complete` | not stale (AC9) |
| E16 | Stamp taken, then a source file is committed alongside roadmap state | stale, today's message |
| E17 | Two mutations in the same second in the same checkout | each writes atomically and commits its own pathspec; the second sees the first's content |
| E18 | Nudge conditions met but roadmap.json unreadable | no nudge, no error (never break `complete`) |
