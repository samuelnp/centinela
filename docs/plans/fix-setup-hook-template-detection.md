# Plan: fix-setup-hook-template-detection

## Scope

Fix setup hook project detection so roadmap prompting does not depend on keeping
`PROJECT.md.template`, and add a plain directive line for stronger LLM behavior.

## Tasks

1. Update setup hook detection logic for template/project file combinations.
2. Emit concise directive lines before boxed setup/roadmap/prod-readiness output.
3. Add regression tests for renamed-template scenario and directive output.
4. Run full test suite and `centinela validate`.

## Non-Goals

- Reworking all UI panels outside setup-related hook output.
