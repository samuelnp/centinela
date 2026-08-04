### Planner Report: roadmap-state-hygiene
**Date:** 2026-08-04

#### Problem

Roadmap state (`.workflow/roadmap.json` + generated `ROADMAP.md`) is the one
mutable file every parallel session writes, and Centinela still treats it as a
single-writer artifact. Mutations end in a bare file write and print "Remember to
sync ROADMAP.md", so the tree is left dirty and the markdown drifts — the exact
pair that aborted an ff-merge mid-delivery and nearly orphaned a completed branch
in the 2048 field test (retrospective §7.2). One week of Phase 13 deliveries
added harder evidence: EVERY delivery hit a merge conflict in both files, always
resolved by the same manual procedure (take main's side, replay this branch's
deferrals, regenerate), and that procedure has already lost data once — commit
`f3886dd` exists solely to re-record "a verifier finding lost in merge".
Verifiers additionally face a false choice: they must regenerate after recording
deferrals, but `artifact stamp` must be their last action and `ROADMAP.md` sits
outside the `.workflow/` digest exclusion, so regenerating after stamping voids
their own verification — one verifier resolved it by skipping the regeneration.
Meanwhile the Backlog has grown from 12 findings (retrospective) to **106**, with
no age, no view, and no closure prompt: a one-way valve.

#### Scope

- **In:** one post-mutation choke point (regenerate in-process + pathspec commit,
  honouring `disable_auto_commit`); a stdlib-only leaf owning the roadmap-state
  pathspec so ONE list drives the commit, the tree-digest exclusion, and the
  freshness revision-range exemption; `roadmap backlog [--stale|--older-than|
  --json]` aged on the existing `deferredAt`; a completion nudge; `roadmap
  resolve` (semantic Backlog union + regenerate, refuse on real divergence);
  canonical one-object-per-line rendering for every phase.
- **Out:** a `merge=union` gitattribute or a custom merge driver (analysed and
  rejected — see Risks/D3 in the plan); a JSONL Backlog sidecar (deferred); full
  semantic 3-way merge of schedulable phases (v1 refuses); auto-committing
  anything but roadmap state; changing `complete`'s `git add -A` step commit;
  changing the `roadmap_drift` gate or its severity; a config knob for the
  staleness threshold.

#### Dependencies & Assumptions

- `Feature.DeferredAt` already exists and is populated on **106/106** current
  Backlog findings — the aging clock needs no schema change.
- `roadmap.RenderMarkdown` is the drift gate's own oracle; folding it into
  mutations cannot disagree with the gate.
- `treestate.Digest` already excludes `.workflow/` (D3a), so `roadmap.json`
  churn is invisible today; only `ROADMAP.md` and the moved HEAD are new.
- `VerificationFresh` is already handed a `treestate.Runner` git seam — the
  revision-range check needs no new dependency or new I/O path.
- `internal/roadmap` is unmapped in the `import_graph` layer config; the new
  `internal/roadmapstate` leaf must be added to the leaf layer + PROJECT.md G2.
- `internal/roadmap` imports `internal/workflow`, so `treestate → roadmap` would
  cycle — hence a shared leaf rather than exporting the path list from `roadmap`.
- Builds on: roadmap-doc-sync (drift gate), roadmap-crud/edit-move/phase-ops (the
  mutation surface), deferred-findings-roadmap-capture (defer + `deferredAt`),
  adversarial-validate-verifier (the stamp/freshness contract).
- Assumes `git commit -- <pathspec>` partial-commit semantics (leaves the index
  for other paths untouched) and that it refuses mid-merge — both asserted in
  tests against a real temp repo, never assumed from documentation.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Freshness exemption becomes a hole (real change hidden in ROADMAP.md) | High | Low | Exemption is a strict subset test over the whole range diff, not a path filter; the drift gate byte-compares ROADMAP.md every validate |
| Auto-commit fires inside a conflicted/rebasing tree | Medium | Medium | Explicit merge/rebase detection; best-effort failure policy; pathspec commit never touches other paths |
| `roadmap resolve` silently drops a finding | High | Low | Union-by-slug with an explicit one-sided-deletion rule, per-direction tests, and per-side counts printed so the arithmetic is visible |
| Canonical re-render (S6) collides with every in-flight branch | Medium | High | Land S6 last, right after `resolve` can absorb it semantically; flag in CHANGELOG |
| `--no-verify` normalizes hook bypass | Medium | Low | Scoped to this one commit path (generated state only, drift clean by construction); no user-facing flag |
| `promote`'s extra artifact writes drift from the declared pathspec | Medium | Medium | Pathspec is supplied by the mutation and asserted against the files promote actually writes |
| Parallel Phase 13 merges break main invisibly (standing repo lesson) | Medium | High | Rebase before validate, re-run the full gate suite on the merged tree, keep aggregate coverage ≥97% |

#### Rollout

- **S1** `internal/roadmapstate` leaf (pathspec + commit-message composer) +
  `internal/gitutil` commit primitives. No behavior change.
