# Example: Adding "Detect Blockers" Feature

> **Sauron-specific example.** Commands (`npm run`, `tsc`), file extensions (`.ts`, `.tsx`), paths (`src/kernel/`, `src/hooks/`), and tools (Vitest, Cucumber) are specific to Sauron's stack. Adapt to your project's language, framework, and folder structure.

A complete walkthrough of all workflow steps for a real Sauron feature.

## Step 1: plan (write the plan)

Create or update `docs/plans/blocker-detection.md`:

```markdown
# Blocker Detection
## Goal: Identify blocked subtasks and compute blocker stats for dashboard KPIs.
## Domain: BlockerInfo value object, BlockerDetector service
## Application: DetectBlockers use case, BlockerReportDto
## Infrastructure: none (uses existing Jira data)
## Hooks: useBlockerSummary
## UI: BlockerBadge, SummaryBar updates
```

```bash
centinela complete detect-blockers
# Step 'plan' completed. Next step: spec
```

## Step 2: spec (write Gherkin)

Create `specs/blocker-detection.feature`:

```gherkin
Feature: Blocker Detection
  As a squad lead
  I want to see which subtasks are blocked
  So I can unblock my team quickly

  Scenario: Subtask in BLOCKED status is detected
    Given a story "US-101" with subtask "SUB-1001"
    And "SUB-1001" has status "BLOCKED"
    When I detect blockers for the sprint
    Then "SUB-1001" should appear in the blocker report

  Scenario: Subtask stuck too long is flagged
    Given a story "US-102" with subtask "SUB-1002"
    And "SUB-1002" has been in "IN REVIEW" for 5 days
    And the threshold for "IN REVIEW" is 3 days
    When I detect blockers for the sprint
    Then "SUB-1002" should appear as a stale item

  Scenario: Summary stats are computed correctly
    Given a sprint with 3 blocked subtasks across 2 stories
    And total blocked days is 12
    When I get the blocker summary
    Then I should see 2 stories with blocks
    And I should see 3 blocked subtasks
    And I should see 12 total blocked days
```

```bash
centinela complete detect-blockers
# Step 'spec' completed. Next step: domain
```

## Step 3: domain (entities, value objects, ports)

Create `src/kernel/domain/value-objects/BlockerInfo.ts` (~20 lines):

```typescript
export type BlockerInfo = {
  readonly subtaskKey: string;
  readonly reason: string;
  readonly linkedIssueKey: string | null;
  readonly daysSince: number;
};
```

Create `src/kernel/domain/services/BlockerDetector.ts` (~40 lines):

```typescript
import type { Subtask } from '../entities/Subtask';
import type { BlockerInfo } from '../value-objects/BlockerInfo';
import { StoryStatus } from '../value-objects/StoryStatus';

export class BlockerDetector {
  constructor(
    private readonly staleThresholds: Record<string, number>
  ) {}

  detect(subtasks: readonly Subtask[]): BlockerInfo[] {
    // Check explicit BLOCKED status + stale items
  }
}
```

```bash
centinela status detect-blockers   # confirm step before writing files
# ... write the files ...
centinela complete detect-blockers
# Step 'domain' completed. Next step: application
```

## Step 4: application (use cases, DTOs)

Create `src/kernel/application/dtos/BlockerReportDto.ts` (~15 lines):

```typescript
export type BlockerReportDto = {
  readonly blockedSubtasks: BlockedSubtaskDto[];
  readonly totalBlockedDays: number;
  readonly affectedStoryCount: number;
  readonly affectedPeople: string[];
};
```

Create `src/kernel/application/use-cases/blockers/DetectBlockers.ts` (~35 lines):

