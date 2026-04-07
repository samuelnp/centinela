package workflow

import "testing"

func TestHasAcceptanceExecutionCommand(t *testing.T) {
	if hasAcceptanceExecutionCommand([]string{"go vet ./..."}) {
		t.Fatal("unexpected acceptance execution detection")
	}
	if !hasAcceptanceExecutionCommand([]string{"go test ./..."}) {
		t.Fatal("expected go test ./... to qualify")
	}
	if !hasAcceptanceExecutionCommand([]string{"npx cucumber-js"}) {
		t.Fatal("expected cucumber command to qualify")
	}
	if !hasAcceptanceExecutionCommand([]string{"go test ./tests/acceptance/..."}) {
		t.Fatal("expected tests/acceptance command to qualify")
	}
	if !hasAcceptanceExecutionCommand([]string{"npm run test:acceptance"}) {
		t.Fatal("expected acceptance + test command to qualify")
	}
}
