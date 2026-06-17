<!-- centinela:doc-version=1 template=docs/architecture/edge-case-tester-prompt.md -->
# Edge-Case Tester Subagent — Invocation Guide

## Purpose

Use this subagent during the `tests` step to detect hard paths and edge cases before completing the step.

## How to Invoke

See [agent-invocation.md](agent-invocation.md) for the canonical Agent
invocation pattern. Replace `<FEATURE_NAME>` in the template below.

## Prompt Template

```
You are the Centinela Edge-Case Tester.

Analyze feature "<FEATURE_NAME>" and produce a hard-path report.

Required analysis:
1. Invalid inputs, empty data, and boundary values.
2. Dependency failures (network, API, DB, malformed payloads).
3. State-transition errors and invalid operation ordering.
4. Retry/idempotency behavior and duplicate requests.
5. i18n/config/environment mismatches.
6. Security-adjacent misuse paths (permission checks, unsafe defaults).

Output format:
### Edge-Case Report: <FEATURE_NAME>
**Date:** <current date>

#### Risk Matrix
- **Case:** <name>
- **Impact:** Low|Medium|High
- **Likelihood:** Low|Medium|High
- **Why:** <short reason>

#### Missing or Weak Scenarios
- List concrete scenarios currently untested or weakly tested.

#### Proposed/Added Tests
- Unit:
- Integration:
- Acceptance:

#### Residual Risks
- Risks still not covered and mitigation suggestions.

#### Deferred Findings
- For every Residual Risk you are deferring rather than covering with a
  test now, run:
  `centinela roadmap defer <slug> --summary "<one line>" --source <feature>/edge-case-tester`
- List the recorded slugs here, or state "none".
```

## Required Artifact

Save report to:

```
.workflow/<feature-name>-edge-cases.md
```

`tests` step cannot complete without this artifact.
