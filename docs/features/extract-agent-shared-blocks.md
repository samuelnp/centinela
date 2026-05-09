# extract-agent-shared-blocks

## Problem

The agent-performance audit identified ~440 tokens (~60 lines) of
repeated boilerplate across the 10 prompt files in `docs/architecture/`:
near-identical "How to Invoke" blocks, duplicate trigger / decision
tables in `gatekeeper-prompt.md`, and a multi-stack example matrix in
`production-readiness-prompt.md.template` that is only relevant for one
stack at runtime. Every prompt invocation pays this overhead in context
tokens with no behavioural benefit.

## User Stories

- As a Centinela user paying LLM API costs, I want repeated boilerplate
  removed so each agent invocation uses fewer context tokens.
- As a Centinela maintainer, I want a single canonical "Agent tool
  invocation" reference so future prompt edits update one place, not ten.
- As a new-project bootstrap consumer, I want my generated
  `production-readiness-prompt.md` to contain only my own stack's
  examples, not a four-language matrix.

## Acceptance Criteria

- A new file `docs/architecture/agent-invocation.md` exists and
  documents the canonical Agent-tool invocation pattern.
- Every prompt file under `docs/architecture/` whose previous "How to
  Invoke" body contained the boilerplate now references
  `agent-invocation.md` in one line (or section is removed and replaced
  by a top-of-file reference).
- The duplicate `Decision Rules` table in `gatekeeper-prompt.md` is
  removed; the equivalent decisions remain expressed in the Output
  Format Recommendation section.
- A new file `docs/architecture/stack-checks-reference.md` exists and
  contains the multi-stack example matrix previously inlined in
  `production-readiness-prompt.md.template`.
- `production-readiness-prompt.md.template` keeps a single-stack
  placeholder pointing to `stack-checks-reference.md`.
- All edits are mirrored byte-identically in
  `internal/scaffold/assets/docs/architecture/`.
- No Go code is modified. No runtime behaviour changes.
- Existing acceptance test
  `TestPromoteOrchestrationAgents_RequiredSections` still passes (the
  three required headings remain).
- Existing acceptance test
  `TestEdgeCaseSubagentPrompt_DocIncludesRequiredSections` still passes
  (Output Format strings are unchanged).
- New acceptance test asserts `agent-invocation.md` exists and is
  referenced by every prompt file that previously contained the
  invocation boilerplate; asserts `stack-checks-reference.md` exists.

## Edge Cases

- The unified-output-format Tier 2 item is **out of scope** because
  acceptance tests assert specific Output-Format substrings.
- Removing `<!-- centinela:doc-version=… -->` HTML comments is **out of
  scope** because `internal/migration/header.go` parses them as part
  of the managed-doc tracking system. Tracked separately as a future
  manifest-based refactor.
- The canonical, project-rendered `production-readiness-prompt.md` (the
  Centinela project's own filled-in copy) is left as-is — only the
  `.template` is slimmed.
- `documentation-generator-prompt.md` has no formal "How to Invoke"
  section and needs no edit for item B.

## Risks

- Drift between `docs/architecture/` and the scaffold mirror; mitigated
  by acceptance test asserting byte-identity.
- Future prompt authors might reintroduce the boilerplate; mitigated
  by documenting the convention in `agent-invocation.md` itself.
- Decoupling stack examples from the production-readiness template
  could orphan them if `stack-checks-reference.md` is later deleted;
  mitigated by acceptance test asserting both files coexist.
