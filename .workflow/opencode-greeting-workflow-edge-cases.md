# Edge Cases: opencode-greeting-workflow

- Greeting-only first prompt with missing `PROJECT.md`: generated OpenCode instructions require the agent to mention Centinela setup before casual conversation.
- Greeting-only first prompt with roadmap or migration required: the same Centinela-first rule applies before normal conversation.
- Setup already complete: the rule is conditional on required setup, migration, or workflow guidance and should not force noise for unrelated conversation.
- Custom unmanaged OpenCode files: existing sync behavior still requires manual review instead of overwriting user changes.
