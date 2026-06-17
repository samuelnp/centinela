# Documentation Specialist — spec-traceability-gate

## KB entry

Authored `docs/project-docs/kb/spec-traceability-gate.md` (audience: end-user,
status: done) describing the new opt-in gate from an operator's point of view —
someone who configures gates in `centinela.toml` and runs `centinela validate`.

- **Summary**: an opt-in gate that runs during `centinela validate` and flags
  any spec scenario with no acceptance test backing it, with configurable
  warn-or-fail severity and diff-aware scoping so only changed specs are checked.
- **What it does / When you'd use it / How it behaves** all present (required by
  the generator).
- **How it behaves** covers every spec scenario as observable behavior: covered
  scenario passes and reports a count; uncovered scenario is named (with its spec
  file); name matching tolerates trailing periods, spacing, and case; a header's
  trailing annotation after the filename is ignored; a Scenario Outline counts
  once; `warn` severity reports gaps without blocking (Centinela's own default);
  diff-aware mode only checks changed specs; the gate skips when nothing is in
  scope; an unknown severity value is rejected at config load; and the gate is
  enabled on Centinela itself in `warn` mode.
- **Examples**: the `[gates.spec_traceability]` toml snippet (enabled +
  severity=warn) and the acceptance-test convention (`// Acceptance:
  specs/<slug>.feature` header + `// Scenario: <name>` comment above the test).

Prose stays in plain operator language — no Given/When/Then, no Go package or
function names. House style matches the existing `security-gate.md` /
`g2-import-graph-gate.md` entries.

## Generated outputs (all confirmed to exist)

- `docs/project-docs/kb/spec-traceability-gate.md` — KB source (3.8K)
- `docs/project-docs/kb/spec-traceability-gate.html` — rendered KB page (7.6K)
- `docs/project-docs/index.html` — regenerated portal index (127.7K)
- `kb/index.html` also regenerated as part of `docs generate`.

`centinela docs validate` passes (inputs valid) and `docs generate` reported the
index written successfully.

## Mermaid

None added. The gate is an internal verification mechanism, not a
project-domain feature/spec relationship, so a Mermaid diagram would add no
reader value here.

## Note on docs-step weight

For an internal gate feature like this, the full HTML-portal regeneration is
heavier than its reader value — the single KB markdown entry is the load-bearing
artifact, while re-rendering the whole portal index is incidental cost. This
nods to the roadmap's right-size-docs-step item: internal-surface features could
justify a lighter docs path than full-portal regen.
