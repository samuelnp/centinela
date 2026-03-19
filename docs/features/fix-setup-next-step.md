---
feature: fix-setup-next-step
type: fix
---

# Fix: PROJECT.md Setup Suggests Wrong Next Step

After Claude writes `PROJECT.md`, it incorrectly tells the user to run `centinela start <feature>`.
The correct next step is to define the roadmap. This fix updates the closing instruction in
`RenderSetupNeeded()` to guide the user toward the roadmap conversation instead.
