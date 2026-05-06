# Edge Cases: opencode-native-subagents

- Empty or missing `opencode.json`: Centinela creates `agent` config with all native specialist subagents.
- Existing custom OpenCode agents: user-defined agents remain present after merge.
- Existing Centinela agent overrides: user prompt/config for a role is preserved instead of overwritten.
- Existing build Task permissions: custom task permissions remain while Centinela subagents are allowed.
