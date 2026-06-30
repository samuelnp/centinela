# mcp-governance-server — validation-specialist

## Gates Run

`centinela validate` (diff-aware, 31 files changed) — all green:
- ✓ G1: File Size · ✓ G-Build: Cross-Compile (6 targets)
- ✓ `go test ./...` · ✓ `go test ./tests/acceptance/...`
- ✓ `./scripts/check-coverage.sh` (95.0% ≥ 95.0%; `internal/mcp` 100%) · ✓ fmt
- ⚠ `import_graph`, ⚠ `spec-traceability-gate` — empty-body, non-blocking,
  pre-existing in diff-aware mode. roadmap_drift in sync.

## Synthesis

Additive MCP governance server on the official Go SDK (v1.2.0), reusing
`verdict.AssembleVerdict` so MCP and native verdicts are structurally identical.
Gatekeeper: SAFE. The decisive properties are verified end-to-end against a
zero-integration harness (SDK client over `mcp serve`): tools/list shows the 4
versioned tools, the shim maps block→exit 2 / allow→exit 0, and
`TestAccMcpParityWithNative` confirms MCP `Combine` ≡ native `Decide` on the same
packet. The new dependency is pinned and isolated for one-package upgrades.

## Decision

PASS → hand off to documentation-specialist.
