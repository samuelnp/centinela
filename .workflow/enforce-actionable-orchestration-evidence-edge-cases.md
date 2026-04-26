# Edge Cases — enforce-actionable-orchestration-evidence

- `big-thinker` or `feature-specialist` evidence lists free-text summaries instead of repo-relative files.
- `senior-engineer` evidence points only to `.workflow/` files, docs artifacts, or tests instead of implementation files.
- `qa-senior` evidence lists edge cases but omits concrete test files or the required `.workflow/<feature>-edge-cases.md` report.
- `documentation-specialist` evidence remains unchanged and should not be blocked by the new actionable-output rule.
