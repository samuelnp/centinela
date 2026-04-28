package acceptance_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Acceptance: specs/readme-centinela-usage.feature
func TestReadmeCentinelaUsageHowtoCoversWorkflow(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "HOWTO.md"))
	if err != nil {
		t.Fatalf("read HOWTO: %v", err)
	}
	howto := string(data)

	for _, want := range []string{
		"## 3. Plan Step",
		"## 4. Code Step",
		"## 5. Tests Step",
		"## 6. Validate Step",
		"## 7. Docs Step",
		"Step plan complete",
		"executable acceptance tests",
	} {
		if !strings.Contains(howto, want) {
			t.Fatalf("HOWTO missing %q", want)
		}
	}
}
