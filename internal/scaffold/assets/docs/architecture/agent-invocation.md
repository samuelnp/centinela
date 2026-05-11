<!-- centinela:doc-version=1 template=docs/architecture/agent-invocation.md -->
# Agent-Tool Invocation — Shared Reference

This file is the single canonical description of how Centinela orchestration
prompts are invoked. Every other prompt under `docs/architecture/` references
this file from its `## How to Invoke` section instead of repeating the same
boilerplate.

## How to invoke any Centinela orchestration prompt

1. Use the **Agent** tool. For most read-and-report subagents,
   `subagent_type: Explore` is correct. The plan / code / tests / validate /
   docs role prompts that may need write access to the workflow tree run as
   plain `Agent` invocations from the orchestrator.
2. Pass the prompt below verbatim, replacing every occurrence of
   `<FEATURE_NAME>` with the kebab-case feature name passed to
   `centinela start`.
3. The subagent reads the inputs declared in the prompt, performs its
   analysis, and saves its report to the artifact path specified in the
   prompt's `## Required Artifact` section.

## Artifact contract

Reports are saved to the `.workflow/` directory using the convention
documented in `cmd/centinela/hook_orchestration.go`:

- `.workflow/<feature>-<role>.md` — human-readable report.
- `.workflow/<feature>-<role>.json` — structured evidence consumed by
  `centinela complete`'s strict-evidence validator. Required keys:
  `feature`, `step`, `role`, `status`, `generatedAt`, `inputs`, `outputs`,
  `edgeCases`, `handoffTo`.
- `.workflow/<feature>-edge-cases.md`, `.workflow/<feature>-gatekeeper.md`,
  `.workflow/<feature>-production-readiness.md` for the report-only
  subagents that don't have an orchestration role.

`centinela complete <feature>` will not advance the workflow until the
required artifacts for the current step exist.

## Why this file exists

The audit feature `agent-performance-audit` measured ~440 tokens of
repeated invocation boilerplate across the prompt files. Extracting the
shared block to one file means future prompt edits update one place
rather than ten, and per-invocation context cost drops because each
prompt now contains a single-line reference instead of the full
paragraph.
