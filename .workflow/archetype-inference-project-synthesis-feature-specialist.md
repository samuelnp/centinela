# archetype-inference-project-synthesis — feature-specialist

## Behavior Summary

`centinela synthesize` loads `.workflow/analysis.json` (via a new
`analyze.Load`), runs a deterministic `Inferer` that scores the five archetypes
from a swappable rule table over the Inventory (PrimaryLanguage + Manifests/
frameworks/deps + Packages + Graph), picks `Best` with `Confidence` and
`Ambiguous`, then `Draft(inv, inf)` assembles a full `PROJECT.md` string:
archetype-specific sections (from an `archetypeProfile` table + layer mapping
bucketed from `Packages`) plus inventory-derived sections (tech stack, folder
tree, locales, language-keyed naming). `WriteDraft` writes `PROJECT.md` only if
absent, else `PROJECT.draft.md`. New aggregator package `internal/synthesize`
(~11 source files ≤100 lines) + `cmd/centinela/synthesize.go` (thin, mirrors
analyze.go) + `internal/ui/render_synthesize.go`. Full file plan in
`docs/plans/archetype-inference-project-synthesis.md`.

## Acceptance Criteria (Gherkin)

See `specs/archetype-inference-project-synthesis.feature` (7 scenarios): Rails→
rails-native draft; Go→n-tier; game→ecs; ambiguous→low-confidence + rationale;
missing analysis.json→actionable error (no file written); existing PROJECT.md→
PROJECT.draft.md, original unchanged; deterministic re-run byte-identical.

## UX States

Command output (`RenderInferenceSummary`): best archetype + confidence
(high|medium|low) + top signals/rationale + the written path; a notice when the
existing PROJECT.md was preserved and a draft written instead. `--json` emits the
`Inference` for tooling. Errors: missing inventory → "run centinela analyze
first"; schema drift / malformed JSON → distinct messages.

## Edge Cases

Polyglot (primary-language naming, small-margin confidence); no manifest
(folder/graph only → n-tier low); unknown framework (ignore signal); monorepo
(union signals + note); exact tie (Ambiguous, low, runners-up); missing/
malformed/schema-drifted analysis.json (distinct errors, no write); existing
PROJECT.md (draft, no clobber); pre-existing PROJECT.draft.md (overwritable);
un-writable dir (no partial file); no packages (TODO stubs). Mirrored in JSON.

## Out-of-Scope

LLM-backed inference/refinement; monorepo per-module PROJECT files;
`centinela.toml` validate-command synthesis; auto-promoting draft → PROJECT.md.

## Handoff

→ senior-engineer: implement in plan order; `analyze.Load` first (contract
owner), then the inference engine (table-driven, pure), then the synthesizer
templates, then the command. Keep the package an aggregator (imports only
`internal/analyze` + stdlib); add `internal/synthesize/**` to centinela.toml's
aggregator layer and register it in PROJECT.md G2.
