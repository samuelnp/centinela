---
feature: docs-knowledge-base-pages
summary: Every feature now ships a plain-language guide for end users, generated from a short markdown the docs step requires.
audience: end-user
status: done
---

## What it does
The final step of every Centinela workflow now produces a per-feature guide written for non-technical readers. Instead of stopping at lists of specs and evidence, the docs step builds a small knowledge base: one page per feature plus an index that links them all together. The page is rendered from a short markdown file that the documentation specialist writes against a fixed template.

## When you'd use it
Open the knowledge base whenever you want to know what Centinela can do without reading code or Gherkin scenarios. The KB index is the answer to "what features does this project have, and what does each one do for me?" — useful for new collaborators, stakeholders, or anyone evaluating the tool.

## How it behaves
- The KB lives under `docs/project-docs/kb/`. Each feature has a markdown source (`<feature>.md`) and a generated HTML page (`<feature>.html`); the directory's `index.html` lists them all with a status badge.
- The main project docs (`docs/project-docs/index.html`) link to the KB from the top navigation as "Knowledge Base".
- When you finish the docs step for a feature, Centinela requires both the KB markdown and the rendered HTML to exist; otherwise it fails with a message naming the missing path.
- Features that don't have a KB markdown yet show up as "Guide not yet written" cards on the index — the index never fails on backfill gaps.
- A KB markdown must contain the sections "What it does", "When you'd use it", and "How it behaves"; an "Examples" section is optional. Missing required sections fail generation with the feature and section named in the error.

## Examples
After finishing a feature, run:

    centinela docs generate --out docs/project-docs/index.html

The same command writes the main report, the KB index, and one HTML page for each KB markdown present.
