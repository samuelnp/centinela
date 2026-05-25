package acceptance_test

import (
	"regexp"
	"strings"
	"testing"
)

// Scenario: Absolute OG and Twitter meta tags are set
func TestLandingPage_OGMetaTags(t *testing.T) {
	html := loadIndex(t)
	const imgURL = "https://samuelnp.github.io/centinela/assets/social-preview.png"
	const pageURL = "https://samuelnp.github.io/centinela/"

	checks := []struct{ label, want string }{
		{"og:image", `content="` + imgURL + `"`},
		{"twitter:image", `content="` + imgURL + `"`},
		{"og:url", `content="` + pageURL + `"`},
		{"twitter:card", `content="summary_large_image"`},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("meta tag %s: %q not found", c.label, c.want)
		}
	}
}

// Scenario: No external runtime CDN or JS framework dependency
func TestLandingPage_NoExternalDeps(t *testing.T) {
	html := loadIndex(t)

	// No <script src="http...">
	scriptExtRe := regexp.MustCompile(`(?i)<script[^>]+\bsrc\s*=\s*["']https?://`)
	if scriptExtRe.MatchString(html) {
		t.Error("found <script> with external http src")
	}

	// No <link rel="stylesheet" href="http...">  (exclude data: favicon)
	linkExtRe := regexp.MustCompile(`(?i)<link[^>]+\brel\s*=\s*["']stylesheet["'][^>]+\bhref\s*=\s*["']https?://`)
	if linkExtRe.MatchString(html) {
		t.Error("found <link rel=stylesheet> with external http href")
	}
	// Also catch reversed attribute order: href first, then rel=stylesheet
	linkExtRe2 := regexp.MustCompile(`(?i)<link[^>]+\bhref\s*=\s*["']https?://[^"']*["'][^>]+\brel\s*=\s*["']stylesheet["']`)
	if linkExtRe2.MatchString(html) {
		t.Error("found <link href=http...> with rel=stylesheet")
	}
}

// Scenario: demo.gif is lazy-loaded with explicit dimensions and a placeholder
func TestLandingPage_DemoGifAttributes(t *testing.T) {
	html := loadIndex(t)
	checks := []struct{ label, want string }{
		{"demo.gif src", `src="assets/demo.gif"`},
		{"loading=lazy", `loading="lazy"`},
		{"decoding=async", `decoding="async"`},
		{"explicit width", `width="1200"`},
		{"explicit height", `height="700"`},
	}
	for _, c := range checks {
		if !strings.Contains(html, c.want) {
			t.Errorf("demo img: %s — %q not found", c.label, c.want)
		}
	}
}
