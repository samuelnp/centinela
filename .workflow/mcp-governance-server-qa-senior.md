# mcp-governance-server — qa-senior

## Test Inventory

**Colocated:**
- `internal/mcp/`: `decision_test.go` (DecideGates/Verify/Combine/Decide table,
  nil packet, nz), `tools_test.go` (4 handlers via stub Deps: schema stamp,
  nil-slice coalesce, error propagation, NewServer) → **100%**.
- `cmd/centinela/`: `mcp_test.go` (rules surface, enabledGateNames,
  workflowForFeature, mcpVerdict assemble + no-active-feature), `mcp_shim_test.go`
  (runMcpShim block→exit2 / allow→no-exit / tool-error propagates, via in-memory
  transport + injected connection seam).

**Tier:** `tests/unit` (decision mapping), `tests/integration` (real in-memory
client/server → run_gates → block), `tests/acceptance` (binary-driven SDK client
= zero-integration harness: tools/list + versioned read_rules; shim block→2 /
allow→0; MCP==native parity). All test files ≤100 lines.

## Coverage Gaps

Total **95.0% ≥ 95.0%** gate. `internal/mcp` 100%. The only uncovered cmd lines
are `runMcpServe` (blocks on stdio) and `mcpConnectSelf` (spawns the binary) —
genuine subprocess seams, exercised functionally by the acceptance tests.

## Acceptance Wiring

`specs/mcp-governance-server.feature` scenarios map to
`TestAccMcpZeroIntegrationHarness` (tools/list + versioned verdict),
`TestAccMcpShimBlockAndAllow` (deny/allow), and `TestAccMcpParityWithNative`
(parity). `centinela.toml` already runs `go test ./tests/acceptance/...`.

## Handoff

→ validation-specialist: full suite + gates green; produce the gatekeeper report.
Note the new pinned dependency `modelcontextprotocol/go-sdk v1.2.0` (+ go.sum).
