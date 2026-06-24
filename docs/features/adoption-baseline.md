# Feature Brief — adoption-baseline

> Phase 9: Brownfield Onboarding, capstone. `audit-baseline-ratchet` already ships
> the machinery — `audit.Record`/`Save`/`Load`/`Ratchet`, the `audit_baseline`
> validate gate, and `centinela audit baseline` to write the snapshot. What is
> still missing is the **adoption moment**: a first-class, discoverable step in
> the brownfield flow that records the snapshot *once, deliberately, before the
> first `validate`* and shows the team the legacy debt they are consciously
> accepting. This feature is that step — `centinela adopt` — not a second way to
> reset the ratchet.

## Problem

A team adopts Centinela on a mature repo. They run the brownfield flow (`analyze`
→ `synthesize` → `reconstruct` → `roadmap brownfield`), then `centinela start`
their first real gap. Day one, `centinela validate` reports thousands of
pre-existing G1 file-size / G2 layer / secret / spec violations that predate
adoption. The gates are unusable; the team turns them off and Centinela's value
evaporates exactly where it was supposed to land.

The cure exists — record a baseline so only *new* work is gated — but it is
**undiscoverable and out of sequence**. Nothing in the brownfield flow tells the
adopter to run `centinela audit baseline`, and `audit baseline` is framed as a
ratchet *re-baseline* (a deliberate overwrite you re-run as debt shrinks), not as
the one-time *adoption* act it needs to be on day one. So the adopter either
never records a baseline (and drowns), or stumbles onto `audit baseline` with no
guidance, no visibility into what they just accepted, and a command that will
silently overwrite the baseline if run again.

## Who / why

The **brownfield adopter** (team lead / staff engineer turning Centinela on for
an existing codebase) and the operating agent driving the onboarding flow. They
need one obvious command, run once, at the right moment, that records the legacy
debt as accepted and **shows them the bill** — "you are starting with N findings
across these gates; ratchet them to zero over time." **Why now:** every
dependency is shipped and proven — `deep-codebase-analysis` and
`audit-baseline-ratchet` are both DONE — so the only remaining gap is sequencing
and the deliberate, visible first-adoption act.

## Net-new value (the delta over `centinela audit baseline`)

`centinela audit baseline` is the **ratchet re-baseline** primitive: overwrite the
snapshot on demand as debt shrinks. `centinela adopt` is the **one-time adoption**
act layered on the same `Record`/`Save`, and it differs in three deliberate ways:

1. **First-adoption safety.** `adopt` *refuses* to overwrite an existing baseline
   (skip-if-exists) unless `--force`. The opposite default from `audit baseline`,
   which always overwrites. Adoption is a one-time act; re-baselining is the
   ongoing one. This prevents an adopter from silently *widening* an established
   baseline (re-tolerating debt that had been paid down).
2. **An adoption report.** A human-facing summary of the accepted debt —
   per-gate counts, total, and the "you are starting with N findings, ratchet to
   zero over time" framing — so the team *sees and consciously accepts* the debt.
   `audit baseline` only prints a terse one-line write confirmation.
3. **Flow integration & discoverability.** `adopt` is the named, documented step
   between `roadmap brownfield` and the first `centinela start`, so the adopter
   records the baseline *before* the first ruinous `validate`.

## In / Out scope

**In (v1):**
- New `centinela adopt` command (thin `cmd/` wrapper) that composes the existing
  `audit.Load` (existence check) + `audit.Record` + `audit.Save`.
- Skip-if-exists default with `--force` to re-adopt; a `--json` verdict for
  agent consumption, mirroring `audit --json`.
- An adoption report (per-gate counts, total, ratchet-to-zero framing) rendered
  via a new `internal/ui` function over the existing `audit.Baseline` type.
- Docs sequencing: brownfield onboarding guide names `adopt` as the step before
  the first `centinela start`.

**Out:**
- Auto-running adoption inside `centinela init` / brownfield without explicit
  consent — adoption is a deliberate act, never silent.
- Ratchet-down automation, per-gate selective adoption, or re-baselining as debt
  shrinks — that is `audit baseline`'s job and stays there.
- Any change to gate logic, fingerprinting, or the baseline file format.
- A second on-disk artifact — `adopt` writes the same
  `.workflow/audit-baseline.json` the ratchet already reads.

## Acceptance summary

1. On a repo with no baseline, `centinela adopt` records
   `.workflow/audit-baseline.json` capturing every current violation across
   participating gates (full-repo scan), and prints an adoption report with
   per-gate counts and total.
2. With a baseline already present, `centinela adopt` refuses to overwrite
   (skip-if-exists), explains why, and exits without modifying the file.
3. `centinela adopt --force` re-records over an existing baseline.
4. After adoption, an unchanged-repo `validate`/`audit` run reports all
   violations as baselined (non-blocking) — adoption made the gates usable.
5. `--json` emits a machine-readable adoption verdict for the onboarding agent.
6. All new source files ≤100 lines; no cross-layer import violations (reuses the
   already-mapped `internal/audit` aggregator + `internal/ui`).

## Surface classification

Internal CLI developer tool (a `centinela` subcommand for adopters), not a
shipped end-user product surface. The docs step stays lightweight: a sequencing
note in the brownfield onboarding guide, no new product/UX surface.
