# spec-conflict-false-positives — gatekeeper

### Adversarial Verifier Report: spec-conflict-false-positives
**Date:** 2026-07-30
**Status:** WARNING

#### Inputs Read
- `git diff origin/main...HEAD` (28 files; working tree clean at `a258234`) plus
  `git status --porcelain` (clean before I started).
- `specs/parallel-feature-worktrees.feature` — the contract. Hotfix archetype
  (code → tests → validate), so there is no plan step and no
  `docs/features/spec-conflict-false-positives.md`; none was expected.
- The hotfix's roadmap entry in `.workflow/roadmap.json` / `ROADMAP.md`.
- Source under test: `internal/worktree/{spec_conflicts,spec_collect,spec_pairing,spec_parser}.go`
  and the call site `cmd/centinela/merge.go:63-65`; for scope questions also
  `internal/gates/{file_size,file_size_scan,spec_traceability}.go` and
  `internal/ui/render_gates.go`.
- Tests under test: `internal/worktree/spec_conflicts{,_more,_falsepos,_baseline,_format,_collect}_test.go`,
  `internal/worktree/coverage_merge_helpers_test.go`, `cmd/centinela/merge_test.go`,
  `tests/acceptance/parallel_feature_worktrees_test.go`,
  `tests/acceptance/spec_conflict_binary{,_helper}_test.go`.
- `.workflow/spec-conflict-false-positives-{senior-engineer,qa-senior,edge-cases}.md`
  — read only to harvest claims to attack, never treated as evidence.
- Output of the commands I executed myself (below), including end-to-end runs of
  scratch binaries against throwaway repos with real `git worktree` checkouts.
- Contamination note: the delegation prompt carried the slug, the diff base, path
  hints and operational notes — no narrative of the implementation. Not
  contaminated.

#### Analyzed Specs
`specs/parallel-feature-worktrees.feature`, both scenarios this hotfix rewrote or
added:
- "Spec conflict across in-flight worktrees is detected before merging" (rewritten
  to state same-file / same-scenario / both-edited-away-from-main, reported at
  most once)
- "Superseding and identical specs never block a merge" (new)

The other ten scenarios in that file are untouched by this branch.

#### Refutation Attempts
- **Claim attacked:** "The narrowed rule still catches a divergence when main has
  no baseline copy — two features that both ADD the same new spec file."
  **How:** Real repo, real worktrees; `zeta` and `eta` each create
  `specs/newthing.feature` with scenario `fresh` resolved to `outcome A` /
  `outcome B`; main has no such file. `centinela merge zeta`.
  **Result:** could not refute — exit 1, blocked once:
  `newthing.feature (zeta) ↔ newthing.feature (eta) — scenario "fresh"`.
  `bothEdited`'s `!ok` branch is genuinely load-bearing.

