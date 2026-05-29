---
feature: claim-verification
summary: Centinela independently re-derives ground truth for each evidence claim and hard-blocks step completion when any claim cannot be confirmed.
audience: end-user
status: done
---

## What it does

When a subagent completes a step it writes evidence — structured JSON that records what it claims it did, including that tests pass, coverage improved, outputs contain real work, and edge cases are covered by tests. Before this feature, Centinela trusted those claims at face value. Now it verifies them.

`centinela verify <feature>` re-runs the configured test commands, re-measures per-package coverage, inspects every output file for substantive content, and cross-checks each edge case against actual test names. It produces a per-claim PASS / FAIL / SKIP / WARN / TIMEOUT report and exits non-zero whenever something does not check out.

`centinela complete <feature>` runs the same checks automatically during the validate step. Any hard failure — failing tests, stub outputs, or an inflated coverage number — blocks the workflow from advancing until the underlying problem is resolved.

## When you'd use it

Use `centinela verify <feature>` whenever you want to confirm that a subagent's handoff report reflects reality before you approve the next step. It is especially useful after the `tests` or `validate` step, when agents are most likely to produce optimistic summaries. Running it standalone lets you see exactly which claims hold and which do not before the automatic gate in `centinela complete` makes the decision for you.

## How it behaves

- When all four claim checks pass, the report shows PASS for each one and `centinela complete` advances the step without any interruption.
- When a subagent claims tests pass but the configured test command exits with an error, the report shows FAIL, names the failing command, and `centinela complete` hard-blocks completion until the tests actually pass.
- When a subagent claims a coverage figure that exceeds the re-measured per-package coverage by more than the configured tolerance (default 0.1%), the report shows FAIL with both the claimed and measured numbers, and `centinela complete` hard-blocks.
- When a coverage claim is within the tolerance window, the report shows PASS and the workflow is not blocked.
- When an output file listed in evidence contains only an empty test function body with no assertions, the report shows FAIL, names the file, and `centinela complete` hard-blocks.
- When an output file is legitimately tiny — for example, a Go interface definition under 40 lines — it is not flagged as a stub.
- When an edge case entry in evidence has no matching test name in the feature's test files, the report shows WARN (not FAIL). The warning appears in `centinela complete` output but does not block step advancement on its own.
- When all edge cases are matched by test names, the edge-cases check shows PASS.
- When no evidence files exist for a step yet, the report shows SKIP for all checks with the message "no claims to verify". Nothing is blocked.
- When evidence omits a coverage field entirely, the coverage check is skipped. Absence of a claim is not a failure.
- When evidence has an empty edge-cases list, the edge-cases check is skipped.
- When `validate.commands` is empty or points to a binary that is not installed, the tests-pass check shows CONFIG ERROR rather than FAIL, and the message instructs you to configure `validate.commands`. This is a setup problem, not a fabricated claim.
- When the test suite takes longer than the configured `verify_timeout`, the report shows TIMEOUT, names the command and the timeout value, and `centinela complete` hard-blocks because the claim could not be confirmed.
- When `workflow.use_worktrees` is on, verify reads evidence and runs test commands from inside the feature worktree, not the root checkout.
- The verify output always ends with a summary line (for example, "2 passed, 1 failed, 1 skipped").
- When `centinela complete` encounters a verify failure, it prints the full verify report before the gate error line, and the error message distinguishes a claim failure from a structural evidence failure.

## Examples

Run verification on demand before approving a step:

```bash
centinela verify my-feature
```

Sample output for a clean handoff:

```
PASS  tests-pass       go test ./... exited 0
PASS  coverage         claimed 95.2%, measured 95.2% (within 0.1% tolerance)
PASS  outputs-not-stub all 8 output files contain substantive content
PASS  edge-cases       all 5 edge case entries matched to test names

4 passed, 0 failed, 0 skipped
```

Sample output when a coverage claim is inflated:

```
PASS  tests-pass       go test ./... exited 0
FAIL  coverage         claimed 92.0%, measured 78.3% — exceeds tolerance of 0.1%
PASS  outputs-not-stub all 8 output files contain substantive content
WARN  edge-cases       1 unmatched entry: "timeout while suite runs"

1 passed, 1 failed, 1 skipped, 1 warned
```

Configure verification knobs in `centinela.toml`:

```toml
[verify]
verify_timeout     = 60    # seconds before a test command is killed (default: 60)
coverage_tolerance = 0.001 # maximum allowed gap between claimed and measured coverage (default: 0.001 = 0.1%)
```
