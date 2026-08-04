# Feature: hook-context-panel-diet

**Phase:** 13 — Lighter Centinela
**Archetype:** canonical
**Depends on:** none (follows token-diet, which dieted prompt docs rather than hooks)

## Problem

`centinela hook context` runs on **every user turn** and its output is injected
into the model's context. Measured on this repo with a single active workflow at
the plan step:

| Measure | Bytes |
|---|---|
| Total emitted | **1770** |
| Content after stripping trailing spaces | 1034 |
| **Trailing whitespace** | **736 (41.6%)** |

Every panel line is padded to a fixed width (93 columns here) with spaces, so
that the box renders as a rectangle in a terminal. Injected into a prompt, that
rectangle is invisible — the padding is pure cost. Nothing in the model's
reading of the text depends on a line being 93 characters long.

The remaining 1034 bytes are also worth auditing: the ACTIVE WORKFLOWS panel
redraws a full step ladder (`✓ plan · ✓ code · ✓ tests · ▶ validate · ○ docs`)
plus box chrome on every turn, and the nudge/REVIEW panels restate standing
instructions the model has already been given.

For comparison, `hook statusline` (120 bytes) and `hook session-start`
(1092 bytes) carry **zero** trailing whitespace — this is specific to the
panel renderer, not to hook output generally.

This is the largest remaining recurring token cost in the framework, and unlike
prompt-doc size it is paid per turn rather than per delegation.

## Expected outcome

1. **No trailing whitespace in hook output.** The padding exists to square a
   terminal box; the hook surface is not a terminal. Removing it is a ~41%
   reduction with zero information loss — this alone is the bulk of the win and
   should be the first, independently revertible slice.
2. **The panels earn their remaining bytes.** Audit what ACTIVE WORKFLOWS and
   the nudge panels actually tell the model that it does not already know from
   the directive lines, and cut what is redundant. The governance signal must
   survive: the model still has to learn the active feature, its step, and any
   blocking action.
3. **A measurement guard**, so the size cannot silently regress — a test that
   fails if hook context output exceeds a recorded budget, in the spirit of the
   existing prompt-doc budget checks.
4. **Terminal rendering is unaffected** where a human actually reads a panel
   (`centinela status`, `validate` output, blocked-write refusals). This feature
   changes what the HOOK emits, not how the CLI renders to a TTY.

## Out of scope

- The prompt-doc budget work already done by `token-diet`.
- `hook session-start` rehydration content (measured clean of padding; its size
  is roadmap data, a different question).
- Changing what the hooks *decide* — this is presentation only. No gate, no
  directive semantics, no step logic changes.
- Colour/ANSI handling: measured output already carries no escape codes on this
  path.

## Constraints

- **The governance signal is not negotiable.** A diet that drops the active
  feature, the current step, or a blocking instruction has broken the harness to
  save bytes. Every removal must be justified against what the model needs to
  act correctly.
- Human-facing TTY rendering must not regress; if the same renderer serves both,
  the split has to be explicit rather than incidental.
- 100-line file cap incl. `_test.go` in `cmd/` and `internal/`; per-package
  coverage ≥97% on touched packages.
- Existing tests pin hook output — `token-diet` added byte-identity assertions.
  Expect breakage there and treat those tests as the specification of what must
  stay, updating them deliberately rather than to match whatever is produced.
