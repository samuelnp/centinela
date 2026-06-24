# Implementation Plan — adoption-baseline

> Feature brief: `docs/features/adoption-baseline.md`.
> Spec: `specs/adoption-baseline.feature` (to be authored by feature-specialist).

Phase 9's capstone turns the already-shipped baseline machinery into the
**adoption moment** of the brownfield flow. `audit-baseline-ratchet` gave us
`audit.Record`/`Save`/`Load`/`Ratchet`, the `audit_baseline` validate gate, and
`centinela audit baseline`. This feature adds a thin, deliberate, *one-time*
adoption command — `centinela adopt` — on top of that machinery, plus the
human-facing report and the docs sequencing that make a brownfield adopter
actually record the baseline before the first ruinous `validate`. **No gate
logic, fingerprinting, or file format changes.**

## Delta over `centinela audit baseline` (read this first)

| Axis | `centinela audit baseline` | `centinela adopt` (this feature) |
|------|---------------------------|----------------------------------|
| Purpose | Ratchet **re-baseline** (ongoing) | **One-time adoption** (day one) |
| Existing baseline | Always overwrites | **Refuses** unless `--force` |
| Output | One-line write confirmation | **Adoption report** (per-gate counts, total, ratchet-to-zero framing) |
| Discoverability | Buried under `audit` | Named step between `roadmap brownfield` and first `start` |

The point of `adopt` is the **skip-if-exists default** + the **report** + the
**flow placement**. Strip those and it would be a pure alias — so the plan keeps
all three.

## Decisions (proposed — feature-specialist to confirm)

1. **Dedicated `centinela adopt` subcommand**, not an `init` prompt and not an
   `audit adopt` subverb. Rationale: matches the sibling brownfield surface
   (`centinela roadmap brownfield` is its own command, not folded into `init`);
   keeps adoption a deliberate, top-level, discoverable act; avoids tangling the
   skip-if-exists semantics into `audit`'s always-overwrite group.

2. **Compose existing functions; add one small `audit.Adopt` orchestrator.**
   The cmd must *not* hold the skip-if-exists decision (G7: no business logic in
   `cmd/`). So add a thin `audit.Adopt(cfg, force) (Outcome, error)` in
   `internal/audit` that: `Load`s the configured `BaselinePath`; if it exists and
   `!force`, returns an outcome flagged `Skipped` (no write); else `Record`s and
   `Save`s and returns the recorded `Baseline` for the report. This keeps the
   "should I overwrite?" rule in the aggregator where it belongs and keeps `cmd/`
   a thin wire. (No new on-disk artifact — same `.workflow/audit-baseline.json`.)

