# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Release notes for tagged versions are also published on the [GitHub Releases page](https://github.com/samuelnp/centinela/releases).

## [Unreleased]

### Added
- **G2 layer-boundary gate** (`import_graph`): mechanically enforces per-layer import rules by parsing the Go import graph via `go list -json` (no new dependency) and failing `centinela validate` on forbidden cross-layer imports. Configured via `[gates.import_graph]` with `name`/`paths`/`allow` layer declarations; unmapped packages warn rather than fail; default disabled.
- Per-feature plain-language knowledge base under `docs/project-docs/kb/`, generated alongside the main project docs.
- Canonical orchestration evidence contract at `docs/architecture/evidence-contract.md` and matching role-specific JSON skeletons in every agent prompt.
- Community profile files: `LICENSE`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `CHANGELOG.md`, plus `.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md`, `.github/dependabot.yml`.
- README badges row, 30-second tour, table of contents, and "When *not* to use Centinela" section.
- feat: compose the PR body and a `CHANGELOG` `[Unreleased]` entry automatically from a feature's delivery evidence (brief, plan, gatekeeper verdict) on `centinela deliver --via pr`
- feat: add `centinela dashboard` — a read-only board aggregating in-flight features (step, age, git-derived owner), roadmap burn-down, and gate health across worktrees, with `--json` output
- feat: extract a HarnessAdapter interface + registry (Claude/OpenCode refactored onto it with byte-for-byte parity) and add first-class Aider support via `--agent aider`
- feat(cost): add cost-governance soft gate — a Stop-hook attributes host-harness transcript token spend to the active feature/step, `centinela cost` reports spend vs configurable `[cost]` budgets, and `centinela validate` surfaces a non-blocking ⚠ when over budget (never blocks). Zero config = silent no-op.
- feat(mcp): expose governance as a versioned MCP server (`centinela.mcp/v1`). `centinela mcp serve` (official Go SDK, stdio) offers read_rules/run_gates/verify_claims/workflow_state reusing the verdict packet; it is advisory (allow/warn/block). `centinela mcp shim` maps a block verdict onto the harness pre-write deny (exit 2), so any MCP harness consumes governance with zero Centinela-specific code.
- feat: add `centinela update` self-updater — downloads the os/arch-matched release binary, verifies it against `SHA256SUMS`, and atomically replaces itself; `--check` is read-only and a TTL-cached, fail-silent startup notice flags new versions (never auto-installs)

### Changed
- Repository metadata: description, homepage, and 20 discoverability topics added on GitHub.

---

For pre-release history, refer to `git log` and the [Releases page](https://github.com/samuelnp/centinela/releases).
