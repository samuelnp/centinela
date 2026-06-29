# Edge Cases: host-harness-adapters

## EC-1: Unknown agent name — no panic, typed error with harness list

If `Lookup("vscode")` or any unregistered name is called, the registry returns a
wrapped `ErrUnknownAgent` sentinel. The error message includes all registered
harness names so callers (CLI, tests) can surface an actionable message.
No panic occurs. Covered by `TestLookup_UnknownAgent_TypedError`.

## EC-2: Pre-existing unmanaged `.aider.conf.yml` is never clobbered

If a project already has a hand-written `.aider.conf.yml` without a
`centinela:managed-version=` header, `planManagedFile` classifies it as
`SyncManualReview` and `ApplySync` skips it. The file is left intact and a
warning is surfaced to the user. Covered by
`TestHostHarnessAC5_UnmanagedAiderConfigNotClobbered`.

## EC-3: AGENTS.md shared surface — no double-write between OpenCode and Aider

Both OpenCode and Aider adapters call `planAgentsFile`, which returns `nil`
(no-op) when the file already matches the target. Running both adapters
sequentially (or applying "opencode" then "aider") writes AGENTS.md exactly
once. The managed region appears once only. Verified in
`TestHostHarnessAider_IdempotentReApply` (aider re-apply is a no-op).

## EC-4: Aider scope isolation — Claude and OpenCode files untouched

`BuildSyncPlan("aider")` iterates only the `aiderAdapter`, which calls
`planAgentsFile` and `planAiderConfig`. It never touches `.claude/settings.json`
or `opencode.json`. Covered by `TestHostHarnessAider_DoesNotTouchClaudeFiles`
and `TestHostHarnessAC5_AiderInitWritesFiles`.

## EC-5: Aider idempotency — second apply is a no-op

After a successful `ApplySync`, the managed files match their target content
exactly. A subsequent `BuildSyncPlan("aider")` returns an empty plan (all
`planManagedFile` calls return `nil`). No file is written a second time.
Covered by `TestHostHarnessAC5_AiderInitIdempotent` and
`TestHostHarnessAider_IdempotentReApply`.

## EC-6: blocks-writes capability requires a prewrite hook item

Any adapter that declares `CapBlocksWrites` must emit at least one
`SyncKindPrewriteHook` item from `PlanItems()`. An adapter without a hook
item that claims `blocks-writes` fails `TestCapabilityParity_BlocksWritesRequiresPrewriteHook`
and `TestHostHarnessAC7_BlocksWritesRequiresPrewriteHook`. Aider explicitly
declares no `blocks-writes` and emits no prewrite hook item (verified by
`TestAiderAdapter_NoPrewriteHook` and `TestHostHarnessAC7_AiderNoPrewriteHook`).

## EC-7: "both" selector uses registry composition, not a hardcoded branch

The `adaptersFor("both")` path resolves through the `composites` map to
`["claude", "opencode"]` — no hardcoded `if agent == "both"` branch in
`BuildSyncPlan`. Adding a new composite entry requires only editing the
`composites` map. Covered structurally by
`TestHostHarnessScopeBoth_UnionOfClaudeAndOpenCode` and
`TestHostHarnessAC2_BothComposesClaudeOpenCode`.

## EC-8: `AgentsFor` unknown selector returns ErrUnknownAgent

`AgentsFor("unknown")` delegates to `adaptersFor`, which calls `Lookup` and
returns the wrapped error. Covered by `TestAgentsFor_Unknown`.
