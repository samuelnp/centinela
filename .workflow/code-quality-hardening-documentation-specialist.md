# code-quality-hardening — documentation-specialist

## KB entry written

`docs/project-docs/kb/code-quality-hardening.md` (audience: end-user, status: done).
Summary line:

> Hardens Centinela's everyday surfaces — stable evidence files, a Go formatting
> gate in validate, and loud, specific errors when your config or workflow state
> is broken.

The guide is written for a Centinela user/operator running the CLI, in plain
language about what they observe — not Go internals. It covers all nine spec
scenarios as user-visible behavior: stable evidence key order (incl. `coverage`
between `mobileFirst` and `handoffTo`), unformatted Go failing `validate`,
formatted trees passing quietly, the format check being part of the validate
command list, `centinela start` failing loudly on a corrupt `centinela.toml`
with no half-created feature, the prompt hook warning without breaking the
session, missing-vs-corrupt-vs-unreadable workflow state errors each reporting
distinctly. It includes example snippets of the format gate and the clearer
error messages. No Given/When/Then prose and no package names in the body.

## Generated outputs (confirmed on disk)

| Output | Confirmed |
|--------|-----------|
| docs/project-docs/kb/code-quality-hardening.md | yes (3.2K) |
| docs/project-docs/kb/code-quality-hardening.html | yes (7.0K) |
| docs/project-docs/kb/index.html | yes (18.3K) — KB index regenerated |
| docs/project-docs/index.html | yes (123.8K) — portal regenerated |

`<binary> docs validate` exits 0; `<binary> docs generate` rendered the KB page,
the KB index, and the top-level portal.

## Mermaid

None added. Per policy, Mermaid is reserved for project feature/spec
relationships, not workflow internals. This is an internal-surface code-quality
fix with no new feature-relationship diagram to draw.

## Right-sizing note

This is an internal-surface feature (developer-facing quality fix: evidence
key-order drift, a gofmt gate, config-error policy, workflow-load transparency).
The full HTML-portal regeneration is heavier than the reader value here — the
durable artifact a user actually consults is the one KB page describing the
clearer errors and the new format gate. Flagging this as a nod to the roadmap's
"right-size the docs step" intent: for internal fixes the markdown KB entry is
the high-value output; the portal rebuild is incidental.
