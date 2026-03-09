# i18n Strategy

> This document describes Sauron's i18n setup (next-intl + Next.js). Adapt the library, file paths, and routing strategy to your project per PROJECT.md → Tech Stack and Locales.

## Library
- **next-intl** — integrates natively with Next.js App Router

## Supported Locales

See PROJECT.md → Locales for the authoritative list of locale codes for this project.

## File Structure

```
src/infrastructure/i18n/
  messages/
    en.json
    es.json
  config.ts           # next-intl configuration
```

## Key Naming Convention

Hierarchical, dot-separated, grouped by feature:

```json
{
  "dashboard": {
    "title": "Sprint {sprintNumber} - Blocker Map",
    "subtitle": "Epics > User Stories > Subtasks",
    "summary": {
      "userStories": "User Stories",
      "storiesWithBlocks": "Stories with Blocks",
      "blockedSubtasks": "Blocked Subtasks",
      "totalBlockedDays": "Blocked Days (Total)",
      "affectedPeople": "Affected People"
    }
  },
  "filters": {
    "all": "All",
    "onlyBlocked": "Only blocked"
  },
  "status": {
    "todo": "To Do",
    "doing": "Doing",
    "inReview": "In Review",
    "pendingQa": "Pending QA",
    "inQa": "In QA",
    "pendingPrd": "Pending PRD",
    "readyForRelease": "Ready for Release",
    "done": "Done",
    "blocked": "Blocked"
  }
}
```

## Rules

1. **No hardcoded strings** in `ui/` components. Every visible text uses a translation key.
2. **Hooks layer** resolves translations via `useI18n` hook and passes strings down.
3. **Both locales** must be updated together. Gate keeper G10 validates completeness.
4. **Interpolation** for dynamic values: `"title": "Sprint {sprintNumber}"`.
5. **Plurals** use ICU format: `"{count, plural, one {# story} other {# stories}}"`.

## Integration with Hooks

```typescript
// src/hooks/useI18n.ts
// Wraps next-intl's useTranslations for our specific namespace needs
export function useDashboardI18n() {
  const t = useTranslations('dashboard');
  return { t };
}
```

## Routing

Locale is detected from browser preference or URL prefix:
- `/space/PROJ` → default locale (en)
- `/es/space/PROJ` → Spanish

Using next-intl's middleware for locale detection and routing.
