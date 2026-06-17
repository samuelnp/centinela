# Feature: governance-telemetry

- surface: internal
- status: planned
- roadmap: Phase 7 — Evidence-Driven Governance (foundational; unblocks 5 features)
- fixes: governance friction is invisible. Every block, gate failure, verify
  rejection, and rework cycle happens, is rendered to a transcript, and then
  vanishes. Nobody — human or downstream feature — can see *which* rails bite,
  *how often*, or *for whom*.

## Problem

Centinela enforces discipline by emitting governance signals — it blocks an
out-of-step write, fails a gate, rejects a claim, refuses an advance. Each signal
is a one-shot stderr render aimed at the human in the loop. The moment it scrolls
past, it is gone. There is no durable, queryable record of governance events.

This is the same gap `governed-project-memory` closed for *knowledge* (lessons,
verdicts, decisions), but for *operational friction*. Memory remembers what we
learned; telemetry must remember what the framework *did to us and why*.

Five downstream roadmap features cannot exist without this log:

- **centinela-insights** — "which blocks/gates trigger most" → needs a block &
  gate-failure event stream.
- **failure-ledger-plan-advisor** — feed recurring gate failures into the plan
  advisor → needs gate-failure events keyed by (feature, gate).
- **capability-calibration** — was the model's enforcement profile right? →
  needs per-feature block / gate / verify rates.
- **team-dashboard** — cross-worktree, cross-contributor friction view → needs a
  per-worktree append log it can aggregate.
- **adaptive-skill-synthesis** — detect a step that keeps failing → needs the
  rework signal (repeated complete-rejections before a successful advance).

## Who's hurting

The developer (and the AI agent) who feels friction but can't measure it —
"why does this keep getting blocked?" has no answer today. And the five features
above, each blocked at the roadmap on a data source that does not exist. Because
five independent readers will consume it, the **event schema is a long-lived
contract**: the cost of churning it is multiplied by five.

## The core idea

A local, git-tracked, append-only JSONL event log of governance events, written
as a side effect of the commands that already produce these signals — mirroring
the `governed-project-memory` capture pattern exactly: **non-blocking,
best-effort, opt-out, no external service.** Domain producers
(`hookpolicy`, `gates`, `verify`, `workflow`) stay side-effect-free (G7); the
thin `cmd/` call-sites that already decide to block / fail / reject also emit.

## Scope

### In (v1)

- New leaf package `internal/telemetry`: a versioned `Event` struct (the
  contract), a non-blocking `Record(cfg, e)`, typed constructors per event type,
  and a `Read(dir)` reader for tests + the 5 downstream readers.
- 5 event types: `block`, `gate-failure`, `verify-rejection`,
  `complete-rejected`, `step-advanced`.
- Storage: append-only `.workflow/telemetry/events.jsonl`, one JSON object per
  line, `O_APPEND|O_CREATE`.
- `[telemetry] enabled` config (`*bool`, default ON / opt-out), wired like
  `[memory]`.
- Emission wired into the existing chokepoints in `cmd/`:
  `hook_prewrite.go` (block), `validate.go` (gate-failure), `complete.go`
  (verify-rejection, complete-rejected, step-advanced).
- `internal/telemetry/**` added to the `leaf` layer in `centinela.toml`.

### Out (v1, deferred)

- Any external sink, daemon, or network emission.
- Reading / aggregation / reporting surfaces — that *is* centinela-insights and
  team-dashboard; v1 only writes and exposes a reader API.
- A workflow attempt counter or backward step transitions — "rework" is a
  **derived** metric (N `complete-rejected` before a `step-advanced` for the same
  feature+step), NOT new workflow state.
- Schema migration tooling — v1 is `v1`; the version field exists so a future
  reader can branch, nothing more.
- Per-event redaction / PII policy — events carry only paths, gate names, and
  claim metadata already visible in the repo.

## Hard invariant (must not regress)

Telemetry is **best-effort and non-fatal**, exactly like `memory.Capture`. No
emission may change an exit code, block a write, or fail an advance. In the
prewrite hook (the hottest path) the event is written *before* `os.Exit(2)` and
any I/O error is swallowed to stderr. Default-on must not alter any existing
flow's observable behavior beyond appending a file.

## Dependencies

- `governed-project-memory` (shipped) — the non-blocking capture contract,
  `*bool` opt-out config pattern, `.workflow/` git-tracked storage. Telemetry
  copies it line-for-line.
- `headless-governance` (shipped) — the `schema: "centinela.telemetry/v1"`
  string-versioning precedent (matches `centinela.verdict/v1`).
- None blocking. Emits only from `cmd/`; reads `gates.Result`, `verify.Check`,
  `hookpolicy.PrewriteDecision` read-only.
