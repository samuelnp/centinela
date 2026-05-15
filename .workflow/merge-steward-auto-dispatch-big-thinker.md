### Big-Thinker Report: merge-steward-auto-dispatch

**Date:** 2026-05-15

#### Problem

The `parallel-feature-worktrees` feature shipped the *scaffolding* for
automated merges — `centinela merge <feature>`, the `merge-steward`
prompt, the evidence-contract entry, and the `RoleMergeSteward`
validator branch in `internal/orchestration/output_rules.go` — but not
the *behaviour*. Today `runMerge` in `cmd/centinela/merge.go` detects a
git text conflict or a post-merge `centinela validate` failure, prints
`ui.RenderStep("Merge Steward required", …)`, and returns an error
pointing at `outcome.StewardHint()`. That is the end of the line: the
Go binary cannot call an LLM, nothing dispatches the Steward, nothing
ever reads `.workflow/<feature>-merge-steward.json`, and there is no
command that can finalize a merge once a human/agent has resolved the
conflict. The operator hurts because they must (a) notice the printed
hint, (b) manually hand the merge-steward prompt to an agent with the
right inputs, and (c) hand-finish the git merge themselves — exactly
the manual, conflict-prone flow the parent feature set out to remove.
Now is the right time because the contract and validator already exist
and are unused; this feature is the thin layer that activates them.

#### Scope

- **In:**
  - On conflict/validate-failure, `centinela merge` writes a pending
    marker `.workflow/<feature>-merge-pending.json` and prints a
    structured `CENTINELA DIRECTIVE:` that tells the orchestrator to
    invoke the merge-steward subagent (prompt path + the exact inputs
    from `merge-steward-prompt.md`) and resume with `centinela merge
    --continue <feature>`.
  - A new UserPromptSubmit hook (`centinela hook merge`) that
    re-emits that directive every prompt while a pending marker exists
    without valid steward evidence — so the dispatch survives across
    turns and is not a one-shot print.
  - `centinela merge --continue <feature>` reads + validates the
    steward evidence through the existing orchestration validator and
    *gates finalization on it*: APPLY/`handoffTo:complete` → finalize
    (remove worktree, clear marker); ESCALATE/`handoffTo:user`/invalid
    → stay blocked, print escalation note + proposed diff to stderr,
    exit non-zero, keep worktree + marker.
  - State logic in `internal/worktree/`; cmd layer stays thin (G7).
- **Out:**
  - Resolving arbitrary semantic conflicts perfectly — the Steward
    proposes a diff, humans approve. No auto-apply without an explicit
    APPLY + complete handoff.
  - Multi-feature merge trains / queued merges.
  - Any GUI or non-CLI surface.
  - Changing the merge-steward prompt's analysis contract (only its
    invocation path changes).

#### Dependencies & Assumptions

- Builds directly on `parallel-feature-worktrees`:
  `internal/worktree/merger.go` (`MergeOutcome`, `Merge`),
  `steward.go` (`StewardHint`, `StewardJSONPath`, `StewardReason`),
  `merger_git.go` (`isDirty`, `parseConflictedPaths`).
- Reuses the orchestration evidence machinery untouched:
  `orchestration.ValidateEvidence(path, feature, "merge",
  RoleMergeSteward, nil)`, `output_rules.go` merge-steward branch,
  `evidence-contract.md` merge-steward entry.
- Reuses the directive/evidence pattern proven in `hook_setup.go`,
  `hook_context.go`, `hook_orchestration.go`, and the hook
  registration path in `internal/setup/hooks.go` (`ensurePrompt`).
- Assumption: the orchestrating Claude session reads stdout
  `CENTINELA DIRECTIVE:` lines and acts on them (the entire Centinela
  control model already depends on this).
- Assumption: the merge-steward prompt at
  `docs/architecture/merge-steward-prompt.md` already documents the
  ESCALATE/APPLY contract and the JSON output; auto-dispatch must not
  weaken it.
- Layering (PROJECT.md G2/G7): worktree layer must not import
  `internal/orchestration`; the cmd layer passes an evidence-validator
  adapter function down into `internal/worktree`.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Wrong dispatch mechanism (Go tries to "spawn" an agent) | High | Low | Decision locked: directive + state-file + `--continue`, mirroring existing hooks; no LLM call in-binary. |
| Silent auto-resolution bypasses the parent escalation contract | High | Low | `--continue` finalizes ONLY on APPLY + `handoffTo:complete`; ESCALATE/invalid keep worktree, exit non-zero, print note+diff to stderr. Unit + acceptance tests assert the block. |
| Stale/forged evidence treated as resolution | High | Low | `--continue` re-runs `ValidateEvidence` (mismatched feature/step/role, bad RFC3339, missing report all reject) and re-checks clean tree before finalizing. |
| Directive printed once, lost across turns | Medium | Medium | UserPromptSubmit hook re-emits while pending marker exists and evidence absent/invalid (same idempotent pattern as `hook_setup`). |
| Layer violation (worktree → orchestration import) | Medium | Medium | Inject a `validateEvidence func(path,…) error` adapter from cmd; worktree stays leaf-clean. Gatekeeper checks G2. |
| New hook regresses other directive hooks (ordering/noise) | Medium | Low | Add via `ensurePrompt` like the rest; hook no-ops when no pending marker; integration test covers presence/absence. |
| `centinela merge` exit-code change breaks CI expectations | Low | Low | Conflict already exits non-zero today; behaviour preserved. `--continue` ESCALATE also non-zero per contract. |
| File-size (G1) creep splitting worktree logic | Low | Medium | Pre-planned file split: `pending.go`, `finalize.go`, additions to `steward.go`; all ≤100 lines. |

#### Rollout

- Step 1 (smallest correct slice): `internal/worktree/pending.go` —
  write/load/clear the pending marker + `PendingPath`. Pure, unit-
  tested, no behaviour change yet.
- Step 2: `StewardDirective(MergeOutcome)` in `steward.go` (the exact
  directive string) and `internal/worktree/finalize.go`
  (`ResolveMerge` with injected evidence validator + clean-tree
  re-check, APPLY vs ESCALATE classification).
- Step 3: wire `cmd/centinela/merge.go` — write pending + print
  directive on conflict; add `--continue` calling `ResolveMerge` with
  the `orchestration.ValidateEvidence` adapter; render APPLY success
  vs ESCALATE block.
- Step 4: `cmd/centinela/hook_merge.go` UserPromptSubmit hook +
  register in `internal/setup/hooks.go`; `internal/ui/render_merge.go`.
- Step 5: tests (unit → integration → acceptance) and edge-case doc.
- Step 6 (docs step, can wait): refresh `merge-steward-prompt.md`,
  `evidence-contract.md` note, `workflow-enforcement.md`, README.

Deferrable: a `centinela merge --status` convenience view and any
queueing of multiple pending merges — not needed for v1 correctness.

#### Handoff

- Next role: feature-specialist
- Outstanding questions:
  - Confirm `--continue` is preferred over a separate `centinela merge
    finalize <feature>` subcommand (lean: `--continue` keeps one
    command, matches `git rebase --continue` mental model).
  - Confirm the pending-marker JSON shape (reason, conflictedPaths,
    worktreePath, generatedAt) is sufficient for the Steward prompt's
    stdin expectations, or whether conflicted-path payload should also
    be passed on the directive line.
  - Confirm whether a spec/contract conflict (third Steward class) is
    in scope here — it is currently blocked *before* `Merge` runs by
    `DetectSpecConflicts`, so it never reaches the pending path; treat
    as out of scope for auto-dispatch v1.
