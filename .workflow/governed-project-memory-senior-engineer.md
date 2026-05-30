### Senior-Engineer Report: governed-project-memory
**Date:** 2026-05-29

#### Implementation Outline
1. `internal/config/memory.go` — `MemoryConfig` (`enabled *bool` default-true, `recall_max_entries`, `recall_max_bytes`) + normalizers; wired into `Config`, `applyDefaults`, and `defaultConfig` (the no-toml path bypasses `applyDefaults`, so defaults are applied there too).
2. `internal/memory/entry.go` — `Entry` type, content-hash id (`feature+type+body`, source excluded so the same fact is identical across worktrees), title derivation, byte sizing.
3. `internal/memory/serialize.go` — markdown + frontmatter marshal/unmarshal round-trip.
4. `internal/memory/dedupe.go` — ledger paths, `writeIfAbsent` (O_EXCL, per-entry files), `loadEntries` (skips malformed files).
5. `internal/memory/parse.go` — three parsers: lesson (edge-cases), verdict (gatekeeper), decisions (one entry per `## Decisions` bullet).
6. `internal/memory/capture.go` — step→source map + `Capture(feature, step, cfg)` orchestrator; non-blocking warnings to stderr; disabled config / non-capture steps are no-ops.
7. `internal/memory/index.go` — regenerate `index.json` from entry files (entries are source of truth).
8. `internal/memory/recall.go` + `rank.go` — `Recall(Query, cfg)`: deterministic ranking (dependency match 100 > shared tag 10 each > recency tie-break) + count/byte caps; `FeatureTags` tag profile.
9. `cmd/centinela/complete.go` — one-line `memory.Capture(feature, current, cfg)` after `saveWorkflow`, using the pre-advance `current` step.
10. `internal/planadvisor/{context.go,memory.go,context_summary.go,advisor.go}` — `bundle.Memory`, `recalledMemory` via `memory.Recall`, rendered through `contextLines` as a `🛡️👁️ MEMORY` line.
11. `internal/ui/render_memory.go` — pure terminal render block for recalled facts (plain-string input, no logic).

#### Files Touched
| Path | Reason |
|------|--------|
| internal/config/memory.go | New MemoryConfig + normalizers (gate, recall caps) |
| internal/config/config.go | Wire Memory field + memory defaults into Config/applyDefaults/defaultConfig |
| internal/memory/entry.go | Entry type + stable content-hash id |
| internal/memory/serialize.go | Markdown+frontmatter round-trip |
| internal/memory/dedupe.go | Idempotent per-entry write + load |
| internal/memory/parse.go | Three typed source parsers |
| internal/memory/capture.go | Step→source orchestrator, non-blocking warnings |
| internal/memory/index.go | Regenerate index.json from entries |
| internal/memory/recall.go | Recall + deterministic rank/score |
| internal/memory/rank.go | Caps, FeatureTags, set helper (split for G1) |
| cmd/centinela/complete.go | Thin capture wiring after saveWorkflow |
| internal/planadvisor/context.go | Add Memory to bundle, pass cfg |
| internal/planadvisor/memory.go | Recall integration (domain→domain) |
| internal/planadvisor/context_summary.go | Render memory line in contextLines |
| internal/planadvisor/advisor.go | Pass cfg to buildBundle |
| internal/ui/render_memory.go | Pure memory render block |

#### Architecture Compliance
- Boundary checks passed:
  - `internal/memory` imports only `internal/config` (+ stdlib). It does NOT import `cmd/`, `internal/ui`, or `internal/planadvisor`.
  - `internal/config` imports nothing internal.
  - `cmd/centinela/complete.go` wiring is one call (`memory.Capture`); all decisions live in the domain (G7 honored).
  - `internal/planadvisor` → `internal/memory` is a new domain→domain edge; `go build`/`go vet` confirm no import cycle (memory never imports planadvisor).
  - `internal/ui/render_memory.go` takes pre-formatted plain strings only — no memory/domain import, no logic.
- G1 file size: every new/modified file ≤ 100 lines (largest new file `serialize.go` at 91; `recall.go` split into `rank.go` to stay under).
- G7 outer-layer rule: no business logic in `cmd/`.

**Architecture-boundary note (flag for gatekeeper):** PROJECT.md → G2 enumerates allowed internal imports for `workflow`/`gates`/`ui` but does not list `planadvisor` or the new `memory` package explicitly. The new edges (`planadvisor → memory`, `memory → config`) follow the existing domain-imports-config pattern (planadvisor already imports `config` and `roadmap`). No silent violation; recommend G2 prose be updated to name `internal/memory` as a domain package that may import `internal/config`, and to acknowledge `internal/planadvisor → internal/memory`.

#### Type-Safety Notes
- No `interface{}`/`any` in business logic. `warn(format string, args ...any)` mirrors the stdlib `fmt.Fprintf` variadic signature for formatting only — not a typed-data shortcut.
- `MemoryConfig.Enabled` is `*bool` so an unset TOML value defaults to enabled (true) while `enabled = false` is honored, rather than a zero-value false ambiguity.
- `Recall` takes a typed `Query` struct (Feature/Dependencies/Tags) rather than loose string args.
- Frontmatter parse maps known keys via a typed `assignField` switch; unknown keys are ignored, bad timestamps fall back to zero time without panicking.

#### Trade-Offs
- **Body-only content hash (excludes source path):** chosen so the same fact captured from different worktree paths dedupes; cost is that two genuinely different facts with identical body+feature+type would collide (acceptable for v1).
- **Feature tag profile derived from the feature's own prior entries (`FeatureTags`)** rather than a roadmap/spec tag field, because no explicit feature-tag source exists yet. Deterministic; can be enriched later.
- **Recall summaries injected as a plain `contextLines` text line** (the directive is a prompt string, not lipgloss output); the `internal/ui` block is provided as the pure terminal renderer for any future cmd surface.
- **Capture never returns an error** (logs warnings) to guarantee it can never block `centinela complete`, per D6.

#### Verification
- `go build ./...` → Success.
- `go vet ./...` → No issues found.
- Existing `internal/planadvisor` + `internal/config` tests still pass (35 tests).
- Throwaway smoke test (removed) confirmed: dedupe idempotence, 3-decision parse, index regeneration, frontmatter round-trip, SC-09 ranking order (dep-feat > other-a > other-b), SC-11 caps, SC-12 disabled no-op.
- `centinela evidence validate governed-project-memory` → evidence ok.

#### Handoff
- Next role: qa-senior
- Outstanding TODOs:
  - Tests step owns unit/integration/acceptance coverage for all 13 scenarios; keep each `_test.go` ≤ 100 lines (G1 applies to tests).
  - Gatekeeper: please decide on the G2 prose update for `internal/memory` / `planadvisor → memory` (flagged above).
