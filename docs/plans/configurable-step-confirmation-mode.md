# Plan: Configurable Step Confirmation Mode

1. Extend workflow config with `step_confirmation_mode` and safe defaults.
2. Add mode parsing/normalization helpers in config package.
3. Apply mode in `hook context` review prompt decisions.
4. Add tests for config parsing/defaults and prompt behavior by mode.
5. Update workflow docs and scaffold templates with mode options.
