# Edge Cases: roadmap-edit-move

## Covered

### edit
- Only the flags provided change; unspecified fields (archetype, dependsOn) stay intact.
- `--depends-on` Changed sentinel: omitted = preserve deps; explicit empty = clear; a value = replace.
- `--name` rename rewrites every dependent's `dependsOn` across ALL phases (same-phase + cross-phase).
- Rename collision refused, naming the owning phase; roadmap.json byte-identical.
- Rename to an invalid (non-kebab) slug refused; byte-identical.
- Rename to the SAME name: no dependents rewritten, write is idempotent (see Residual Risks re byte-identity).
- Unknown-dep and self/multi-hop cycle refused post-mutation; byte-identical.
- Unknown slug ("not found") and missing/malformed roadmap.json refused; byte-identical / still absent.
- Archetype field validated (unsupported archetype refused via ValidateDependencies).
- Malformed target/sibling feature JSON surfaces a decode error instead of a silent partial write.

### move
- Cross-phase relocation appends by default; untouched phases byte-identical (from canonical on-disk form).
- Anchor placement first/last/middle via `--before`/`--after`.
- Backlog/Baseline refused as source or target; unknown phase and unknown anchor refused; byte-identical.
- Not-found slug refused; byte-identical.
- Draft flag and name-keyed quality entries preserved verbatim across the move.
- Move of a depended-on feature is allowed (dependency is by name, not phase); validate stays PASS.
- Post-move typed re-decode failure (malformed sibling) refuses the write.

### reorder
- Within-phase and cross-phase (anchor in another phase) reposition.
- No-op reorder (order unchanged) performs NO write — byte-identical, guaranteed by the order-snapshot guard.
- Backlog/Baseline anchor, not-found slug, unknown anchor, and missing `--before/--after` refused; byte-identical.

### raw helpers
- `rewriteDependents` multi-dependent/multi-phase exact-byte rewrite; untouched phases not re-rendered; no-match no-op.
- `anchorPos` before/after/append/unknown-anchor and malformed-name error.
- `insertFeatureAt` head/tail insert and out-of-range rejection.
- `phaseOrder`/`schedulablePhaseIndex` decode-error branches on malformed later phases.

## Residual Risks

- **Spec vs code deviations (deferred to validation-specialist), all data-safe (roadmap.json byte-identical):**
  1. `move --before/--after <self>` errors "anchor ... not found" (removes before resolving the anchor) rather than
     exiting 0 as spec scenario "move self-anchor no-op" aspires. No silent mutation.
  2. `edit --name <same>` is a semantic no-op but re-renders the target's phase one-per-line, so it is NOT byte-identical
     against a `json.Indent`-canonical file (spec scenario "same-name byte-identical"). It IS idempotent once settled.
     This is inherent to the mutate-and-render design: only rejected ops and the reorder no-op guard are truly byte-identical.
  3. Error wording differs from the spec's example substrings: Backlog/Baseline targets say "non-schedulable" (spec: "unknown phase");
     unknown anchors/slugs say "not found" (spec: "unknown feature"). Behavior (non-zero exit + byte-identical) is correct.
- **Untouched-phase byte-identity** only holds when the on-disk file is already in canonical render form; a phase previously
  written one-per-line is re-indented on the next write (pre-existing rawrender behavior, acknowledged by the senior engineer).
- **Unreachable defensive guards** (decodePhase errors in `requireSchedulablePhaseIdx`/`insertFeatureAt`/`replaceFeatureAt`,
  `compactBytes` failures) are left uncovered: the phase was already decoded upstream, so these branches cannot fire in practice.
