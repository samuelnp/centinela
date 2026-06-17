---
feature: security-gate
summary: An opt-in gate that runs during `centinela validate` to catch leaked secrets and vulnerable dependencies mechanically — secret findings block validation, dependency warnings surface without blocking.
audience: end-user
status: done
---

## What it does

The security gate is an opt-in check that runs as part of `centinela validate`. When enabled, it scans your working tree for leaked secrets using gitleaks — and any secret it finds **blocks** validate so the leak can never slip through to a merge. At the same time, it audits your project's dependencies for known vulnerabilities using govulncheck and osv-scanner; those findings show up as **warnings** but do not block validate, because a fix is often unavailable, transitive, or needs human judgement. If a scanner isn't installed on the machine, that check is simply **skipped with a clear message** — never a crash, and never a silent false "all clear".

## When you'd use it

Turn this gate on when you want committed secrets and vulnerable dependencies caught automatically at validate time, rather than hoping a human or review step notices them. It is the mechanical safety net for the two failure modes that are easy to miss in review: an API key or `.env` value accidentally committed, and a dependency that quietly carries a known CVE.

## How it behaves

- When a detectable secret is present in a scanned file, the `G-Secrets` result is `Fail`, the details name the file and the matched rule, and `centinela validate` blocks (exits non-zero).
- When no secrets are found, the `G-Secrets` result is `Pass` and the secrets check does not affect the exit code.
- When a dependency with a known vulnerability is found, the `G-Vuln` result is `Warn` and names the affected package and vulnerability ID, but validate still passes — the warning surfaces the risk without blocking unrelated work.
- When a scanner (for example gitleaks) is not installed, its result is `Skip` with a message naming the missing tool — validate does not crash and is never failed just because a tool is absent.
- When the gate is off (the default — `enabled = false` or no `[gates.security]` block at all), no security results are produced and existing projects behave exactly as before.
- When a secret match is listed in your `secrets.allowlist` (by rule ID or path glob), that match is excluded; if it was the only finding, `G-Secrets` reports `Pass` — so known false positives don't push you to disable the gate.
- When you run validate locally, the secret scan is scoped to your changed files (the diff-aware behavior, matching the other file-scoped gates); in CI it full-scans the whole tree, so a secret hiding in an unchanged file is still caught there.

## Examples

Enable the gate in `centinela.toml`:

```toml
[gates.security]
enabled = true

# Secret scanning (hard-fail). Optional allowlist for known false positives.
[gates.security.secrets]
tool = "gitleaks"            # v1: gitleaks
allowlist = []               # gitleaks rule IDs or path globs to ignore

# Dependency vulnerability audit (warn-only in v1).
[gates.security.vuln]
tools = ["govulncheck", "osv-scanner"]   # run those present; Skip+warn if absent
```

Sample output when a secret is committed:

```
G-Secrets  FAIL  Secret detected:
  · internal/config/loader.go  rule: generic-api-key
```

Sample output when a vulnerable dependency is found (validate still passes):

```
G-Vuln  WARN  Known vulnerabilities in dependencies:
  · golang.org/x/net  GO-2024-1234
```

Sample output when a scanner is not installed:

```
G-Secrets  SKIP  gitleaks not found on PATH — secret scanning skipped.
```
