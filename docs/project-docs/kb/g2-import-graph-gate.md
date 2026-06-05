---
feature: g2-import-graph-gate
summary: Centinela automatically checks that your Go project's packages respect the architectural layering rules you've defined and fails validation the moment a forbidden dependency sneaks in.
audience: end-user
status: done
---

## What it does

The import-graph gate adds a new check that runs during `centinela validate`. You declare your project's layer boundaries in `centinela.toml` — for example, that a low-level `config` package may never reach up into `ui` code — and Centinela reads the actual Go source to see whether any package has broken those rules. If a forbidden dependency is found, the gate reports `Fail` and names the exact offending pair, so you know which package is importing what it shouldn't. When all dependencies are clean, the gate reports `Pass` with no noise.

The gate is opt-in: if you haven't added a `[gates.import_graph]` block to your config, it simply doesn't run. When you do enable it, it always loads the entire module rather than only the files you just changed, because a dependency you added elsewhere can break the graph even if you didn't touch the violating file.

## When you'd use it

Enable this gate once you've settled on a layer structure for your project and want those boundaries enforced automatically — rather than relying on code review to catch a forbidden cross-layer import after the fact. It is especially useful in teams or AI-agent workflows where many changes land quickly and architectural invariants are easy to miss in review.

## How it behaves

- When every package in your module imports only packages that its layer is allowed to use, the gate reports `Pass` and `centinela validate` exits 0.
- When a package imports a package in a forbidden layer, the gate reports `Fail` and lists each offending dependency on its own line, for example: `internal/config -> internal/ui (config may not import ui)`. Multiple violations are all listed. `centinela validate` exits non-zero.
- When a package doesn't match any layer you've defined in the config, the gate reports `Warn` and names the unmapped package — the layer matrix must stay exhaustive so nothing slips through silently.
- When no `[gates.import_graph]` block is present, or the block has `enabled = false`, the gate is omitted entirely from the report and validate proceeds without it.
- When the config block has a problem — such as an empty `paths` list for a layer, or an allow-list that references a layer name you haven't defined — the gate reports `Fail` with a message starting with `import_graph config:`, distinct from an import violation.
- When the block is present and enabled but defines zero layers, the gate reports `Warn` rather than silently passing, so an empty matrix can't masquerade as a clean check.
- When the module contains code that cannot be compiled, the gate reports `Fail` with the load error from the Go toolchain rather than returning a false `Pass`.
- Standard-library and third-party packages (`fmt`, `golang.org/x/tools/…`, etc.) are never evaluated against the layer matrix — only packages inside your own module are checked.
- Test files (`_test.go`) are in scope: a test package that imports across a forbidden boundary is flagged just like production code, using the same layer as the package it belongs to.
- Imports between packages in the same layer are always allowed; only cross-layer edges are checked against the allow list.
- If a violation exists in a file outside your current git diff, the gate still reports it — the whole-module load means no edge is hidden by what you happened to change today.

## Examples

Add a `[gates.import_graph]` block to your `centinela.toml` and declare each layer with the package-path globs it covers and the other layers it is allowed to depend on:

```toml
[gates.import_graph]
enabled = true

[[gates.import_graph.layers]]
name  = "config"
paths = ["internal/config/**"]
allow = []                        # leaf layer — may not import anything internal

[[gates.import_graph.layers]]
name  = "domain"
paths = ["internal/workflow/**", "internal/gates/**"]
allow = ["config"]                # domain may import config, nothing else

[[gates.import_graph.layers]]
name  = "cmd"
paths = ["cmd/**"]
allow = ["domain", "config"]      # entry point may import domain and config
```

Sample output when a violation is detected:

```
import_graph  FAIL  1 forbidden import(s):
  · internal/config -> internal/ui (config may not import ui)
```
