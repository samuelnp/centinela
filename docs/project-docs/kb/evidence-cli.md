---
feature: evidence-cli
summary: A typed CLI for authoring and validating workflow evidence artifacts, eliminating hand-written JSON and automating evidence formatting.
audience: end-user
status: done
---

## What it does
The evidence CLI provides a set of commands—`centinela evidence` and `centinela artifact new`—that let you author evidence files (the JSON + markdown proof that each workflow step was completed) without ever writing raw JSON by hand. Instead of using Python one-liners, heredocs, or jq filters to escape strings and merge fields, you use straightforward CLI commands that guarantee your evidence conforms to the schema and stays properly formatted.

## When you'd use it
Every time a Centinela workflow step runs—whether you're planning a feature, writing code, running tests, validating gates, or writing documentation—you'll use these commands to record what was done. Before this feature, agents would hand-write JSON and risk format errors, duplicate fields, or schema violations. Now the CLI handles the mechanics, so you can focus on the substance of your evidence.

## How it behaves
- `centinela evidence init <feature> <role>` drops a schema-valid skeleton with all required fields for that role (e.g., `big-thinker`, `qa-senior`, `documentation-specialist`).
- `centinela evidence set <feature> <role> <field> <value>` updates a single scalar field (like `status` or `generatedAt`) in one atomic write.
- `centinela evidence append <feature> <role> <field> <value>` extends a list field (like `inputs`, `outputs`, or `edgeCases`) without creating duplicates.
- `centinela evidence read <feature> <role> --field <name>` retrieves a single field so you can inspect what a previous step produced.
- `centinela evidence validate <feature>` scans all evidence for the feature and reports each missing or malformed field with the exact command to fix it.
- `centinela evidence repair <feature>` cleans up orphaned temporary files if a write was interrupted.
- `centinela artifact new <feature> <kind>` creates pre-filled markdown templates for edge-cases, gatekeeper, production-readiness, or documentation-specialist artifacts.
- All writes use atomic temp-file-plus-rename so a crash mid-operation never leaves corrupted JSON on disk.
- Concurrent writes to the same evidence file serialize via advisory lock, so two agents can safely run `centinela evidence append` at the same time.
- If you accidentally hand-write JSON using the Write or Edit tool, a PostToolUse hook automatically reformats it to pretty-print with stable key order.
- Free-form attachments go in an `extra` object, so you can store notes without losing schema strictness.
- Unknown fields from older or newer binaries are preserved on round-trip, preventing version-skew data loss.
- The hook respects worktree boundaries: if you're working on feature "alpha" in a worktree, the formatter only touches alpha's `.workflow/` files and leaves other features' evidence untouched.

## Examples
To author big-thinker evidence for a new feature called `my-feature`:

```bash
centinela evidence init my-feature big-thinker
centinela evidence set my-feature big-thinker status done
centinela evidence append my-feature big-thinker inputs docs/features/my-feature.md
centinela evidence append my-feature big-thinker outputs docs/features/my-feature.md
centinela evidence append my-feature big-thinker outputs docs/plans/my-feature.md
centinela evidence set my-feature big-thinker generatedAt 2026-05-28T14:30:00Z
```

To drop a markdown template for edge-cases:

```bash
centinela artifact new my-feature edge-cases
```

To verify all evidence is valid before advancing the workflow:

```bash
centinela evidence validate my-feature
```

If the validator finds a missing field, it tells you exactly which command to run:

```
centinela evidence append my-feature big-thinker edgeCases "concurrent writes may deadlock"
```