- **Claim attacked:** "The production failure shape now passes."
  **How:** Seeded main with `login.feature` plus an `archetypes.feature` carrying
  two companion scenarios sharing one `Given`; gave the merging feature `eta` an
  unrelated code commit and added two idle bystander worktrees holding
  byte-identical spec copies. Ran the same repo shape through a binary built from
  `origin/main` and one built from this branch.
  **Result:** could not refute — `origin/main` binary: exit 1 with 2,425 bytes of
  companion-scenario noise (the reported bug, reproduced). This branch: exit 0,
  merge completes, bystander untouched. The mirror case (`zeta` and `eta`
  diverging on main's `clash` scenario) still blocks at exit 1 in 539 bytes.

- **Claim attacked:** "Reports stay deduped and bounded no matter how many
  divergent pairs exist."
  **How:** Forced 90 distinct `(file, scenario, otherOwner)` divergences — 30
  scenarios in a shared `many.feature` against three other worktrees.
  **Result:** could not refute — total output 1,831 bytes; 10 entries per error
  print plus `… and 80 more`. Cobra prints the error twice, so 20 arrows across
  2 prints, never 90.

- **Claim attacked:** "The unit suite has teeth — it would catch a regression of
  the fix rather than just ratifying it."
  **How:** Five source mutations, each run against
  `go test ./internal/worktree/ ./cmd/centinela/` and byte-exactly reverted.
  **Result:** could not refute for the load-bearing logic. `bothEdited`→`true`
  fails 2 tests; `diverges` ignoring `Then` fails 8; removing the `seen` dedup
  fails `RepeatedScenarioReportedOnce`; removing the format cap fails
  `CapsLongReports`. Removing `if len(merging) == 0 { return nil }` fails nothing,
  but that mutant is semantically equivalent (an empty merging set yields no
  pairings anyway), so it is not a test gap. A sixth mutation
  (`appendDivergences` appending unconditionally) exposed a real, non-blocking
  weakness — Finding 4.

- **Claim attacked:** "The fix did not UNDER-block anything that matters."
  **How:** Two directions. (a) Delete-vs-edit: `eta` edits main's `clash`
  scenario, `zeta` deletes `specs/login.feature`; `centinela merge zeta`. (b)
  Cross-file contradiction: `zeta` adds `zed.feature` and `eta` adds
  `etaf.feature`, both `Given user has account` with opposite `Then`;
  `centinela merge zeta`.
  **Result:** partially refuted, non-blocking. (a) exits 0 and main loses the
  spec — already captured as Backlog `spec-conflict-scenario-deletion-detection`,
  and git raises a modify/delete conflict when the other side merges, so nothing
  is lost silently. (b) exits 0 — the documented trade-off, but the documented
  *backstop* for it does not exist (Finding 2).

- **Claim attacked:** "Detection cannot be silently disarmed."
  **How:** Built the genuine two-worktree divergence (main→`checkout`,
  `zeta`→`dashboard`, `eta`→`onboarding`), then removed `zeta`'s worktree with
  `git worktree remove --force` before merging — the state left by a prior partial
  merge, `--force-remove`, or raw-PR delivery. Also ran the variant where the
  worktree exists but its `specs/` directory was deleted.
  **Result:** REFUTED. Both variants exit 0 and merge unblocked; the `origin/main`
  binary blocked the same shape. Finding 1.

- **Claim attacked:** "`centinela validate` and the full suite pass on this tree."
  **How:** Ran `centinela validate` twice and `go test ./...` once. Run 1 of
  validate FAILED (exit 1,
  `TestAcceptance_SpecConflict_TwoWorktreesDivergeBlocksRealMergeOnce`) — but that
  run overlapped my own mutation experiments, and the mutation live at that moment
  was `diverges` ignoring `Then`, which is exactly the axis that acceptance test
  diverges on. I reverted every mutation, confirmed `git status --porcelain`
  clean, reproduced that test's exact repo shape by hand with a freshly rebuilt
  binary (blocked correctly, exit 1, 539 bytes), then re-ran both commands
  serially with nothing else touching the tree.
  **Result:** could not refute — `go test ./...` exit 0 (344.0 s; acceptance tier
  341.9 s), `centinela validate` exit 0, "All gates passed" (338.3 s). Run 1's
  failure was verifier-induced contamination, not a defect; it is recorded in the
  verification block rather than hidden.

- **Claim attacked:** "The two rewritten `.feature` scenarios have honest
  executing markers."
  **How:** Extracted all 12 scenario names from
  `specs/parallel-feature-worktrees.feature` and every
  `// Acceptance: specs/parallel-feature-worktrees.feature` + `// Scenario:`
  marker pair under `tests/acceptance/`.
  **Result:** could not refute — "Spec conflict across in-flight worktrees is
  detected before merging" and "Superseding and identical specs never block a
  merge" both match exactly, each at two tiers (function-driven and
  binary-driven), and all four tests executed inside the passing acceptance
  package. The `spec-traceability-gate` warning is attributable entirely to nine
  pre-existing uncovered scenarios in that same file, not to this hotfix.

- **Claim attacked:** "G1 is satisfied for everything this branch touched."
  **How:** `wc -l` over every touched source and test file; read
  `internal/gates/file_size.go`.
  **Result:** could not refute. Every touched file under `internal/` and `cmd/` is
  ≤ 98 lines. `tests/acceptance/parallel_feature_worktrees_test.go` is 186 lines,
  but `tests/` is outside G1's `sourceRoots` (`src internal cmd lib app pkg`) and
  the file was already 140 lines on `origin/main` — pre-existing convention, not a
  regression introduced here.

#### Commands Run
All from `/Users/samuelnp/projects/personal/centinela/.worktrees/spec-conflict-false-positives`
at `a258234`. Scratch binary `/tmp/centinela-verify-scfp` built from
`./cmd/centinela`; baseline binary `/tmp/centinela-old-scfp` built from a
`git archive origin/main` export.

- `go build -o /tmp/centinela-verify-scfp ./cmd/centinela` — exit 0 (untimed).
- `go build -o /tmp/centinela-old-scfp ./cmd/centinela` (origin/main export) — exit 0 (untimed).
- `centinela validate` (run 1) — exit 1, 350,015 ms. **Invalid:** overlapped my own
  in-tree source mutations; recorded for honesty, not relied on.
- `go test ./...` (clean tree, serial) — exit 0, 344,009 ms.
- `centinela validate` (run 2, clean tree, serial) — exit 0, 338,331 ms, "All gates passed".
- Six end-to-end scenarios driven through `/tmp/centinela-verify-scfp merge <feature>`
  in throwaway `git init` repos with real `git worktree add` checkouts (new-file
  divergence, delete-vs-edit, missing merging `specs/`, removed merging worktree,
  90-pair report bounding, production shape), plus two runs of
  `/tmp/centinela-old-scfp merge <feature>` for the before/after comparison — exit
  codes and byte counts quoted inline above. Described in prose because each is a
  multi-step shell sequence, not a single argv.
