# Edge Cases: migrate-full-sync

## Covered

- Missing setup files produce `create` actions and are created on apply.
- Managed setup files with known legacy content upgrade to current managed format.
- Custom unmanaged `AGENTS.md` and plugin content are flagged as `manual-review`.
- Invalid `--agent` values are rejected for both full and setup migrate commands.
- Hook migration output remains silent outside Centinela project context.

## Residual Risks

- Manual-review detection is content-based and intentionally conservative.
- Future template shifts should keep legacy fingerprints to preserve auto-upgrade paths.
