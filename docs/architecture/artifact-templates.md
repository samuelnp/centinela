<!-- centinela:doc-version=1 template=docs/architecture/artifact-templates.md -->
# Artifact Templates

Use these templates when Centinela asks for missing setup or workflow artifacts.

## Setup Artifacts

### `ROADMAP.md`

```md
# Roadmap

## Phase 0: Bootstrap

- project-bootstrap

## Phase 1

- feature-slug
```

### `.workflow/roadmap.json`

```json
{
  "phases": [
    {
      "name": "Phase 0: Bootstrap",
      "features": [{ "name": "project-bootstrap" }]
    },
    {
      "name": "Phase 1",
      "features": [{ "name": "feature-slug" }]
    }
  ]
}
```

### `.workflow/roadmap-analysis.md`

```md
# Roadmap Analysis

- Role: senior-product-manager
- Feature: project-bootstrap
- Dependencies: none
```

### `.workflow/roadmap-analysis.json`

```json
{
  "role": "senior-product-manager",
  "features": [
    { "name": "project-bootstrap", "dependsOn": [] },
    { "name": "feature-slug", "dependsOn": ["project-bootstrap"] }
  ]
}
```

### `.workflow/roadmap-quality.md`

```md
# Roadmap Quality Evaluation

- Role: roadmap-quality-evaluator
- Threshold: 9
- Feature: feature-slug
- Summary: Ready after refinement.
```

### `.workflow/roadmap-quality.json`

```json
{
  "role": "roadmap-quality-evaluator",
  "threshold": 9,
  "features": [
    {
      "name": "feature-slug",
      "scores": {
        "acceptanceCriteria": 9,
        "userValue": 9,
        "definitionClarity": 9,
        "dependencies": 9,
        "effortEstimation": 9,
        "overall": 9
      },
      "summary": "Ready to build."
    }
  ]
}
```

## Per-Feature Artifacts

### `docs/features/<feature>.md`

Sections: `## Problem`, `## User Stories`, `## Acceptance Criteria`, `## Edge Cases`, `## Risks`, `## Decomposition`.

### `docs/plans/<feature>.md`

Ordered implementation steps for the feature.

### `specs/<feature>.feature`

Gherkin scenarios matching the user-visible behavior.

### `.workflow/<feature>.json`

Workflow state file created by `centinela start <feature>`.

### `.workflow/<feature>-<role>.md`

Roles: `big-thinker`, `feature-specialist`, `senior-engineer`, `ux-ui-specialist`, `qa-senior`, `documentation-specialist`.

```md
# Orchestration Evidence: <role>

- Feature: `<feature>`
- Step: `<plan|code|tests|docs>`
- Outcome: Short summary.
- Handoff: `<next-role>`
```

### `.workflow/<feature>-<role>.json`

```json
{
  "feature": "feature-slug",
  "step": "code",
  "role": "ux-ui-specialist",
  "status": "done",
  "generatedAt": "2026-04-24T00:00:00Z",
  "inputs": ["docs/features/feature-slug.md"],
  "outputs": ["src/ui/page.tsx"],
  "edgeCases": ["mobile-first", "visual-hierarchy", "typography-hierarchy", "responsive-layout", "loading-state", "empty-state", "error-state", "motion-and-reduced-motion"],
  "mobileFirst": true,
  "handoffTo": "qa-senior"
}
```

For strict orchestration roles, `outputs` must be project-relative file paths that already
exist on disk. Free-text summaries are not valid outputs.

- `big-thinker` and `feature-specialist`: include a real `docs/plans/...` or `specs/...` file.
- `senior-engineer`: include at least one real non-evidence implementation file.
- `ux-ui-specialist`: required only for features whose brief includes `surface: user-facing`; include at least one real UI file under configured `ui_paths`, set `mobileFirst: true`, and include these tags in `edgeCases`: `mobile-first`, `visual-hierarchy`, `typography-hierarchy`, `responsive-layout`, `loading-state`, `empty-state`, `error-state`, `motion-and-reduced-motion`.
- `qa-senior`: include at least one real `tests/...` file and `.workflow/<feature>-edge-cases.md`.

### Other workflow outputs

- `.workflow/<feature>-edge-cases.md`
- `.workflow/<feature>-gatekeeper.md`
- `.workflow/<feature>-production-readiness.md` when enabled
- `docs/project-docs/index.html` after the docs step

### Plan advisor behavior

Plan advisor mode does not create a new orchestration evidence role. It changes prompt behavior
during the `plan` step and orchestrates between `big-thinker` and `feature-specialist`.

- `workflow.plan_advisor_mode = "missing_info"` (default): ask only the missing high-value questions.
- `workflow.plan_advisor_mode = "always"`: always ask an advisor round during `plan`.
- `workflow.plan_advisor_mode = "off"`: disable advisor prompting.
- `workflow.plan_question_limit` defaults to `4` and caps each advisor round.
- Advisor context reads the current feature brief, plan, spec, and local edge-case report when present.
- If roadmap analysis exists, dependencies are considered before same-phase siblings.
- Related edge-case lessons and roadmap quality notes may influence questions through compact summaries rather than raw artifact dumps.
