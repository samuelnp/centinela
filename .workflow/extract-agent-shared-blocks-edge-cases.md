### Edge-Case Report: extract-agent-shared-blocks
**Date:** 2026-05-11

#### Risk Matrix

- **Case:** Future prompt author reintroduces the invocation paragraph
  - **Impact:** Low (extra ~4 lines per file)
  - **Likelihood:** Medium (cognitive habit; copy-paste from other docs)
  - **Why:** Asserted by `TestExtractAgentSharedBlocks_PromptsReferenceShared` — every affected prompt MUST contain the substring `agent-invocation.md`, so a paragraph that doesn't reference it would still fail the test only if it replaces the reference. Improvement opportunity: also assert the canonical phrasing.

- **Case:** `agent-invocation.md` deleted by accident
  - **Impact:** High (every prompt's `How to Invoke` becomes a dangling link)
  - **Likelihood:** Low
  - **Why:** Asserted by `TestExtractAgentSharedBlocks_SharedFilesExist`.

- **Case:** Gatekeeper `## Decision Rules` heading reintroduced
  - **Impact:** Low (re-creates the duplicate)
  - **Likelihood:** Low
  - **Why:** Asserted by `TestExtractAgentSharedBlocks_GatekeeperDecisionRulesRemoved`.

- **Case:** `production-readiness-prompt.md.template` reabsorbs the
  multi-language matrix in a future regeneration
  - **Impact:** Medium (re-bloats the template by ~13 lines, blocks the
    win for new-project bootstraps)
  - **Likelihood:** Medium (template files attract incremental
    additions over time)
  - **Why:** Asserted by `TestExtractAgentSharedBlocks_StackMatrixMovedOut`
    which trips if three or more language entries reappear together.

- **Case:** Scaffold mirror drifts after a future canonical edit
  - **Impact:** Medium (new projects bootstrap with stale prompts)
  - **Likelihood:** High (two parallel trees; only canonical reviewed
    in PRs by default)
  - **Why:** Asserted by
    `TestExtractAgentSharedBlocks_ScaffoldMirrorParity`.

- **Case:** Removing the gatekeeper Decision Rules table leaves no
  guidance for the "what to do after a WARNING" case
  - **Impact:** Low — the Output Format Recommendation block at
    `gatekeeper-prompt.md:46-52` still names SAFE / WARNING / BLOCKING
    and what each means.
  - **Likelihood:** Low (reviewers will read the Recommendation block).
  - **Why:** Verified by `TestExtractAgentSharedBlocks_GatekeeperDecisionRulesRemoved`
    asserting the status words still appear.

#### Missing or Weak Scenarios

- No check that `documentation-generator-prompt.md` was correctly
  *skipped* (it has no `## How to Invoke` section). A future iteration
  could assert the file remains unchanged across this feature.
- No automated check that the per-prompt reference is the canonical
  phrasing rather than any string containing `agent-invocation.md`.
  Could harden by asserting the exact one-line wording.

#### Proposed/Added Tests

- Unit: none — doc-only change.
- Integration: none — no runtime behavior changes.
- Acceptance: `tests/acceptance/extract_agent_shared_blocks_acceptance_test.go`
  with six sub-tests: shared-files existence, agent-invocation contract
  content, per-prompt reference to shared file, gatekeeper Decision
  Rules removed, stack matrix moved out of template, scaffold mirror
  parity for the affected files.

#### Residual Risks

- The total document-tree size grew by ~89 lines (two new shared
  files) while individual hot-path prompts shrank by ~16 lines. The
  per-invocation context win is real (gatekeeper especially) but the
  *static* repository surface grew slightly. Acceptable trade-off:
  shared references are loaded lazily, not on every invocation.
- The `<!-- centinela:doc-version=… -->` HTML comments remain on
  affected files because `internal/migration/header.go` parses them.
  A manifest-based refactor is queued as a follow-up feature and would
  shave another ~3 lines × N files.
