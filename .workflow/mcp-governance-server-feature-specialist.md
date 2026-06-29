# mcp-governance-server — feature-specialist

## Behavior Summary

`centinela mcp serve` runs a stdio MCP server registering 4 versioned tools
(read_rules, run_gates, verify_claims, workflow_state) that reuse the verdict
assembler. Any MCP client obtains a structured `allow|warn|block` verdict with
no Centinela-specific code. `centinela mcp shim <feature>` is a client that
combines the tool verdicts and exits 2 on `block` (deny) / 0 on `allow`. The
server is advisory: it only reads, never writes or blocks.

## Acceptance Criteria (Gherkin)

See `specs/mcp-governance-server.feature` — zero-integration harness gets a
verdict via tool calls; versioned `centinela.mcp/v1`; shim block→exit 2 /
allow→exit 0; MCP verdict == native verdict (parity); advisory-only (no mutation).

## UX States

- **Server**: starts on stdio, lists 4 tools, answers tool calls with structured
  verdicts; no stdout chatter outside the protocol.
- **Shim block**: exits 2, mirroring `RenderBlocked` semantics of the native hook.
- **Shim allow/warn**: exits 0.
- **No active feature / unknown feature**: shim returns a clear error, exit ≠ 2
  (not a deny — a usage error).

## Edge Cases

- Zero-Centinela-code harness obtains a verdict purely via MCP tool calls.
- Shim: block → exit 2 (deny); allow → exit 0.
- Parity: MCP verdict == native-hook verdict for the same diff + workflow state.
- Verdict payload carries schema id `centinela.mcp/v1` (versioned, pinnable).
- Advisory: tools only read gates/claims/workflow; never perform/abort a write.

## Out-of-Scope

- Tools beyond the 4 named; MCP resources/prompts; HTTP/SSE transports.
- Server-side enforcement (stays advisory by design).

## Handoff

→ senior-engineer: add the SDK dep (pinned), build `internal/mcp` (schema,
decision, deps, tool handlers, server) + `mcp serve`/`mcp shim` cmds; reuse
`verdict.AssembleVerdict`. Keep each file ≤100 lines; add `internal/mcp` to the
import_graph aggregator layer.
