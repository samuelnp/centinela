# Feature: evidence-schema-skeleton-legacy-handoff

**Phase:** 13 — Lighter Centinela
**Archetype:** canonical
**Depends on:** binding-evidence-gates (shipped)

## Problem

`binding-evidence-gates` made `handoffTo` a real gate: the expected successor is
derived from the workflow's own contract, and a wrong value refuses `complete`.
It fixed the `evidence init` prefill (`handoffForRole`) to derive the same way,
so a scaffolded stub is never seeded with a value the gate then rejects.

`centinela evidence schema <role>` was not fixed, because it *cannot* be: it
takes only a role, has no feature, and therefore no workflow contract to derive
from. It falls back to `legacyHandoffForRole`, the pre-derivation static chain —
so it prints, for example, `"handoffTo": "documentation-specialist"` for the
gatekeeper, which the gate refuses on any workflow where the docs step requires
no role evidence (every internal feature) and on archetypes that omit the step
entirely (hotfix, spike).

This matters more than a stale default normally would, because **every role
prompt instructs the agent to run this command** ("Use `centinela evidence
schema <role>` to print the JSON skeleton — the CLI is the single source of
truth"). The CLI hands out a value its own gate rejects, and the agent discovers
it only when `complete` fails minutes later. It was found by the verifier on
`beg-docstring-debt`, whose own evidence had to be corrected by hand.

## Expected outcome

`centinela evidence schema` never emits a `handoffTo` the chain gate would
refuse:

1. When a feature is known, the skeleton derives `handoffTo` exactly as
   `evidence init` and the gate do — one derivation, three callers.
2. When no feature is known, the skeleton does not guess. It emits a value that
   is obviously a slot to fill rather than a plausible-but-wrong role, so the
   failure mode is "the author must decide" instead of "the CLI was confidently
   wrong". The rendered form must remain schema-valid JSON, since the output is
   embedded in prompts and piped.
3. The command's help states which of the two applies.

The exact CLI surface for (1) — positional feature, `--feature` flag, or
deriving from the active workflow when the CWD resolves one — is the plan's
decision. Whatever it is, the no-feature path must stay usable, because the
prompts that call this command do not always know the feature slug.

## Out of scope

- Changing the chain-derivation rule itself, or the tolerance in
  `acceptsHandoff`. This feature makes the skeleton agree with the gate; it does
  not renegotiate what the gate wants.
- The other role prompts' wording beyond the minimum needed for accuracy.
- `evidence-doc-comment-overclaims-handoff` (a separate Backlog entry).
- Retrofitting `.workflow/*.json` written before this change.

## Constraints

- No gate weakened: the fix moves the CLI toward the gate, never the reverse.
- `legacyHandoffForRole` must keep working for the genuine no-workflow-state
  case that `evidence init` still relies on.
- Output stays valid JSON and stable enough to embed in prompts.
- 100-line file cap incl. `_test.go` in `cmd/` and `internal/`; per-package
  coverage ≥97% on touched packages; scaffold-mirror lockstep for any
  `docs/architecture/` prompt edit.
