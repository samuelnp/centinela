# Edge Cases: lean-evidence-footprint

Each case is covered by a test in the unit / integration / acceptance tiers.

## Covered

| # | Edge case | Handling | Test |
|---|-----------|----------|------|
| 1 | `roadmap.json` must stay tracked despite the broad `*.json` ignore | `!` negation re-includes it | `integration:TestKbAndRoadmapNotIgnored`, `acceptance:TestAccEvidenceIgnoreMatrix` |
| 2 | Negation ordering — `!roadmap` before the `*.json` ignore would be a no-op | Unit asserts negation line follows the ignore line | `unit:TestRoadmapNegationAfterJSONIgnore` |
| 3 | Per-feature root `<feature>.json` (not just `-<role>.json`) is machine-only | Broad `.workflow/*.json` covers it | `integration:TestEvidencePlumbingIgnored` (`f.json`) |
| 4 | `.lock` files are advisory and never deleted in code | Ignored by `.workflow/*.lock`; not removed from disk | `integration:TestEvidencePlumbingIgnored`, `acceptance:TestAccEvidenceIgnoreMatrix` |
| 5 | `-<role>.md` narratives must remain tracked (reviewer + LLM KB) | Not matched by any ignore pattern | `integration:TestKbAndRoadmapNotIgnored`, `acceptance:TestAccEvidenceIgnoreMatrix` |
| 6 | Retroactive `git rm --cached` must not delete local files | `--cached` clears index only; disk copy survives | `acceptance:TestAccRetroactiveUntrack` |
| 7 | The shipped `.gitignore` (not a hand-written fixture) is asserted | All tiers read repo-root `.gitignore` | `unit:TestGitignoreHasEvidencePatterns` |
| 8 | A previously-committed plumbing file is genuinely untracked after cleanup | Force-add → commit → `rm --cached` → `ls-files` empty | `acceptance:TestAccRetroactiveUntrack` |

## Residual Risks

- No change to lock semantics, so no concurrency test is added — the
  unlink-after-unlock race is avoided by *not* deleting locks in code.
  Local 0-byte locks persist but are untracked and harmless.
- No coverage-gate movement: this feature adds no `internal/` or `cmd/`
  source, only `.gitignore` plus test files.
