# g2-import-graph-gate — documentation-specialist

## Summary

Authored the end-user KB entry for the import-graph gate feature. The KB page explains that Centinela can now automatically check that a Go project's packages respect configured architectural layering rules, and fails `centinela validate` when a forbidden dependency is detected. The page covers the opt-in `[gates.import_graph]` config block, all significant behavioral scenarios, and includes a sample `centinela.toml` snippet with a working example failure line.

## Spec coverage

The spec (`specs/g2-import-graph-gate.feature`) contains **12 scenarios** covering:

1. All imports respect the layer matrix — Pass
2. Forbidden cross-layer import (Scenario Outline with 3 examples) — Fail with edge details
3. Multiple forbidden edges all listed — Fail
4. No `[gates.import_graph]` block — gate omitted
5. `enabled = false` — gate omitted
6. Unmapped package — Warn
7. Malformed config (Scenario Outline with 3 examples) — Fail with `import_graph config:` prefix
8. Empty matrix (block present, zero layers) — Warn
9. Uncompilable code — Fail with load error
10. Standard-library / third-party imports — ignored
11. Test file (`_test.go`) external package — same-layer assignment, still checked
12. Intra-layer import — always allowed
13. Violation outside current diff — still reported (whole-module load)

All 12 scenarios are represented in the KB's "How it behaves" bullet list, rewritten in plain end-user language without Gherkin or internal technical jargon.

## Workflow status

- Inputs read: docs/features/g2-import-graph-gate.md, docs/plans/g2-import-graph-gate.md, specs/g2-import-graph-gate.feature, docs/project-docs/kb/cross-platform-build-gate.md (tone reference)
- Outputs written: docs/project-docs/kb/g2-import-graph-gate.md
- HTML generated: docs/project-docs/kb/g2-import-graph-gate.html, docs/project-docs/kb/index.html, docs/project-docs/index.html
- Evidence fields: inputs, outputs, handoffTo=complete all populated
- handoffTo: complete
