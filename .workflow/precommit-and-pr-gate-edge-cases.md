# Edge Cases: precommit-and-pr-gate

## Covered

- **Staged fail-gate violation** (oversized staged `.go`) → `centinela precommit`
  exits non-zero, names G1 (commit blocked).
- **Only clean staged** → exit 0.
- **Unstaged working-tree changes ignored** — an unstaged oversized file is not
  in `ChangedFilesStaged()` and does not block (integration test).
- **Not a git repo / nothing staged** → `Summary.Degrade` / empty set → precommit
  exits 0 cleanly, no crash/stack trace.
- **Build/cross-compile gate skipped** under precommit by default (`skip_build`
  default true via `RawSkipBuild *bool`; explicit `false` honored).
- **Warn-severity gate** under precommit → reported, non-blocking.
- **Installer**: writes executable (`0o755`) `.git/hooks/pre-commit` with the
  marker block; idempotent (re-install changed=false); preserves a pre-existing
  user hook; `MkdirAll` missing dir; surfaces write errors.
- **Uninstall**: removes only the marker block, preserves user lines; deletes a
  centinela-only hook; missing-file no-op.
- **`splice`/`removeBlock`**: append/replace-in-place/idempotent/marked-region-only.
- **`gitdiff.ChangedFilesStaged`**: success Set (`Base=="STAGED"`), empty index,
  git-error degrade (nil set, nil err).
- **`RenderGatesMarkdown`**: `<!-- centinela:pr-gate -->` marker, fail `<details>`,
  pass/warn headers, byte-deterministic, Details capped with overflow note.
- **`pr-gate`**: Markdown to stdout + exit (Fail → non-zero; `fail_on_warning`
  makes Warn fail, default does not); degrade → full-scan notice; outside PR
  context still renders, posts nothing.
- **Config**: Normalize/validate for both sections; TOML decode of explicit
  `skip_build=false` stays false.
- **Determinism**: identical staged content → identical verdict + exit.

## Residual Risks

- Aggregate coverage sits near the 95% gate; pure packages (githooks,
  render_markdown, staged) are at/near 100% to hold margin. If unrelated code
  drifts the total down, push those error branches higher — never lower the gate.
- The Go binary is network-free; PR-comment posting lives in CI (`gh pr
  comment`), so the GitHub round-trip is not unit-tested here (out of scope —
  verified by the yaml step, exercised in real PR CI).
- Windows shell/git-hook paths are `t.Skip`-ped; the POSIX path is fully covered.
