# Feature Brief — adversarial-validate-verifier

> Phase 13 "Lighter Centinela", from the centinela-v2 port-back analysis
> (2026-07-28). Evidence: V2's implement → mechanical check → fresh-context
> adversarial verification → harden loop found a real defect in 20/20 rounds
> (all 9 kernel phases + all 11 roadmap items). Self-review by the
> implementing context reliably misses what a fresh refuting context finds.

## Problem — what pain, who
The validate step today runs two trust-shaped artifacts: a
validation-specialist subagent that *narrates* that checks passed, and a
gatekeeper subagent whose prompt is a compliance checklist ("verify the
gates"). Both are artifacts ABOUT work. The project's own history shows the
gap: false "build ok / tests pass" claims from code-step agents, a subagent
dying mid-write leaving a stub report, acceptance suites silently skipping
every scenario while the gate stayed green. A checklist auditor primed with
the implementer's narrative confirms; a fresh context primed to REFUTE finds.
Operators need validate to be the step where an independent adversary tries
to break the completion claim — and fails — before `complete` advances.

## Scope (this feature ONLY)
- **In:** rewrite the gatekeeper into an adversarial verifier (refutation
  stance); evidence-only input contract (paths, not narrative); verifier
  re-runs `centinela validate` + the test suite itself and records what it
  ran; drop the validation-specialist role from the validate step; CRITICAL
  verdict blocks `complete`.
- **Out:** the plan-role merge (unified-plan-specialist); changes to
  `centinela verify` claim-verification mechanics (it stays the mechanical
  layer the verifier complements); production-readiness gate; merge-steward.

## User Stories
- As an operator, when a feature reaches validate, a fresh-context verifier
  that has seen none of the implementation conversation tries to refute
  "this feature is complete" against the diff, the spec, and gates it re-ran
  itself — and its verdict is what gates `complete`.
- As an orchestrating agent, I delegate validate verification by passing only
  the feature slug and file paths; I am forbidden from summarizing the work
  into the verifier's prompt.
- As a maintainer, the validate step costs one subagent instead of two, and
  the one that remains distrusts by default.

## Acceptance Criteria (→ Gherkin)
1. `docs/architecture/gatekeeper-prompt.md` (and scaffold mirror) is rewritten
   to a refutation stance: the verifier's task is "find the way the completion
   claim is false; default to NOT-VERIFIED when uncertain", with an explicit
   input contract — diff vs base, feature brief + spec, gate/check outputs —
   and an explicit prohibition on accepting orchestrator summaries or role
   reports as evidence.
2. The verifier MUST itself execute `centinela validate` and the project test
   suite and record each command (argv, exit code, duration) in a
   machine-readable section of `.workflow/<feature>-gatekeeper.md`; a report
   without a non-empty commands-run record fails evidence validation
   (dead-subagent stubs and narrated verdicts are refused).
3. `RequiredRolesForFeature(f, "validate")` no longer includes
   validation-specialist; validate-step required evidence is the gatekeeper
   report (+ existing `centinela validate` pass). Legacy in-flight workflows
   with validation-specialist evidence already written still complete.
4. Verdict semantics parsed from the structured `**Status:**` line (first
   token, existing PR #59 contract): `SAFE`/`WARNING` allow `complete`;
   `CRITICAL` or a missing/unparseable status BLOCKS `complete` with the
   verdict reason echoed. No prose scan.
5. After a blocked validate and subsequent fixes, re-verification spawns a
   FRESH verifier context (never "continue" the previous one); the report is
   overwritten, and the previous verdict cannot satisfy the gate for a tree
   that changed (report records the verified revision; gate compares it to
   HEAD).
6. The hook directive for the validate step names the adversarial verifier,
   its reasoning-tier model, and the paths-only input contract.
7. `centinela validate`, full suite, and scaffold-parity test pass.

## Edge Cases
- Verifier cannot execute commands (harness without Bash for subagents) →
  report must state it and carry NO commands-run record → evidence validation
  fails → fail-closed, never a narrated pass.
- Verifier dies mid-write (known failure mode) → stub/partial report lacks
  the commands-run record or Status line → blocked, orchestrator re-spawns.
- Wall-clock: verifier re-running the full suite adds a run on top of
  validate-step `complete`'s own runs (known ~3x + claim-verification cost);
  brief the verifier to run `centinela validate` once, not per-claim, and
  document the budget interaction with `verify_timeout`.
- Revision skew: report's verified revision ≠ HEAD at `complete` (fixes
  landed after verification) → gate refuses, demands fresh verification.
- Contaminated delegation: orchestrator pastes narrative anyway → prompt
  requires the verifier to list the inputs it actually read; a report citing
  orchestrator-provided summaries as evidence is a WARNING-level smell the
  prompt tells the verifier to flag (mechanical enforcement is out of scope).
- WARNING verdict → advances but the finding must land in the edge-cases /
  memory ledger (existing behavior for gatekeeper WARNINGs preserved).
- Legacy: a workflow mid-validate with old checklist-style report but no
  commands-run record → treated as legacy-complete only if written before
  migration; a fresh run demands the new format (same either-set discipline
  as unified-plan-specialist, decided in plan step).

## Data Model
No persisted schema change. Gatekeeper report gains a required
machine-readable commands-run section and a verified-revision field (inside
the existing `.md` artifact; format decided in plan — fenced JSON block
preferred so `centinela verify` can cross-check argv/exit codes against
ground truth).

## Integration Points
- `docs/architecture/gatekeeper-prompt.md` + `internal/scaffold/assets` mirror
- `internal/orchestration/policy.go` (drop RoleValidationSpec from validate)
- Validate-step gate: report parsing (`**Status:**` first-token contract),
  new commands-run + revision checks; `cmd/centinela` validate/complete path
- Hook directives (`hook_setup`/context) for the validate step
- `centinela verify`: optional cross-check of the verifier's recorded argv
  against its own claim-verification runs
- Delivery composer: consumes the same Status contract (unchanged)
- `internal/setup` agent-config emitters (verifier agent description)

## Risks
- **Verifier theater** (High): a refutation prompt can still rubber-stamp if
  the report isn't grounded — the commands-run record + revision binding is
  the mechanical teeth; both must be enforced by evidence validation, not
  convention.
- **Cost regression** (Med): one more full-suite run per validate; mitigated
  by dropping the validation-specialist context and by single-run briefing.
- **Blocking loop** (Med): an over-zealous verifier could CRITICAL-loop a
  healthy feature; WARNING semantics + the standing revise path are the
  escape valve — never a skip flag (no such command exists, by design).
- **Coverage** (Low): colocated `_test.go` ≤100 lines, ≥97% on touched
  packages.

## Decomposition
Single feature; no split. Independent of unified-plan-specialist (can land in
either order; if planner lands first its prompt table edits touch the same
CLAUDE.md section — trivial merge).
