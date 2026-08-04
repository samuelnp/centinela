# Plan — roadmap-state-hygiene

> Brief: [docs/features/roadmap-state-hygiene.md](../features/roadmap-state-hygiene.md).
> Spec: [specs/roadmap-state-hygiene.feature](../../specs/roadmap-state-hygiene.feature).
> Source: retrospective.md WS1.4 + WS6 + failure #10, plus one week of Phase 13
> field evidence (every delivery conflicted on roadmap state; one finding lost).

## Goal

Make roadmap state safe to mutate from many concurrent sessions: every mutation
regenerates its markdown and commits itself on a tight pathspec, that churn stops
voiding verification stamps, conflicts are resolved by a command instead of by
hand, and the Backlog grows an age and a closure prompt.

---

## The five decisions (and why)

### D1 — Which mutations auto-commit; is `generate` folded in?

**Auto-commit: every command that writes `.workflow/roadmap.json`** — `defer`,
`add`, `remove`, `edit` (incl. rename), `move`, `reorder`, `promote` (in-place and
into a phase), `phase add|rename|remove`. **Not** `generate`, `show`, `validate`,
`ready`, `iterate` (writes a gitignored marker), `brownfield` (writes a DRAFT to a
separate path and never touches the canonical file).

**Regeneration is folded into each mutation** (an in-process call to
`RenderMarkdown`, not a subprocess), and `generate` **stays** as a standalone
idempotent command that does **not** commit. The rule to remember:

> A commit is a side effect of *mutating* state, never of *rendering* it.

Justification: `generate` is the repair path — it is the command you run in the
middle of a conflict resolution, in a detached/merging tree where a partial commit
is impossible, and in CI where a surprise commit is wrong. Folding it into
mutations removes the "Remember to sync" footgun at the source; keeping it
standalone preserves the drift gate's remediation and `roadmap resolve`'s
building block.

### D2 — Commit boundary, pathspec, and the three hostile environments

- **Boundary:** one commit per CLI invocation, after the atomic roadmap.json
  write and the ROADMAP.md write both succeed. Never a commit per record.
- **Pathspec:** exactly `.workflow/roadmap.json` and `ROADMAP.md`, via
  `git commit --no-verify -q -m <msg> -- <paths>`. A pathspec commit is a
  *partial commit*: it composes the new tree from HEAD plus those paths only, so
  it neither reads nor disturbs the index for anything else.
  - `promote` additionally rewrites `.workflow/roadmap-analysis.json` /
    `-quality.json` (+ their `.md` companions). Those are roadmap state too, so
    the pathspec is supplied **by the mutation**, not hardcoded per command:
    `roadmapstate.Paths()` (the two always-present files) plus the extra paths the
    mutation reports writing. A mutation that writes a file outside its declared
    pathspec is the exact footgun this feature exists to kill, so the declared
    set is asserted in tests against the files each mutation actually touches.
- **Unrelated staged changes:** untouched, still staged (AC3). This is the whole
  reason for a pathspec commit over `git add -A` + commit (what `complete` does).
- **Worktree vs primary checkout:** the commit is made in the **current** tree
  and lands on its HEAD. Inside `.worktrees/<feature>` that is the feature branch
  — correct: the deferral travels with the branch and reaches main with the
  feature. We deliberately do **not** reach into the primary checkout; writing
  another tree is what produced the Backlog finding
  `merge-pending-marker-dirties-primary-tree`.
- **`disable_auto_commit`:** regeneration still happens; the commit is skipped
  with one line of notice. Regeneration is correctness (the drift gate); the
  commit is VCS policy. The knob governs policy only.
- **Merge/rebase in progress** (`MERGE_HEAD`, `rebase-merge/`, `rebase-apply/`,
  `CHERRY_PICK_HEAD`): skip the commit with a notice — git refuses partial
  commits mid-merge, and a mutation run during a conflict resolution is exactly
  when you do not want a surprise commit.
