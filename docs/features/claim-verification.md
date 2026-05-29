# claim-verification

> Independently verify that orchestration evidence claims match ground truth before a step can complete.

## Problem — what pain does this solve? Who is the user?

Centinela's central promise is *trustworthy* enforcement of the plan → code → tests → validate → docs cycle. Today the orchestration validator (`internal/orchestration/`) checks the **form** of evidence JSON — that fields match context, `outputs` point at real files, `inputs` include the doc snapshot — but never checks whether the **claims are true**. A subagent can write a perfectly-shaped `qa-senior.json` asserting "tests pass" and "coverage rose" with neither being real: the files exist, so the gate passes.

This pushes the framework's core guarantee onto the human operator. The maintainer's own working rules already encode "verify agent claims independently" and "close failing gates with real work, never by gaming them" as **manual disciplines** — meaning the most important check is the one Centinela does not automate.

**User:** the developer running Centinela who delegates steps to subagents and relies on `centinela complete` to mean the work is genuinely done — not merely well-described.

## User Stories

- As a maintainer, I want `centinela complete` to reject a step when the agent's evidence claims diverge from reality, so a fabricated or optimistic report cannot advance the workflow.
- As a developer, I want a standalone `centinela verify <feature>` I can run on demand to see exactly which claims hold and which don't, before I attempt to complete.
- As someone paying for agent work, I want empty-stub outputs and unran tests caught automatically, so I don't have to re-audit every handoff by hand.

## Acceptance Criteria — concrete, testable (→ Gherkin scenarios)

Verification compares each evidence JSON in `.workflow/<feature>-<role>.json` against independently re-derived ground truth. Four claim checks are in scope for v1:

1. **Tests actually pass.** When `qa-senior` / `validation-specialist` evidence asserts a passing suite, verification re-runs the configured test command(s) (`validate.commands`) and confirms exit 0. A non-zero exit fails verification with the failing command named.
2. **Coverage actually moved.** When coverage is claimed, verification re-derives per-package coverage (consistent with the project's per-package, no-`-coverpkg` model) and confirms the claimed figure/delta holds. A claim above measured coverage fails.
3. **Outputs aren't empty stubs.** Every `outputs` file is checked for substantive content (non-empty beyond boilerplate; test outputs must contain real assertions, not empty `func Test…(){}` bodies). A stub that only satisfies file-exists fails.
4. **Edge cases map to tests.** Each `edgeCases` entry in evidence is cross-checked against an existing test name or assertion in the feature's test files. An edge case with no corresponding test fails (warning vs. hard-fail policy decided in plan, but divergence is reported).

Behavioral guarantees:
- `centinela verify <feature>` prints a per-claim PASS/FAIL report and exits non-zero if any check fails.
- `centinela complete <feature>` runs verification as part of its gate and **hard-blocks** advancement on any failure (no warn-only bypass).
- A clean, truthful workflow verifies green and completes unchanged — verification must not produce false positives on honest evidence.

## Edge Cases — invalid input, concurrency, empty state, limits

- **No evidence files yet** (step not delegated): verification reports "no claims to verify" and does not block steps that legitimately have no evidence contract (e.g. `code`).
- **Claim type absent** (evidence omits coverage or edgeCases): skip that check; absence of a claim is not a failure.
- **Test command itself missing/misconfigured** in `validate.commands`: surface as a configuration error, distinct from a failed test claim.
- **Worktree context:** verification must run against the active worktree's tree and per-feature state, not the root checkout, when `use_worktrees` is on.
- **Non-deterministic / slow suites:** re-running tests may be expensive; need a documented timeout and a way to scope the run.
- **Stub detection false positives:** legitimately tiny files (interfaces, single-line helpers) must not be flagged as stubs.
- **Coverage re-derivation drift:** re-measured coverage may differ slightly from the claim due to ordering; define an exact-vs-tolerance rule.

## Data Model — entities, key fields, relationships

- **Evidence (existing):** `.workflow/<feature>-<role>.json` — `status`, `inputs`, `outputs`, `edgeCases`, claim fields. Input to verification.
- **VerificationResult (new):** per-feature aggregate of per-claim `Check` results — `{ claim, role, status: pass|fail|skip, detail }`. Lives in `internal/` (domain), rendered by `internal/ui`, never decided in `cmd/`.
- **Ground truth sources:** test command exit codes, per-package coverage output, output-file contents, test-file AST/symbol names.

## Integration Points — APIs, events, external services

- `internal/orchestration/` — verification plugs into the existing `complete` gate alongside structural evidence validation.
- `internal/config/` — reads `validate.commands` for the test/coverage commands to re-run.
- `internal/worktree/` — resolves the correct tree/state when worktrees are active.
- `cmd/centinela/` — thin `verify` command + wiring into `complete`; no business logic.
- `internal/ui/` — renders the PASS/FAIL report.

## Risks — performance, security, unclear requirements

- **Performance:** re-running the suite inside `complete` doubles test cost. Mitigation: reuse the run already done at the `validate` step where possible; allow `verify` to consume cached results.
- **Stub heuristics:** content/assertion detection is heuristic and language-aware (Go first). Over-aggressive rules cause false failures and erode trust — must be conservative and well-tested.
- **Coverage tolerance:** an exact-match rule risks flaky failures; a loose tolerance risks letting gaming through. Needs a deliberate, documented threshold.
- **Scope creep:** v1 is four checks, Go-first. Multi-language stub/coverage detection is explicitly out of scope.

## Decomposition — if large, list sub-feature slugs to split into

Single feature is workable, but if the plan step finds it too large, split along claim type:
- `claim-verification-core` — VerificationResult model + `verify` command + complete-gate wiring + tests-pass and outputs-stub checks.
- `claim-verification-coverage` — coverage re-derivation check.
- `claim-verification-edgecases` — edge-case-to-test mapping check.
