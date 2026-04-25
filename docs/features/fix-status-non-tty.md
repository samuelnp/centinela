---
feature: fix-status-non-tty
type: fix
---

# Fix: `centinela status` Without a TTY

`centinela status <feature>` currently fails in non-interactive environments because it
always launches a Bubble Tea program that tries to open `/dev/tty`.
This fix keeps the interactive view for real terminals and falls back to static output
when stdin or stdout is not a TTY.
