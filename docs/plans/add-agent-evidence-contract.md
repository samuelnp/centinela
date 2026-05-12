# Plan: Make Agent Prompts Spell Out the Evidence Contract

1. Author `docs/architecture/evidence-contract.md` covering:
   - Full JSON schema (`feature`, `step`, `role`, `status`, `generatedAt`,
     `inputs`, `outputs`, `edgeCases`, `mobileFirst?`, `handoffTo`).
   - Global rules: `status="done"`, RFC 3339 timestamp, outputs must be
     real file paths on disk (except documentation-specialist), non-empty
     inputs/outputs/handoffTo.
   - Per-role rules in compact tables:
     * big-thinker / feature-specialist (plan): snapshot of every
       `docs/features/*.md`; output must include a plan or spec file;
       feature-specialist needs non-empty edgeCases.
     * senior-engineer (code): outputs must include at least one
       implementation file outside `.workflow/`, `tests/`, `docs/`, `specs/`.
     * qa-senior (tests): outputs must include at least one `tests/` file
       AND `.workflow/<feature>-edge-cases.md`; edgeCases non-empty.
     * ux-ui-specialist: `mobileFirst: true` AND edgeCases containing all
       eight required UX tags.
     * validation-specialist / documentation-specialist: basics only.
2. Update each of the seven prompts to:
   - link to `evidence-contract.md`,
   - embed a role-specific JSON skeleton populated with placeholder values,
   - include a short bullet list of the rules that apply to that role.
3. Mirror every prompt change to
   `internal/scaffold/assets/docs/architecture/<prompt>.md`.
4. Add `tests/acceptance/agent_evidence_contract_acceptance_test.go` that
   asserts each prompt contains the canonical schema markers plus the
   role-specific rule statements.
5. Add a paragraph to CLAUDE.md "Gatekeepers Checklist" pointing
   maintainers at `evidence-contract.md`.
