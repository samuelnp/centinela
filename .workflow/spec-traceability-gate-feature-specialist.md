### Feature-Specialist Report: spec-traceability-gate
**Date:** 2026-06-10

#### Behavior Summary
The spec-traceability gate is a config-gated, diff-aware built-in gate that maps every in-scope `Scenario:`/`Scenario Outline:` in `specs/*.feature` to a covering acceptance test via the `// Acceptance: specs/<slug>.feature` header plus a `// Scenario: <name>` comment, keyed by (slug, normalized name). It honors the same diff filter as G1/G11 — locally diff-aware (only changed specs in scope), in CI full-scan (nil filter walks all specs). On Centinela it ships `enabled=true, severity="warn"`: CI surfaces the 397-scenario legacy backlog as warnings without blocking merge, while the local diff-aware dogfood requires this branch's own 10 scenarios to be genuinely covered. Matching parses defensively (trailing header annotations ignored, case-insensitive, trim/collapse whitespace, strip one trailing period) so the loosely-followed convention is enforced going forward without retroactively breaking unchanged specs.

#### Gherkin Scenarios
References `specs/spec-traceability-gate.feature` (10 scenarios):
- **A scenario with a matching acceptance test passes the gate** — Given a spec scenario with a matching acceptance comment / When the gate runs over that spec / Then it passes and the message reports the covered count. (asserts `Result.Status==Pass` + `Message`).
- **A scenario with no acceptance test fails the gate** — Given an uncovered scenario / When the gate runs / Then it fails and Details name the scenario + spec file. (asserts `Result.Status==Fail` + `Details`).
- **Matching normalizes trailing period, spacing, and letter case** — Given scenario "Start the watcher" and comment "// Scenario:  start the WATCHER ." / When the gate runs / Then it passes and the scenario is covered. (asserts normalization yields Pass — observable, not an internal helper).
- **An acceptance header with a trailing annotation still matches its spec** — Given header "// Acceptance: specs/spec-traceability-gate.feature (AC4, AC5)" + a matching `// Scenario:` comment / When the gate runs / Then the annotation is ignored and the scenario is covered. (asserts defensive header parse via Pass).
- **A Scenario Outline counts as one covered scenario** — Given a spec with a Scenario Outline + examples table / When the gate evaluates coverage / Then the outline is one scenario for matching. (asserts a single covered/uncovered unit, not N examples).
- **Warn severity reports gaps without failing** — Given severity=warn (Centinela's dogfood default) and an uncovered scenario / When the gate runs / Then status is warn rather than fail and the uncovered scenario is still in Details. (asserts `Result.Status==Warn` + `Details`).
- **Diff-aware scope gates only changed spec files** — Given an unchanged spec with an uncovered scenario + a changed spec all covered / When the gate runs diff-aware / Then the unchanged spec is not gated and the gate passes. (asserts filter excludes unchanged → Pass).
- **No spec files in scope skips the gate** — Given no spec files in scope / When the gate runs / Then it is skipped with an explanatory message. (asserts `Result.Status==Skip` + `Message`).
- **An unknown severity value is rejected at config load** — Given a centinela.toml with an unsupported severity / When config loads / Then loading fails naming the severity field. (asserts `validateSpecTraceability` returns an error).
- **The gate is registered and enabled for Centinela in warn mode** — Given Centinela's own centinela.toml / When configured gates are read / Then the gate is enabled and its severity is warn. (asserts `Enabled==true` AND `Severity=="warn"` — not a false `fail` claim).

#### UX States
| State | Trigger | Surface |
|-------|---------|---------|
| Pass | All in-scope scenarios covered | `centinela validate`: `spec_traceability` PASS line with covered count |
| Warn | Uncovered scenario(s) and `severity="warn"` (Centinela default) | `centinela validate`: `spec_traceability` WARN line; Details list each `specs/<slug>.feature: "<scenario>"` |
| Fail | Uncovered scenario(s) and `severity="fail"` | `centinela validate`: `spec_traceability` FAIL line; Details list each uncovered scenario + spec; non-zero exit |
| Skip | No `.feature` files in scope (e.g. diff-aware local run with no spec changes) | `centinela validate`: `spec_traceability` SKIP line with diff-aware explanatory message |
| Config error | Unknown `severity` value | n/a at gate runtime — surfaced at config load, naming the severity field; validate aborts before gate execution |

#### Out-of-Scope
- Per-scenario runtime execution proof (no `go test -json` test-name correlation; v1 proves a scenario maps to a test in the executed suite, not that that test ran/passed).
- Step-level (Given/When/Then) traceability — scenario-level only.
- A whole-repo ratchet/baseline — that is `audit-baseline-ratchet`'s job; this gate only coexists with it via diff-aware scoping.
- Scenario-Outline example expansion — an outline counts once.
- Auto-generating missing acceptance tests — the gate reports gaps, it does not write tests.
- Non-Go acceptance runners — matching targets `tests/acceptance/*.go` comments only.

#### Handoff
- Next role: senior-engineer
- Open clarifications: None blocking. Implementation is fixed by the plan — config leaf in `internal/config/spec_traceability.go` (Normalize + validate, reject unknown severity); parse/match/entry split across `internal/gates/spec_traceability_parse.go`, `spec_traceability_match.go`, `spec_traceability.go` (each ≤100 lines, G1); register in `gates.go` `RunWithFilter` honoring the same `*gitdiff.Set` filter as G1; enable in `centinela.toml` with `enabled=true, severity="warn"`. Defensive matcher must strip trailing header annotations after the filename and normalize names case-insensitively (trim, collapse whitespace, strip one trailing period). qa-senior's `tests/acceptance/spec_traceability_gate_test.go` must honestly cover all 10 scenarios above (dogfood closure) so the diff-aware local run reports them covered.
