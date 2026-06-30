# local-harness-support — documentation-specialist

## KB Pages

Regenerated `docs/project-docs/index.html` (56 KB) via `centinela docs generate`; the
per-feature knowledge base under `docs/project-docs/kb/` is regenerated alongside it from
the committed `docs/features/` and `docs/plans/` artifacts. No new KB page was hand-authored.

## project-docs Entries

Extended the user-facing config reference in `README.md` (section "3. Configure
centinela.toml") with a new subsection **"Point Centinela at a local model"** documenting
the `[orchestration.local]` block:

- The four fields — `provider` (`ollama` | `openai-compatible`), `endpoint`, `model`,
  optional `api_key_env` — in a table with required/optional and meaning.
- Two TOML examples: an Ollama block (`http://localhost:11434/v1`, no API key) and a generic
  `openai-compatible` block (llama.cpp / vLLM / LM Studio) using `api_key_env`.
- The managed-provider wiring behavior: `init`/`migrate` add a managed `opencode.json`
  provider block (npm `@ai-sdk/openai-compatible`, `options.baseURL` from the endpoint,
  `options.apiKey = "{env:NAME}"` for openai-compatible with `api_key_env`, model under
  `models`); owns only its own key (no clobber); idempotent; zero-config output unchanged.
- The capability default: a declared local `model` with no explicit class defaults to
  `limited` → `strict` as the strictly-lowest precedence tier; explicit `--profile` and
  global `enforcement_profile` still win; `centinela status` shows
  `Profile  strict  (local default: <model> → limited → strict)`.
- The opaque-strings note: Centinela validates only shape; `endpoint`/`model`/`api_key_env`
  are never connected-to, verified, or resolved — availability is the runner's job.

The README was chosen over editing the plan-step feature brief so plan-step evidence
snapshots stay valid; it is the genuine user-facing config home (it already documents
`[orchestration]`, `[workflow]`, and `[validate]` blocks).

## Outcome

Changelog one-liner written to `.workflow/local-harness-support-changelog.md` (Added,
Keep-a-Changelog feat line). Project docs HTML regenerated at
`docs/project-docs/index.html`. README config reference extended. Documentation accurately
reflects the shipped code in `internal/config/orchestration_local.go` and
`internal/setup/opencode_provider.go` (provider key = provider name; ollama gets no apiKey;
only openai-compatible with `api_key_env` writes `{env:NAME}`).
