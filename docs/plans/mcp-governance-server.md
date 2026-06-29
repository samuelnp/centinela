# Plan: mcp-governance-server

## Summary

A versioned MCP server (`centinela.mcp/v1`) over stdio, built on the official
Go SDK, exposing 4 governance tools that reuse `internal/verdict.AssembleVerdict`
as the wire payload. Advisory-by-protocol: it returns `allow|warn|block`; a
`centinela mcp shim` client maps `block`→exit 2 to deny harness-side.

## Dependency

Add `github.com/modelcontextprotocol/go-sdk v1.2.0` (stdio). All SDK types are
confined to `internal/mcp` + `cmd/centinela/mcp*.go` so a future SDK bump touches
one package. API: `mcp.NewServer`, `mcp.AddTool(server, &mcp.Tool{...}, handler)`,
`server.Run(ctx, &mcp.StdioTransport{})`; client `mcp.CommandTransport` for tests.

## Architecture / layer placement (G2)

- `internal/mcp` (NEW, aggregator): imports `verdict` (aggregator), `gates`,
  `verify`, `workflow`, `config` + the external SDK; imported only by `cmd/`.
  Add `internal/mcp/**` to the aggregator layer `paths`.
- `cmd/centinela`: `mcp.go` (`mcp serve`), `mcp_shim.go` (`mcp shim`).

## Components (each ≤100 lines)

| File | Responsibility |
|------|----------------|
| `internal/mcp/schema.go` | `SchemaVersion = "centinela.mcp/v1"`; typed tool in/out structs with `jsonschema` tags |
| `internal/mcp/decision.go` | `Decide(*verdict.Packet) string` → allow/warn/block (pass→allow; any gate fail / verify FAIL→block; warns-only→warn) |
| `internal/mcp/deps.go` | `Deps` (verdict assembler + rules reader + workflow loader) so handlers are injectable/testable |
| `internal/mcp/tools_gates.go` | `read_rules`, `run_gates` handlers |
| `internal/mcp/tools_verify.go` | `verify_claims`, `workflow_state` handlers |
| `internal/mcp/server.go` | `NewServer(Deps) *mcp.Server` — registers the 4 tools |
| `cmd/centinela/mcp.go` | `centinela mcp serve` — build real Deps (gates.RunAll, verify, EvidenceIndex), run stdio |
| `cmd/centinela/mcp_shim.go` | `centinela mcp shim <feature>` — MCP client over `serve`, combine tool verdicts, exit 2 on block |
| `internal/setup/hooks.go` | optional: document `mcp shim` as a harness adapter (no forced wiring) |

## Tool surface (`centinela.mcp/v1`)

- `read_rules` → the governing rule surface (config knobs + archetype rules).
- `run_gates` → `{gates: [...], decision}` from `gates.RunAll`.
- `verify_claims{feature, step}` → `{checks: [...], decision}` from `verify.Verify`.
- `workflow_state{feature}` → `{run, evidence}` from the packet's RunInfo + EvidenceIndex.

The shim calls `run_gates` + `verify_claims`, combines (block if either blocks),
and exits 2/0 — the verdict reached purely through MCP tool calls.

## Test strategy

- **unit**: `Decide()` mapping table; schema-version constant; tool output structs
  marshal with expected JSON fields.
- **integration**: construct `NewServer` with stub `Deps` and invoke handlers
  directly (no transport) → assert tool outputs + decision for pass/fail/warn.
- **acceptance** (binary-driven, SDK client = "zero-Centinela-code harness"):
  1. `CommandTransport` spawns the built `centinela mcp serve`; `tools/list`
     shows the 4 tools; `run_gates`/`verify_claims` return a verdict.
  2. **shim**: `centinela mcp shim` exits 2 on a block-producing repo, 0 on allow.
  3. **parity**: the MCP-derived verdict equals the native `centinela verdict`
     packet's decision for the same feature/step + repo state.

## Rollout

Additive: new command tree + new dep. No migration. The server is opt-in (a
harness must invoke `mcp serve`); the shim is documented, not force-wired.

## Risks

- Young SDK → pin v1.2.0, isolate behind `internal/mcp`; acceptance test exercises
  the real protocol so a breaking bump fails loudly.
- Parity drift between MCP and native verdict → the parity test is the guard;
  both paths share `AssembleVerdict`, so divergence can only come from `Decide`.
