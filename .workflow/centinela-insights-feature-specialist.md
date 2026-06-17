# Feature-Specialist Report — centinela-insights

## Spec
`specs/centinela-insights.feature` — 36 scenarios, 1:1 acceptance-traceable: per-metric ranking + --top + tie-breaking + empty-field buckets; mean steps-to-green worked examples (1.00/2.00/n/a); empty/whitespace/missing log; malformed-line resilience; --json shape stability (incl. empty); determinism (byte-identical reruns); non-TTY no-ANSI; single-type-only logs degrade per section; span/eventCount metadata.

## Edge cases added to brief
Single event of each type; three-way tie ordering; --top N > available buckets; only-step-advanced log; only-block log; step-advanced excluded from rework; span + eventCount in human output.

Handoff → senior-engineer.
