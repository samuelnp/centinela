# Feature: deterministic-artifact-scaffolds

- surface: internal
- status: planned
- roadmap: Phase 6 — Capability-Adaptive Governance
- fixes: low-capability models fail artifact contracts on SHAPE, not substance —
  burning retries hand-looping mechanically-derivable boilerplate (e.g. ~80
  `docs/features/*.md` paths into a plan-role `inputs` array every single time)

## Problem

The evidence CLI (`evidence-cli`, shipped) removed *JSON authoring by heredoc*,
but it did not remove *shape toil*. Two gaps remain:

1. **`centinela evidence init <feature> <role>` pre-fills only scalars.** It
   leaves `inputs` and `outputs` empty. For the two plan roles (big-thinker,
   feature-specialist) the validator (`validatePlanSnapshotInputs`) then demands
   that `inputs` snapshot **every** `docs/features/*.md` plus the feature's own
   plan file. Today that is ~80 paths the agent must hand-loop into `inputs` on
   every run — pure mechanical derivation, identical to what the validator
   already computes internally via `requiredPlanInputs(feature)`.

2. **`centinela artifact new` and the companion markdown ship italic prose
   placeholders** (`_List each edge case…_`). A strong model reads these as
   instructions; a weak model under the `strict` profile mis-shapes the
   artifact, fails the contract on structure, and burns retries inventing a
   skeleton that the framework could have stamped out deterministically.

This is the exact failure mode Phase 6 targets: under `strict` / `limited`-class
driver models, rails must be **physical** (stamped structure + pre-filled
derivable values), not prose. The proven template is the docs CLI-fallback
pattern (`internal/docgen`): compute everything mechanically derivable first,
leave only genuine substance for the LLM.

## Who is hurting

- **`limited`-capability / local driver models** under the `strict` profile —
  they spend their scarce tokens satisfying shape and retry on shape failures.
- **Plan-role agents** of any capability — the 80-path `inputs` hand-loop is
  toil for everyone and a frequent validator miss.

## The core idea

`evidence init` and `artifact new` become **shape-complete by construction**.
Two distinct fill strategies, never mixed:

- **Mechanical pre-fill (real, derivable values)** — only for `inputs` of the
  two plan roles, derived from the single source of truth the validator already
  uses. Pre-filled values are *real* and pass validation as-is.
- **`<FILL: …>` slots (substance the agent must supply)** — only in **markdown**
  bodies (companion `.md` and `artifact new` `.md`). Greppable, survives the
  no-HTML-escape marshal, and is never valid in finished content so a leftover
  slot is trivially caught.

## Scope

### In

- Promote `requiredPlanInputs` → exported `orchestration.RequiredPlanInputs`
  (single source of truth shared with the validator — prediction and validation
  can never drift).
- `evidence init` pre-fills `inputs` with `RequiredPlanInputs(feature)` for
  big-thinker and feature-specialist **only**; other roles' `inputs` stay empty.
- A canonical `<FILL: …>` marker (one constant + helper in `internal/evidence`).
- Per-role companion markdown skeletons seeded with `<FILL: …>` slots, replacing
  the one-line placeholder, for the report-bearing roles.
- `artifact new` markdown bodies: italic prose → explicit `<FILL: …>` slots,
  plus mechanical pre-fill where derivable (e.g. gatekeeper "Analyzed Specs"
  lists `specs/*.feature`).
- Pre-fill is default-on; existing existence-check / `--force` semantics guard
  overwrite.

### Out (v1)

- **No pre-fill of `outputs` in evidence JSON.** The validator
  (`validateActionableOutputs`) requires every `outputs` entry to be a **real
  file on disk** at validate-time. At `init` time the predicted outputs
  (`docs/plans/<feature>.md`, `specs/<feature>.feature`, the edge-cases report)
  do not exist yet, and pre-seeding them risks listing files the agent never
  creates. Pre-filling `outputs` converts a benign "you forgot to list it" miss
  into a hard "you listed a file you didn't create" failure. Outputs stay a
  genuine fill slot. (See plan → Divergence.)
- **No `<FILL: …>` in any evidence JSON list field** — a FILL string in
  `outputs`/`inputs` would fail the real-files / snapshot validators.
- **No `PredictedOutputs` API.** Deferred with `outputs` pre-fill; revisit only
  if a later feature finds a validator-safe seam (e.g. predicted-but-unvalidated
  hint field).
- **No `--minimal` opt-out flag.** Pre-fill is cheap and always wanted; skip.
- **No new model-awareness.** Scaffolds are *more valuable* under strict/limited
  but are unconditional here; profile-gating is out of scope.

## Dependencies

- `enforcement-profiles` (shipped) — scaffolds are most valuable under `strict`.
- `model-capability-profiles` (shipped) — `limited`-class driver models are the
  prime beneficiary. Neither is a code dependency; both are the *why*.
- `evidence-cli` (shipped) — provides `evidence init`, `artifact new`,
  `Skeleton`, companion writer, and the orchestration validator this extends.

## Back-compat invariant (must not regress)

- Pre-existing `.workflow/*.json` still validate (no schema change; `inputs`/
  `outputs` remain `[]string`).
- The orchestration validator (`internal/orchestration`, `internal/verify`,
  `internal/gates`) is **not loosened**. Pre-filled `inputs` must *pass*
  `validatePlanSnapshotInputs` by construction (same source fn). No FILL marker
  ever lands in a validated JSON list.
