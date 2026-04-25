---
feature: fix-status-non-tty
---

# Plan: Fix `centinela status` Without a TTY

## Problem

`cmd/centinela/status_model.go` always runs Bubble Tea through `tea.NewProgram(...).Run()`.
In non-interactive shells this opens `/dev/tty` and fails before rendering status output.

## Change

Add a non-interactive path for the status command:

```go
// before:
func runStatusModel(wfs []*workflow.Workflow) error {
	p := tea.NewProgram(statusModel{workflows: wfs})
	_, err := p.Run()
	return err
}

// after:
func runStatusModel(wfs []*workflow.Workflow) error {
	if !hasTTY(os.Stdin) || !hasTTY(os.Stdout) {
		printStatus(wfs)
		return nil
	}
	...
}
```

## Files Changed

- `cmd/centinela/status_model.go` - add TTY detection and static rendering fallback
- `cmd/centinela/status_model_test.go` - cover static rendering helper behavior
- `cmd/centinela/status_runner_test.go` - add regression coverage for non-TTY execution
