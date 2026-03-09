# Hexagonal Architecture Guide

> This document describes the **hexagonal archetype** as implemented in Sauron. It is one of five supported architecture archetypes. See [architecture-overview.md](architecture-overview.md) to understand which archetype fits your project before reading this. Adapt all file paths to your project per PROJECT.md → Folder Structure.

## Why Hexagonal?

Sauron depends on external APIs (Jira, GitHub) that are:
- Rate-limited
- Slow to respond
- Subject to change

Hexagonal architecture isolates our business logic from these externalities,
making the app testable, maintainable, and adaptable.

## Layer Details

### Domain (`src/kernel/domain/`)

The core. Zero dependencies on frameworks, APIs, or libraries.

```
domain/
  entities/          # Core business objects
    Epic.ts
    Story.ts
    Subtask.ts
    Sprint.ts
    Person.ts
  value-objects/     # Immutable descriptors
    StoryStatus.ts
    Priority.ts
    Discipline.ts    # QA, BE, FE, UX
    StatusTransition.ts
    BlockerInfo.ts
    TimeInStatus.ts
  ports/             # Interfaces for external access
    repositories/
      EpicRepository.ts
      StoryRepository.ts
      SprintRepository.ts
      SubtaskRepository.ts
      PersonRepository.ts
    services/
      IssueTrackerClient.ts
      VersionControlClient.ts
      CacheStore.ts
  services/          # Pure domain logic
    BlockerDetector.ts
    BounceCounter.ts
    StatusGroupClassifier.ts
```

**Rules:**
- Entities are created via factory methods or constructors, never raw objects.
- Value objects are immutable (all properties `readonly`).
- Ports are interfaces only — no implementations here.
- Domain services contain logic that doesn't belong to a single entity.

### Application (`src/kernel/application/`)

Orchestrates domain objects to fulfill use cases. One class per use case.

```
application/
  use-cases/
    sprints/
      GetActiveSprint.ts
      GetSprintEpics.ts
    epics/
      GetEpicWithStories.ts
    stories/
      GetStoryDetail.ts
      GetStoryTimeline.ts
    subtasks/
      GetSubtaskHistory.ts
    blockers/
      DetectBlockers.ts
      GetBlockerSummary.ts
    spaces/
      ListSpaces.ts
      ToggleFavoriteSpace.ts
    pull-requests/
      GetPendingPullRequests.ts
      GetStoryPullRequests.ts
  dtos/
    SprintSummaryDto.ts
    EpicSummaryDto.ts
    StoryCardDto.ts
    SubtaskProgressDto.ts
    BlockerReportDto.ts
    PullRequestDto.ts
  services/
    WorkflowAnalyzer.ts    # Computes time-in-status, bounce detection
```

**Rules:**
- Use cases receive ports via constructor injection.
- Use cases return DTOs, never domain entities directly.
- Each use case has exactly ONE public `execute()` method.
- Use cases must not call other use cases — extract shared logic to domain services.

### Infrastructure (`src/infrastructure/`)

Concrete implementations of domain ports + framework-specific code.

```
infrastructure/
  jira/
    JiraApiClient.ts         # Low-level HTTP client for Jira REST API
    JiraEpicRepository.ts    # Implements EpicRepository port
    JiraStoryRepository.ts
    JiraSprintRepository.ts
    JiraSubtaskRepository.ts
    mappers/
      JiraIssueMapper.ts     # Maps Jira API response → domain entities
      JiraChangelogMapper.ts
  github/
    GitHubApiClient.ts
    GitHubPullRequestRepository.ts
    mappers/
      GitHubPrMapper.ts
  persistence/
    prisma/
      schema.prisma
      PrismaSpaceRepository.ts
      PrismaCacheStore.ts
  i18n/
    messages/
      en.json
      es.json
    config.ts
  di/
    container.ts             # Dependency injection wiring
```

**Rules:**
- Infrastructure NEVER exports domain types — only implements ports.
- All external API responses are mapped to domain entities via dedicated mappers.
- Mappers handle null/undefined/missing fields defensively.
- DI container is the ONLY place where ports are bound to implementations.

### Hooks (`src/hooks/`)

React hooks that bridge UI and application layer.

```
hooks/
  useSprintEpics.ts
  useStoryDetail.ts
  useSubtaskHistory.ts
  useBlockerSummary.ts
  usePendingPrs.ts
  useSpaces.ts
  useFavoriteSpaces.ts
  useEpicFilter.ts
  useI18n.ts
```

**Rules:**
- Hooks call use cases from application layer (injected via context/provider).
- Hooks manage React state (loading, error, data).
- Hooks contain UI logic (filtering, sorting, formatting for display).
- Hooks NEVER import from infrastructure directly.

### UI (`src/ui/`)

Pure presentational React components. Zero business logic.

```
ui/
  components/
    dashboard/
      SummaryBar.tsx
      EpicSection.tsx
      StoryCard.tsx
      SubtaskRow.tsx
      WorkflowStepper.tsx
      BlockerBadge.tsx
    common/
      Avatar.tsx
      Badge.tsx
      PillFilter.tsx
      StatusBadge.tsx
    layout/
      Navbar.tsx
      SpaceSwitcher.tsx
    pull-requests/
      PrCard.tsx
      PrList.tsx
  providers/
    KernelProvider.tsx       # Provides DI container to React tree
    I18nProvider.tsx
```

**Rules:**
- Components receive ALL data and actions via props or hooks.
- No `fetch`, no API calls, no direct use case instantiation.
- Components are testable with just props — no context needed for unit tests.
- Prefer composition over prop drilling.

### App (`app/`)

Next.js App Router. Thin routing layer only.

```
app/
  layout.tsx                 # Root layout with providers
  page.tsx                   # → renders SpaceSelector from ui/
  [locale]/
    space/
      [projectKey]/
        page.tsx             # → renders SprintDashboard from ui/
        epic/
          [epicKey]/
            page.tsx         # → renders EpicDrilldown from ui/
        story/
          [storyKey]/
            page.tsx         # → renders StoryDetail from ui/
        prs/
          page.tsx           # → renders PrReview from ui/
  api/
    jira/
      route.ts               # API routes (proxy to avoid CORS)
    github/
      route.ts
```

**Rules:**
- Pages import ONE component from `ui/` and render it.
- Server components handle data fetching at page level if needed.
- API routes are thin proxies — logic lives in application layer.
