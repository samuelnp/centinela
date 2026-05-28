### Big-Thinker Report: evidence-cli
**Date:** 2026-05-28

#### Problem

Every workflow step (plan, code, tests, validate, docs) and every role
(big-thinker, feature-specialist, senior-engineer, ux-ui-specialist,
qa-senior, validation-specialist, gatekeeper, production-readiness,
documentation-specialist) currently requires the agent to hand-write a
`.workflow/<feature>-<role>.json` + `.md` pair plus templated companions
(`edge-cases.md`, `gatekeeper.md`, `production-readiness.md`,
`documentation-specialist.{md,json}`). There is no typed CLI for any of
this, so agents reach for `python3 -c`, `cat <<EOF` heredocs, and `jq`
one-liners to escape strings, merge fields, and produce JSON that
conforms to the validator's schema in `internal/orchestration/`. The
output is unreadable, frequently invalid, and forces costly human
round-trips. The earlier `add-agent-evidence-contract` feature
documented the schema in prompts but left the authoring burden on the
agent. This feature closes that gap with a typed
`centinela evidence` CLI plus a postwrite formatter and prompt mandates,
so agents never hand-author evidence JSON again.

#### Scope (In / Out for v1)

- **In**
  - `centinela evidence init|set|append|read|validate|repair|schema`
    subcommands covering every role in the contract.
  - Atomic writes (temp-file + rename) with advisory file locking.
  - Reuse of the existing validator in `internal/orchestration/` —
    no rewrite of schema logic.
  - `centinela artifact new` for the templated companions
    (`edge-cases`, `gatekeeper`, `production-readiness`,
    `documentation-specialist`).
  - PostToolUse Write/Edit hook auto-pretty-prints `.workflow/*.json`,
    scoped to the active feature's `.workflow/` prefix
    [[project_worktree_operational_model]].
  - Every agent prompt in `docs/architecture/*-prompt.md` and its
    `internal/scaffold/assets/` mirror mandates the CLI and forbids
    hand-written JSON; parity test extended to cover prompts
    [[project_scaffold_mirror_partial_parity]].
- **Out**
  - Multi-machine distributed locking (single-host advisory `flock`
    is sufficient for v1).
  - Schema migration tooling — schema is bound to the binary via
    `_meta.cli_version`; unknown fields preserved, missing required
    fields rejected.
  - Rewriting the orchestration validator itself (we delegate to it).
  - GUI / TUI inspector for `.workflow/` (text CLI only).
  - Rewriting `centinela start/complete/status/validate` flags or
    workflow state file format.

#### Dependencies & Assumptions

- Hard dependency on the existing validator in
  `internal/orchestration/` (`evidence.go`, `output_rules.go`,
  `plan_snapshot.go`, `evidence_ux.go`) — single source of truth for
  schema rules. The CLI must call into it, not duplicate it.
- Hard dependency on `docs/architecture/evidence-contract.md` as the
  human-readable schema. New role structs must mirror it field-for-field.
- Hard dependency on `internal/hookpolicy/` for the postwrite formatter
  injection point (Slice 2 extends `cmd/centinela/hook_postwrite.go`).
- Hard dependency on `internal/scaffold/assets/docs/architecture/`
  mirror — prompt edits must update both locations or the parity test
  fails.
- Assumes Go `flock` (`golang.org/x/sys/unix`) is acceptable for
  advisory locking on darwin/linux; Windows is not a current target.
- Assumes one feature per worktree under `.worktrees/<feature>/`
  [[project_worktree_operational_model]] — the postwrite formatter
  scopes to the active feature's `.workflow/` prefix only.
- Assumes per-package coverage stays ≥95%
  [[project_coverage_per_package_no_coverpkg]] — every new
  `internal/` and `cmd/` file ships a colocated `_test.go`.
- Assumes the 100-LOC source-file cap applies to test files too
  [[project_g1_applies_to_test_files]] — split tests when they grow.

#### Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| CLI surface growth (8+ roles × 7 verbs) explodes the binary and tests. | High | Medium | Generate per-role command set from Go schema structs; one role = one file ≤100 LOC. |
| Schema drift between `evidence-contract.md`, the new Go structs, and the validator in `internal/orchestration/`. | High | Medium | Delegate validation to the existing validator; add an acceptance test asserting every role declared in the contract has a matching Go struct and CLI verb. |
| Postwrite formatter corrupts other features' `.workflow/` files. | High | Low | Scope path matching to the active feature's prefix; regression test exercises a foreign-feature path and asserts no-op. |
| Agents keep hand-writing JSON anyway. | Medium | Medium | Slice 3 acceptance test scans every prompt and `internal/scaffold/assets/` mirror for `python3 -c`, heredoc patterns, and raw JSON examples. Catches drift in CI. |
| Schema bound to binary version causes friction on upgrades. | Medium | Low | `_meta.cli_version` recorded per file; unknown fields preserved on round-trip; missing required fields rejected with a fix hint pointing at `centinela evidence set`. |
| Concurrent subagents target the same `<feature>-<role>.json` and one clobbers the other. | Medium | Low | Advisory `flock` with short timeout; clear error directs the agent to `centinela evidence read` before retry. |
| `centinela evidence repair` deletes a file an agent is mid-writing. | Medium | Low | Only orphaned temp files (`*.tmp.<pid>` style, mtime > N seconds) are removed; idempotent and safe under retry. |
| Pre-existing `.workflow/` JSON files written before this feature fail under the new validator. | High | Low | AC7 covered explicitly — no breaking schema change; only `_meta` is additive and optional on read. |
| Hook ordering bugs — postwrite formatter runs before another hook expecting raw content. | Medium | Low | Formatter is invariant on a no-op input; runs last in the postwrite chain. |

#### Rollout

The three slices ship in dependency order; only Slice 1 is the blocker
for value. Slices 2 and 3 may split into follow-up features if the
plan-step subagents flag sequencing pressure.

- **Slice 1 — `evidence-cli-core` (BLOCKER for value).**
  Ship the typed CLI: `centinela evidence init|set|append|read|validate|repair|schema`
  with one Go struct per role in `internal/evidence/`. Atomic writes,
  advisory locking, validator delegation to `internal/orchestration/`.
  No prompt changes yet — agents can opt in to use the CLI but the
  old path still works. Unblocks every subsequent role and every other
  slice.
- **Slice 2 — `evidence-cli-artifacts`.**
  `centinela artifact new <feature> <kind>` for the templated
  companions, plus the postwrite formatter in
  `internal/hookpolicy/format_evidence.go` wired into
  `cmd/centinela/hook_postwrite.go`. Active for all features once
  merged; worktree-scoped to avoid touching other features'
  evidence. Depends on Slice 1's IO layer.
- **Slice 3 — `evidence-cli-prompts`.**
  Rewrite every agent prompt in `docs/architecture/*-prompt.md` and
  its `internal/scaffold/assets/` mirror to mandate the CLI; remove
  embedded raw JSON skeletons in favour of
  `centinela evidence schema <role>`. Acceptance test asserts no
  prompt embeds forbidden authoring instructions. Closes the loop —
  agents are pushed onto the CLI and drift is caught in CI.

What can wait: a TUI inspector, multi-machine locking, schema
migration tooling, and Windows support. None are blockers for the
v1 value proposition (kill the heredoc, ship typed evidence).

#### Handoff

- **Next role:** feature-specialist.
- **Outstanding questions for confirmation before feature-specialist
  proceeds:**
  1. Confirm Slice 2's postwrite formatter is in-scope for v1 (vs.
     deferred follow-up) — the plan currently keeps it in.
  2. Confirm we should keep the per-role JSON skeleton authoritative
     in Go structs (Slice 1) and remove the duplicated skeletons from
     the prompt bodies (Slice 3). This is the proposed single source
     of truth.
  3. Confirm advisory `flock` (single-host) is acceptable for v1;
     multi-machine distributed locking is out of scope.
  4. Confirm `_meta.cli_version` semantics — record it, but do NOT
     gate validation on a version match. Older binaries reading newer
     files: tolerated. Older files validated by newer binaries:
     tolerated if all required fields present.
  5. Confirm Slice 3's acceptance test scope: scan
     `docs/architecture/*-prompt.md` + `internal/scaffold/assets/`
     mirror only, not the full repo.
