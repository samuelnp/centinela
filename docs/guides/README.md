# Centinela Documentation

Guides for setting up and operating Centinela. New here? Start with the [README](../../README.md) for the 30-second tour, then come back.

## Guides

| Guide | What it covers |
|-------|----------------|
| [Getting Started](getting-started.md) | Full setup: `init`, `PROJECT.md`, roadmap bootstrap, your first feature, `migrate` |
| [Configuration Guide](configuration.md) | Copy-paste `centinela.toml` recipes by use case (solo, team+CI, local models, regulated, fleet) |
| [Configuration Reference](configuration-reference.md) | Every `centinela.toml` key — type, default, allowed values |
| [Workflow & Hooks](workflow-and-hooks.md) | The enforced five-step workflow and the agent hooks behind it |
| [Quality Gates](gates.md) | Built-in and opt-in gates, diff-aware mode, claim verification |
| [MCP Governance](mcp.md) | Consume governance from any MCP-speaking harness |
| [Concepts](concepts.md) | Harness engineering — why Centinela exists, and when *not* to use it |

## Reference (bundled with `centinela init`)

The `docs/architecture/` directory holds the framework's own reference material:

- [Architecture Archetypes](../architecture/architecture-overview.md) — Hexagonal, Rails, N-Tier, ECS, Modular
- [Gatekeepers](../architecture/gatekeepers.md) — full gate reference (G1–G11)
- [Testing Strategy](../architecture/testing-strategy.md), [i18n Strategy](../architecture/i18n-strategy.md), [Workflow Enforcement](../architecture/workflow-enforcement.md), and more

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md).
