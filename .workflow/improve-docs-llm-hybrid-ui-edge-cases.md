# Edge Cases: improve-docs-llm-hybrid-ui

## Covered

- Feature dependency Mermaid graph renders with nodes that have no dependencies.
- Spec coverage Mermaid graph still renders when no specs are present.
- Generated HTML remains navigable on mobile via responsive sidebar collapse.
- Workflow internals remain documented via tables, not Mermaid graphs.
- Prompt guidance enforces LLM-first flow while preserving CLI fallback command.

## Residual Risks

- Large PROJECT.md and ROADMAP.md files are intentionally truncated in source context excerpts.
- Mermaid rendering depends on external CDN availability in browser runtime.
