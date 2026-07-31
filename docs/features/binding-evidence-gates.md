# Feature: binding-evidence-gates

**Phase:** 13 — Lighter Centinela
**Archetype:** canonical
**Depends on:** none

## Problem

Three gates in the evidence/artifact layer accept input they claim to check.
Each was found by an adversarial verifier on a different feature, and each is
fail-OPEN: the gate reports success on data it was written to reject. For a
framework whose thesis is "no false success", a gate that does not bind is worse
than an absent one — it produces a green signal nobody has earned.

1. **`handoffTo` is unvalidated.** Every role prompt documents a handoff chain
   (planner → senior-engineer → qa-senior → validation-specialist → complete)
   and the evidence contract states `handoffTo` MUST be the successor role. The
   validator only checks the field is non-empty: `handoffTo: banana` completes a
   step. The documented chain is decoration.
   *(was Backlog `handoff-chain-unvalidated`)*

2. **`artifact stamp` accepts any commands shape.** The stamp writes the
   `centinela:verification` block that the validate gate later reads to prove the
   verifier actually ran the gates. Nothing validates the `commands` array's
   schema at stamp time, so a malformed or hand-shaped record is only discovered
   later — or not at all, if it happens to satisfy the reader's narrower checks.
   The trust anchor's own plumbing is unchecked.
   *(was Backlog `stamp-validates-commands-schema`)*

3. **An unfilled changelog template passes the docs gate.**
   `validateChangelog` requires a non-empty line; `centinela artifact new
   <feature> changelog` writes `- <FILL: type>: <FILL: one-line summary>`, which
   is non-empty. The docs step therefore completes on a stub nobody wrote. This
   was hit in practice during `dynamic-model-routing`.
   *(was Backlog `changelog-template-placeholder-passes-gate`)*

## Expected outcome

Each gate rejects what it already claims to reject, with an actionable error:

1. `handoffTo` is validated against the role chain the workflow's own contract
   defines (archetype- and contract-aware, so legacy and subset workflows keep
   working). An unknown or out-of-chain value fails `complete` naming the
   expected successor. The terminal role's `complete` value stays valid.
2. `centinela artifact stamp` validates the `centinela:verification` commands
   schema — entry shape, required keys, types — and fails loudly at stamp time
   rather than deferring the discovery. The gate that reads it keeps its own
   checks; this closes the write side.
3. `validateChangelog` rejects unreplaced template placeholders, so a stub is
   not a changelog. The error says what to write.

## Out of scope

- `revise-to-plan-sheds-no-evidence` — adjacent (evidence invalidation) but on
  the rewind path, with its own state-transition semantics.
- `spec-conflict-precheck-requires-merging-worktree` — same fail-open class, but
  a different subsystem (merge).
- Reworking the evidence contract or role chain itself; this feature enforces
  what is already documented, it does not redesign it.
- Retrofitting existing `.workflow/*.json` in this repo or downstream projects.

## Constraints

- **Fail-closed, and retroactively brittle only where exactness is the point.**
  Legacy and archetype-subset workflows (hotfix has no plan step; internal
  features skip docs roles) must keep completing, and validation must derive the
  expected successor from the workflow's own contract rather than a hardcoded
  five-role list. TWO classes are deliberately NOT accommodated, both seeded by
  the old hardcoded prefill this feature also fixes:
  - naming a role from a LATER step, skipping a required same-step role
    (`senior-engineer → qa-senior` on a user-facing feature, where the derived
    successor is the same-step `ux-ui-specialist`);
  - naming a role where the chain has ENDED (`gatekeeper →
    documentation-specialist` on an internal feature, where the docs step
    requires no role evidence so the derived successor is `complete`).
  Accepting either would give back the property that makes the gate meaningful,
  so both are refused — and the error prints the one command that repairs it.
  Any repo with an in-flight workflow at those points will hit this once.
- No gate may be weakened to make another pass.
- 100-line file cap incl. `_test.go` in `cmd/` and `internal/`; per-package
  coverage ≥97% on touched packages; scaffold-mirror lockstep for any
  `docs/architecture/` edit.
- Each of the three is independently testable and independently revertible —
  they share a layer, not an implementation.
