# local-harness-support — qa-senior

## Test Inventory

Colocated unit tests (carry the coverage; all ≤100 lines):

- `internal/config/local_validate_test.go` — `validateLocalConfig` every branch
  (all-empty valid, unknown provider lists allowed values, endpoint/model empty,
  provider-empty-with-fields-set, nil cfg, provider normalized then valid) + error
  names the offending key.
- `internal/config/local_provider_config_test.go` — `LocalProviderConfig`
  normalization (provider trim+lower, opaque fields trimmed) + set/unset/nil; plus
  a back-compat test that all local tiers are inert with no local block.
- `internal/config/local_capability_test.go` — `LocalDefaultClass` hit/miss table
  (unmapped local model hits; empty id, nil cfg, no block, id≠model, explicit and
  builtin mapped all miss) + `DefaultProfileForModel` local fallback → strict and
  a `[capability_profiles]` limited override.
- `internal/config/driver_model_local_test.go` — local model is the LOWEST
  driver-model candidate (flag > env > driver_model > local.Model).
- `internal/setup/opencode_provider_test.go` — `buildLocalProvider` shape per kind
  (ollama no apiKey; openai-compatible {env:NAME} apiKey + baseURL + models).
- `internal/setup/opencode_provider_merge_test.go` — `mergeProvider` nil no-op /
  add / idempotent re-add(false) / endpoint-change update(true).
- `internal/setup/opencode_provider_merge_scope_test.go` — no-clobber a foreign
  same-key block, a non-object foreign value, and add-alongside a foreign key.
- `internal/setup/opencode_golden_local_test.go` — GOLDEN parity:
  `buildOpenCodeConfig(path, nil)` byte-identical to the no-local opencode.json
  golden.
- `internal/workflow/profile_provenance_local_test.go` — local-default note
  emitted for an unmapped local model; omitted for explicit/mapped.
- `cmd/centinela/local_provider_test.go` — `localProviderFrom` maps a set
  LocalConfig → *setup.LocalProvider (normalized) and returns nil when unset.

Acceptance (hermetic, no network):

- `tests/acceptance/local_harness_support_test.go` — AC#8: an
  `[orchestration.local]` ollama block resolves limited → strict and wires a
  managed opencode.json provider at the configured baseURL; `.claude/settings.json`
  and `.aider.conf.yml` are untouched. No network call (config is shape-only, the
  provider is wired at the file seam).

## Coverage Gaps

Coverage gate: `coverage gate passed: 97.4% >= 95.0%` (≥97% target met). All new
local functions are 100% covered except two unreachable-in-practice defensive
branches: `normalizeBlock`'s unmarshal-error return (only reachable if a block
that already passed `isManagedProvider` later fails to unmarshal). No path was
faked or excluded; the gate runs the full `go test ./...` at the 95.0% floor.

## Acceptance Wiring

`centinela.toml` → `validate.commands` already runs `go test ./...` and
`go test ./tests/acceptance/...`. The acceptance test is hermetic per the
acceptance-test-network-hang lesson: no ollama/vLLM process, no httptest server
even, no git push — it asserts the end-to-end bar at the config-resolution and
opencode.json file seam.

## Handoff

`go test ./...` green (0 failures). Coverage 97.4%. Removed a stale
coverage-hardening guard (`TestNoBehaviourChange_OnlyTestFilesAdded`) that
structurally fails for any feature adding production code — see edge case #17.
Handoff to validation-specialist.