- Six source mutations of `internal/worktree/spec_pairing.go` /
  `spec_conflicts.go`, each followed by `go test ./internal/worktree/ ./cmd/centinela/`
  and a byte-exact revert verified with `git status --porcelain` — results in prose
  above.
- `centinela roadmap defer spec-conflict-precheck-requires-merging-worktree --summary "…" --source spec-conflict-false-positives/gatekeeper` — exit 0.
- `centinela roadmap defer spec-conflict-cross-file-contradiction-unbacked --summary "…" --source spec-conflict-false-positives/gatekeeper` — exit 0.
- `centinela roadmap generate` — exit 0 (keeps `ROADMAP.md` in sync with my two
  deferrals; the hotfix archetype has no docs step in which to clear drift later).
- `centinela evidence init spec-conflict-false-positives gatekeeper` — exit 0.
- `centinela artifact new spec-conflict-false-positives gatekeeper` — exit 1
  ("artifact already exists"). `evidence init` had already written the companion
  skeleton, which I then overwrote with this report. `--force` deliberately not used.
- `centinela artifact stamp spec-conflict-false-positives` — last action.

#### Findings
- **Affected spec:** `specs/parallel-feature-worktrees.feature`
  **Affected scenario:** Spec conflict across in-flight worktrees is detected before merging
  **Risk (MEDIUM):** The pre-check is silently disarmed whenever
  `.worktrees/<merging-feature>/specs` is absent. `DetectSpecConflicts`
  (`internal/worktree/spec_conflicts.go:37-40`) reads the merging worktree first
  and returns `nil` before touching any other worktree. Demonstrated end to end:
  main holds `clash → checkout`, `zeta` resolves it to `dashboard`, `eta` to
  `onboarding` — a textbook true positive — but after
  `git worktree remove --force .worktrees/zeta`, `centinela merge zeta` exits 0
  and merges with no block at all. Same result when the worktree exists but its
  `specs/` directory does not. That state is routine: a prior partial merge,
  `--force-remove`, or raw-PR delivery all leave the branch without its worktree.
  The `origin/main` binary blocked this shape, so it is a narrow behavioural
  regression — though the old block fired on companion-scenario noise rather than
  on the real divergence, and git still raises a text conflict when `eta` later
  merges, so nothing is lost silently. The edge-cases report documents this as an
  accepted residual risk; the qa-senior report asserts it "is already captured as
  a Backlog item", which was true of neither existing entry.
  **Suggestion:** Fall back to the merging branch's committed tree
  (`git show <feature>:specs/…`) when the worktree is gone, or say so loudly
  ("pre-check skipped — no worktree for `<feature>`") instead of returning `nil`
  silently. Deferred as `spec-conflict-precheck-requires-merging-worktree`.

- **Affected spec:** `specs/parallel-feature-worktrees.feature`
  **Affected scenario:** Superseding and identical specs never block a merge
  **Risk (LOW-MEDIUM):** The documented backstop for the dropped cross-file
  contradiction detection does not exist. Both the senior-engineer report ("the
  merged-tree `centinela validate` run remains the real semantic gate") and the
  edge-cases Residual Risks repeat this. Verified otherwise: no gate compares
  Gherkin semantics across files (`grep -rni contradict internal/ scripts/` → 0
  hits; the gate set is file_size, cross-compile, import_graph, spec_traceability,
  i18n, custom commands, roadmap_drift; `validate.commands` is `go test` +
  `check-coverage.sh` + `check-fmt.sh`). Proven end to end: `zeta` adds
  `zed.feature` and `eta` adds `etaf.feature`, both asserting opposite outcomes
  for `Given user has account`; `centinela merge zeta` exits 0 and the merged
  tree's validate passes. The trade-off itself is defensible — the old
  Given-bucketing pairer was the entire false-positive source — but the
  justification overstates the safety net that remains.
  **Suggestion:** Amend the trade-off text to state the class is accepted as
  undetected, or add a real cross-spec check. Deferred as
  `spec-conflict-cross-file-contradiction-unbacked`.

- **Affected spec:** `specs/parallel-feature-worktrees.feature`
  **Affected scenario:** (test-suite accuracy)
  **Risk (LOW):** `internal/worktree/spec_conflicts_collect_test.go:TestDetectSpecConflicts_IncompleteScenarioIgnored`
  contradicts itself. Its name says "Ignored" and its doc comment says "Scenarios
  missing a Given or Then are incomplete and never compared", but the body asserts
  `len(got) == 1` — a missing `Then` IS reported as a divergence. The edge-cases
  scenario-to-test map cites this test under the same "incomplete scenario
  ignored" heading. The behaviour is a deliberate change (the deleted
  `scenariosConflicts` skipped records with empty fields; `diverges` does not), but
  three artifacts now describe it backwards, which will mislead the next reader of
  this detector.
  **Suggestion:** Rename to `…_MissingThenIsDivergence`, fix the doc comment and
  the edge-cases heading. No code change needed.

