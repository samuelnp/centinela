# lean-evidence-footprint — validation-specialist

## Gates Run

`centinela validate` (diff-aware, 14 files changed) — all green:
- ✓ G1: File Size · ✓ G-Build: Cross-Compile (6 targets) · ✓ roadmap_drift
- ✓ `go test ./...` · ✓ `go test ./tests/acceptance/...`
- ✓ `./scripts/check-coverage.sh` · ✓ `./scripts/check-fmt.sh`
- ⚠ `import_graph`, ⚠ `spec-traceability-gate` — both emitted with empty
  bodies (no packages/scenarios listed), non-blocking, pre-existing in
  diff-aware mode.

## Synthesis

The change is repository-configuration only (`.gitignore` + 751 index
removals). Gatekeeper verdict: SAFE. The decisive evidence is this run
itself: the validate gate passed on a branch whose own per-role evidence
`.json` is now gitignored (0 tracked) yet still present on disk and read by
`centinela complete` — proving the workflow is unaffected. Coverage gate is
unmoved (no `internal/`/`cmd/` source added). Knowledge base intact: 673
`-<role>.md` files remain tracked; `roadmap.json` remains tracked.

## Decision

PASS → hand off to documentation-specialist. Per-feature committed
`.workflow/` footprint drops ~1,419 → ~672 files; future features commit
only their readable `.md` narratives.
