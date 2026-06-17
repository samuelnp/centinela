package evidence

import (
	"os"
	"strings"
	"testing"
)

func chdirSpecsTemp(t *testing.T, specs ...string) {
	t.Helper()
	d := t.TempDir()
	o, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(o) })
	if err := os.Chdir(d); err != nil {
		t.Fatal(err)
	}
	if len(specs) > 0 {
		if err := os.MkdirAll("specs", 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, s := range specs {
		if err := os.WriteFile("specs/"+s, []byte("Feature: x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAnalyzedSpecsListSortedWhenPresent(t *testing.T) {
	chdirSpecsTemp(t, "b.feature", "a.feature")
	got := analyzedSpecsList()
	if got != "- specs/a.feature\n- specs/b.feature" {
		t.Fatalf("analyzedSpecsList not sorted/expected: %q", got)
	}
	if strings.Contains(got, "<FILL:") {
		t.Fatalf("populated list must carry no fill row: %q", got)
	}
}

func TestAnalyzedSpecsListEmptyWhenNoSpecs(t *testing.T) {
	chdirSpecsTemp(t)
	got := analyzedSpecsList()
	if !strings.HasPrefix(got, "- <FILL:") {
		t.Fatalf("empty case should render one fill row: %q", got)
	}
}

func TestGatekeeperBodyKeepsStatusAndDateAndFill(t *testing.T) {
	chdirSpecsTemp(t, "a.feature")
	body := string(gatekeeperBody("demo"))
	for _, want := range []string{"**Status:**", "**Date:**", "<FILL:", "specs/a.feature"} {
		if !strings.Contains(body, want) {
			t.Fatalf("gatekeeperBody missing %q in %s", want, body)
		}
	}
}

func TestProdReadyBodyKeepsStatusAndDate(t *testing.T) {
	body := string(prodReadyBody("demo"))
	for _, want := range []string{"**Status:**", "**Date:**", "<FILL:"} {
		if !strings.Contains(body, want) {
			t.Fatalf("prodReadyBody missing %q", want)
		}
	}
}

func TestEdgeAndChangelogBodiesUseFill(t *testing.T) {
	if !strings.Contains(string(edgeCasesBody("demo")), "<FILL:") {
		t.Fatal("edgeCasesBody missing fill marker")
	}
	if !strings.Contains(string(changelogBody("demo")), "<FILL:") {
		t.Fatal("changelogBody missing fill marker")
	}
}
