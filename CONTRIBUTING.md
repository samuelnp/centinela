# Contributing to Centinela

Thanks for considering a contribution. Centinela is small, opinionated, and developed using its own workflow — that gives every change a paper trail and keeps quality predictable.

## How we develop Centinela on Centinela

Every change goes through the same 5-step workflow Centinela enforces on your projects:

```bash
centinela start <feature-name>          # required before any file write
# 1. plan      → write docs/features/<feature>.md, docs/plans/<feature>.md, specs/<feature>.feature
centinela complete <feature-name>
# 2. code      → implement the smallest correct change
centinela complete <feature-name>
# 3. tests     → add unit, integration, acceptance + edge-case coverage
centinela complete <feature-name>
# 4. validate  → run gatekeeper + centinela validate; all gates must pass
centinela complete <feature-name>
# 5. docs      → write the KB markdown and regenerate project docs
centinela complete <feature-name>
```

The orchestration validator rejects evidence JSONs that don't match the contract — read [`docs/architecture/evidence-contract.md`](docs/architecture/evidence-contract.md) before writing one.

## Build from source

**Prerequisites:** Go 1.21+

```bash
git clone https://github.com/samuelnp/centinela
cd centinela
go build -o centinela ./cmd/centinela/
```

Cross-compile for other platforms:

```bash
GOOS=linux   GOARCH=amd64 go build -o centinela-linux-amd64  ./cmd/centinela/
GOOS=darwin  GOARCH=arm64 go build -o centinela-darwin-arm64 ./cmd/centinela/
GOOS=windows GOARCH=amd64 go build -o centinela-windows.exe  ./cmd/centinela/
```

## Pull-request expectations

- **One feature per branch.** Branch name should match the feature slug used in `centinela start`.
- **Conventional commits.** `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`. The PR title follows the same convention.
- **All gates pass.** `centinela validate` must exit 0 — no skipping G1, no skipping the test suite.
- **Coverage stays ≥ 95%.** Adjust by raising coverage, not by lowering the threshold.
- **Files stay ≤ 100 lines** (or have an explicit, justified G1 exception in `centinela.toml`).
- **No business logic in the outer layer** (`cmd/centinela/` for this repo).
- **Strict typing.** No `interface{}` shortcuts in Go; no dynamic-typing escape hatches in any language for projects that adopt Centinela.

## What kinds of contributions are welcome

- **Bug fixes** to existing workflow enforcement, gates, or scaffold assets.
- **New archetypes** (e.g., serverless, event-driven) — open an issue first so we can align on layer rules.
- **New gates** that catch concrete, real-world quality problems.
- **Integration plugins** for additional AI coding agents (today: Claude Code + OpenCode).
- **Documentation improvements** — especially in the knowledge base under `docs/project-docs/kb/`.

## Filing an issue

Pick the relevant template from [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/). Bug reports without a reproduction are usually closed.

## Code of Conduct

This project follows the [Contributor Covenant v2.1](CODE_OF_CONDUCT.md). Be kind. Engage in good faith.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
