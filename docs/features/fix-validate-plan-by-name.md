---
feature: fix-validate-plan-by-name
type: fix
---

# Fix: validatePlan Checks Content Instead of Filename

`validatePlan` scans every file in `docs/plans/` looking for the feature slug in the file
content. This fails when the plan file is correctly named `<feature>.md` but doesn't happen
to contain the slug as a literal string.

The fix checks for the presence of `docs/plans/<feature>.md` by filename instead.
