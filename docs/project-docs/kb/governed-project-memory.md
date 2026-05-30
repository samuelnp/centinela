---
feature: governed-project-memory
summary: A git-tracked ledger automatically captures and recalls your project's hard-won lessons, gate verdicts, and architectural decisions across features.
audience: end-user
status: done
---

## What it does

When you complete a Centinela step, the framework automatically harvests three types of knowledge: edge-case lessons from the tests step, gatekeeper verdicts from the validate step, and architectural decisions from the plan step. These facts are written as reviewable, git-tracked files in `.workflow/memory/entries/` with frontmatter linking each entry back to its source artifact. When you start planning a new feature, the most relevant entries are automatically injected into your plan context, ranked by deterministic signals (dependencies, shared tags, recency) — no semantic vectors, no fuzzy matching.

## When you'd use it

Every time you start a new feature. Without this, lessons learned the hard way during feature X disappear when you move to feature Y. You might re-discover that your coverage tests need mutation testing (a lesson from X), or re-run into the same gate verdict on unsafe patterns. With governed memory, you're automatically reminded. You can also browse the ledger as a git history — each memory entry is a single, auditable file, so you can see what you learned and when.

## How it behaves

- **Automatic capture at step boundaries**: When you run `centinela complete <feature>`, the tool captures edge-case lessons (from `.workflow/<feature>-edge-cases.md`), gatekeeper verdicts (from `.workflow/<feature>-gatekeeper.md`), and decisions (from a `## Decisions` section in the feature brief or plan) and writes each as a timestamped, git-tracked file.
- **Idempotent writes**: Completing a step twice does not duplicate entries — the ledger uses a stable content hash to deduplicate. Re-completing is always safe.
- **Graceful on missing or malformed artifacts**: If a source artifact is missing or malformed, capture logs a warning and moves on — it never blocks `centinela complete`.
- **Plan-time recall**: When you start the plan step for a new feature, relevant memory entries are injected into the plan advisor context. Relevance is computed by explicit signals: entries from the feature's declared dependencies rank highest, then entries sharing the feature's tags, then recent entries. No embeddings.
- **Configurable caps**: You can set a maximum count and byte budget for recalled entries via the `[memory]` section of your Centinela config; they default to safe values.
- **Config gating**: The entire system is controlled by the `[memory] enabled` flag in your config (default on). Disable it to opt out of both capture and recall.

## Examples

After you complete a feature with a lesson file `.workflow/my-auth-feature-edge-cases.md` containing a finding about JWT rotation pitfalls, it's automatically harvested into `.workflow/memory/entries/<content-hash>.md` with frontmatter like:

```
feature: my-auth-feature
type: lesson
title: JWT rotation must happen before expiry
tags: [auth, security, timing]
sourceArtifact: .workflow/my-auth-feature-edge-cases.md
```

When you start planning the next feature, if that feature depends on the auth feature or shares the `auth` tag, this entry will appear in your plan advisor context — a one-line reminder of a lesson you paid for in blood.
