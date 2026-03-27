# Edge Cases: generate-html-project-docs

## Covered

- `docs validate` fails when `PROJECT.md`, `ROADMAP.md`, or `.workflow/roadmap.json` is missing.
- Generator works when roadmap analysis artifacts are absent by falling back to roadmap features.
- Generator includes Mermaid blocks and traceability tables even with minimal evidence data.
- Command tests verify HTML output file is created at configurable `--out` path.

## Residual Risks

- Markdown content is embedded as escaped text blocks; advanced markdown rendering is out of scope.
- Dependency semantics rely on roadmap-analysis quality when that file is present.
