# Feature: spec-traceability-gate

- surface: internal
- status: planned
- roadmap: Phase 3 — Close the Mechanical-Verification Gap
- fixes: spec and acceptance tests drift apart; scenarios silently go unimplemented

## Problem

Centinela's spec-first gates (G5/G6) require a `.feature` file to exist before
code, but nothing mechanically verifies that each **scenario** inside it is
actually exercised by an acceptance test. Scenarios are added to a spec and
then silently never implemented — the spec says the behavior is guaranteed, the
test suite never checks it, and no gate catches the gap. This is the last
"requested, not enforced" hole in the mechanical-verification phase, after
g2-import-graph-gate and security-gate.

## What it verifies

Every `Scenario:` / `Scenario Outline:` in `specs/*.feature` maps to an
acceptance test that runs as part of the executed acceptance suite. The match
uses the convention this repo already follows (and which this gate makes
**enforced** rather than incidental):

- A test file under `tests/acceptance/` declares which spec it covers with a
  header comment: `// Acceptance: specs/<slug>.feature`
- Each covering test carries a `// Scenario: <exact scenario name>` comment
  immediately above the test function.

A scenario is "covered" when some acceptance test file whose header points at
its spec carries a `// Scenario:` comment whose normalized text equals the
scenario name. Because `tests/acceptance/` is run by `go test ./tests/acceptance/...`
in `[validate] commands`, a matched scenario is mapped to a test that the suite
actually executes.

## The central constraint: dogfooding without a baseline explosion

Centinela's own repo has 74 spec files / ~208 scenarios, but only ~36 spec
files have any acceptance test, and per-scenario coverage within those is
partial. A strict, whole-repo version of this gate would fail Centinela's own
`centinela validate` on 100+ pre-existing scenarios the day it ships — the
classic "turn on a new gate, drown in legacy findings, disable the gate"
failure (the same problem `audit-baseline-ratchet` exists to solve generally).

This feature MUST ship in a way that does not force us to either (a) disable it
or (b) game it by back-filling stub tests. The chosen answer: the gate is
**diff-aware** (reusing the existing G1/G11 diff machinery) and **default
disabled**. Enabled on Centinela itself in diff-aware mode, it gates only
scenarios in spec files changed on the current branch — so new and modified
scenarios must be traceable, while the legacy backlog is left for a later
ratchet/backfill rather than blocking every unrelated feature.

## Goal

A configurable built-in gate (`[gates.spec_traceability]`) that, when enabled,
fails `centinela validate` naming each uncovered scenario, scoped by the same
diff-aware rules as the other file-walking gates, and dogfooded on Centinela's
own repo without back-filling or disabling.

## Non-goals (v1)

- **Per-scenario runtime execution proof.** v1 proves a scenario maps to a test
  in the executed acceptance suite, not that that specific test ran and passed
  at runtime (would require `go test -json` test-name correlation). Noted as a
  future enhancement.
- **Step-level (Given/When/Then) traceability.** Scenario-level only.
- **A whole-repo ratchet/baseline.** That is `audit-baseline-ratchet`'s job;
  this gate only needs to coexist with it (diff-aware scoping is the bridge).
- **Auto-generating missing acceptance tests.** The gate reports gaps; it does
  not write tests.
- **Non-Go acceptance runners.** Matching targets `tests/acceptance/*.go`
  comments; other languages are a later extension.
