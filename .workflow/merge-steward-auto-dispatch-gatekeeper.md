### Gatekeeper Report: merge-steward-auto-dispatch
**Date:** 2026-05-20
**Status:** SAFE

#### Analyzed Specs

Reviewed every `.feature` file under `specs/` (61 files in total). The closest neighbours — and the only ones that touch the merge surface — are:

- `specs/parallel-feature-worktrees.feature` (parent feature; shipped the merge command and the Merge Steward seam)
- `specs/merge-steward-auto-dispatch.feature` (new spec under review)
- `specs/add-agent-evidence-contract.feature`, `specs/enforce-actionable-orchestration-evidence.feature`, `specs/promote-orchestration-agents.feature` (orchestration evidence contract — re-validated for any merge-steward role drift)
- `specs/diff-aware-gatekeeper.feature` (validate gate that runs from `centinela validate`, which the post-merge path invokes)

All other specs sit in unrelated surfaces (init/setup, plan-advisor, status, docs, opencode, semver release, etc.) and were inspected only to confirm they neither reference `centinela merge` nor depend on the hook list shape.

Domain code reviewed per `PROJECT.md` → Gatekeeper Paths:

- `internal/worktree/` — `merge_pending.go`, `finalize.go`, `steward.go`, `merger.go` (+ their tests).
- `internal/orchestration/` — `policy.go` (`RoleMergeSteward`, `RequiredRoles`), `evidence.go`, `output_rules.go` (steward output-shape rule).
- `internal/setup/hooks.go` and `hooks_test.go` (registered `centinela hook merge`, hook-count assertion updated 6 → 7).
- `cmd/centinela/` — `merge.go`, `merge_dispatch.go`, `merge_continue.go`, `merge_evidence.go`, `hook_merge.go` and their tests.

#### Findings

- **Affected spec:** `specs/parallel-feature-worktrees.feature`
  **Affected scenario:** *Text conflict invokes the Merge Steward* / *Semantic conflict after a clean text merge invokes the Steward* / *Merge Steward escalates uncertain resolutions to the user*
  **Risk:** The parent spec asserts that `centinela merge` "invokes the Merge Steward agent" and that an evidence file is written to `.workflow/<feature>-merge-steward.json`. The new feature does **not** synchronously invoke the agent — it now writes a pending marker, prints a CENTINELA DIRECTIVE, exits non-zero, and waits for the orchestrator session to run the steward and the user to call `centinela merge --continue`.
  **Suggestion:** This is a refinement of the same contract, not a contradiction: the parent scenarios still hold once the directive-driven dispatch completes (worktree kept, steward evidence ultimately written, escalation flow surfaces a proposed diff without modifying main). No parent scenario was assertively pinned to a synchronous invocation. The senior-engineer step also updated the only test surface the parent feature locked to a specific error message (the "steward review" hint), and `merge_steward_test.go` was tightened to additionally assert the new pending marker. No stale tests were found; treat the parent spec scenarios as covered by the new spec's more precise wording. Documentation step should consider adding a cross-reference between the two `.feature` files, but that is not a gate.

- **Affected spec:** `specs/add-agent-evidence-contract.feature`, `specs/enforce-actionable-orchestration-evidence.feature`
  **Affected scenario:** Merge Steward evidence shape rules (out-of-band role)
  **Risk:** The new `centinela merge --continue` flow injects `orchestration.ValidateEvidence(..., RoleMergeSteward, ...)` from `cmd/centinela/merge_evidence.go`. If the evidence contract had drifted, the gate would either over-block (refuse APPLY) or under-block (accept malformed evidence).
  **Suggestion:** Verified `RoleMergeSteward` is still declared in `internal/orchestration/policy.go`, its actionable-output rule (`.workflow/<feature>-merge-steward.md` required) is unchanged in `output_rules.go`, and the evidence-contract doc (`docs/architecture/evidence-contract.md`) already documents `step:"merge"` + `handoffTo:"complete"|"user"` for this role and is bit-identical to its scaffold mirror. No drift; gate remains correctly wired.

- **Affected spec:** `specs/merge-steward-auto-dispatch.feature` (self)
  **Affected scenario:** *The hook stops re-emitting once valid steward evidence is present*
  **Risk:** When valid ESCALATE evidence is on disk, the hook deliberately falls silent (treating it as a resolved-by-evidence state). An operator who only reads UserPromptSubmit hook output could miss the escalation.
  **Suggestion:** Accounted-for: `centinela merge --continue` re-surfaces the escalation loudly (stderr panel via `RenderMergeEscalated` + non-zero exit), and the original CENTINELA DIRECTIVE printed during dispatch already names `centinela merge --continue <feature>` as the resume path. Not a gate violation; the spec scenario explicitly approves this behavior.

- **Affected spec:** `specs/diff-aware-gatekeeper.feature`
  **Affected scenario:** Built-in validate gates (file size, coverage)
  **Risk:** The post-merge validate path calls into the same gate runner that diff-aware-gatekeeper governs.
  **Suggestion:** No new gate added, no gate signature changed. `centinela validate` was re-run on the feature branch and passes with exit 0 (G1 file-size diff-aware skip; `go test ./...` green; coverage gate green at 95.1%).

- **Affected file:** `internal/setup/hooks.go` and `internal/setup/hooks_test.go`
  **Affected scenario:** Existing setup/migrate scenarios that assert hook idempotence (e.g. `specs/migrate-full-sync.feature`)
  **Risk:** Adding the `cmdMerge` UserPromptSubmit hook increases the prompt-hook count from 6 to 7. If migrate or wizard tests asserted an exact count without being updated, hook installation would fail.
  **Suggestion:** Verified — only `internal/setup/hooks_test.go` line 10 pinned a numeric expectation, and it was updated from 6 to 7 in the code step. All other setup/migrate tests (`sync_test.go`, `sync_managed_files_test.go`, `migrate_*_test.go`, `settings_test.go`) assert on idempotence and presence of individual commands, not on the total count, so the new hook is additive without breakage. Full suite green.

#### Recommendation

- **SAFE**: No spec contradictions, no broken existing tests, no domain entity or port signature was widened or narrowed in a way that strands existing adapters. The parent-feature contract (`parallel-feature-worktrees.feature`) is refined by this feature rather than contradicted — the new spec encodes the *mechanism* (marker + directive + `--continue`) that the parent spec left abstract ("invokes the Merge Steward"). The lone test pinned to a hook count was updated. The hook-silent-on-ESCALATE behavior is an approved-by-spec decision with a loud surfacing path through `centinela merge --continue`. Proceed.

##### Pre-existing notes (out of scope, not blocking)

- `diff -r docs/architecture internal/scaffold/assets/docs/architecture` shows pre-existing drift in five unrelated files (`gatekeepers.md`, `new-project-guide.md`, `production-readiness-prompt.md`, `testing-strategy.md`, `workflow-enforcement.md`). None of them belong to this feature's surface; `merge-steward-prompt.md` and `evidence-contract.md` (the two docs this feature depends on) are bit-identical to their scaffold mirrors. Recorded for future docs-consistency pass.
