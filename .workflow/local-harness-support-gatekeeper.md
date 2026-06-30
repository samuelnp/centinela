<!-- centinela:doc-version=1 template=docs/architecture/gatekeeper-prompt.md -->
### Gatekeeper Report: local-harness-support
**Date:** 2026-06-30
**Status:** SAFE

#### Analyzed Specs
Scanned all specs at the PROJECT.md → Gatekeeper Paths surface. Specs whose
domains intersect this feature's shared surfaces were reviewed line-by-line:

- `specs/local-harness-support.feature` (the new spec — 22 scenarios, AC#1–AC#8 + edge cases)
- `specs/model-capability-profiles.feature` (capability class → default profile precedence)
- `specs/enforcement-profiles.feature` (explicit `--profile` / global `enforcement_profile` precedence)
- `specs/configurable-model-routing.feature` + `specs/configurable-subagent-models.feature` (OrchestrationConfig shape, driver_model)
- `specs/host-harness-adapters.feature` (BuildSyncPlan / adapter registry / golden parity)
- `specs/adapt-opencode-support.feature`, `specs/opencode-setup-priority.feature`, `specs/harden-opencode-plugin-compat.feature`, `specs/opencode-hook-parity.feature` (managed opencode.json output)
- `specs/coverage-hardening.feature` (the removed self-referential guard test)
- `specs/g2-import-graph-gate.feature`, `specs/spec-traceability-gate.feature`, `specs/g1-justified-file-size-exceptions.feature` (the gates run here)
- `specs/claude-status-line.feature` (status provenance rendering)

The remaining specs were swept for references to the changed symbols
(`InjectOpenCodeConfig`, `BuildSyncPlan`, `OrchestrationConfig`, `SyncItem`,
`DefaultProfileForModel`, `DriverModelFrom`) — none consume them in a way the
changes break.

#### Findings

All seven adversarial checks plus the removed-test assessment came back clean.
No BLOCKING or WARNING-level conflicts. Detail per check:

**1 — OpenCode config signature changes (SAFE).**
`InjectOpenCodeConfig`, `planOpenCodeConfig`, and `buildOpenCodeConfig` each
gained a `*LocalProvider` param. Every caller across `cmd/` (init_agent.go,
migrate.go, migrate_setup.go, hook_migrate.go) and every test
(internal/setup, tests/unit, tests/integration, tests/acceptance) was updated —
production call sites pass `localProviderFrom(cfg)`; pre-existing tests pass
`nil`. `BuildSyncPlan(agent)` delegates to `BuildSyncPlanWithLocal(agent, nil)`
(sync.go:8-9). The regression tripwire `TestBuildOpenCodeConfigNilLocalGoldenParity`
(opencode_golden_local_test.go) asserts `buildOpenCodeConfig(path, nil)` is
byte-for-byte identical to the committed no-local golden. `mergeProvider(raw, nil)`
returns false immediately, so the zero-config opencode/Claude/Aider output is
unchanged. `go build`, `go vet`, and the full suite confirm every caller compiles
and behaves.

**2 — DefaultProfileForModel / capability precedence (SAFE).**
The back-compat ladder is preserved. `DefaultProfileForModel` still calls
`CapabilityClassFor` first (explicit `[orchestration.capabilities]` override →
builtin map). Only on a class MISS does it consult `LocalDefaultClass`, which
returns `("", false)` unless the id is non-empty, `cfg != nil`, the id EQUALS the
configured `[orchestration.local].model`, AND the id has no
explicit/builtin class. Consequences:
  - Existing cloud/Anthropic models (claude-*, anthropic/*) resolve via the
    builtin/override map exactly as before — the new tier is never reached.
  - Zero-config (no local block) → `LocalProviderConfig` returns `ok=false` →
    `LocalDefaultClass` returns false → `DefaultProfileForModel` returns
    `("", false)`, identical to pre-feature; the caller does not engage the tier.
  - An explicitly mapped local model (`capabilities["qwen2.5-coder"]="capable"`)
    short-circuits at `CapabilityClassFor` and never hits the local default
    (spec scenario "An explicitly mapped local model id wins").
The local default is strictly the lowest tier; explicit `--profile` and global
`enforcement_profile` still win — confirmed by `ProfileProvenance` (tiers 1→4)
and by the AC#4 spec scenarios. Verified by driver_model_local_test,
local_capability_test, and profile_provenance_local_test (all green).

**3 — OrchestrationConfig.Local field (SAFE).**
The new `Local LocalConfig` (`toml:"local"`) is an additive optional block.
Configs without `[orchestration.local]` unmarshal it to the zero value, which
`LocalProviderConfig` reports as `ok=false`. No existing OrchestrationConfig
reader (model tiers, overrides, model_map, ui_paths, capabilities) touches it.
Orchestration parity / routing tests pass.

**4 — SyncItem.Local field (SAFE).**
`Local *LocalProvider` is `nil` for every non-opencode item and for the
zero-config opencode path (set only in planOpenCodeConfig when a local block is
present). Its sole consumer is `applyItem` → `InjectOpenCodeConfig(it.Path,
it.Local)`. No other SyncItem consumer (golden parity, sync_branches, hooks)
reads it; nil flows through harmlessly.

**5 — Layer rules / import graph (SAFE).**
`go list -deps ./internal/config` and `./internal/setup` show NEITHER package
imports any other internal package — the new files (orchestration_local.go,
local_capability.go, local_validate.go, opencode_provider.go,
opencode_provider_merge.go) import only stdlib. The config→setup mapping lives in
`cmd/centinela/local_provider.go` (the outer layer), keeping internal/setup
"imports nothing internal". The `import_graph` gate result is **Warn with
`unmapped`-packages-only, zero forbidden edges** — i.e. no cross-layer violation
was introduced (the Warn is the pre-existing advisory for packages not assigned
to a layer in the matrix; non-failing).

**6 — i18n / status provenance string (SAFE).**
The new note `"local default: <id> -> limited -> strict"` (profile_provenance.go:32)
is a read-only status *diagnostic* annotation rendered muted in
`profileLine`. It is emitted exactly like its sibling provenance literals in the
same function — `"--profile"`, `"global"`, `"driver: <id> -> <class>"`,
`"default"` — which are likewise plain English. These developer-facing status
provenance notes are NOT part of the localized user-facing string surface, so the
new note is consistent with the established convention and introduces no missing
i18n key. The i18n gate is unaffected.

**7 — G1 file size (SAFE).**
Every changed/new `.go` file is <=100 lines (capability.go is exactly 100). The
G1 gate reports "All files under 100 lines"; G-Build cross-compile passes all 6
release targets.

**Removed test `TestNoBehaviourChange_OnlyTestFilesAdded` — justified, not a weakening (SAFE).**
The deleted guard asserted `git diff --diff-filter=A main...HEAD` adds no
production `.go` files. That invariant only held while coverage-hardening (a
deliberately test-only feature) was itself HEAD; it structurally fails for ANY
later feature that legitimately adds production code, and local-harness-support
is the first such feature. It is redundant with the live coverage gate
(`./scripts/check-coverage.sh`, which passes) that enforces real coverage of new
code on the merged tree. Removal eliminates a permanent self-referential landmine
rather than dropping a real behavioural invariant; the sibling tests in the same
file (TestDeferredPaths_InRoadmapBacklog etc.) are retained, and a comment records
the rationale in place.

#### Deferred Findings
- none

#### Recommendation
- SAFE: No conflicts detected. The feature is purely additive behind an
  all-or-nothing `[orchestration.local]` block; the zero-config path is proven
  byte-identical (golden parity tripwire) and the new capability tier is strictly
  lowest-precedence and gated on a declared local model. Proceed.

##### Verification evidence observed
- `go build ./...` — success; `go vet ./...` — no issues.
- `gofmt -l internal/ cmd/ tests/` — empty (no diffs).
- `go test ./... -count=1` — 3127 passed in 43 packages, exit 0, no FAIL/panic.
- `centinela validate --changed` (diff-aware: 52 files) — exit 0, "All gates passed":
  - PASS G1 File Size · PASS G-Build Cross-Compile · PASS roadmap_drift
  - WARN import_graph (unmapped packages only — zero forbidden edges; non-failing)
  - WARN spec-traceability-gate (warn-severity advisory; non-failing)
  - Validate Commands: PASS go test ./... · PASS go test ./tests/acceptance/... ·
    PASS ./scripts/check-coverage.sh · PASS ./scripts/check-fmt.sh
