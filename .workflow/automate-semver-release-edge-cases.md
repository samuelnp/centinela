### Edge-Case Report: automate-semver-release
**Date:** 2026-03-24

#### Risk Matrix
- **Case:** Non-conventional commit messages default to incorrect bump type
- **Impact:** Medium
- **Likelihood:** Medium
- **Why:** Commit parsing relies on regex and falls back to patch.

- **Case:** CI loop caused by release commit triggering workflow repeatedly
- **Impact:** High
- **Likelihood:** Low
- **Why:** Guard depends on `[skip ci]` and actor checks staying intact.

#### Missing or Weak Scenarios
- No tags exist yet (`v0.0.0` bootstrap behavior)
- Empty commit range after latest tag
- Feature and breaking-change precedence in mixed commit history

#### Proposed/Added Tests
- Unit: version-bump workflow file exists and is named correctly
- Integration: workflow includes branch trigger and conventional-commit parsing
- Acceptance: workflow commits, tags, and pushes release version

#### Residual Risks
- Tag push release artifact workflow is not yet covered in this step.
- Installer download/checksum behavior needs dedicated tests once implemented.
