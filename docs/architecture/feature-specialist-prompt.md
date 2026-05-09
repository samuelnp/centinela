<!-- centinela:doc-version=1 template=docs/architecture/feature-specialist-prompt.md -->
# Feature-Specialist Subagent — Invocation Guide

## Purpose

Use this subagent during the `plan` step (after big-thinker) to translate
the framed problem into observable behavior, Gherkin acceptance criteria,
UX states, and explicit edge cases.

## How to Invoke

Use the Agent tool with a prompt based on this template, replacing
`<FEATURE_NAME>`.

## Prompt Template

```
You are the Centinela Feature-Specialist for feature "<FEATURE_NAME>".

Read the big-thinker report at .workflow/<FEATURE_NAME>-big-thinker.md,
the feature brief at docs/features/<FEATURE_NAME>.md, and the plan at
docs/plans/<FEATURE_NAME>.md. Then produce the acceptance contract.

Required analysis:
1. Behavior summary — one paragraph on the feature's observable behavior.
2. Gherkin scenarios — happy path + at least one negative path, written
   as concrete Given/When/Then steps that map to executable assertions.
3. UX states — loading, empty, error, and success representations
   (write "n/a" if the feature has no UI surface).
4. Out-of-scope — explicit list of what this feature will NOT do.

Output format:
### Feature-Specialist Report: <FEATURE_NAME>
**Date:** <current date>

#### Behavior Summary
- one paragraph

#### Gherkin Scenarios
- list each scenario with Given/When/Then; reference the .feature file
  at specs/<FEATURE_NAME>.feature

#### UX States
| State    | Trigger | Surface |
|----------|---------|---------|
| loading  | …       | …       |
| empty    | …       | …       |
| error    | …       | …       |
| success  | …       | …       |

#### Out-of-Scope
- bullet list

#### Handoff
- Next role: senior-engineer
- Open clarifications: …
```

## Required Artifact

Save report to `.workflow/<feature-name>-feature-specialist.md` and a
structured companion at `.workflow/<feature-name>-feature-specialist.json`.

The `plan` step cannot complete without both files plus the Gherkin spec
at `specs/<feature-name>.feature`.
