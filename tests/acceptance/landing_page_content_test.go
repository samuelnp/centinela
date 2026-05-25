package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: Full interactive render shows all required above-the-fold elements
func TestLandingPage_HeroElements(t *testing.T) {
	html := loadIndex(t)
	checks := []struct{ label, want string }{
		{"logo-banner src", `src="assets/logo-banner.png"`},
		{"h1 Centinela", `<h1>`},
		{"Get Started link", `>Get Started<`},
		{"Star on GitHub link", `>★ Star on GitHub<`},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("hero missing %s: %q not found", c.label, c.want)
		}
	}
}

// Scenario: Install command exactly matches the README canonical string
func TestLandingPage_InstallCommand(t *testing.T) {
	html := loadIndex(t)
	const exact = "go install github.com/samuelnp/centinela@latest"
	if !strings.Contains(html, exact) {
		t.Errorf("install command %q not found in page", exact)
	}
}

// Scenario: Value prop text is present
func TestLandingPage_ValueProp(t *testing.T) {
	html := loadIndex(t)
	// The value prop is split across spans; assert the key identifiable fragment.
	const want = "plan →"
	if !strings.Contains(html, want) {
		t.Errorf("value prop arrow text %q not found", want)
	}
	if !strings.Contains(html, "enforced") {
		t.Error("value prop 'enforced' text not found")
	}
}

// Scenario: Pipeline diagram section is present with five step labels
func TestLandingPage_PipelineStepLabels(t *testing.T) {
	html := loadIndex(t)
	for _, step := range []string{"plan", "code", "tests", "validate", "docs"} {
		if !strings.Contains(html, `class="pname">`+step) {
			t.Errorf("pipeline step label %q not found", step)
		}
	}
}

// Scenario: Greenfield roadmap section is present
func TestLandingPage_GreenfieldSection(t *testing.T) {
	html := loadIndex(t)
	checks := []struct{ label, want string }{
		{"roadmap word", "roadmap"},
		{"Phase label", "Phase"},
		{"feature chip", "feature"},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("greenfield section missing %s: %q not found", c.label, c.want)
		}
	}
}

// Scenario: Enforcement "aha" panel is present
func TestLandingPage_EnforcementPanel(t *testing.T) {
	html := loadIndex(t)
	if !strings.Contains(html, "write blocked") {
		t.Error("enforcement panel: 'write blocked' message not found")
	}
	if !strings.Contains(html, "`plan`") {
		t.Error("enforcement panel: plan step context not found")
	}
}

// Scenario: Footer contains real non-empty outbound links
func TestLandingPage_FooterLinks(t *testing.T) {
	html := loadIndex(t)
	// Extract footer section
	footerStart := strings.Index(html, "<footer")
	footerEnd := strings.Index(html, "</footer>")
	if footerStart < 0 || footerEnd < 0 {
		t.Fatal("footer element not found")
	}
	footer := html[footerStart : footerEnd+9]

	if strings.Contains(footer, `href=""`) || strings.Contains(footer, `href="#"`) {
		t.Error("footer contains empty or '#' href")
	}
	if !strings.Contains(footer, "github.com/samuelnp/centinela") {
		t.Error("footer has no link containing github.com/samuelnp/centinela")
	}
}
