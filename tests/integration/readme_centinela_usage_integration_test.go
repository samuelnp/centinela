package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadmeCentinelaUsageLinksLandingPageHowto(t *testing.T) {
	readmeData, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	howtoData, err := os.ReadFile(filepath.Join("..", "..", "HOWTO.md"))
	if err != nil {
		t.Fatalf("read HOWTO: %v", err)
	}
	readme := string(readmeData)
	howto := string(howtoData)

	if !strings.Contains(readme, "HOWTO.md`](HOWTO.md)") {
		t.Fatalf("README does not link to HOWTO.md")
	}
	for _, want := range []string{"landing-page-mvp", "centinela start", "centinela validate", "centinela docs generate"} {
		if !strings.Contains(howto, want) {
			t.Fatalf("HOWTO missing %q", want)
		}
	}
}
