# Consuming Governance via MCP

> Use Centinela's governance from any MCP-speaking harness — with zero Centinela-specific code.

The native hooks are harness-specific (Claude, OpenCode). For **any** other MCP-speaking harness, Centinela exposes the same governance as a versioned **Model Context Protocol** server — so a host with *zero* Centinela-specific code can obtain a verdict purely through tool calls.

```bash
centinela mcp serve   # runs the MCP server on stdio (schema: centinela.mcp/v1)
```

Register it with any MCP client. For a project-scoped `.mcp.json` (Claude Code, and most MCP hosts):

```json
{
  "mcpServers": {
    "centinela": { "command": "centinela", "args": ["mcp", "serve"] }
  }
}
```

## Tools (`centinela.mcp/v1`)

| Tool | Arguments | Returns |
|------|-----------|---------|
| `read_rules` | _(none)_ | profile, archetype, file-size limit, enabled gates, locales |
| `run_gates` | `feature?` | gate results + a `decision` (`allow`/`warn`/`block`) |
| `verify_claims` | `feature?` | claim-verification checks + a `decision` |
| `workflow_state` | `feature?` | active feature run provenance + on-disk evidence index |

`feature` is optional — omit it to use the active feature. Every result carries `"schema": "centinela.mcp/v1"` so harnesses can pin a compatibility level. A `run_gates` result looks like:

```json
{ "schema": "centinela.mcp/v1", "decision": "block",
  "gates": [ { "name": "G1: File Size", "status": "fail", "message": "internal/big/big.go: 134 lines (>100)" } ] }
```

## Advisory by protocol + the enforcement shim

The server is **advisory**: it returns `allow | warn | block` and *cannot itself stop a write*. Enforcement stays harness-side via a thin shim that maps a `block` verdict onto the harness's existing pre-write deny — the same `exit 2` contract the native hook uses:

```bash
centinela mcp shim            # active feature; exit 2 on block, 0 on allow/warn
centinela mcp shim my-feature # explicit feature
```

Wire it as a deny hook in the harness. The shim runs the full gate + claim suite, so prefer a once-per-turn `Stop` hook over a per-write `PreToolUse` matcher unless your gates are fast:

```json
{ "hooks": { "Stop": [
  { "hooks": [ { "type": "command", "command": "centinela mcp shim" } ] } ] } }
```

A `block` exits non-zero and the harness acts on the deny; `allow`/`warn` exit 0 and it proceeds. The MCP verdict is identical to the native-hook verdict for the same diff and workflow state (a parity test enforces this), so you can adopt MCP without changing how governance decides.

---

← Back to the [documentation index](README.md) · [Quality gates](gates.md)
