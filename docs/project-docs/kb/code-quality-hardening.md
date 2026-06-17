---
feature: code-quality-hardening
summary: Hardens Centinela's everyday surfaces — stable evidence files, a Go formatting gate in validate, and loud, specific errors when your config or workflow state is broken.
audience: end-user
status: done
---

## What it does

This change makes Centinela behave predictably when something is wrong on disk. Evidence files written by the hooks now keep a stable field order, so they no longer churn between runs. `centinela validate` gains a formatting check, so unformatted Go can no longer slip through. And when your `centinela.toml` config or a workflow's saved state is corrupt, missing, or unreadable, Centinela tells you exactly what and where — instead of failing silently or blaming the wrong thing.

## When you'd use it

You benefit from this any time you run Centinela day to day: starting a feature, letting the prompt hook inject context, or running `centinela validate` before a merge. It matters most on the bad days — a half-edited `centinela.toml`, a truncated state file, or a permissions glitch — where the old behavior would either swallow the problem or report a confusing one.

## How it behaves

- Evidence files keep a stable, canonical field order on every write — including the `coverage` field, which always sits between `mobileFirst` and `handoffTo` — so re-running the hooks never reshuffles your evidence JSON.
- If any Go file in the project isn't formatted, `centinela validate` now fails and prints the path of the offending file, so the fix is obvious.
- If every Go file is already formatted, the format check passes quietly and adds no noise to your validate run.
- The format check is wired into the `centinela validate` command list, so it runs automatically as part of validation without any extra step from you.
- Running `centinela start` with a `centinela.toml` Centinela can't parse now fails immediately with an error that names `centinela.toml` — and crucially, no half-created feature or workflow file is left behind.
- The prompt context hook never breaks your session over a bad config: it still exits cleanly so your work continues, and it injects a visible config warning naming the failure so you know something needs fixing.
- Loading a feature that has no saved workflow reports a clear "no workflow found" for that feature, so absence reads as absence.
- Loading a feature whose state file is corrupt reports the state file's path and the underlying parse failure, so you can find and fix the broken file.
- A state file that exists but can't be read is no longer mistaken for a missing feature — the error names the file's path instead of falsely claiming the workflow doesn't exist.

## Examples

`centinela validate` now flags formatting alongside its other checks:

```
$ centinela validate
...
$ ./scripts/check-fmt.sh
internal/ui/render_status.go
exit status 1
```

Clearer errors when something on disk is broken:

```
$ centinela start my-feature
error: failed to load centinela.toml: ...
# nothing was created — no workflow state file written

# missing vs. corrupt vs. unreadable are now distinct:
no workflow found for "my-feature"
failed to parse workflow state .workflow/my-feature.json: invalid character ...
failed to read workflow state .workflow/my-feature.json: permission denied
```
