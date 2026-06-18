# Edge Cases — deep-codebase-analysis

Enumerates the edge/boundary cases the acceptance spec
(`specs/deep-codebase-analysis.feature`) guarantees for `centinela analyze`.
Each row is the case, the expected behavior, and the spec scenario that pins it.
These map to executable Go assertions at the tests step (`tests/acceptance/`,
`internal/analyze/*_test.go`, `tests/integration/analyze_test.go`).

| # | Edge case | Expected behavior | Spec scenario |
|---|-----------|-------------------|---------------|
| 1 | Empty / docs-only repo | Valid inventory; `primaryLanguage` is `""`; manifests and graph empty; exit 0 (not an error) | "An empty or docs-only repo yields a valid empty inventory and exits 0" |
| 2 | Polyglot repo (Go + JS + Ruby) | All languages counted; `primaryLanguage` = highest file count; equal counts broken alphabetically (deterministic); every ecosystem's manifest still listed | "A polyglot repo counts every language and picks the highest-count primary deterministically" |
| 3 | `go list` fails (uncompilable Go / no toolchain) | Go graph recorded as best-effort empty `edges` with a `note`; rest of inventory still emits; exit 0 (diagnostic, not a gate) | "When go list fails the Go graph is recorded as best-effort empty with a note and the rest still emits" |
| 4 | `package.json` present but malformed JSON | npm manifest recorded as detected-but-unparsable (no parsed build/test/deps); scan continues; exit 0 | "A malformed package.json is recorded as detected-but-unparsable and the scan continues" |
| 5 | Multiple manifests across ecosystems | All manifests detected and listed; `primaryLanguage` decides only the headline | "A polyglot repo counts every language…" + "Analyzing a Node project detects the npm manifest…" |
| 6 | No i18n at all | `locales` empty list; summary shows 0; not an error | "A repo with no i18n reports an empty locale list and exit 0" |
| 7 | Huge repo / deep tree | `packages` layout is depth-bounded so output stays small/bounded | covered by the layout assertions in the happy-path + no-manifest scenarios (bounded `packages`) |
| 8 | Re-run after output already exists | Deterministically overwritten; byte-identical when nothing changed; clean git diff | "Re-running analyze on an unchanged repo produces a byte-identical inventory" |
| 9 | Symlinks / unreadable individual files | Skipped without aborting the scan; inventory still emitted; exit 0 | covered by the read-only skip-set scenario + best-effort sub-detector semantics |
| 10 | Skip set + gitignored paths | `vendor/`, `node_modules/`, `.git/`, `.workflow/`, `dist/`, `build/`, and gitignored paths excluded from counts | "The scan skips dependency and build directories so counts reflect real source" |
| 11 | No recognized manifest (unfamiliar repo) | Languages + layout populated; manifests empty; graph `none`/empty; exit 0 (never hard-fails) | "A repo with no recognized manifest still produces a valid inventory and exits 0" |
| 12 | Un-writable output path | Hard error: non-zero exit; clear stderr message; NO partial/corrupt `.workflow/analysis.json` left on disk | "Running analyze with an un-writable output path fails clearly…" |
| 13 | Non-existent / unreadable root | Hard error: non-zero exit; clear stderr naming the root; no inventory written | "Running analyze against a non-existent or unreadable root fails clearly and writes no inventory" |
| 14 | Read-only guarantee | Only `.workflow/analysis.json` (or `--out` target) is created/modified; no source file mutated | "The scan skips dependency and build directories…" (read-only assertion) |
| 15 | `--out` override | Inventory written to the custom path; default `.workflow/analysis.json` not created that run | "The --out flag redirects the inventory to a custom path" |
| 16 | Schema stability | `schemaVersion` present (= 1) so downstream consumers detect format changes | "Analyzing a Go module writes a complete inventory…" (schemaVersion assertion) |

## Negative / failure cases (non-zero exit)

- **Un-writable output path** (#12) — the only sub-detector-independent hard
  failure; everything else degrades best-effort.
- **Non-existent / unreadable analysis root** (#13) — clear error, no artifact.

All other irregularities (uncompilable Go, malformed manifests, missing
toolchain, symlinks, unreadable individual files) are **non-fatal**: analyze is
diagnostic, records a best-effort/empty result with a reason where relevant, and
exits 0.