- **Failure policy:** the commit is best-effort. No git / no repo / no HEAD /
  hook failure / signing failure → warn on stderr, exit 0. A non-zero exit after
  a successful on-disk mutation would invite a retry and a duplicate record.
- **`--no-verify`:** the commit contains only generated state; the one gate that
  governs it (`roadmap_drift`) is clean by construction because we just
  regenerated. Running the full precommit gate suite on every `roadmap defer`
  would make deferring cost a full validate — and could *block* it.

### D3 — Should roadmap.json be structurally merge-friendlier? (the higher-leverage half)

Yes, but not by changing the file's identity. Options considered:

| Option | Verdict |
|---|---|
| `.gitattributes: .workflow/roadmap.json merge=union` | **Rejected.** Union on a JSON array of one-object-per-line records produces `{A}\n{B}` with no separating comma → invalid JSON; a leading-comma render would fix syntax but union also silently merges *semantic* edits (two sides editing one feature) into duplicate entries. Fail-loud-later is still corruption. |
| Custom merge driver (`centinela roadmap merge-driver %O %A %B`) | **Rejected for v1.** Needs per-clone `git config` installation (every worktree, every CI checkout, every contributor); a driver that is not installed fails *silently back to* today's conflict. Same logic as `resolve`, worse delivery. Revisit once `resolve` is proven. |
| Backlog moved to a `merge=union`-safe `.jsonl` sidecar | **Rejected for v1, worth revisiting.** It genuinely eliminates the append-vs-append class, but it is a schema split touching every reader (Load, view, mdgen, validators, promote, brownmap, doctor, migration) and it cannot help `ROADMAP.md`, which conflicts just as often. Deferred as a Backlog finding. |
| **Canonical one-record-per-line for every phase (S6) + semantic `roadmap resolve` (S5)** | **Chosen.** The render change makes each conflict hunk one line instead of a reformatted phase; `resolve` mechanizes the exact procedure humans ran all week, covers `ROADMAP.md` (which no union driver can), needs no per-clone installation, and refuses rather than guesses when the divergence is real. |

Explicitly acknowledged: this **reduces** conflicts rather than eliminating them —
two tail appends still collide. What changes is that resolution becomes one
command with a deterministic, tested, data-preserving rule instead of a manual
replay that has already dropped a finding.

### D4 — `roadmap backlog --stale`

- **Clock:** the `deferredAt` RFC 3339 field records already carry (`internal/
  roadmap/defer.go` sets it; all 106 current findings have it). No schema change
  is needed. Missing/unparseable → `unknown`, sorted first, counted stale — the
  fail-loud direction, so hand-edited entries surface instead of hiding.
- **Threshold:** `--stale` = older than **14 days** (one delivery cadence);
  `--older-than N` overrides. No config knob in v1 (deferred).
- **Output:** oldest-first table `AGE  SLUG  SOURCE  SUMMARY` (summary truncated
  to terminal-friendly width), then a footer:
  `106 findings · 91 older than 14d · oldest 53d (rawio-reformat-diff-churn)`.
  `--json` emits `{"threshold_days":14,"total":106,"stale":91,"findings":[{"slug",
  "summary","source":{"feature","role"},"deferredAt","ageDays"|null,"stale"}]}`
  for Magallanes' Plan page. `--stale` filters; the JSON shape is identical.
- **Nudge trigger:** in `centinela complete`, after a workflow reaches `done`,
  when `FirstIncomplete(roadmap)` reports no incomplete schedulable feature AND
  `BacklogFeatures` is non-empty. Wording:
  ```
  Roadmap complete — 106 deferred findings remain in the Backlog (91 older than 14d).
    Review:   centinela roadmap backlog --stale
    Schedule: centinela roadmap promote <slug> --phase "<phase>"
  ```
  Any error loading the roadmap suppresses the nudge silently: a hint must never
  break `complete`.

### D5 — Interaction with the freshness stamp and `complete`'s auto-commit

