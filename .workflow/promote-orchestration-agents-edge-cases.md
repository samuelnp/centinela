### Edge-Case Report: promote-orchestration-agents
**Date:** 2026-05-10

#### Risk Matrix

- **Case:** Scaffold mirror drift after a future edit to a canonical prompt
  - **Impact:** Medium
  - **Likelihood:** High (every future prompt edit must update both trees)
  - **Why:** Two parallel files; only the canonical is reviewed in PRs by default. Asserted by `TestPromoteOrchestrationAgents_MirrorByteIdentical`.

- **Case:** A future edit pushes a prompt file over the 70-line budget
  - **Impact:** Low (prompt usability), Medium (per-invocation context cost)
  - **Likelihood:** Medium (incremental additions accrete over time)
  - **Why:** No automation enforces the budget except the new test
    `TestPromoteOrchestrationAgents_LineBudget`.

- **Case:** Heading text drift (`## Purpose` → `### Purpose`, etc.)
  - **Impact:** Medium (acceptance test breaks; downstream tooling that
    parses sections may miss content)
  - **Likelihood:** Low (humans rarely change heading levels intentionally)
  - **Why:** Asserted by `TestPromoteOrchestrationAgents_RequiredSections`.

- **Case:** New orchestration role added to `policy.go` without a
  matching prompt file
  - **Impact:** High (the new role gets terse embedded guidance only,
    re-creating the asymmetry we just closed)
  - **Likelihood:** Medium (easy to forget; not enforced by the build)
  - **Why:** Out of scope for this feature's tests, but worth a
    follow-up (e.g. a Go test that diffs `policy.go` roles against
    `docs/architecture/*-prompt.md` filenames).

- **Case:** `ux-ui-specialist` prompt invoked for a non-user-facing
  feature
  - **Impact:** Low (extra noise; orchestrator can ignore)
  - **Likelihood:** Low (`RequiredRolesForFeature` gates invocation by
    the `surface` field in the feature brief)
  - **Why:** Behaviour is enforced by `policy.go:31-37`; the prompt
    file documents the conditional clearly under `## Purpose` and
    `## How to Invoke`.

- **Case:** `validation-specialist` restates gatekeeper / readiness
  content instead of cross-linking
  - **Impact:** Medium (context bloat; the very thing the parent audit
    aimed to reduce)
  - **Likelihood:** Low (the prompt explicitly says "Do NOT restate")
  - **Why:** Conventional review during prompt edits should catch drift.

#### Missing or Weak Scenarios

- No automated check that `policy.go` roles correspond 1:1 with
  `docs/architecture/*-prompt.md` files. Suggested follow-up.
- No check that the canonical and mirror copies are simultaneously
  updated in the same commit (only that they are equal at any point in
  time).

#### Proposed/Added Tests

- Unit: none — this is a doc-only change.
- Integration: none — no runtime behavior changes.
- Acceptance: `tests/acceptance/promote_orchestration_agents_acceptance_test.go`
  with four assertions: existence, required headings, scaffold mirror
  byte-identity, per-file line budget.

#### Residual Risks

- A new role added to `policy.go` without a prompt file would be
  invisible to the new tests. Mitigate via a follow-up feature or by
  expanding `policy.go` with a runtime check that prompt files exist
  for each registered role.
- The `## How to Invoke` boilerplate now repeats across nine prompt
  files. The Tier 2 plan (extract to a shared
  `docs/architecture/agent-invocation.md`) becomes more valuable now,
  not less; tracked as a future feature.
