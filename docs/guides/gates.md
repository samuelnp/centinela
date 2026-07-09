# Gate Checks

> The quality checks that must pass before a feature can ship.

Gates are quality checks that must pass before a feature can ship. They run during `centinela validate` and automatically when completing the `validate` step. Every gate toggle and its options are listed in the [configuration reference](configuration-reference.md#gates).

## Claim verification at the validate step

When `centinela complete <feature>` advances through the `validate` step it runs claim verification as a HARD block. Any failing claim — tests that do not actually pass, a coverage figure that exceeds the measured result beyond tolerance, or an output file that contains only an empty stub — stops completion and names the failing claim. Edge-case mapping failures emit a warning and surface in the output but do not hard-block on their own.

Run verification at any time with:

```bash
centinela verify <feature>
```

See the [`[verify]` configuration block](configuration-reference.md#verify) to adjust the timeout and coverage tolerance.

## Built-in gates

| Gate | Rule | Config |
|------|------|--------|
| **G1: File Size** | Default max 100 lines, with optional justified exceptions up to 130 lines | `[gates] file_size = true` |
| **G2: Layer Boundaries** | Parses the Go import graph and fails on any import that violates your declared per-layer allow-matrix | `[gates.import_graph]` |
| **G11: i18n** | All locale files have identical keys (no missing translations) | `[gates] i18n = true` |
| **G-Build: Cross-Compile** | Cross-compiles every configured release target and fails naming the broken `GOOS/GOARCH` | `[gates.build] enabled = true` |

Centinela ships additional opt-in gates — security scanning, spec-traceability, roadmap-drift, audit-baseline ratchet, and fully custom command gates. See the [configuration reference](configuration-reference.md#gates) for the full list and every key.

### Cross-compile build gate

The `G-Build: Cross-Compile` gate runs each target in your `[gates.build] targets` list through the configured build command, sets `GOOS`, `GOARCH`, and `CGO_ENABLED=0` automatically, and collects any failures. If any target fails, the gate reports `Fail` with a detail line per broken platform. Default is **disabled**.

```toml
[gates.build]
enabled = true
command = "go build ./cmd/myapp"   # executed once per target; no shell expansion
targets = [
  { goos = "linux",   goarch = "amd64" },
  { goos = "linux",   goarch = "arm64" },
  { goos = "darwin",  goarch = "amd64" },
  { goos = "darwin",  goarch = "arm64" },
  { goos = "windows", goarch = "amd64" },
  { goos = "windows", goarch = "arm64" },
]
```

The `command` is argv-parsed (`strings.Fields`) and executed directly — never via a shell — so spaces in paths are safe and shell injection is not possible.

### Layer-boundary (import-graph) gate

The `G2: Layer Boundaries` gate turns your architecture's layer rules into a mechanical check. It parses the Go import graph with `go list -json` (no extra dependency) and fails `centinela validate` when a package imports another package its layer is not allowed to import — for example a leaf `config` package reaching up into `ui`. You declare the matrix in `centinela.toml`: each layer has a name, path globs, and the list of layers it may import. Standard-library and third-party imports are ignored; a package that matches no configured layer produces a non-failing **warning** (so the matrix can be adopted incrementally) rather than passing silently. Default is **disabled**.

```toml
[gates.import_graph]
enabled = true
# module = "github.com/you/yourapp"   # optional; defaults to `go list -m`

[[gates.import_graph.layers]]
name  = "leaf"
paths = ["internal/config/**"]
allow = []                            # a leaf layer imports no other layer

[[gates.import_graph.layers]]
name  = "domain"
paths = ["internal/service/**"]
allow = ["leaf"]

[[gates.import_graph.layers]]
name  = "cmd"
paths = ["cmd/**"]
allow = ["domain", "leaf"]
```

A forbidden edge is reported as `internal/config -> internal/ui (leaf may not import domain)`. Same-layer and self imports are always allowed.

A companion parity test (`TestBuildMatrixParity`) keeps `[gates.build] targets` in `centinela.toml` in sync with the release matrix in `.github/workflows/release.yml`. If either list drifts, `go test ./...` fails during `centinela validate` and names the missing targets.

### i18n gate formats

G11 supports two formats natively — use **one** `[i18n]` block:

```toml
# JSON locale files (next-intl, i18next, vue-i18n)
[i18n]
format  = "json"
dir     = "src/i18n/messages"
locales = ["en", "es", "fr"]
```

```toml
# GNU gettext .po files (Godot, Qt)
[i18n]
format  = "gettext"
dir     = "i18n"
locales = ["en", "es"]
```

For other formats (Unity CSV, Android XML, iOS `.lproj`), set `format = "none"` and add a custom command to `[validate] commands`.

## Diff-aware mode

`centinela validate` can scope the file-walking gates (G1, G11) to files changed on the current branch, so the report flags only violations introduced by your work — not pre-existing ones in untouched files.

Default behavior (`diff_mode = "auto"`):

- **Locally** (no `CI` env var): diff-aware. Header reads `Built-in Gates (diff-aware: N files changed since main)`.
- **In CI** (`CI=true` or `CI=1`): full scan. Header reads `Built-in Gates (full scan)`. The ship gate stays strict.

Configure via `centinela.toml`:

```toml
[validate]
diff_mode = "auto"   # "auto" | "always" | "off"
diff_base = "main"   # any git ref (e.g. "master", "develop")
```

Override per invocation:

```bash
centinela validate --changed   # force diff-aware
centinela validate --full      # force full scan
```

Flags beat config, config beats CI detection. `--changed` and `--full` are mutually exclusive.

How the change set is built:

- `git diff --name-only --diff-filter=ACMR $(git merge-base HEAD <diff_base>)` for tracked changes.
- `git ls-files --others --exclude-standard` for untracked files (new code is gated before `git add`).
- Renamed files appear via the new path. Deleted files are naturally skipped.

G1 walks only files in the change set. G11 runs the full key-completeness comparison when any locale file is in the change set, and short-circuits with a "no locale changes" Pass otherwise (partial-locale comparison is not meaningful).

User `[validate] commands` are **not** scoped by the diff — they always run in full.

Degrade paths: non-git directory, missing diff base, shallow clone, or any git failure prints a one-line `notice:` and falls back to full scan.

CI systems that don't set `CI=true` (uncommon — GitHub Actions, GitLab CI, CircleCI, Travis, Buildkite, and Drone all do) need either `diff_mode = "off"` or `--full` in the pipeline.

## Manual gates (code review)

| Gate | Rule |
|------|------|
| **G2: Layer Dependencies** | No imports cross forbidden layer boundaries (archetype-specific) |
| **G3: Type Safety** | Strictest static analysis — no `any`, no untyped variables |
| **G5: Spec First** | Every feature has a `.feature` file before implementation starts |
| **G6: Plan First** | Every feature has a plan document before implementation starts |
| **G7: No Business Logic in Outer Layer** | UI components and adapters contain no domain logic |
| **G8: Single Responsibility** | Each file exports one thing and does one thing |

Full gate documentation: [`docs/architecture/gatekeepers.md`](../architecture/gatekeepers.md)

---

← Back to the [documentation index](README.md) · [Configuration reference](configuration-reference.md)
