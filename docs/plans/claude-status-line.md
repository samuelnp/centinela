# Plan: Claude status line for Centinela

## Scope
Implement a Centinela-owned status line command and wire it into Claude settings so users see compact workflow context continuously.

## Work Items
1. Add status line hook command
- Introduce `centinela hook statusline` command in `cmd/centinela`.
- Build compact token output from active workflow state.

2. Add UI rendering helper
- Create a small renderer in `internal/ui` for deterministic token lines.
- Include no-workflow fallback tokens.

3. Derive status fields
- Feature, step, progress (`n/total`), next action.
- Blocker code from artifact checks for current step.
- Risk status from validate-step warning helpers.

4. Wire setup into Claude settings
- Extend hook injection to add `statusLine` command when missing.
- Keep idempotent behavior and preserve unrelated settings keys.

5. Add tests
- Unit tests for renderer and field derivation.
- Integration tests for settings injection and command output behavior.

## Validation
- Run targeted tests for `cmd/centinela`, `internal/ui`, and `internal/setup`.
- Run `go test ./...`.

## Rollout
- Keep backward compatibility: hooks remain unchanged.
- Status line is additive and safe if disabled by user config.
