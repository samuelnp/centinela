# Plan: Improve Docs LLM Hybrid UI

1. Update documentation generation prompts to define hybrid behavior:
   LLM-first synthesis, `centinela docs generate` fallback.
2. Refactor `internal/docgen` renderer into small files under 100 lines each with:
   layout, theme, navigation, sections, examples, and feature-focused graphics.
3. Remove workflow-specific Mermaid rendering and add feature/spec graphs only.
4. Add richer HTML sections: overview cards, artifact summaries, examples, and
   linked navigation anchors.
5. Update unit/command tests to assert UI structure and graph policy.
6. Update acceptance tests/spec expectations for polished docs behavior.
