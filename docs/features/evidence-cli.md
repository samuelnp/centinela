# Feature Brief: `centinela evidence` — Typed Artifact CLI

## Problem

Every workflow step (plan, code, tests, validate, docs) and every role
(big-thinker, feature-specialist, senior-engineer, ux-ui-specialist,
qa-senior, validation-specialist, gatekeeper, production-readiness,
documentation-specialist) requires the agent to hand-write a
`.workflow/<feature>-<role>.json` + `.md` evidence pair plus several
templated companions (`edge-cases.md`, `gatekeeper.md`,
`documentation-specialist.{md,json}`).

There is no typed CLI for any of this. Agents reach for `python3 -c`,
heredocs, and `jq` one-liners to escape strings, merge fields, and produce
JSON that conforms to the validator's schema. Outputs land as minified or
escape-noisy JSON that humans can't read, fail the validator on missing or
misspelled fields, and force human round-trips. The prior
`add-agent-evidence-contract` feature documented the schema in prompts but
did not remove the authoring burden — agents still hand-write the JSON.

## User Stories

- As an **LLM agent** executing a step, I want to author my role's evidence
  with `centinela evidence` subcommands so I never write JSON by heredoc.
- As a **developer reviewing `.workflow/`**, I want pretty-printed,
  schema-conforming JSON I can audit at a glance.
- As **Centinela's gate logic**, I want schema enforcement inside the
  binary so malformed evidence is caught before `centinela complete`, with
  a precise fix hint.
- As an **agent prompt author**, I want one mandate ("use the CLI; do not
  hand-write JSON") instead of repeating the schema in every prompt.

## Acceptance Criteria

- AC1: For every required `.workflow/<feature>-<role>.{json,md}` artifact,
  a `centinela evidence` subcommand produces a passing file end-to-end
  with zero raw JSON authoring.
- AC2: `centinela evidence validate <feature>` reports each missing or
  malformed field with the exact subcommand to fix it; exit non-zero on
  failure.
- AC3: All writes are atomic (temp-file + rename); a crash mid-append
  never leaves invalid JSON on disk.
- AC4: The PostToolUse Write/Edit hook auto-pretty-prints any
  `.workflow/*.json` written outside the CLI, scoped to the current
  feature's `.workflow/` prefix.
- AC5: `centinela artifact new <feature> <kind>` produces pre-filled
  templates for `edge-cases`, `gatekeeper`, `production-readiness`, and
  `documentation-specialist` (md + json).
- AC6: Every agent prompt in `docs/architecture/*-prompt.md` and its
  `internal/scaffold/assets/` mirror mandates the CLI and forbids
  hand-written JSON. Drift caught by an acceptance test.
- AC7: Pre-existing `.workflow/` files written before this feature still
  validate (no breaking schema change).

## Edge Cases

- **Concurrent writes** — two subagents may target the same
  `<feature>-<role>.json`. v1 uses advisory file locks with a short
  timeout and a clear error pointing the agent at
  `centinela evidence read` before retry.
- **Schema/version skew** — older binary, newer prompt (or vice versa).
  Unknown fields preserved on round-trip; missing required fields
  rejected; `_meta.cli_version` recorded per JSON for diagnosis.
- **Partial/aborted runs** — temp-file + atomic rename guarantees
  on-disk JSON is never half-written; `centinela evidence repair
  <feature>` drops orphaned temp files.
- **Free-form fields** — each schema declares an explicit `extra: object`
  slot; anything else fails validation. One escape hatch without losing
  strictness.
- **Hooks outside the worktree** — postwrite formatter must scope to the
  current feature's `.workflow/` prefix and never touch other features'
  files.

## Data Model

- **EvidenceFile**: `.workflow/<feature>-<role>.json` — fields per role
  per `docs/architecture/evidence-contract.md`, plus
  `_meta: { cli_version, written_at, role, feature }`.
- **EvidenceCompanion**: `.workflow/<feature>-<role>.md` — human-readable
  narrative written by the same subcommand call as the JSON.
- **ArtifactTemplate**: typed templates for `edge-cases`, `gatekeeper`,
  `production-readiness`, `documentation-specialist` produced by
  `centinela artifact new`.
- **Schema**: Go structs in `internal/evidence/` (one per role) — single
  source of truth for runtime validation and prompt-embedded skeletons.

## Integration Points

- `centinela hook prewrite` / `postwrite` — auto-format `.workflow/*.json`
  on Write/Edit, surface schema errors as block messages.
- `centinela complete <feature>` — reuse the `validate` codepath; reject
  step advance with the existing required-evidence message format.
- Agent prompts in `docs/architecture/*-prompt.md` and the
  `internal/scaffold/assets/docs/architecture/` mirror.
- Documentation generator (`docs/project-docs/`) — read JSON via the
  typed reader instead of ad-hoc parsing.

## Risks

- **CLI surface growth** — 8+ roles × multiple verbs is nontrivial.
  Mitigate by generating the per-role command set from Go schema structs.
- **Schema bound to binary version** — every schema change is a release.
  Mitigate with `_meta.cli_version` and explicit "unknown field"
  tolerance.
- **Agent non-compliance** — agents may keep hand-writing JSON.
  Mitigate with the postwrite hook (catches and reformats) and a strict
  pre-commit/CI check that JSON conforms to schema.
- **Worktree scoping bugs** — postwrite formatter must not corrupt other
  features' evidence. Scope to the active feature's `.workflow/` prefix.

## Decomposition

Ships as one feature with three dependency-ordered sub-slices:

1. **`evidence-cli-core`** — `centinela evidence init|set|append|read|validate`
   covering all role schemas. The blocker for value.
2. **`evidence-cli-artifacts`** — `centinela artifact new` templates +
   postwrite auto-format hook.
3. **`evidence-cli-prompts`** — rewrite every agent prompt + scaffold
   mirror to mandate the CLI; acceptance test asserts no prompt embeds
   raw JSON authoring instructions.

If plan-step subagents flag sequencing pressure, slices 2 and 3 may split
into follow-up features; slice 1 is the only blocker for value.
