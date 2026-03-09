# Dependency Injection Strategy

> Examples in this document are Sauron-specific. Adapt types and names to your project per PROJECT.md → Domain Language and Layer Mapping.

## Overview

We use a simple DI container pattern — no heavy frameworks.
The container is created once at app startup and provided via React Context.

## Container Structure

```typescript
// src/infrastructure/di/container.ts
interface Container {
  // Repositories (ports)
  epicRepository: EpicRepository;
  storyRepository: StoryRepository;
  sprintRepository: SprintRepository;
  subtaskRepository: SubtaskRepository;
  pullRequestRepository: PullRequestRepository;
  spaceRepository: SpaceRepository;

  // Use cases
  getActiveSprint: GetActiveSprint;
  getSprintEpics: GetSprintEpics;
  getEpicWithStories: GetEpicWithStories;
  getStoryDetail: GetStoryDetail;
  detectBlockers: DetectBlockers;
  getPendingPullRequests: GetPendingPullRequests;
  listSpaces: ListSpaces;
  toggleFavoriteSpace: ToggleFavoriteSpace;
}
```

## Wiring

```typescript
export function createContainer(config: AppConfig): Container {
  // 1. Create infrastructure (adapters)
  const jiraClient = new JiraApiClient(config.jira);
  const githubClient = new GitHubApiClient(config.github);

  // 2. Create repositories (bind ports to adapters)
  const epicRepository = new JiraEpicRepository(jiraClient);
  const storyRepository = new JiraStoryRepository(jiraClient);
  // ...

  // 3. Create use cases (inject repositories)
  const getSprintEpics = new GetSprintEpics(epicRepository, sprintRepository);
  // ...

  return { epicRepository, storyRepository, getSprintEpics, /* ... */ };
}
```

## React Integration

```typescript
// src/ui/providers/KernelProvider.tsx
const KernelContext = createContext<Container | null>(null);

export function KernelProvider({ children, container }) {
  return (
    <KernelContext.Provider value={container}>
      {children}
    </KernelContext.Provider>
  );
}

export function useKernel(): Container {
  const ctx = useContext(KernelContext);
  if (!ctx) throw new Error('KernelProvider not found');
  return ctx;
}
```

## In Hooks

```typescript
// src/hooks/useSprintEpics.ts
export function useSprintEpics(projectKey: string) {
  const { getSprintEpics } = useKernel();
  // ... call use case, manage state
}
```

## Testing

For tests, create a container with mock implementations:

```typescript
const mockContainer = createMockContainer({
  epicRepository: new InMemoryEpicRepository(testEpics),
});
```
