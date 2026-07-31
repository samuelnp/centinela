---
id: fabd37be9d8eb127
feature: docstring-gate
step: tests
type: lesson
title: - Undocumented exported function in a changed file fails, naming file:line,
tags: edge-cases, lesson
sourceArtifact: .workflow/docstring-gate-edge-cases.md
createdAt: 2026-07-31T11:36:54Z
---

- Undocumented exported function in a changed file fails, naming file:line,
- Fully documented changed file passes, message names the inspected count —
- Warn severity keeps `centinela validate` green while still reporting the
- A legacy unchanged file with undocumented exports is never opened when a
- An empty changed-file scope reports Skip, never a confident Pass —
- A changed-file scope with no Go files (e.g. a docs-only branch) reports
- `_test.go` and `// Code generated ... DO NOT EDIT.` files are excluded;
- An unresolvable merge base (no baseline branch) reports Skip naming the
- CI full-scan mode hands every gate a nil filter; the docstring gate
- A grouped `const ( A; B )` block's single doc comment covers every spec
- `//centinela:nodoc` exempts an identifier AND the exemption is listed on
- Undocumented exported struct fields do not fail the gate while
- An unparseable `.go` file is reported by path, never silently dropped from
- An unknown `severity` value is rejected at config load, naming the field
- `centinela docs lint` exits 1 on Fail and 0 on Warn while reporting the
- `centinela docs lint --full` is report-only: always exit 0, prints the
- `centinela docs lint --json` shape (`scope`, `status`, `undocumented`,
- This repository's own `centinela.toml` ships `[gates.docstring]
- `docs/architecture/senior-engineer-prompt.md` states the doc-comment duty
- Colocated (pre-existing) unit coverage the acceptance tier does not
- The 171-item whole-repo legacy backlog (`centinela docs lint --full`
- `package-doc-comments` (a per-package, not per-file, doc-comment rule) and
- Multi-language scanning is a named seam (`docstring.Register`/`For`) with
- Hunk-level (line-range) diff scope was rejected at plan time in favor of
- No new coverage gaps were found beyond what the plan and senior-engineer
