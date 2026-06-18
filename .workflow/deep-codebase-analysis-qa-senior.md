### QA-Senior Report: deep-codebase-analysis
**Date:** 2026-06-18

Adds the full three-tier test suite for `centinela analyze`. Per-package coverage
is moved by **colocated** `_test.go` files inside each new package (the gate is on
the overall total via `go test ./...`, but `tests/`-tier files do not lift a new
package's number, so colocated tests are mandatory to keep the total ≥95%).
`internal/golist` reaches **100%** and `internal/analyze` **95.1%**. All G1-gated
`_test.go` files under `internal/**` and `cmd/**` are ≤100 lines.

#### Test Inventory

| Tier | File | Scenarios / focus |
|------|------|-------------------|
| unit | `internal/analyze/languages_test.go` | ext→language mapping; count-desc/name-asc sort; empty⇒primary "" |
| unit | `internal/analyze/manifests_test.go` | detection + sort-by-path; npm scripts/deps/framework; malformed package.json detected-but-unparsable; none-when-absent |
| unit | `internal/analyze/extract_misc_test.go` | go.mod module, Cargo/Gemfile/pyproject/requirements deps, Makefile build/test |
| unit | `internal/analyze/locales_test.go` | root+nested locale dirs, subdir codes, no-i18n⇒empty |
| unit | `internal/analyze/walk_test.go` | skip set; depth bound; gitignore; symlink skip; unreadable-root hard error |
| unit | `internal/analyze/gitignore_test.go` | absent⇒empty; path/name/dir-prefix matching; negation lines ignored |
| unit | `internal/analyze/graph_test.go` | Go fixture edges; go-list-failure best-effort + Note; non-Go declared-deps; none |
| unit | `internal/analyze/inventory_test.go` | byte-stable Save across re-runs; schemaVersion present; un-writable path errors |
| unit | `internal/analyze/analyze_test.go` | polyglot assembly; failing sub-detector still valid; empty repo; unreadable root |
| unit | `internal/golist/golist_test.go` | ModulePath/Packages decode; error surfacing; firstStderrLine |
| unit | `internal/golist/golist_fake_test.go` | PATH-stubbed go: empty-stderr fallback + decode-error paths |
| unit | `internal/ui/render_analyze_test.go` | full summary fields; empty⇒(none) + graph note |
| unit | `cmd/centinela/analyze_test.go` | happy-path writes JSON+summary; `--out` override |
| unit | `cmd/centinela/analyze_errors_test.go` | un-writable out; unreadable root (0o311) hard errors |
| integration | `tests/integration/analyze_test.go` | mini Go module + package.json + locales/: Go primary, module path, npm scripts, locales, non-empty edges |
| integration | `tests/integration/analyze_determinism_test.go` | byte-identical re-run (AC-3/4); no source mutated (AC-6, sha256 witness) |
| acceptance | `tests/acceptance/analyze_happy_test.go` | scan-writes-inventory; deterministic-rerun; `--out` override |
| acceptance | `tests/acceptance/analyze_edge_test.go` | no-manifest-still-valid; skips-vendor-deps + read-only; un-writable-out-fails; unreadable-root-fails |
| acceptance | `tests/acceptance/analyze_helper_test.go` | one-shot binary build + Go-module fixture helpers |

Each acceptance `Test*` carries a `// Scenario: <name>` mapping 1:1 to a Gherkin
scenario title. The edge→scenario→test map is in
`.workflow/deep-codebase-analysis-edge-cases.md`.

#### Coverage Gaps

None blocking. The 15 spec scenarios are each asserted by at least one executable
test (acceptance for binary-level behavior, integration for in-process AC-3/4/6,
unit for the detection tables and best-effort branches). Two implementation
branches unreachable with the real `go` toolchain — `golist.Packages` decode
error and `runGo`'s empty-stderr fallback — are covered deterministically by a
PATH-stubbed fake `go` (`golist_fake_test.go`), following the repo's established
`internal/gates/security_fake_bin_test.go` pattern, taking `internal/golist` to
100%.

Note on in-process integration: `golist` operates on the process CWD, so the
integration tests chdir into the fixture root before calling `Analyze(".")` —
mirroring exactly how the binary is invoked (the acceptance tests exec the real
binary with the fixture as cwd).

#### Acceptance Wiring

`centinela.toml` `validate.commands` already runs the acceptance tier and the
coverage gate; no edit was needed:

```
"go test ./tests/acceptance/...",
"./scripts/check-coverage.sh",
```

#### Regression Guards

- Determinism (`TestSave_ByteStableAcrossReruns` + integration byte-identical
  re-run) guards against map-order / missing-sort regressions that would dirty
  git diffs on re-analysis.
- Read-only sha256 witness guards against the scan ever writing into source.
- `golist` error-surfacing tests guard the "never a false empty-but-valid graph
  on unloadable code" invariant shared with the import_graph gate.

#### Deferred Findings

none. (The big-thinker already deferred non-go-source-import-graphs,
brownfield-framework-fingerprinting, incremental-codebase-analysis, and
codebase-metrics-enrichment; no new gaps introduced at the tests step.)

#### Handoff

- Next role: validation-specialist
- `go build ./...` green; `go test ./...` = 2373 passed / 0 failed;
  `./scripts/check-coverage.sh` = passed (95.1% ≥ 95.0%); per-package
  golist 100% / analyze 95.1%.
- Edge-case report: `.workflow/deep-codebase-analysis-edge-cases.md` (enriched
  with the edge→test map).
