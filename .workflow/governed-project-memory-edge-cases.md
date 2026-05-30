# Edge Cases: governed-project-memory

## Covered

- **Missing source artifact** (SC-06) — capturing a step whose source file is
  absent writes no entry and does not block `centinela complete`.
  Tested: `internal/memory/capture_test.go`, `tests/acceptance/..._c_test.go`.
- **Malformed source artifact** (SC-07) — an edge-cases file with no bullets (or
  an empty gatekeeper report) yields no entry and no blocking error.
  Tested: `internal/memory/parse_test.go` (empty bullets / empty verdict),
  `tests/acceptance/..._c_test.go`.
- **No `## Decisions` section** (SC-04) — plan-step capture is a clean no-op.
  Tested: `internal/memory/parse_test.go`, `tests/acceptance/..._b_test.go`.
- **Idempotent re-capture** (SC-05) — re-completing a step does not duplicate
  entries; dedupe keys on content hash.
  Tested: `internal/memory/dedupe_test.go`, integration + acceptance suites.
- **Empty ledger recall** (SC-10) — `Recall` over an empty ledger returns no
  entries and raises no error; the plan-advisor directive omits the MEMORY block.
  Tested: `internal/memory/recall_test.go`.
- **Deterministic ranking ties** (SC-09) — dependency match > shared tag >
  recency, with a stable tie-break so output order is reproducible.
  Tested: `internal/memory/rank_test.go`, `rank_extra_test.go`.
- **Recall caps** (SC-08/SC-11) — count (`recall_max_entries`) and byte
  (`recall_max_bytes`) budgets bound the injected slice.
  Tested: `internal/memory/recall_test.go`.
- **Disabled / nil config** (SC-12) — capture and recall are full no-ops when
  `[memory] enabled = false` or config is nil.
  Tested: `internal/memory/capture_test.go`, `capture_more_test.go`,
  `internal/config/memory_test.go`.
- **Non-capture step** — completing the `code` step (no mapped source) writes
  nothing. Tested: `internal/memory/capture_more_test.go`.
- **Concurrent completes across worktrees** (SC-13) — per-entry files keyed by
  content hash mean simultaneous captures for different features do not clobber
  each other. Tested: `tests/integration/..._recall_test.go` (goroutine race).

## Residual Risks

- **Hash collision across distinct facts** — id is a content hash; two different
  facts colliding is astronomically unlikely but not impossible. Mitigation:
  collisions would silently dedupe one entry; acceptable for v1 given the input
  domain (short curated facts). Revisit if the ledger grows large.
- **Tag extraction quality** — v1 derives tags from simple keyword signals;
  poor tags degrade recall *relevance* (never correctness). Out of scope to
  improve here; recall is capped so noise stays bounded.
- **Index/entries drift** — `index.json` is a regenerable cache; if hand-edited
  it can diverge from entry files (the source of truth). Mitigation: regenerate
  on every capture. Not separately tested beyond regen coverage.
- **Very large ledgers** — recall scans entries linearly; fine at expected
  scale, unmeasured at thousands of entries. Deferred (no semantic store in v1).
