package integration_test

import (
	"strings"
	"testing"

	"github.com/samuelnp/centinela/internal/ui"
)

func TestPersonaAppearsAcrossMessageTones(t *testing.T) {
	success := ui.RenderSuccess("workflow started")
	info := ui.RenderStep("Next step", "tests")
	errorLine := ui.RenderBlocked("code", "plan", "f", "/tmp/a.go")

	if !strings.Contains(success, "CENTINELA says") || !strings.Contains(success, "^_^") {
		t.Fatal("success output should include success persona")
	}
	if !strings.Contains(info, "CENTINELA says") || !strings.Contains(info, "o_o") {
		t.Fatal("info output should include info persona")
	}
	if !strings.Contains(errorLine, "CENTINELA says") || !strings.Contains(errorLine, "ò_ó") {
		t.Fatal("error output should include error persona")
	}
}
