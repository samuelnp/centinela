# Edge Cases: agent-performance-audit

- OpenCode config without existing agents: `validation-specialist` is generated with `mode: subagent`.
- OpenCode config with existing custom agents: custom entries remain preserved while `validation-specialist` is added when missing.
- Build Task permissions: `validation-specialist` is allowed without removing existing task permissions.
- Validate step orchestration: strict mode now requires validation-specialist evidence before completing validate.
