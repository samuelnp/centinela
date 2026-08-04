# Edge Cases: version-lock-profile-contract

## Covered

- The version lock itself is the executable check, and it is already red→green
  here: on the merged tree `TestWorkflowStructFieldsAreVersionLocked` failed
  naming `profileContract` as present in the struct but absent from the golden
  list; with the list updated it passes. No new test is written because the test
  that caught this IS the coverage — a second assertion over the same reflected
  field list would restate it, not strengthen it.
- The reflected list is derived from struct tags at run time, so it cannot drift
  from the real marshalled shape without failing. A future field added without a
  decision fails the same way.

## Residual Risks

- **The lock catches shape, not meaning.** Redefining what an existing key means
  while keeping its name and type passes silently. That is the case the comment
  names as owing a SchemaVersion bump, but nothing mechanical enforces it —
  it rests on the author reading the rule.
- **Version 1 now describes two shapes across time** (with and without
  `profileContract`). That is sound because the field is back-compat-by-absence,
  but it means "schemaVersion: 1" is not a byte-level shape identity — only a
  compatibility class. A reader expecting the former would be misled.
- **The conflict class recurs.** Any feature adding a `Workflow` field while
  another is in flight hits this again, and only a gate run on the MERGED tree
  reveals it — neither PR's own CI can. That is a property of parallel delivery,
  not of this fix.
