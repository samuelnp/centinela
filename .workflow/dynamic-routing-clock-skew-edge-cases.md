# Edge Cases: dynamic-routing-clock-skew

## Covered

- An evidence artifact stamped a few milliseconds BEFORE `StartedAt` — the coarse
  kernel clock case that failed CI — counts as written during this run
  (`TestRoleEvidenceFromThisRun_ToleratesCoarseMtimeSkew`).
- A stub predating the run by 30 minutes (an earlier run of the same slug, or a
  rewind that never cleaned up) still does NOT close the routing window
  (`TestRoleEvidenceFromThisRun_StillRejectsAnEarlierRunsStub`) — the tolerance
  must not swallow what the check exists for.
- Both sides of the grace boundary: 1s inside counts, 1s outside does not
  (`TestRoleEvidenceFromThisRun_GraceWindowBoundary`).
- The skew is forced with `os.Chtimes` rather than inferred from the host clock,
  so these tests fail on macOS when the fix is reverted — verified by reverting
  it and observing 2 of 3 go red, then restoring.
- Pre-existing granularity-dependent cases (`...CurrentStepNeedsEvidenceOnDisk`,
  `...JSONStubFromThisRunCounts`) keep passing and now do so on both platforms.

## Residual Risks

- **A clock jump larger than the 2s grace.** An NTP step backwards of more than
  two seconds between `start` and the first evidence write would still misjudge
  the stub. Mitigation: the failure is fail-open on the routing window only (a
  downgrade stays permitted slightly longer); floors, the reason requirement, and
  the machine gates are all unaffected. Not worth clock-source machinery.
- **A filesystem with second-or-coarser mtime granularity** (e.g. some network
  mounts) is comfortably inside the grace window, but a filesystem that rounds
  mtimes UP would not be helped by a backward-only tolerance. No such filesystem
  is in the supported set.
- **The tolerance is one-directional by design.** An artifact stamped in the
  future still counts as "from this run", which is the conservative reading — it
  closes the routing window rather than opening it.
