# Feature Brief: Generate HTML Project Docs

## Problem

Centinela data exists across roadmap, workflow, evidence, and specs, but there is
no single human-readable document that explains project status and traceability.

## Goal

Add a command that generates an HTML documentation report from Centinela
artifacts with understandable narrative sections and Mermaid diagrams.

## Scope

- Add `centinela docs generate` to render HTML documentation.
- Add `centinela docs validate` to verify required documentation inputs.
- Include roadmap, features, specs, workflow states, and orchestration evidence.
- Include Mermaid graphs for roadmap dependencies and evidence handoffs.