- **S2** The mutation choke point: regenerate + pathspec commit wired into all
  ten mutating commands (the roadmap entry's literal contract).
- **S3** Freshness/digest exemption for roadmap-state paths — must land with S2
  so auto-commit can never stale a stamp.
- **S4** `roadmap backlog [--stale]` + the completion nudge.
- **S5** `centinela roadmap resolve` — the conflict procedure, mechanized.
- **S6** Canonical one-object-per-line render for every phase; closes the
  existing Backlog finding `rawio-reformat-diff-churn`.
- If scope must shrink: drop S6, never S5.

#### Behavior Summary

After this feature, any `centinela roadmap <mutation>` writes roadmap.json
atomically, immediately re-renders `ROADMAP.md` from it in the same process, and
then makes exactly one conventional commit containing only the roadmap-state
paths — leaving every other staged or dirty file exactly as it was, committing to
the current worktree's own branch, skipping the commit (with a stated reason)
when `disable_auto_commit` is set, a merge or rebase is in progress, or git is
unusable, and never failing the mutation because of the commit. That roadmap-state
churn — committed or not — is exempt from the verification freshness comparison,
so a verifier can record deferrals before or after stamping without voiding its
own verdict, while any non-roadmap path in the same range still stales it.
`centinela roadmap backlog` gives the Backlog an age (from the `deferredAt` each
finding already carries), `--stale` filters to findings older than 14 days, and
completing the last schedulable feature prints how much deferred debt remains and
the two commands that act on it. When two branches do collide on roadmap state,
`centinela roadmap resolve` performs the merge the operators were doing by hand —
union the Backlog findings from both sides, regenerate ROADMAP.md from the result,
stage both — and refuses, with markers intact, when the divergence is real.

#### Gherkin Scenarios

All scenarios live in `specs/roadmap-state-hygiene.feature` (14 ACs, happy and
negative path for each). Headline shapes:

- *a deferral regenerates ROADMAP.md and commits roadmap state by itself* —
  Given a governed repo with ROADMAP.md in sync, When `roadmap defer …`, Then
  exit 0 AND the drift gate is clean without `roadmap generate` AND exactly one
  commit whose changed paths are exactly the two roadmap-state files.
- *unrelated staged and unstaged changes survive the mutation commit* — Given a
  staged Go file and a dirty README, When a mutation runs, Then neither is in the
  commit and both retain their prior index/worktree state.
- *a hostile git environment warns but never fails the mutation* (Scenario
  Outline over: no repo, no HEAD, merge in progress, rebase in progress, commit
  fails) — Then exit 0, state correct on disk, warning names the reason.
- *a deferral after the verification stamp does not stale the verification* vs
  *a source change committed after the stamp still stales the verification* —
  the exemption is proved in both directions, plus fail-closed on a git error.
- *resolve unions Backlog findings from both sides* vs *resolve refuses when a
  schedulable phase diverged on both sides* (non-zero exit, markers intact,
  nothing staged).
- *a finding with no deferredAt is reported as unknown age and counted stale*.

#### UX States

| State | Trigger | Surface |
|-------|---------|---------|
| loading | n/a — every command is synchronous and local | n/a |
| empty | Backlog absent or empty | `roadmap backlog`: explicit "no deferred findings", exit 0; no nudge |
| error | `resolve` on a real phase divergence or an unparseable side | non-zero exit, message names the phase/side, conflict markers untouched, nothing staged |
| degraded | commit skipped (policy, merge in progress, no git) | mutation still succeeds; one stderr line naming the reason |
| success | mutation committed | `Deferred "x" to the Backlog.` + `roadmap state committed (.workflow/roadmap.json, ROADMAP.md)` |

Surface is declared `internal` in the brief, matching every prior CLI feature in
this repo (`deferred-findings-roadmap-capture`, `docstring-gate`, …); the only
human-visible surface is CLI text, English-only (i18n gate disabled per
PROJECT.md → Locales).

#### Out-of-Scope

- `.gitattributes merge=union` / a custom git merge driver (rejected with reasons
  in the plan; both either corrupt JSON or fail silently when uninstalled).
- Moving the Backlog into a JSONL sidecar (deferred: `roadmap-backlog-jsonl-sidecar`).
- Full semantic 3-way merge of schedulable phases (deferred:
  `roadmap-resolve-schedulable-phase-merge`).
- A config knob for the staleness threshold (deferred:
  `roadmap-backlog-stale-threshold-config`).
- Auto-committing non-roadmap state; altering `complete`'s step commit; altering
  the drift gate's checks or severity; squashing the `chore(roadmap)` commits.

#### Deferred Findings

Recorded with `--source roadmap-state-hygiene/planner`:

- `roadmap-backlog-jsonl-sidecar`
- `roadmap-resolve-schedulable-phase-merge`
- `roadmap-backlog-stale-threshold-config`
- `roadmap-state-commit-squash-on-deliver`

Deliberately NOT deferred (already in the Backlog, folded into this feature's
S6): `rawio-reformat-diff-churn`. Adjacent and left alone:
`summary-digest-hashes-nonschedulable-phases`, `gitignore-durable-state-guard`.

#### Handoff

- Next role: **senior-engineer**.
- Outstanding questions: (1) confirm empirically, in a temp repo, that
  `git commit -- <pathspec>` preserves an unrelated staged path's index entry on
  this git version before building S2 on it; (2) `promote`'s exact written-file
  set must be read from the code, not from this plan, when wiring its pathspec;
  (3) S6's normalization commit should be produced and landed alone, after S5.
