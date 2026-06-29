# Feature: mcp-governance-server

## Problem

Every new host harness needs a bespoke hook/parity adapter, so Centinela's
governance can only be consumed by harnesses it has explicitly integrated. The
rules, gate engine, and claim-verification are reachable only through the CLI +
harness-specific hooks.

## Goal

Expose Centinela's governance as a versioned **MCP server** (`centinela.mcp/v1`)
so any MCP-speaking harness obtains a verdict with **zero Centinela-specific
code** — purely through MCP tool calls. The server is **advisory-by-protocol**:
it returns a structured verdict (`allow` | `warn` | `block`) and cannot itself
stop a write. Enforcement stays harness-side via a thin **shim** that maps a
`block` verdict onto the harness's existing pre-write deny (exit 2 — the same
contract the native hook uses).

## Decisions (locked)

- **Official Go MCP SDK** `github.com/modelcontextprotocol/go-sdk` (v1.2.0,
  stdio transport) — protocol-correct, future-proofed for spec changes.
- **Ship the shim + parity tests** in v1 (full roadmap acceptance).

## Wire payload

Reuse the existing headless-governance verdict packet (`internal/verdict`,
`centinela.verdict/v1`, `AssembleVerdict`) as the structured payload. A new
`Decision(packet) → allow|warn|block` maps `pass`→`allow`, any gate `fail` /
verify `FAIL`→`block`, warnings-only→`warn`.

## Scope (v1)

1. **`internal/mcp`** (aggregator): build an `mcp.Server`, register 4 versioned
   tools, each a thin handler over an existing engine:
   - `read_rules` — the governing rule surface (config + archetype rules).
   - `run_gates` — runs the gate engine → gate lines + decision.
   - `verify_claims` — runs claim verification for a feature/step → check lines.
   - `workflow_state` — active feature, step, and on-disk evidence.
   Plus `Decision()` and the versioned tool/verdict schema.
2. **`centinela mcp serve`** — runs the stdio MCP server.
3. **`centinela mcp shim`** — connects as an MCP client, obtains the verdict,
   and exits 2 on `block` / 0 on `allow|warn` (mirrors the native hook deny).
4. Wire `centinela mcp shim` into harness setup as an optional Stop/PreToolUse
   adapter (documented; not forced on).

## Out of scope (v1)

- Tool surface beyond the 4 named tools; resources/prompts MCP features.
- Non-stdio transports (HTTP/SSE) — stdio only.
- Auto-enforcement inside the server (it stays advisory by design).

## Acceptance (see spec)

- A harness with zero Centinela code obtains a verdict via MCP tool calls (the
  SDK client drives the `serve` binary in-test).
- The shim aborts the write on `block` (exit 2) and proceeds on `allow` (exit 0).
- **Parity**: the MCP verdict and the native-hook verdict are identical for the
  same diff + workflow state.

## Risks

- New (young) dependency: pin v1.2.0; isolate all SDK types behind `internal/mcp`
  so a future SDK change touches one package.
- Schema drift: the tool surface + verdict schema are explicitly versioned so
  harnesses pin a compatibility level.
