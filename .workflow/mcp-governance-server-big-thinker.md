# mcp-governance-server — big-thinker

## Problem

Every new host harness needs a bespoke hook/parity adapter, so Centinela's
governance is consumable only by harnesses it has explicitly integrated. The
rules, gate engine, and claim-verification are reachable only via the CLI +
harness-specific hooks.

## Scope

Expose governance as a versioned MCP server (`centinela.mcp/v1`) over stdio
(official Go SDK) with 4 tools — read_rules, run_gates, verify_claims,
workflow_state — reusing `internal/verdict.AssembleVerdict` as the wire payload.
Advisory-by-protocol: it returns `allow|warn|block`. A `centinela mcp shim`
client maps `block`→exit 2 so enforcement stays harness-side via the existing
pre-write deny contract.

## Dependencies & Assumptions

- Reuses `verdict.AssembleVerdict` + `verdict.Deps{Gates: gates.RunAll, Verify,
  Evidence: EvidenceIndex}` (the same wiring `centinela verdict` uses), so the
  MCP and native paths share one assembler — parity is structural.
- Dep `headless-governance` (verdict packet) is done.
- New dep `github.com/modelcontextprotocol/go-sdk v1.2.0` (stable); confined to
  `internal/mcp` + the two cmd files.

## Risks

- Young SDK → pin v1.2.0, isolate behind `internal/mcp`; the acceptance test runs
  the real protocol so a breaking bump fails loudly.
- Parity drift → guarded by a parity test; divergence can only come from `Decide`
  (the allow/warn/block mapping), since both paths share `AssembleVerdict`.

## Rollout

Additive: new `mcp` command tree + one dep. Opt-in — a harness must invoke
`mcp serve`; the shim is documented, not force-wired. No migration.

## Handoff

→ feature-specialist: encode the zero-integration-harness, shim block/allow,
parity, versioned-schema, and advisory-only behaviors as acceptance scenarios.
