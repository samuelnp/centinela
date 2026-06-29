# mcp-governance-server — senior-engineer

## Files Touched

- **`internal/mcp`** (NEW, aggregator): `schema.go` (SchemaVersion + typed tool
  in/out + `nz` nil-slice coalescer), `decision.go` (DecideGates/DecideVerify/
  Combine/Decide → allow|warn|block), `deps.go` (injected engines), `tools.go`
  (4 handlers), `server.go` (`NewServer` registers the 4 tools).
- **cmd**: `mcp.go` (`mcp serve` + `mcpVerdict` reusing `verdict.AssembleVerdict`
  with the same Deps as `centinela verdict`), `mcp_rules.go` (rule surface),
  `mcp_shim.go` (`mcp shim` — MCP client over `serve`, combines tool decisions,
  exit 2 on block).
- **deps**: `github.com/modelcontextprotocol/go-sdk v1.2.0` (direct, pinned).
- **config**: `internal/mcp/**` added to the import_graph aggregator layer.

## Architecture Compliance

`internal/mcp` imports the `verdict` aggregator + gates/verify/workflow/config
engines + the external SDK; imported only by `cmd/`. All SDK types are confined
to `internal/mcp` + the two cmd files, so a future SDK bump is one-package. The
server is advisory: every handler only reads; none mutate or block.

## Type-Safety Notes

Strict typing; tools use the SDK's typed In/Out generics. One protocol gotcha
found and fixed in smoke: the SDK validates tool output against the inferred
schema and rejects nil slices (`null` vs `array`) — `nz()` coalesces every slice
field to `[]`. Verified end-to-end via a throwaway SDK client (since removed):
`tools/list` → 4 tools; `read_rules` → `{schema:"centinela.mcp/v1",...}`; shim
`allow`→exit 0, `block`→exit 2 (oversized file under `internal/` trips G1).

## Trade-Offs

- Each tool returns its scope decision; the shim `Combine`s run_gates + verify
  decisions. Equals `Decide` on the same packet → parity is structural (both
  paths share `AssembleVerdict`; divergence can only come from the mapping).
- Tier/rules surface kept minimal (profile, archetype, file limit, enabled
  gates, locales) — enough for a harness to pin compatibility.

## Handoff

→ qa-senior: unit (Decide/Combine table, nz), integration (NewServer with stub
Deps → tool outputs), acceptance (SDK client = zero-integration harness;
shim block→2/allow→0; MCP==native parity). Use an oversized `internal/<d>/*.go`
to force a block deterministically.
