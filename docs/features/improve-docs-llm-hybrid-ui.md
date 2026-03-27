# Feature Brief: Improve Docs LLM Hybrid UI

## Problem

The generated HTML report is technically correct but visually weak, hard to scan,
and too centered on Centinela internals instead of project-facing documentation.

## Goal

Produce polished documentation with navigation, examples, and meaningful graphics,
while using an LLM-first authoring flow with a deterministic CLI fallback.

## Scope

- Upgrade `centinela docs generate` HTML output to a structured, responsive docs UI.
- Add clear navigation, section anchors, summary cards, and example blocks.
- Keep Mermaid graphs focused on project features and specs only.
- Remove workflow-specific Mermaid visuals (state/evidence handoff internals).
- Update docs prompt to instruct LLM-led narrative synthesis first, command fallback
  second.

## Non-Goals

- Changing Centinela workflow semantics.
- Introducing external build tooling for docs rendering.
