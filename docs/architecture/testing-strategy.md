# Testing Strategy

> This document describes Sauron's testing setup (TypeScript + Vitest + Cucumber). Adapt tools, file extensions, and paths to your project per PROJECT.md → Tech Stack and Folder Structure.

## Framework
- **Vitest** for unit and integration tests
- **MSW** (Mock Service Worker) for HTTP mocking in integration tests
- **Testing Library** for React hook tests
- **@cucumber/cucumber** for executable Gherkin acceptance tests

## Test Structure

```
specs/
  jira-connection.feature
  sprint-data-fetch.feature
  sprint-dashboard.feature
  epic-filter.feature
  pr-linking.feature
  ...
tests/
  unit/
    kernel/
      domain/
        entities/
          Epic.test.ts
          Story.test.ts
        value-objects/
          StoryStatus.test.ts
          StatusTransition.test.ts
        services/
          BlockerDetector.test.ts
          BounceCounter.test.ts
      application/
        use-cases/
          GetSprintEpics.test.ts
          DetectBlockers.test.ts
    hooks/
      useSprintEpics.test.ts
      useBlockerSummary.test.ts
  integration/
    infrastructure/
      jira/
        JiraEpicRepository.test.ts
        JiraStoryRepository.test.ts
      github/
        GitHubPullRequestRepository.test.ts
      persistence/
        PrismaSpaceRepository.test.ts
  acceptance/
    steps/
      jira-connection.steps.ts
      sprint-data-fetch.steps.ts
      sprint-dashboard.steps.ts
      ...
    support/
      world.ts                 # Cucumber World with test container
      hooks.ts                 # Before/After scenario setup
  fixtures/
    jira-responses/
      epic.json
      story-with-changelog.json
      sprint.json
    github-responses/
      pull-request.json
    domain/
      epic.factory.ts        # Factory functions for test entities
      story.factory.ts
```

## What to Test per Layer

### Domain (unit)
- Entity creation and validation
- Value object immutability and equality
- Domain service logic (blocker detection, bounce counting)
- Edge cases: missing data, boundary values

### Application (unit)
- Use case orchestration with mocked ports
- DTO mapping correctness
- Error handling (missing sprint, empty results)

### Hooks (unit)
- State management (loading → data → error flows)
- Correct use case invocation
- Data transformation for UI consumption

### Infrastructure (integration)
- API client request/response mapping (with MSW)
- Repository implementations return correct domain entities
- Mapper edge cases (null fields, unexpected formats)
- Cache behavior

## Test Naming Convention

```typescript
describe('GetSprintEpics', () => {
  it('should return epics with story counts for active sprint', () => {});
  it('should return empty array when sprint has no epics', () => {});
  it('should throw when no active sprint exists', () => {});
});
```

### Acceptance (BDD)
- Gherkin scenarios execute against the application layer
- Infrastructure is mocked (in-memory repositories, MSW for APIs)
- Each `.feature` file has a matching `.steps.ts` file
- Scenarios validate full use case flows end-to-end
- Use Cucumber World to share state between steps

```typescript
// tests/acceptance/support/world.ts
export class AppWorld {
  container: Container;  // DI container with mocked infra
  result: unknown;
  error: Error | null;
}
```

```typescript
// tests/acceptance/steps/sprint-data-fetch.steps.ts
Given('a project {string} with an active sprint', function(projectKey) {
  this.container = createMockContainer({
    sprintRepository: new InMemorySprintRepository([activeSprint]),
  });
});

When('I fetch the sprint epics', async function() {
  const useCase = this.container.getSprintEpics;
  this.result = await useCase.execute(this.projectKey);
});

Then('I should see {int} epics', function(count) {
  expect(this.result).toHaveLength(count);
});
```

## Running Tests

```bash
npm test                  # ALL tests (unit + integration + acceptance)
npm run test:unit         # Unit tests only
npm run test:integration  # Integration tests only
npm run test:acceptance   # Gherkin acceptance tests only
npm run test:watch        # Watch mode (unit + integration)
```

## Test Factories

Use factory functions to create test entities, avoiding fixture bloat:

```typescript
// tests/fixtures/domain/story.factory.ts
export function createStory(overrides?: Partial<StoryProps>): Story {
  return Story.create({
    key: 'US-101',
    summary: 'Test story',
    status: StoryStatus.DOING,
    priority: Priority.MEDIUM,
    ...overrides,
  });
}
```