```typescript
import type { SubtaskRepository } from '../../../domain/ports/repositories/SubtaskRepository';
import { BlockerDetector } from '../../../domain/services/BlockerDetector';
import type { BlockerReportDto } from '../../dtos/BlockerReportDto';

export class DetectBlockers {
  constructor(
    private readonly subtaskRepository: SubtaskRepository,
    private readonly blockerDetector: BlockerDetector
  ) {}

  async execute(sprintId: string): Promise<BlockerReportDto> {
    const subtasks = await this.subtaskRepository.getBySprintId(sprintId);
    const blockers = this.blockerDetector.detect(subtasks);
    return this.toDto(blockers);
  }

  private toDto(blockers: BlockerInfo[]): BlockerReportDto { /* ... */ }
}
```

```bash
centinela complete detect-blockers
# Step 'application' completed. Next step: infrastructure
```

## Step 5: infrastructure (skip — uses existing repos)

```bash
# No skip command in current flow; document rationale in plan and continue
# Step 'infrastructure' skipped. Next step: hooks
```

## Step 6: hooks

Create `src/hooks/useBlockerSummary.ts` (~30 lines):

```typescript
export function useBlockerSummary(projectKey: string) {
  const { detectBlockers } = useKernel();
  // state management, call use case, return formatted data
}
```

```bash
centinela complete detect-blockers
# Step 'hooks' completed. Next step: ui
```

## Step 7: ui

Create `src/ui/components/dashboard/BlockerBadge.tsx` (~20 lines):

```typescript
type Props = { count: number };
export function BlockerBadge({ count }: Props) {
  if (count === 0) return null;
  return <span className="...">{count} blocked</span>;
}
```

```bash
centinela complete detect-blockers
# Step 'ui' completed. Next step: tests
```

## Step 8: tests (unit + integration)

Create `tests/unit/kernel/domain/services/BlockerDetector.test.ts`:

```typescript
describe('BlockerDetector', () => {
  it('should detect subtask with BLOCKED status', () => { /* ... */ });
  it('should detect subtask stale beyond threshold', () => { /* ... */ });
  it('should not flag subtask within threshold', () => { /* ... */ });
});
```

Create `tests/unit/kernel/application/use-cases/DetectBlockers.test.ts`:

```typescript
describe('DetectBlockers', () => {
  it('should return blocker report with correct stats', () => { /* ... */ });
  it('should return empty report when no blockers', () => { /* ... */ });
});
```

```bash
npm run test:unit  # verify passing
centinela complete detect-blockers
# Step 'tests' completed. Next step: acceptance
```

## Step 9: acceptance (Gherkin step definitions)

Create `tests/acceptance/steps/blocker-detection.steps.ts`:

```typescript
import { Given, When, Then } from '@cucumber/cucumber';

Given('a story {string} with subtask {string}', function(storyKey, subtaskKey) {
  // Set up in-memory test data
});

When('I detect blockers for the sprint', async function() {
  const useCase = this.container.detectBlockers;
  this.result = await useCase.execute(this.sprintId);
});

Then('{string} should appear in the blocker report', function(subtaskKey) {
  expect(this.result.blockedSubtasks.map(b => b.key)).toContain(subtaskKey);
});
```

```bash
npm run test:acceptance  # verify passing
centinela complete detect-blockers
# Step 'acceptance' completed. Next step: gatekeeper
```

## Step 10: gatekeeper (invoke subagent)

Invoke Agent tool with Gatekeeper prompt (see gatekeeper-prompt.md).
Save report to `.workflow/detect-blockers-gatekeeper.md`.

Example output:
```
### Gatekeeper Report: detect-blockers
**Status:** SAFE
No conflicts with existing specs.
```

```bash
centinela complete detect-blockers
# Step 'gatekeeper' completed. Next step: validate
```

## Step 11: validate (run all tests)

```bash
npm test
# All tests pass: unit (12) + integration (0) + acceptance (3)
centinela complete detect-blockers
# WORKFLOW COMPLETE for 'detect-blockers'!
```

## Key Takeaways

- Every step produced a concrete artifact before advancing
- Infrastructure was explicitly skipped with a reason
- Tests were written AFTER implementation but BEFORE validation
- The gatekeeper ran even though it found no conflicts (still mandatory)
- Total files created: ~8, all under 100 lines, all in correct layers
