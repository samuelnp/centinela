# Edge Cases: docs-knowledge-base-pages

## Covered

- KB markdown missing required sections fails generation with feature + section names.
- KB markdown with unterminated frontmatter fails with an explicit error.
- Features without a KB markdown render as placeholder cards in the index, while populated features link to their per-feature HTML page.
- Validator distinguishes between "kb markdown missing" and "kb html missing" with paths in the error.
- Frontmatter parser tolerates optional fields (audience, status) when absent.
- Examples section is omitted from the rendered page when empty and rendered when present.
- Generator continues to write the main index.html even when no KB markdowns exist (KB index still emitted with all-placeholder cards).

## Residual Risks

- Markdown body uses a minimal renderer (paragraphs and dash-bullets only). Rich content (tables, code fences) is not transformed.
- Status badge on cards relies on `.workflow/<feature>.json` being current.
