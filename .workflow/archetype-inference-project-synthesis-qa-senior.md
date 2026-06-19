# archetype-inference-project-synthesis — qa-senior

## Test Inventory

**Unit (colocated, `internal/synthesize`):** infer_test (all 5 archetypes from
fixture inventories + empty→custom + ambiguous tie + deterministic order +
Reasons), draft_test (full n-tier section assertions, empty→TODO stubs +
byte-stability, ambiguous banner + unknown-language naming, matchPaths no-hit),
write_test (fresh write, never-clobber, un-writable path, draftPathFor). All
fixture-driven — no LLM, no network.

**Contract (`internal/analyze`):** load_test (round-trip via Save, missing→
ErrNoInventory, malformed JSON, schema drift, directory→read error).

**Command (`cmd/centinela`):** synthesize_test (writes draft, --json) +
synthesize_errors_test (missing→guides to analyze, existing PROJECT.md preserved,
malformed distinct error).

**Integration (`tests/integration`):** synthesize_pipeline_test — real
analyze.Analyze → Save → Load → Infer → Draft → WriteDraft on a Go n-tier fixture
(file-system boundary).

**Acceptance (`tests/acceptance`):** synthesize_{helper,happy,edge}_test drive
the real built binary as subprocesses; carry
`// Acceptance: specs/archetype-inference-project-synthesis.feature` +
a `// Scenario:` per the 7 spec scenarios. All 7 pass.

## Coverage Gaps

Per-package: **synthesize 99.4%**, **analyze 95.3%**, cmd/centinela 93.3% (its
baseline as a large multi-command package; the new synthesize.go is well
covered). The only intentionally-uncovered synthesize lines are unreachable map
defaults. Aggregate `check-coverage.sh` gate verified ≥95%.

## Acceptance Wiring

`centinela.toml` `validate.commands` already runs `go test ./tests/acceptance/...`,
so the suite executes in validate. The spec-traceability gate (warn mode) now
maps a `// Scenario:` to each of the 7 feature scenarios.

## Handoff

→ validation-specialist: full suite green, coverage gate ≥95%, gofmt clean, all
test files ≤100 lines, dogfooded `centinela synthesize` produces a correct n-tier
draft and honestly infers `custom` on this repo's unconventional layout. Ready
for the gatekeeper + validate gate run.