- **Affected spec:** `specs/parallel-feature-worktrees.feature`
  **Affected scenario:** Superseding and identical specs never block a merge
  **Risk (LOW):** Four of the regression tests pass vacuously. Under a mutation
  that makes `appendDivergences` append a conflict unconditionally, only three
  `_NoFlag` tests fail. The survivors never reach the pairing logic they claim to
  exercise: `TestDetectSpecConflicts_SameGivenSameThen_NoFlag` (merges "alpha") and
  `TestDetectSpecConflicts_SameOwnerIsNotConflict` (merges "solo") hit the
  empty-merging early return because neither feature has a worktree, while
  `TestDetectSpecConflicts_CrossFileSameGivenOnMain_NoFlag` and
  `TestDetectSpecConflicts_DifferentFilesDifferentOwners_NoFlag` have no other
  worktree to compare against at all, so the loop body never runs. The shipped
  behaviour is fine — the genuinely load-bearing false-positive classes (bystander
  baseline, merging baseline, identical copies + companion scenarios) ARE covered
  by tests with teeth — but four entries in the claimed regression matrix carry no
  signal.
  **Suggestion:** Give each of the four a merging worktree AND at least one other
  worktree so the assertion exercises the intended branch.

- **Affected spec:** (repo hygiene — pre-existing, not this branch)
  **Risk (LOW):** `ui.RenderGateResult` renders `Details` only for `Fail`, never
  for `Warn` (`internal/ui/render_gates.go:21-22`), so both warning gates print a
  dangling `Packages match no configured layer:` / `Scenarios without acceptance
  coverage:` with every item suppressed. I had to re-derive the traceability gaps
  by hand to confirm they were pre-existing. Not caused by this hotfix and not
  deferred under it.
  **Suggestion:** Render `Details` for `Warn` as well as `Fail`.

#### Deferred Findings
- `spec-conflict-precheck-requires-merging-worktree` — recorded via
  `centinela roadmap defer spec-conflict-precheck-requires-merging-worktree --summary "…" --source spec-conflict-false-positives/gatekeeper`
  (exit 0).
- `spec-conflict-cross-file-contradiction-unbacked` — recorded via
  `centinela roadmap defer spec-conflict-cross-file-contradiction-unbacked --summary "…" --source spec-conflict-false-positives/gatekeeper`
  (exit 0).
- `ROADMAP.md` regenerated with `centinela roadmap generate` so `roadmap_drift`
  stays green — the hotfix archetype (code → tests → validate) has no docs step in
  which to clear it later.
- Not re-deferred (already in the Backlog, verified present):
  `spec-conflict-scenario-deletion-detection`, `spec-conflict-deep-gherkin-diff`.

#### Recommendation
WARNING — proceed, with the two deferrals tracked.

The hotfix does what it claims where it claims it. I reproduced the reported
production failure against an `origin/main` binary and watched this branch clear
it; the true positives still block, once each; the baseline rule is real and
mutation-proved, including the no-baseline branch; the report is deduped at
detection and capped at format, holding 90 divergences to 1.8 KB;
`centinela validate` and `go test ./...` both pass on a clean tree; and both
rewritten `.feature` scenarios have exact, executing acceptance markers at two
tiers.

It is not SAFE because one refutation landed: the pre-check is silently disarmed
whenever the merging feature's worktree is gone, and a genuine two-worktree
divergence then merges unblocked — a shape `origin/main` did block. It is not
CRITICAL because git still raises a text conflict on the other side's merge, so
nothing is lost silently, and because the remaining findings are documentation and
test-quality issues rather than behavioural defects. Fix the disarm case in the
deferred follow-up, and correct the two artifact claims (the non-existent validate
backstop, and the inverted "incomplete scenario ignored" test name) before they
mislead the next reader.

```json centinela:verification
{
  "revision": "a2582340c6613a72b494d9bcfac470ab1e7fa95b",
  "treeDigest": "sha256:1b35c190260c605c0f69a9d6b3d86fbdc366fac7f6a1ae684ca456880d8924ae",
  "commands": [
    {"argv": ["centinela", "validate"], "exitCode": 1, "durationMs": 350015},
    {"argv": ["go", "test", "./..."], "exitCode": 0, "durationMs": 344009},
    {"argv": ["centinela", "validate"], "exitCode": 0, "durationMs": 338331}
  ]
}
```
