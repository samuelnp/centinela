# Plan: Generate HTML Project Docs

1. Create a `docs` command group with `generate` and `validate` subcommands.
2. Implement an internal `docgen` package to load project/workflow artifacts.
3. Build HTML output with sections, traceability tables, and Mermaid diagrams.
4. Add validation for required roadmap and workflow inputs.
5. Add unit, integration-like command, and acceptance tests.
