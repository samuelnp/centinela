<!-- centinela:managed-version=1 template=AGENTS.md -->
# Centinela Rules

This project uses Centinela workflow enforcement.

## Mandatory
- Read and follow CLAUDE.md for framework rules.
- Read PROJECT.md before planning or coding.
- Bootstrap before features: if PROJECT.md is missing, do not suggest centinela start <feature>, and do not ask what to work on.
- When PROJECT.md is missing, ask setup questions and write PROJECT.md before any feature or roadmap work.
- If roadmap setup is required, define the roadmap before asking for feature work.
- Start every feature with centinela start <feature> only after project setup and roadmap bootstrap are complete.
- Do not bypass workflow order: plan -> code -> tests -> validate -> docs.

## OpenCode Integration
- Centinela prewrite checks are enforced by .opencode/plugins/centinela.js.
- Treat Centinela setup and migration directives as higher priority than casual chat.
- If setup or roadmap is required, do not reply to greetings first; start the required setup flow immediately.
- On a greeting-only first prompt, first state any required Centinela setup, migration, or workflow step before normal conversation.
- If a write is blocked, fix step alignment instead of forcing the write.

## Commands
- centinela start <feature>
- centinela complete <feature>
- centinela status <feature>
- centinela validate
- centinela docs validate
- centinela docs generate --out docs/project-docs/index.html
