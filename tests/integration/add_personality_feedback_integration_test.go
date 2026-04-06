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

	if !strings.Contains(success, "🛡️👁️") {
		t.Fatal("success output should include emoji prefix")
	}
	if !strings.Contains(info, "🛡️👁️") {
		t.Fatal("info output should include emoji prefix")
	}
	if !strings.Contains(errorLine, "🛡️👁️") {
		t.Fatal("error output should include emoji prefix")
	}
}
