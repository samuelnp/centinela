# g2-multi-language-import-graph — feature-specialist

## Behavior Summary

A new leaf package `internal/importgraph` holds the `GraphProvider` abstraction,
four backends (go/node/python/script), manifest-based selection, and an
injectable `Runner` exec seam. `internal/gates` keeps the matrix/edge-check
logic and delegates graph *loading* to the leaf. Backends pre-scope to
module-relative `Pkg{Path, Imports}`, so the gate's existing `checkEdges` runs
unchanged. Config gains `Provider string` + `ScriptCommand []string`; unset =
auto-select (Go repo stays byte-identical). Full file-by-file breakdown — 14
source + ~12 test files, each ≤100 lines — is in
`docs/plans/g2-multi-language-import-graph.md`; the create/modify set is in the
JSON `outputs` field.

## Acceptance Criteria (Gherkin)

See `specs/g2-multi-language-import-graph.feature` (7 scenarios): Go enforced
(auto-select), Node enforced (provider=node), Python enforced (provider=python),
no-manifest → WARN + exit 0, custom-script enforced, tool-missing → WARN + exit
0, empty-matrix → WARN before selection.

## UX States

Gate outcomes surfaced to the operator: **Pass** (no forbidden edges), **Fail**
(forbidden edge, or real load error / malformed output / non-zero script exit),
**Warn** (no provider matched, external tool missing, or empty matrix). Warn
messages name the cause and the remedy (e.g. "configure
gates.import_graph.provider", "install dependency-cruiser or set provider=script").

## Edge Cases

No/multiple manifests; explicit provider but tool absent; malformed tool output;
custom-script nonzero exit / empty-valid output; cyclic graph; monorepo
(top-level only); Go-repo parity; broken go.mod still Fails; empty matrix;
backend timeout; OS-specific/absolute paths normalized. (Mirrored in JSON
`edgeCases`.)

## Out-of-Scope

Java/Rust/Ruby/PHP/C#/Elixir/Kotlin built-in backends (use custom-script until
`non-go-source-import-graphs` is promoted); monorepo per-subtree selection;
bundler/path-alias resolution; transitive third-party edges; auto config
migration.

## Handoff

→ senior-engineer: implement in plan file order; keep the Go path green at every
step (run `go test ./internal/gates/...` after the seam refactor, before adding
new backends).
