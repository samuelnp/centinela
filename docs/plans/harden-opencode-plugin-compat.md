# Plan: Harden OpenCode Plugin Compatibility

## Scope
Improve generated plugin resilience for tool/event payload variations.

## Work Items
1. Expand file-path extraction to additional common keys.
2. Make write-tool detection tolerant to casing and unknown wrappers.
3. Keep prompt append robust for `output.prompt`, `output.context`, or missing output object.
4. Add tests asserting generated plugin contains compatibility helpers.

## Validation
- `go test ./...`
- `centinela validate`
