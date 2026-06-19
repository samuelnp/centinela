# g2-multi-language-import-graph — big-thinker

## Problem

The G2 import-graph gate — the mechanical enforcement of Centinela's flagship
separation-of-concerns promise — has only one graph backend, hardcoded to Go
(`internal/golist` → `go list -m` / `go list -json ./...`). Any non-Go project
that enables `[gates.import_graph]` hard-fails `centinela validate` on a
`go list` error. We must stop the hard-fail AND make enforcement work for
non-Go languages without regressing the Go path.

## Scope

In: a pluggable `GraphProvider` seam (new **leaf** package
`internal/importgraph`); the Go backend as reference implementation;
manifest-driven auto-selection; JS/TS+Node and Python backends; a custom-script
provider escape hatch; self-skip-with-WARN when no provider matches.
Out: the other 7 built-in languages (served by the custom-script provider until
the `non-go-source-import-graphs` backlog item is promoted), monorepo per-subtree
selection, bundler/alias resolution, auto layer-path migration.

## Dependencies & Assumptions

Builds on the shipped `g2-import-graph-gate` and the existing custom-gate
mechanism. **Layer decision:** `internal/importgraph` is a leaf (stdlib +
`os/exec` + `internal/golist`) with its own minimal manifest detector, so
`internal/gates` (domain) keeps importing leaves only — no domain→domain edge,
no cycle (`analyze` never imports `gates`/`importgraph`). Every backend returns
an already-scoped module-relative `Graph{Module, Pkgs[]Pkg}` so the gate matrix
is provider-agnostic. Assumes external tools (depcruise/madge, python3) may be
absent on CI → backends take an injectable `Runner`.

## Risks

Shelling out to external parsers (env coupling, flaky output) — mitigated by the
injectable `Runner` + skip-guarded integration tests. Per-package 95% coverage —
pure parse functions carry coverage. Silent edge loss — malformed output and
non-zero exits map to `Fail`, never a false Pass.

## Rollout

The gate's `Warn` status is already non-failing (`AllPassed` only fails on
`Fail`). No-provider / tool-missing → `Warn`; real load failure → `Fail`. This
repo (Go, G2 enabled) auto-selects `go` with `provider` unset and behaves
byte-identically; a parity test locks the regression.

## Handoff

→ feature-specialist: file-by-file plan under the ≤100-line rule, the injectable
`Runner` test seam, and the Gherkin acceptance spec. Edge cases enumerated in the
JSON `edgeCases` field.
