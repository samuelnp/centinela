# Edge Cases: add-agent-evidence-contract

## Covered

- Contract document enumerates all 10 schema fields including the optional `mobileFirst`.
- Test asserts each of the seven role prompts (including documentation-generator) references the contract AND embeds a `"role": "<role>"` JSON skeleton line.
- Plan-step prompts (big-thinker, feature-specialist) explicitly state the `docs/features/*.md` snapshot requirement.
- Senior-engineer prompt requires an implementation-file output outside `.workflow/`, `tests/`, `docs/`, `specs/`.
- QA-senior prompt requires both a `tests/` path AND `.workflow/<feature>-edge-cases.md`.
- UX-UI prompt requires `mobileFirst: true` AND all eight UX edge-case tags.
- Documentation-specialist prompt notes its exemption from the outputs-existence check explicitly so the agent does not over-engineer.
- Scaffold mirrors are byte-identical for every updated prompt — drift fails the acceptance suite.
- Existing `promote-orchestration-agents` per-file line budget raised from 70 to 130 (spec + test) to fit the expanded prompts.

## Residual Risks

- Agents may still skip the contract link if they bypass the `## Required Artifact` section; mitigated by embedding the JSON skeleton directly under that heading.
- The eight UX tags must match string-for-string (case-insensitive); typos in agent output still fail validation. The prompt lists the exact tags.
- `generatedAt` correctness is not enforced by the prompt beyond a placeholder; the validator catches malformed timestamps.
