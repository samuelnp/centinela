# fix-setup-hook-template-detection

## Problem

`centinela hook setup` only runs setup/roadmap guidance when
`PROJECT.md.template` exists, so projects that rename the template to
`PROJECT.md` lose roadmap prompting.

## User Stories

- As a user, I still get roadmap guidance after creating `PROJECT.md`.
- As a user, hook directives remain clear even with boxed UI output.

## Acceptance Criteria

- Setup hook treats project as initialized when either `PROJECT.md.template` or
  `PROJECT.md` exists.
- If `PROJECT.md` exists and `ROADMAP.md` is missing, roadmap guidance is shown.
- Hook output includes a plain directive line before boxed guidance.

## Edge Cases

- Non-centinela repo with neither file returns no output.
- Template missing + project missing still returns setup guidance in centinela
  repos with existing scaffolding.
