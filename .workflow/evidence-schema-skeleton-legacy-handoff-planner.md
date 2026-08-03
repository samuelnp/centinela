# evidence-schema-skeleton-legacy-handoff — planner

## Problem

`centinela evidence schema <role>` is the single feature-less printer that
every one of the 8 role prompts instructs agents to run for the JSON
skeleton. Because it has no feature argument, it cannot ask the workflow's
own contract what `handoffTo` should be — it falls back to
`legacyHandoffForRole`, the pre-derivation static chain. That static chain is
wrong whenever a workflow's contract diverges from the five-role default:
internal features (no docs-step role), hotfix/spike archetypes (no docs step
at all), and any pinned-legacy contract. The agent only discovers the wrong
value when `centinela complete` refuses it minutes later — found in
production by the verifier on `beg-docstring-debt`.

## Scope

In scope: `internal/evidence/repair.go` (`SchemaSkeleton` signature + the two
placeholder consts), a new `internal/evidence/schema_active.go`
(`ResolveActiveFeature`), `cmd/centinela/evidence_schema.go` (wiring +
`--help` text), and the 4 call sites broken by the signature change (1
production, 3 tests). Out of scope (per the feature brief): the
chain-derivation rule itself (`ExpectedHandoff`/`CheckHandoffTo`/
`acceptsHandoff`/`alternateContractRoles` are untouched), prompt wording
beyond correctness (turns out to be none — see plan §4),
`evidence-doc-comment-overclaims-handoff` (separate Backlog entry, already
tracked), and retrofitting `.workflow/*.json` written before this change.

## Dependencies & Assumptions

- Depends on `binding-evidence-gates` (shipped): `ExpectedHandoff`,
  `CheckHandoffTo`, `RequiredEvidenceRoles`, `handoffForRole` all already
  exist and are reused unmodified.
- Assumes `worktree.DetectFeatureFromCwd` and `workflow.ActiveWorkflows` are
  reliable, already-trusted primitives (used today by the prewrite hook and
  `cmd/centinela/active_feature.go`'s cost-attribution `activeWorkflow`).
  `ResolveActiveFeature` reuses the SAME two primitives but combines them
  more conservatively (see plan §1) because this call site is
  gate-consequential, not cosmetic.
- Assumes the project's stated worktree operational model (memory:
  `project_worktree_operational_model`) — "run centinela from inside
  `.worktrees/<feature>/`" — is the dominant real-world invocation shape, so
  worktree-CWD detection covers the common case without a new argument.
- No new external dependency; no config/schema version bump.

## Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|------------|
| Prompt/mirror drift: a future prompt edit adds a `[feature]`-shaped example that this design doesn't need, silently reintroducing the "pass a possibly-stale slug" failure mode | Medium | Low | Zero prompt files are touched by this feature (plan §4); the CLI's own `--help` documents the two behaviors so there is no reason for a future prompt edit to invent an argument. Any such edit would be caught by `TestPromptsLinkToEvidenceContract`'s substring check only if it removes the base invocation — it would NOT catch an added argument, so the code-review gate is the real backstop here, called out explicitly in the plan. |
| Breaking agents that already pipe `centinela evidence schema <role>` output into a file or parser | Medium | Low | Output shape (valid JSON, same field set) is unchanged; only `feature` (now real when derivable) and `handoffTo` (now correct-or-obviously-placeholder) values change. A consumer that previously matched literally on `"<feature-slug>"` or on a specific legacy `handoffTo` string would need updating — none of this repo's own tests do (verified in plan §3's breakage table) and this is the exact wrongness the feature intends to stop shipping. |
| Parallel sessions running other features (this repo's own dev model — multiple worktrees active at once) cause CWD-derivation to resolve the WRONG feature | High if it happened | Low (mitigated by design) | `ResolveActiveFeature` picks a feature ONLY when the signal is unambiguous: worktree-CWD (always unique) or exactly-one active workflow outside worktree mode. Two-or-more active workflows with no worktree signal deliberately falls to the no-feature placeholder rather than guessing — this is the one place this plan diverges from the existing `activeWorkflow(cwd)` helper's "most-recently-touched" heuristic, on purpose. |
| `SchemaSkeleton` becomes CWD/filesystem-sensitive for the first time, so a test that forgets to chdir into an isolated temp dir (or leaves a stray `os.Chdir` without `t.Cleanup`) becomes a hidden environment/order dependency instead of a pure-function assertion | Medium | Medium | Called out explicitly in plan §3's breakage table per-file; new tests MUST use the existing `chdirEvidenceTemp`/`chdirToTemp` + `writeFakeWorkflow` fixtures, never rely on the ambient repo `.workflow/` state. |
| 100-line file cap violated by the signature/derivation additions | Low | Low | Sized in plan §3 (`repair.go` ~55, new `schema_active.go` ~30, `evidence_schema.go` ~45) — all well under 100; `schema_init.go` is untouched (98/100 already, no room, no edit needed there). |

## Rollout

1. Code step: add `placeholderFeature`/`unfilledHandoffSlot` consts and the
   new `SchemaSkeleton` signature in `repair.go`; add
   `internal/evidence/schema_active.go`; wire `cmd/centinela/evidence_schema.go`
   (Getwd → ResolveActiveFeature → SchemaSkeleton) plus `--help` text; fix the
   4 broken call sites in the SAME commit so the tree compiles throughout.
2. Tests step: colocated unit tests for `ResolveActiveFeature` (4-way table)
   and `SchemaSkeleton`'s two branches (incl. the merge-steward exemption);
   an integration test driving the real binary from inside a fabricated
   `.worktrees/<feature>` directory; acceptance tests against
   `specs/evidence-schema-skeleton-legacy-handoff.feature`'s 7 scenarios;
   `.workflow/evidence-schema-skeleton-legacy-handoff-edge-cases.md`.
3. Validate step: gatekeeper report + full suite; re-run the exact
   reproduction in plan §2 against the FIXED binary and confirm
   `centinela evidence schema gatekeeper` (no feature) now prints
   `"<successor-role>"`, and (from inside this worktree) prints
   `"complete"` for the real derived value.
4. Docs step: changelog entry; no `docs/guides/` change is required since no
   user-facing CLI surface (argument/flag) changed — only output content and
   `--help` text.

## Behavior Summary

`centinela evidence schema <role>` keeps its exact `<role>`-only invocation.
Internally it now asks `evidence.ResolveActiveFeature(cwd)` whether the CWD
unambiguously names one feature (worktree-CWD, or the sole active workflow
outside worktree mode). If yes, the skeleton is built exactly as `evidence
init` would build it for that feature — same `Skeleton()` call, same
`handoffForRole` derivation, so `evidence schema`, `evidence init`, and
`centinela complete`'s gate can no longer disagree for a real feature. If no
feature is resolvable, the skeleton's `feature` stays the existing
`"<feature-slug>"` placeholder and its `handoffTo` becomes the NEW
`"<successor-role>"` placeholder — except for `merge-steward`, whose
`handoffTo` is always the literal `"complete"` regardless of feature,
unchanged. `evidence schema bogus` still fails at role-parsing, before any
derivation is attempted, with the existing "unknown role" error.

## Acceptance Criteria (Gherkin)

See `specs/evidence-schema-skeleton-legacy-handoff.feature` — 7 scenarios:
derive-with-feature (canonical internal, hotfix archetype/no docs step,
user-facing same-step ux-ui hop), no-feature (CWD resolves nothing,
ambiguous CWD with 2+ active workflows), unknown role, and merge-steward
out-of-band (both with and without a resolvable feature).

## UX States

- `centinela evidence schema <role>` from inside `.worktrees/<feature>` (or
  the sole active workflow, non-worktree mode): prints valid JSON with the
  real `feature` and the gate-agreeing `handoffTo`. Exit 0.
- `centinela evidence schema <role>` with no resolvable feature (root
  checkout, no active workflow, or 2+ active workflows and no worktree
  signal): prints valid JSON with `"feature": "<feature-slug>"` and
  `"handoffTo": "<successor-role>"` (or `"complete"` for merge-steward).
  Exit 0 — this is expected, documented output, not an error.
  `centinela evidence schema --help` states both branches.
- `centinela evidence schema bogus`: prints nothing to stdout, returns the
  existing "unknown role" error, non-zero exit. Unchanged.
- Downstream: `centinela evidence validate <feature>` / `centinela complete
  <feature>` refuse the placeholder with the SAME specific
  `"evidence handoffTo for %q is %q, but this workflow's contract makes %q
  its successor — fix with: ..."` message they already give any wrong value
  — no new error format introduced.

## Edge Cases

- Ambiguous CWD (2+ active workflows, no worktree signal) must never guess —
  placeholder, not a pick of either.
- Merge-steward is exempt from the no-feature placeholder override; its
  literal `"complete"` answer is feature-independent and must survive both
  the known-feature and no-feature paths unchanged.
- Hotfix/spike archetypes omit the docs step entirely — a derived
  `gatekeeper` `handoffTo` must resolve to `"complete"`, not
  `"documentation-specialist"`.
- `SchemaSkeleton`'s signature change breaks 4 call sites (1 production, 3
  tests) that must be updated in the same commit or the package fails to
  compile.
- Unknown role must still be rejected by `ParseRole` before any CWD
  derivation is attempted (no wasted filesystem probing on a doomed call).

## Out-of-Scope

- Changing `ExpectedHandoff`/`CheckHandoffTo`/`acceptsHandoff`/
  `alternateContractRoles` or the tolerance they implement.
- Role-prompt wording beyond correctness — audited all 8 prompts (13
  invocation lines, 16 files counting scaffold mirrors); none need edits
  under the winning design (plan §4).
- `evidence-doc-comment-overclaims-handoff` (already a separate, tracked
  Backlog entry).
- Retrofitting `.workflow/*.json` files written before this change.

## Handoff

Advancing to **senior-engineer** (code step). Implementation order per plan
§3: `repair.go` consts + signature, new `schema_active.go`, then
`evidence_schema.go` wiring, then fix the 4 broken call sites so the tree
compiles. Full file-by-file plan, prompt/mirror inventory, and test-breakage
table are in `docs/plans/evidence-schema-skeleton-legacy-handoff.md`.
