# spec-reconstruction — qa-senior

**Date:** 2026-06-22

## Test Inventory

| Tier        | File | Scenarios |
|-------------|------|-----------|
| unit        | `internal/reconstruct/select_test.go`, `select_exclude_test.go` | Go n-tier roles+sort, empty/doc-only → 0, polyglot empty-graph, graph in-edge promotion, exclusion precedence, bounding, collisions |
| unit        | `internal/reconstruct/rules_test.go`, `signals_test.go` | promote/exclude predicates + role hints; signal accessors |
| unit        | `internal/reconstruct/feature_test.go`, `brief_test.go`, `templates_test.go`, `slug_test.go` | role-aware Gherkin parses with real `spec_traceability` parser, `# TODO: confirm` present, no fabricated steps; brief stub shape; templates; slugify+disambiguate |
| unit        | `internal/reconstruct/reconstructor_test.go`, `write_test.go` | Reconstruct end-to-end determinism + TodoCount; WriteCorpus skip-if-exists, review-dir write, byte-identical re-run |
| unit        | `cmd/centinela/reconstruct_test.go`, `reconstruct_errors_test.go` | happy + `--json`; no-inventory → ErrNoInventory, non-zero exit, no files |
| unit        | `internal/ui/render_reconstruct_test.go` | summary rendering (targets/written/skipped/TODO) |
| integration | `tests/integration/reconstruct_pipeline_test.go` | real analyze → Save → Load → Reconstruct → WriteCorpus; review-dir corpus + byte-identical re-run |
| acceptance  | `tests/acceptance/reconstruct_{helper,happy,edge}_test.go` | all 9 `.feature` scenarios via the real binary |

## Coverage Gaps

- None at the spec level: all 9 scenarios in `specs/spec-reconstruction.feature`
  have an acceptance test with `// Acceptance:` + `// Scenario:` traceability
  comments (verbatim names).
- Coverage gate: `./scripts/check-coverage.sh` → **passed: 95.2% ≥ 95.0%**.
  `internal/reconstruct` is at **99.3%**. `internal/ui` (91.8%) and
  `cmd/centinela` (93.4%) remain below 95% in isolation — a pre-existing
  baseline the new code does not regress; the total clears the gate.

## Acceptance Wiring

`centinela.toml` `[validate].commands` already runs the acceptance tier:

```toml
commands = [
  "go test ./...",
  "go test ./tests/acceptance/...",
  "./scripts/check-coverage.sh",
  "./scripts/check-fmt.sh",
]
```

`go test ./tests/acceptance/ -run AccRecon` → 9 passed.
`go test ./tests/integration/ -run Reconstruct` → 1 passed.

## Deferred Findings

- none (route/flow extraction already deferred as `brownfield-route-flow-extraction`).

## Handoff

- Next role: validation-specialist
- Edge-case report: `.workflow/spec-reconstruction-edge-cases.md`
