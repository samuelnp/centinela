package acceptance_test

import (
	"strings"
	"testing"
)

// Scenario: No-JS degraded path — all content remains legible
// Proxy: .reveal hidden state is gated behind html.js class only.
func TestLandingPage_NoJS_RevealGating(t *testing.T) {
	html := loadIndex(t)
	if !strings.Contains(html, "html.js .reveal") {
		t.Error("no-JS guard missing: 'html.js .reveal' selector not found")
	}
}

// Scenario: Reduced-motion preference disables CSS animations
func TestLandingPage_ReducedMotionGuard(t *testing.T) {
	html := loadIndex(t)
	if !strings.Contains(html, "prefers-reduced-motion") {
		t.Error("prefers-reduced-motion media query not found")
	}
}

// Regression guard 1 (overflow): nav collapse specificity fix is present.
// The rule .nav-links a.hide-sm{display:none} must beat .nav-links a so
// links collapse on narrow screens without causing horizontal overflow.
func TestLandingPage_NavCollapseSpecificityRegression(t *testing.T) {
	html := loadIndex(t)
	const rule = ".nav-links a.hide-sm{display:none}"
	if !strings.Contains(html, rule) {
		t.Errorf("nav overflow regression: rule %q not found in page", rule)
	}
}

// Scenario: Narrow viewport reflows pipeline and roadmap without horizontal overflow
func TestLandingPage_NarrowViewportMediaQuery(t *testing.T) {
	html := loadIndex(t)
	if !strings.Contains(html, "@media (max-width:375px)") {
		t.Error("narrow-viewport guard '@media (max-width:375px)' not found")
	}
}

// Regression guard 2 (palette): page uses logo azure blue, NOT removed neon green.
func TestLandingPage_PaletteBlueNotGreen(t *testing.T) {
	html := loadIndex(t)
	const azure = "#40a0e0"
	if !strings.Contains(html, azure) {
		t.Errorf("palette regression: logo azure %q not found", azure)
	}
	const neonGreen = "#3fff9f"
	if strings.Contains(html, neonGreen) {
		t.Errorf("palette regression: removed neon green %q still present", neonGreen)
	}
}
