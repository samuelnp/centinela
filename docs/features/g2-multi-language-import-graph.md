# Feature: g2-multi-language-import-graph

**Phase:** Phase 3 — Close the Mechanical-Verification Gap
**Archetype:** n-tier
**Status:** plan

## Problem

The G2 import-graph gate mechanically enforces layer-dependency rules, but its
only graph backend is hardcoded to Go (`go list -m` + `go list -json ./...` in
`internal/golist`). On any non-Go project that enables `[gates.import_graph]`,
the gate shells out to `go`, gets a non-zero exit, and **hard-fails**
`centinela validate` — breaking the gate for the very polyglot projects
Centinela aims to govern. The gate's core value (catching forbidden cross-layer
imports) is currently Go-only.

## User value

Non-Go teams (Node/TS, Python, and — via an escape hatch — any language) can
enforce the same layer-boundary rule that is Centinela's flagship promise,
instead of hitting a cryptic `go list` failure. Go projects are unaffected.

## What ships

1. **Self-skip + WARN** when no graph provider matches the project — stops the
   non-Go hard-fail. A missing external tool also WARNs (CI stays green).
2. A pluggable **`GraphProvider`** seam (new leaf package `internal/importgraph`)
   with an injectable command `Runner` for testability. The existing
   `go list` logic becomes the **reference backend** behind it.
3. **Manifest-driven auto-selection** (`go.mod` → go, `package.json` → node,
   `pyproject.toml`/`requirements.txt` → python) via the leaf's own minimal
   detector, plus an explicit `provider` config override.
4. **JS/TS+Node** backend (shells out to `dependency-cruiser`/`madge`, parses
   JSON) and **Python** backend (AST import-graph walker).
5. A **custom-script** provider: for languages without a built-in backend, the
   user configures a command that emits the import graph in a defined JSON
   contract — tying into Centinela's existing custom-gate philosophy.

## Explicitly deferred

- The remaining built-in languages (Java, Rust, Ruby, PHP, C#, Elixir, Kotlin)
  — these are served by the custom-script provider until promoted from the
  backlog item `non-go-source-import-graphs`.
- Monorepo per-subtree provider selection (top-level manifest only).
- Bundler/path-alias resolution beyond what the JS tool emits; transitive
  third-party edge analysis; automatic layer-path migration.

## Backward compatibility

This repo is a Go project that already enables G2. With `provider` unset it
auto-selects the `go` backend and behaves byte-identically to today; all
existing `import_graph_*_test.go` stay green. A parity test locks this.
