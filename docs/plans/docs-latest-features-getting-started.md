# Plan: Latest Features Docs Refresh

1. Update `README.md` to add a concise latest-features section and a getting-started
   workflow tutorial that reflects the current CLI and integration behavior.
2. Fix stale README references so command lists, hook behavior, and workflow
   examples match the current five-step process and migration surface.
3. Update the HTML doc generator to render presentation-oriented sections for
   latest features and workflow onboarding instead of relying on stale raw source
   excerpts.
4. Regenerate `docs/project-docs/index.html` and verify docs commands plus the Go
   test suite still pass.
