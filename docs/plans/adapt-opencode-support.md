# Plan: Adapt Centinela for OpenCode

## Scope
Implement OpenCode support while preserving existing Claude behavior and keeping workflow engine unchanged.

## Work Items
1. Add init target selection
- Extend `centinela init` with `--agent` (`claude`, `opencode`, `both`).
- Keep default as `both` for smooth onboarding.

2. Add OpenCode setup artifacts
- Write or merge `opencode.json` with project-safe defaults.
- Scaffold `.opencode/plugins/centinela.js` plugin for enforcement hooks.
- Keep behavior non-destructive when files already exist.

3. Extract shared enforcement logic
- Move step/file classification and block/allow decisions into reusable internal package.
- Reuse this logic in Claude hook handlers and OpenCode plugin contract.

4. Wire OpenCode behavior
- Enforce prewrite checks for edit/write/patch-like operations.
- Emit workflow progress/context hints in OpenCode prompt lifecycle.
- Reuse setup checks (`PROJECT.md`, `ROADMAP.md`, production readiness prompt setup).

5. Update docs and scaffold templates
- Document OpenCode support in `README.md`.
- Keep `CLAUDE.md` plus OpenCode guidance in scaffold assets.

6. Add tests
- Unit tests for init target parsing and config merge behavior.
- Integration tests validating OpenCode artifact generation.
- Acceptance-level parity checks for block/allow behavior.

## Validation
- Run `go test ./...`.
- Run `go run ./cmd/centinela init --agent opencode` in a temp fixture and verify created files.

## Rollout
- Incremental: first setup and shared logic extraction, then docs/tests cleanup.
- Backward compatibility gate: existing Claude-only projects continue working unchanged.
