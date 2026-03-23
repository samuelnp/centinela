# Feature Brief: Add CI Validate Workflow

## Problem
Coverage and validation are enforced locally, but there is no repository CI workflow to enforce `centinela validate` on pull requests.

## Goal
Add a GitHub Actions workflow that runs `centinela validate` (and therefore coverage gate + built-in gates) on pushes and PRs.

## Acceptance Criteria
- CI workflow file exists under `.github/workflows/`.
- Workflow runs on push and pull_request.
- Workflow installs Go, runs tests, and runs `go run ./cmd/centinela validate`.
- Failing validate command fails the job.