3. **Report shape.** Render the recorded `audit.Baseline`: a header
   ("Adopted baseline — N findings across M gate(s)"), a per-gate line
   (`gate: K findings`) sorted by gate name, and a closing framing line
   ("These are accepted as legacy debt — ratchet to zero over time; new
   violations are gated strictly."). On skip, render a distinct "baseline already
   exists at <path> — re-run with --force to re-adopt" message. Reuse existing
   `ui` styles (`StyleBold`/`StyleMuted`/`StyleGreen`) as `render_audit.go` does.

4. **`--json` verdict** for the onboarding agent: `{adopted: bool, skipped: bool,
   path, total, per_gate: {gate: count}}`, mirroring `audit --json`'s
   cmd-layer-marshals-an-Outcome pattern (see `audit_render.go`). Exit 0 on both
   adopt and skip (skip is not an error — it is the safe default); reserve
   non-zero for real failures (load/save errors).

5. **Skip-if-exists is the default; `--force` re-records.** Opposite of `audit
   baseline`. Skip is non-blocking, exit 0, with a clear message. This is the
   single most important behavioral guard against silently widening an
   established baseline.

## Proposed surface

```
centinela adopt [--force] [--json]
```
- bare: record the baseline if none exists; print the adoption report.
- `--force`: re-record over an existing baseline.
- `--json`: emit the machine-readable adoption verdict.

## File layout (each ≤100 lines)

| File | Layer | Role |
|------|-------|------|
| `internal/audit/adopt.go` | aggregator | `Adopt(cfg, force) (Outcome, error)` + `Outcome` type (Skipped flag + recorded Baseline + path); the skip-if-exists rule lives here |
| `internal/audit/adopt_test.go` | test | unit: skip-if-exists, force-overwrite, fresh-adopt, load/save error paths |
| `internal/ui/render_adopt.go` | presentation | `RenderAdoption(Outcome) string` over the `audit.Baseline`/`Outcome` types |
| `internal/ui/render_adopt_test.go` | test | renders fresh + skip cases |
| `cmd/centinela/adopt.go` | cmd (outer) | thin Cobra command: load cfg, call `audit.Adopt`, render or `--json` |
| `cmd/centinela/adopt_render.go` | cmd (outer) | `--json` verdict marshalling (mirrors `audit_render.go`) |
| `cmd/centinela/adopt_test.go` | test | cmd-level: skip exit 0, force, json shape |

If `adopt.go` + `Outcome` crowd 100 lines, split the `Outcome`/counts helpers
into `internal/audit/adopt_outcome.go`. Keep `RenderAdoption` lean by reusing the
`auditSection` helper pattern.

## cmd wiring

`cmd/centinela/adopt.go` registers `adoptCmd` on `rootCmd` in its `init()` (same
pattern as `auditCmd`). It calls `config.Load()`, then `audit.Adopt(cfg, force)`,
then either `printAdoptJSON(cmd, outcome)` or
`fmt.Fprintln(cmd.OutOrStdout(), ui.RenderAdoption(outcome))`. No business logic
in cmd — the skip decision is inside `audit.Adopt`.

## Import-graph / layer check (expect NO new edges)

- `internal/audit` is **already** in the `aggregator` layer
  (`centinela.toml` → `paths = [... "internal/audit/**" ...]`,
  `allow = ["domain","leaf","aggregator"]`) and **already** imports
  `internal/gates` (domain) + `internal/config` (leaf). `Adopt` adds no new
  import — `Load`/`Record`/`Save` are same-package.
- `internal/ui` **already** imports `internal/audit` (`render_audit.go`), so
  `RenderAdoption` over `audit.Outcome`/`Baseline` adds no new edge.
- `cmd/**` may import `domain`, `leaf`, `aggregator` — already satisfied.
- **Conclusion: zero G2 import-graph matrix changes required.** Confirm in the
  gatekeeper step by running `centinela validate`'s `import_graph` gate.

## Config

No new config. `adopt` reuses `cfg.Gates.AuditBaseline.BaselinePath` (default
`.workflow/audit-baseline.json`) so adoption writes exactly the file the ratchet
gate and `audit` command read. No `[gates.audit_baseline]` schema change.

## Docs sequencing note

The only docs deliverable is a sequencing addition to the brownfield onboarding
guide (e.g. `docs/architecture/new-project-guide.md` / the brownfield section):
insert **`centinela adopt`** as the explicit step between
`centinela roadmap brownfield` and the first `centinela start`, with a one-line
rationale ("record the legacy debt as accepted before your first `validate`").
No new product/UX surface — this stays a lightweight docs step.

## Risks (carry into the spec)

- **Thin/duplicate of `audit baseline`.** Mitigated by the three deliberate
  deltas (skip-if-exists, report, flow placement); the spec must assert all
  three so the value is testable, not cosmetic.
- **Silently overwriting an established baseline.** Mitigated by skip-if-exists
  default + explicit `--force`; assert the refuse-without-force path.
- **Recording a baseline that hides NEW post-adoption violations.** Inherent and
  correct: the snapshot is point-in-time; the ratchet (already shipped) fails on
  anything new after adoption. Document, don't "fix."

## Test strategy (for qa-senior, noted here)

Unit (`internal/audit/adopt_test.go`): skip-if-exists returns Skipped + no write;
`force=true` overwrites; fresh repo records full set. Renderer test: fresh report
shows per-gate counts + total; skip message distinct. cmd test: `adopt` on
existing baseline exits 0 + prints skip; `--force` rewrites; `--json` shape.
Acceptance (`specs/adoption-baseline.feature`): the day-one adoption story —
adopt once, see the report, re-run is refused, post-adoption validate is clean.
