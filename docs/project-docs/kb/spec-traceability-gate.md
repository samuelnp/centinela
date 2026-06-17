---
feature: spec-traceability-gate
summary: An opt-in gate that runs during `centinela validate` and flags any spec scenario that has no acceptance test backing it, with configurable warn-or-fail severity and diff-aware scoping so only changed specs are checked.
audience: end-user
status: done
---

## What it does

The spec-traceability gate is an opt-in check that runs as part of `centinela validate`. When enabled, it reads every scenario in your spec files and confirms each one is actually exercised by an acceptance test in the executed suite. If a scenario has no covering test, the gate reports it — either as a hard failure that blocks validate, or as a non-blocking warning, depending on the severity you configure. It closes the gap where a scenario is added to a spec, says a behavior is guaranteed, and then is silently never implemented.

## When you'd use it

Turn this gate on when your project keeps Gherkin spec files and you want a mechanical guarantee that every scenario you write is genuinely tested — rather than trusting code review to notice that a new scenario never got an acceptance test. It is the safety net for spec drift: specs and tests that quietly fall out of sync, leaving "verified" behavior that nothing actually verifies.

## How it behaves

- When a scenario has a matching acceptance test, the gate counts it as covered; if every in-scope scenario is covered, the gate passes and reports the covered-scenario count.
- When a scenario has no acceptance test, the gate reports the gap and names both the uncovered scenario and the spec file it lives in.
- Scenario-name matching is forgiving: a trailing period, extra spaces, and differences in letter case are all normalized away, so a test still matches its scenario despite cosmetic differences.
- A test's spec reference still matches even when the spec filename is followed by a trailing annotation (for example, a note listing which acceptance criteria the test covers) — only the filename portion is used for matching.
- A Scenario Outline with an examples table is treated as a single scenario for coverage, so one covering test is enough — it is not counted once per example row.
- Under `warn` severity, an uncovered scenario is reported as a warning and listed in the details, but validate is not blocked. This is Centinela's own default while a pre-existing backlog of untested scenarios is paid down gradually.
- In diff-aware mode, only spec files changed on the current branch are checked: an unchanged spec with an uncovered scenario is left alone, so unrelated legacy gaps don't block the work in front of you.
- When no spec files fall within the gate's scope, the gate is skipped with an explanatory message rather than passing or failing on nothing.
- An unsupported severity value is rejected when the configuration loads, with an error that names the offending field — so a typo in the config can't silently disable enforcement.
- The gate is enabled on Centinela's own repository in `warn` mode, so its CI surfaces the legacy coverage gaps without blocking every feature.

## Examples

Enable the gate in `centinela.toml`:

```toml
[gates.spec_traceability]
enabled  = true
severity = "warn"   # "warn" surfaces gaps without blocking; "fail" blocks validate
```

An acceptance test declares which spec it covers and which scenario it exercises with two header comments:

```go
// Acceptance: specs/spec-traceability-gate.feature
// Scenario: A scenario with a matching acceptance test passes the gate
func TestScenarioWithMatchingAcceptanceTestPasses(t *testing.T) {
    // ...
}
```

The `// Acceptance:` line points at the spec file, and each `// Scenario:` line above a test names the exact scenario it covers. The gate matches them up — tolerating trailing periods, spacing, and case differences — to decide which scenarios are traceable to a real test.
