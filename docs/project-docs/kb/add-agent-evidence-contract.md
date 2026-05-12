---
feature: add-agent-evidence-contract
summary: Agent prompts now spell out the exact JSON evidence contract so subagents pass the orchestration validator on the first try.
audience: end-user
status: done
---

## What it does
Every agent prompt under `docs/architecture/` now embeds the JSON skeleton it must produce, plus a short checklist of the rules the orchestration validator will check. A single canonical reference, `docs/architecture/evidence-contract.md`, holds the full schema and per-role rules and is linked from each prompt.

## When you'd use it
You'll feel this every time a subagent runs a Centinela step. Before this change, agents commonly wrote prose summaries as `outputs`, skipped the `docs/features/*.md` snapshot inputs, or omitted required edge-case tags — forcing a human to rewrite the JSON before `centinela complete` would succeed. Now the rules are visible in the prompt itself, so the agent's first attempt is usually accepted.

## How it behaves
- The `docs/architecture/evidence-contract.md` document lists the full schema (feature, step, role, status, generatedAt, inputs, outputs, edgeCases, mobileFirst, handoffTo) and the validator's per-role rules.
- Each of the seven agent prompts (big-thinker, feature-specialist, senior-engineer, qa-senior, ux-ui-specialist, validation-specialist, documentation-generator) links to the contract and includes a role-specific JSON skeleton with realistic placeholder paths.
- Plan-step prompts call out that `inputs` must snapshot every `docs/features/*.md` in the repo.
- The qa-senior prompt names `.workflow/<feature>-edge-cases.md` and at least one `tests/` path as required outputs.
- The ux-ui-specialist prompt lists all eight required edge-case tags and the `mobileFirst: true` flag.
- The documentation-specialist prompt notes its exemption from the outputs-existence check.
- Scaffold mirrors under `internal/scaffold/assets/docs/architecture/` are byte-identical so new projects bootstrap with the same prompts; an acceptance test guards against drift.

## Examples
A new agent run can copy the skeleton straight from its prompt and substitute the feature name. For instance, the big-thinker prompt now shows:

    {
      "feature": "<FEATURE_NAME>",
      "step": "plan",
      "role": "big-thinker",
      "status": "done",
      "generatedAt": "<RFC 3339 timestamp>",
      "inputs": ["docs/features/<FEATURE_NAME>.md", "…every other docs/features/*.md…"],
      "outputs": ["docs/features/<FEATURE_NAME>.md", "docs/plans/<FEATURE_NAME>.md"],
      "edgeCases": ["Optional — risks or invariants you flagged"],
      "handoffTo": "feature-specialist"
    }
