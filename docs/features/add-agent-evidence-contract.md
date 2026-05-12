# Feature Brief: Make Agent Prompts Spell Out the Evidence Contract

## Problem

Centinela's orchestration validator enforces strict rules on the per-role
`.workflow/<feature>-<role>.json` evidence files: outputs must be real file
paths on disk, plan-step roles must snapshot every `docs/features/*.md`, the
qa-senior must reference `.workflow/<feature>-edge-cases.md`, the
ux-ui-specialist must declare `mobileFirst: true` plus eight specific edge-case
tags, and so on.

The agent prompts (big-thinker, feature-specialist, senior-engineer,
qa-senior, ux-ui-specialist, validation-specialist, documentation-generator)
only say "save a structured companion JSON" without describing the schema or
those rules. Agents repeatedly write prose summaries as outputs, skip the
feature-doc snapshot, or omit required edge-case tags — every failed
`centinela complete` then forces a human round-trip to fix the JSON.

## Goal

Make the JSON contract self-evident in each agent prompt so the agent
produces correct evidence the first time.

## Scope

- Add one canonical `docs/architecture/evidence-contract.md` documenting the
  full schema and every per-role rule the validator enforces (linked from
  CLAUDE.md and from each role prompt).
- Update each agent prompt to embed a role-specific JSON skeleton and a
  short checklist of the rules that apply to that role.
- Mirror every change to `internal/scaffold/assets/docs/architecture/` so
  new projects scaffold with the corrected prompts.
- Add an acceptance test that asserts each prompt contains the schema and
  its role-specific rules so future drift is caught.

## Non-Goals

- Changing the validator itself.
- Auto-generating evidence JSON for the agent (the agent still authors it).
- Re-running historical features whose evidence already passed validation.
