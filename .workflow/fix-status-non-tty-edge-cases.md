# Edge Cases — fix-status-non-tty

- `centinela status <feature>` runs with stdin and stdout attached to pipes instead of a terminal.
- `centinela status-all` runs in the same non-interactive environment and must still render every workflow.
- Missing workflows should still return the existing not-found error instead of the TTY failure.
