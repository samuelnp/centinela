# Edge Cases: mcp-governance-server

## Covered

| # | Edge case | Handling | Test |
|---|-----------|----------|------|
| 1 | Decision mapping: clean / gate-warn / gate-fail / verify-fail / verify-warn / both-fail | worst-wins allow/warn/block | `mcp:TestDecideScopes`, `unit:TestMcpDecisionMapping` |
| 2 | Nil packet | treated as allow | `mcp:TestDecideNilPacket` |
| 3 | Combine with no decisions | allow | `mcp:TestCombineWorstWins` |
| 4 | Nil output slices (SDK rejects `null` for array schema) | `nz()` coalesces to `[]` | `mcp:TestNzCoalescesNil`, `mcp:TestHandlersStampSchemaAndCoalesce` |
| 5 | Verdict assembler errors inside a handler | propagated as a tool error | `mcp:TestHandlersPropagateVerdictError` |
| 6 | Zero-integration harness (no Centinela code) lists tools + gets a versioned verdict | SDK client over `mcp serve` | `acceptance:TestAccMcpZeroIntegrationHarness` |
| 7 | Shim: block verdict | exit 2 (harness pre-write deny) | `acceptance:TestAccMcpShimBlockAndAllow`, `cmd:TestRunMcpShimBlockExitsTwo` |
| 8 | Shim: allow verdict | exit 0 | `acceptance:TestAccMcpShimBlockAndAllow`, `cmd:TestRunMcpShimAllowDoesNotExit` |
| 9 | Shim: tool/connection error | propagated, no deny | `cmd:TestRunMcpShimToolErrorPropagates` |
| 10 | Parity: MCP verdict == native `centinela verdict` for same repo | shared `AssembleVerdict`; `Decide` == `Combine(scope decisions)` | `acceptance:TestAccMcpParityWithNative`, `integration:TestMcpServerRunGatesInMemory` |
| 11 | `read_rules` takes no args (empty-object schema) | called with `{}`; extra props rejected by the SDK | `acceptance:TestAccMcpZeroIntegrationHarness` |
| 12 | No active feature / unknown feature | resolver returns nil → verdict error (usage error, not a deny) | `cmd:TestMcpVerdictNoActiveFeature`, `cmd:TestWorkflowForFeature` |
| 13 | Advisory only | every handler reads (gates/verify/workflow); none mutate or block | by construction; server has no write path |

## Residual Risks

- `runMcpServe` (blocks on stdio) and `mcpConnectSelf` (spawns the binary) are
  thin subprocess seams left uncovered by the in-process profile; they are
  exercised functionally by the binary-driven acceptance tests.
- Young SDK (v1.2.0): pinned and confined to `internal/mcp` + two cmd files; the
  acceptance test runs the real protocol so a breaking bump fails loudly.
- Cross-feature attribution is the active feature at call time (same as the
  native verdict path) — not a regression introduced here.
