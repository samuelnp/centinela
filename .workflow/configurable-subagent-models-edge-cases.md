# Edge Cases: configurable-subagent-models

The following edge cases are guaranteed by the test suite (unit + integration + acceptance):

- **Tier normalization — casing**: `"Reasoning"` and `"REASONING"` are accepted by `NormalizeTier` and normalize to `"reasoning"`; the acceptance test drives the hook binary with a cased value and asserts the annotation reads `(model: reasoning)`.

- **Tier normalization — whitespace**: `" fast "` (leading/trailing spaces) is accepted and normalizes to `"fast"`; covered in unit and acceptance tests.

- **Empty vs absent `[orchestration.models]` table**: both produce nil from `OrchestrationModels` and allow all defaults to apply without error. Two unit tests cover `nil` error paths for both cases.

- **Unknown role key rejection**: a key like `"backend-wizard"` causes `validateOrchestrationModels` to return an error naming the key; covered by unit test in `internal/config` and the acceptance AC5 test.

- **Invalid tier after normalization**: `" Genius "` normalizes to `"genius"` which is still invalid; the config unit test asserts this is rejected, preventing silent acceptance of non-canonical values.

- **Missing tier→model mapping no-panic**: `ResolveModel` with `RunnerUnknown` (no entry in the per-runner table) returns the tier name + `ok=false` without panicking; covered in both the internal package test and the flat `tests/unit` resolve test.

- **Unknown runner emits tier name as both-ID fallback**: `RunnerUnknown` has no entry in `tierModels[tier]`; the resolver falls back to the tier name string with `ok=false`, so the caller can warn. Verified by `TestResolveModel_UnknownRunnerFallback`.

- **Out-of-band roles not annotated in directive**: `gatekeeper`, `production-readiness`, `edge-case-tester`, `merge-steward` do not appear in the plan/code/tests/validate/docs step role lists and must not appear with `(model:` annotations in the emitted directive; verified by `TestOrchestrationHook_OutOfBandRolesAbsent`.

- **Allow-list parity (config↔domain drift guard)**: every `orchestration.AllowedTiers()` value is accepted as a valid tier by `config.Load()`, and every `orchestration.AllowedRoleSlugs()` value is accepted as a valid role key; asserted in `tests/unit/configurable_subagent_models_config_unit_test.go` via real `config.Load()` calls in temp dirs.

- **`ModelReference` deduplication**: passing the same tier twice produces a single entry in the reference line, not two; the reference line is semicolon-delimited per tier, so one tier → zero semicolons.

- **`ModelReference` stable order**: tiers appear in `AllowedTiers()` order (reasoning, balanced, fast) regardless of the input order; verified by `TestModelReference_StableOrder`.