The stamp binds a verdict to `(HEAD revision, working-tree digest)`; the digest
already excludes `.workflow/` (D3a: "the verifier's own output must not stale its
own stamp"). Roadmap state is the same *kind* of thing — governance bookkeeping,
not product — but half of it (`ROADMAP.md`) sits at the repo root and the other
half now moves HEAD. So this feature extends D3a symmetrically, from ONE
definition:

- `treestate.Digest` drops entries whose paths are all roadmap-state paths
  (covers the `disable_auto_commit` case, where the files stay dirty).
- `workflow.compareStamp` on a revision mismatch asks git for
  `diff --name-only <recorded>..HEAD`; if every changed path is a roadmap-state
  path, the stamp still holds. Any other path → stale, with today's message.
- Lifecycle statement: **regeneration happens inside the mutation, therefore
  always before any subsequent `artifact stamp`, and the mutation's commit is
  exempt from the freshness comparison either way.** A verifier can record
  deferrals before *or* after stamping and never has to choose between drift and
  freshness. The "skip the regeneration to protect my stamp" workaround is gone.
- `complete`'s `git add -A` step commit is unchanged. Roadmap state is simply
  already committed by then, so the step commit no longer sweeps it in.
- Deliberately NOT done: suppressing regeneration to protect a stamp. A stamp is
  supposed to go stale when the *product* changes; the fix is scoping what counts
  as product, not muting the check.

---

## Slices (rollout order — each independently shippable and testable)

### S1 — The pathspec leaf + git primitives (foundation, no behavior change)
- `internal/roadmapstate/paths.go` (~45) — NEW stdlib-only **leaf**: `Paths()`
  (`.workflow/roadmap.json`, `ROADMAP.md`), `Covers(paths []string) bool`,
  `Message(verb, subject string) string` (conventional `chore(roadmap): …`,
  subject truncated to 60 chars). Rationale for a new leaf (the
  `internal/acceptance` / `internal/worktreepath` precedent): `internal/roadmap`,
  `internal/treestate` and `internal/workflow` all need this list, and
  `treestate → roadmap → workflow → treestate` would be a cycle.
- `internal/gitutil/commit.go` (~85) — `CommitPaths(repo, msg string, paths []string) error`,
  `MergeInProgress(repo string) bool`, `PathsChangedSince(repo, rev string) ([]string, error)`,
  `HasHead(repo string) bool`. Leaf (stdlib + `os/exec`), matching the package's
  existing style.
- `centinela.toml` `[gates.import_graph]` leaf layer += `internal/roadmapstate/**`;
  PROJECT.md G2 gains the one-sentence rationale (scaffold mirror not affected).
- Tests: `internal/roadmapstate/paths_test.go` (~60), `internal/gitutil/commit_test.go`
  (~95, real temp repo: commit-only-pathspec, staged-file preservation, no-HEAD,
  merge-in-progress detection). Negative: `CommitPaths` on a non-repo returns an
  error and writes nothing.

### S2 — The mutation choke point (the named contract)
- `internal/roadmap/statesync.go` (~90) — `type Committer interface { Commit(msg
  string, paths []string) error }`; `Sync(opts SyncOptions) Report` where
  `SyncOptions{Verb, Subject string; ExtraPaths []string; Commit bool; C Committer}`:
  reload → `RenderMarkdown` → write `ROADMAP.md` atomically → skip-if-unchanged →
  commit. Returns a `Report` the caller renders (committed / skipped-why).
- `cmd/centinela/roadmap_sync.go` (~55) — one wrapper
  `syncRoadmapState(verb, subject string, extra ...string)` reading
  `config.Load()` for `DisableAutoCommit` and wiring the gitutil committer; each
  mutating command gains **one** line calling it (thin outer layer preserved).
- Edits (2–4 lines each): `roadmap_defer.go`, `roadmap_add.go`, `roadmap_remove.go`
  (drop the "Remember to sync" string), `roadmap_edit.go`, `roadmap_move.go`,
  `roadmap_reorder.go`, `roadmap_promote.go`, `roadmap_phase_add.go`,
  `roadmap_phase_rename.go`, `roadmap_phase_remove.go`.
- `internal/ui/render_roadmap_sync.go` (~45) — the committed / left-uncommitted /
  skipped lines (presentation stays in ui).
- Tests: `statesync_test.go` (~95: regenerates, no-op → no commit, commit
  disabled → still regenerates, committer error → Report says why, extra paths
  forwarded), `roadmap_sync_test.go` (~70: config knob honored, wrapper never
  returns an error on commit failure). Negative both directions: commit attempted
  exactly when enabled AND content changed; never otherwise.

### S3 — Freshness and digest exemption
- `internal/treestate/digest.go` (edit, ≤100 after) — the `excluded` constant
  becomes an exclusion predicate consulting `roadmapstate.Covers`.
- `internal/workflow/validate_freshness.go` (edit, ~85) + `freshness_range.go`
  (~45) — revision mismatch consults `run(root, "diff", "--name-only",
  recorded+"..HEAD")`; state-only → continue to the digest comparison; anything
  else → today's stale error.
- Tests: `digest_test.go` additions (~35), `freshness_range_test.go` (~90) with a
  stubbed `treestate.Runner`: state-only range → fresh; mixed range → stale;
  empty range → fresh; git error on the range → **stale** (fail closed).

### S4 — `roadmap backlog [--stale]` + the completion nudge
- `internal/roadmap/backlog_age.go` (~85) — `type Aged struct{ Feature; AgeDays int;
  KnownAge, Stale bool }`, `AgeBacklog(r *Roadmap, now time.Time, thresholdDays int) []Aged`
  (oldest-first, unknown ages first) + `BacklogStats`.
- `internal/roadmap/backlog_nudge.go` (~45) — pure `NudgeFor(r *Roadmap, now,
  thresholdDays) (Nudge, bool)`; true only when nothing schedulable is incomplete
  and the Backlog is non-empty.
- `internal/ui/render_backlog.go` (~80) — table + footer + nudge rendering.
- `cmd/centinela/roadmap_backlog.go` (~65) — flags `--stale`, `--older-than`,
  `--json`; `cmd/centinela/roadmap_backlog_json.go` (~45) if the payload pushes
  the file over budget.
- `cmd/centinela/complete.go` (edit, +6 lines, stays ≤100) — print the nudge on
  `done`.
- Tests: `backlog_age_test.go` (~95: ordering, threshold boundary at exactly N
  days, missing/garbage `deferredAt`, empty Backlog), `backlog_nudge_test.go`
  (~70: fires / does not fire when schedulable work remains / Backlog empty /
  nil roadmap), `render_backlog_test.go` (~60), `roadmap_backlog_test.go` (~70:
  `--json` shape, `--stale` filtering, exit 0 on empty).

### S5 — `centinela roadmap resolve`
- `internal/roadmap/resolve.go` (~95) — `Resolve(base, ours, theirs []byte)
  ([]byte, error)`: pure 3-way over parsed docs. Backlog = union by slug (dedupe,
  earliest `deferredAt` wins, ordered by `deferredAt` then name; a slug deleted on
  exactly one side stays deleted). Schedulable phases: take the side that differs
  from base; both differ → `ErrPhaseConflict` naming the phase.
- `internal/roadmap/resolve_stages.go` (~60) — the git-stage reader seam
  (`:1:`/`:2:`/`:3:` blobs) behind an injected runner.
- `cmd/centinela/roadmap_resolve.go` (~75) — detect conflicted paths, call
  Resolve, write, regenerate ROADMAP.md from the result, `git add` both, render a
  summary (`kept N findings: 12 from main, 3 from this branch`). Non-zero exit on
  `ErrPhaseConflict`, leaving markers untouched.
- Tests: `resolve_test.go` (~95: union, dedupe same slug, one-sided delete,
  both-sides phase edit → error, unparseable side → error), `resolve_stages_test.go`
  (~55), `roadmap_resolve_test.go` (~80: no-conflict no-op exit 0, ROADMAP.md-only
  conflict, refusal leaves markers). Negative direction is the headline: a real
  divergence must NOT be auto-merged.

### S6 — Canonical render for every phase (diff-churn kill)
- `internal/roadmap/rawrender.go` (edit, ≤100) — every phase renders through
  `renderDirtyPhase`'s one-object-per-line form, not only mutated ones; untouched
  phases keep their key order because the renderer preserves raw member bytes.
- One-time normalization commit of `.workflow/roadmap.json` (mechanical, produced
  by running any no-op-safe mutation path).
- Closes Backlog finding `rawio-reformat-diff-churn` (fold it in — do not
  re-defer it).
- Tests: `rawrender_canonical_test.go` (~80: second write is byte-identical
  (idempotent), a one-field edit produces a one-line diff, unknown fields survive
  the round trip). Negative: a phase with no `features` key still renders valid
  JSON.

**If scope must shrink:** drop S6 (defer it) — never S5. S1–S4 are the roadmap
entry's literal contract; S5 is the half the field evidence says costs the most.

---

## Reuse (do NOT reimplement)
- `roadmap.RenderMarkdown` (the drift gate's oracle — one renderer, one truth).
- `writeAtomic` in `internal/roadmap/rawio.go` for the ROADMAP.md write.
- `finalizeMutation` — the existing post-mutation choke point in the domain; S2
  hangs off the CLI side of it, not inside it (the domain must not run git).
- `Feature.DeferredAt`, `Source`, `BacklogFeatures`, `FirstIncomplete`,
  `IsBacklogPhaseName` — all already exist.
- `config.WorkflowConfig.DisableAutoCommit`; `ui.RenderSuccess`/`StyleMuted`.
- `treestate.Runner` — the git seam `VerificationFresh` is already handed.

## Risks

| # | Risk | Impact | Likelihood | Mitigation |
|---|------|--------|------------|------------|
| R1 | Freshness exemption becomes a hole: a real change hidden in ROADMAP.md escapes the stamp | High | Low | ROADMAP.md is generated from roadmap.json and the drift gate byte-compares it every validate; the exemption is a strict subset test (`Covers`), not a filter that drops paths from a mixed range |
| R2 | Auto-commit fires inside another session's conflicted/rebasing tree | Medium | Medium | Explicit merge/rebase detection (D2) + best-effort failure policy; pathspec commit never touches other paths |
| R3 | `--no-verify` normalizes bypassing hooks | Medium | Low | Scoped to this one commit path, justified in code comment + docs; no flag exposed to users |
| R4 | S6's whole-file normalization collides with every in-flight branch | Medium | High | Land S6 last, immediately after S5 so `roadmap resolve` can absorb it semantically; announce in CHANGELOG |
| R5 | `resolve` silently drops a finding (the failure it exists to prevent) | High | Low | Union-by-slug with an explicit deletion rule + a test per direction; the summary prints per-side counts so a human sees the arithmetic |
| R6 | Parallel Phase 13 merges break main invisibly (standing repo lesson) | Medium | High | Rebase before validate, re-run the full gate suite on the merged tree, re-check aggregate coverage ≥97% |
| R7 | `promote`'s extra artifact paths get out of sync with the declared pathspec | Medium | Medium | The pathspec is supplied by the mutation and asserted in a test against the files promote actually writes |
| R8 | Commit noise: many `chore(roadmap)` commits on a feature branch | Low | High | Accepted — one per mutation is the price of a clean tree; they squash on PR merge |

## Definition of Done
- All 14 ACs green in `specs/roadmap-state-hygiene.feature` with executable
  acceptance coverage; unit + integration tiers per slice.
- `centinela validate` passes; per-package coverage ≥97% on new packages.
- No file over 100 lines (tests included).
- The "Remember to sync ROADMAP.md" string no longer exists in the codebase.
