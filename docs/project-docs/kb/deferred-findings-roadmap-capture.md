---
feature: deferred-findings-roadmap-capture
summary: A command that lets any workflow agent capture an out-of-scope discovery or a not-fixed-now finding directly into a Backlog section of the roadmap, so it is never lost in prose reports and is always visible for triage.
audience: end-user
status: done
---

## What it does

When a workflow agent spots something worth addressing later — an out-of-scope gap, a known risk it won't fix in the current feature, a hardcoded value it noticed but deliberately deferred — it can now record that finding in the roadmap with a single command. The finding lands in a dedicated Backlog section of `roadmap.json`, carries a one-line summary and a source label, and is visible every time you look at `centinela roadmap`. Backlog entries never appear as startable features, are never required to have quality scores, and cannot be started until explicitly promoted. At triage time, a quality-evaluator scoring pass moves a finding out of the Backlog and into a real roadmap phase, where it becomes a first-class feature.

## When you'd use it

Use `centinela roadmap defer` any time a workflow agent — big-thinker, feature-specialist, senior-engineer, qa-senior, or any of the other roles — identifies a finding that is deliberately out of scope for the current feature but should not be forgotten. All eight role prompts now carry a mandatory "Deferred Findings" section: agents are expected to run the command for every such finding and list the recorded slugs in their report. This replaces the old pattern of burying out-of-scope items in prose sections (`Out-of-Scope`, `Residual Risks`, `Outstanding TODOs`) that nobody reads again.

Use `centinela roadmap promote` when you are ready to pull a Backlog finding into active planning. The two-step promote flow (first print evaluator context, then pass scores) ensures the finding earns its place in the roadmap with an honest quality evaluation rather than a rubber-stamp.

## How it behaves

- **Capturing a finding** adds it to the Backlog section of `roadmap.json` with its summary, an optional source (`<feature>/<role>`), and a timestamp. All previously existing roadmap entries are left byte-for-byte unchanged.
- **Source auto-detection**: when run from inside a worktree directory, the command fills in the source feature automatically; `--source` is an explicit override and is optional everywhere.
- **Running from outside a worktree** with no `--source` flag creates the entry without a source field — the finding is still valid and visible.
- **Validation rejections**: an empty summary, a duplicate slug (whether in Backlog or in any other phase), or an invalid slug format all cause the command to exit with an error before touching `roadmap.json`.
- **Backlog entries are always visible**: `centinela roadmap` shows a dedicated Backlog section listing each finding's slug and summary. The section disappears automatically when the Backlog is empty.
- **Backlog entries never appear as ready**: `centinela roadmap ready` and readiness checks always skip Backlog entries, no matter how few dependencies they have.
- **Starting a Backlog feature is refused**: `centinela start <backlog-slug>` exits with an error telling you to promote the finding first.
- **`centinela roadmap validate` stays green**: Backlog entries are exempt from analysis and quality coverage requirements — `roadmap validate` never demands scores or analysis entries for them.
- **That exemption is precise**: a phase named "Pre-Backlog Work" or anything other than the exact "Backlog" name is not exempt and still requires full coverage.
- **Promote without scores prints evaluator context and writes nothing**: the output shows the finding's name, summary, source, target phase, the six scoring dimensions with the ≥9 overall threshold, and the exact re-invocation line — so you can run an honest quality-evaluator agent and then re-invoke with `--scores`.
- **Promote with valid scores** moves the entry out of Backlog, strips provenance metadata from the roadmap feature, appends analysis and quality entries, records a provenance bullet in the companion `.md` files, and runs `roadmap validate` as its final step.
- **Score validation happens before any write**: an overall score below 9, any dimension outside 1–10, a malformed CSV, or an unknown target phase all cause promote to exit cleanly with no files changed.
- **All eight role prompts** (big-thinker, feature-specialist, senior-engineer, qa-senior, ux-ui-specialist, validation-specialist, gatekeeper, edge-case-tester) and their scaffold mirrors contain a byte-identical "Deferred Findings" section requiring agents to run the command and record the slugs.

## Examples

**Capturing a finding mid-feature:**

```bash
centinela roadmap defer hook-timeout-config \
  --summary "Prewrite hook timeout is hardcoded; should be configurable" \
  --source deferred-findings-roadmap-capture/senior-engineer
```

The command exits with code 0. The Backlog section now contains one entry. All other phases are untouched.

**Seeing Backlog findings in the roadmap overview:**

```bash
centinela roadmap
```

The output includes a Backlog section showing each finding's slug and summary alongside the normal phase overview.

**Promoting a finding — first, get the evaluator context:**

```bash
centinela roadmap promote hook-timeout-config --phase "Phase 6: Capability-Adaptive Governance"
```

The command prints the finding details, the six scoring dimensions, the ≥9 overall threshold, and the exact re-invocation line. Nothing is written to disk.

**Promoting a finding — second, pass scores after an honest evaluation:**

```bash
centinela roadmap promote hook-timeout-config \
  --phase "Phase 6: Capability-Adaptive Governance" \
  --scores 9,9,8,7,9,9
```

`hook-timeout-config` moves into Phase 6, gains analysis and quality entries, and `centinela roadmap validate` passes.
