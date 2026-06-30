# Edge Cases: brownfield-setup-detection

## Covered

- Empty `src/` is NOT a signal — a freshly scaffolded repo carrying an empty
  conventional source dir must read as greenfield (HasSource false). Covered by
  detect_test.go (`empty src dir is not a signal`) and the acceptance
  greenfield case.
- Makefile-only is a brownfield signal — a repo whose only manifest is a
  Makefile is detected as existing code. Covered by detect_test.go and
  acceptance `TestAccBrownfield_MakefileAndEnrichConfirm`.
- No subdir walk / depth guard — source that lives only in a deeply nested
  non-root directory (e.g. `deep/nested/app/main.go`), with no root-level
  manifest or root source dir, reads as greenfield. The detector is root-only
  by design (O(entries at root)). Covered by the acceptance `nested only` case.
- PROJECT.md present bypasses both directives — when PROJECT.md exists the hook
  never emits a setup directive (brownfield or greenfield) and proceeds to the
  roadmap checks. Covered by `TestAccBrownfield_ProjectMdBypassesSetup`.
- Greenfield unchanged — a truly empty initialized repo still emits the
  existing six-question setup directive verbatim (no behavior change). Covered
  by `TestRunHookSetupGreenfieldUnchanged` and the acceptance greenfield cases.
- Hook early-return guard — `runHookSetup` returns nil unless one of
  PROJECT.md.template / PROJECT.md / centinela.toml exists, so detection only
  runs in an initialized repo. Tests seed `centinela.toml` to reach the branch.
- dirHasEntry robustness — a non-existent path and a plain file (not a dir) are
  both non-signals. Covered by detect_test.go `TestDirHasEntry`.
- Enrich-then-confirm wording — the brownfield directive instructs
  analyze → synthesize → ENRICH → set `**Project Stage:** existing` → confirm,
  and must NOT carry the greenfield "Do not answer the user's message" wording.

## Residual Risks

- Manifest breadth is limited to the existing `manifestTable` set (go.mod,
  package.json, Cargo.toml, Gemfile, pyproject.toml, requirements.txt,
  Makefile). Unlisted manifests (pom.xml, composer.json, build.gradle) degrade
  to greenfield — no regression, recorded as deferred `brownfield-manifest-breadth`.
- A repo whose source sits only in a deeply nested non-root dir with no root
  manifest reads as greenfield. Accepted by spec scenario 10 (cost guard); the
  user can still run `centinela analyze` manually.
- `dirHasEntry`'s `os.Open` error branch (a dir that stats as a directory but
  cannot be opened, e.g. a permissions race) is not exercised — it degrades to
  "not a signal", the safe default. Total coverage stays above the 95% gate.
