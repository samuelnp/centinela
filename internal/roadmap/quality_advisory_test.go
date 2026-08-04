package roadmap

import (
	"os"
	"strings"
	"testing"
)

// writeQuality lays down a quality report pair in a temp CWD and returns the
// single-feature roadmap it grades.
func writeQuality(t *testing.T, json string) *Roadmap {
	t.Helper()
	t.Chdir(t.TempDir())
	os.MkdirAll(".workflow", 0755)                                  //nolint:errcheck
	os.WriteFile(RoadmapQualityMarkdown, []byte("# quality"), 0644) //nolint:errcheck
	os.WriteFile(RoadmapQualityFile, []byte(json), 0644)            //nolint:errcheck
	return &Roadmap{Phases: []Phase{{Name: "P0", Features: []Feature{{Name: "user"}}}}}
}

func qualityJSON(threshold, overall string) string {
	return `{"role":"roadmap-quality-evaluator","threshold":` + threshold +
		`,"features":[{"name":"user","scores":{"acceptanceCriteria":9,"userValue":9,` +
		`"definitionClarity":9,"dependencies":9,"effortEstimation":9,"overall":` + overall +
		`},"summary":"s"}]}`
}

// TestDeclaredThresholdIgnored: any declared threshold (or none) validates. The
// field still decodes so artifacts written before the deletion keep parsing.
func TestDeclaredThresholdIgnored(t *testing.T) {
	for _, threshold := range []string{"9", "7", "0", "99"} {
		r := writeQuality(t, qualityJSON(threshold, "9"))
		if err := ValidateQuality(r); err != nil {
			t.Fatalf("declared threshold %s must be ignored, got %v", threshold, err)
		}
	}
}

// TestScoreRangeStillRefuses is the ❌ direction: deleting the MINIMUM must not
// have deleted the 1-10 RANGE, and the message must name field and value.
func TestScoreRangeStillRefuses(t *testing.T) {
	for _, overall := range []string{"0", "11", "-3"} {
		r := writeQuality(t, qualityJSON("9", overall))
		err := ValidateQuality(r)
		if err == nil {
			t.Fatalf("overall %s must still be refused", overall)
		}
		if !strings.Contains(err.Error(), `"overall"`) || !strings.Contains(err.Error(), "between 1 and 10") {
			t.Fatalf("overall %s must fail as a range fault naming the field, got %v", overall, err)
		}
	}
}

// TestQualityAdvisories_NoAdviceWhenHealthyOrUnreadable: a report that is fine
// says nothing, and a report that is BROKEN also says nothing here — advisories
// must never speak over (or contradict) ValidateQuality's verdict.
func TestQualityAdvisories_NoAdviceWhenHealthyOrUnreadable(t *testing.T) {
	writeQuality(t, qualityJSON("9", "10"))
	if adv := QualityAdvisories(); len(adv) != 0 {
		t.Fatalf("a healthy report must draw no advice, got %v", adv)
	}
	writeQuality(t, `{"role":"roadmap-quality-evaluator"`)
	if adv := QualityAdvisories(); adv != nil {
		t.Fatalf("an unreadable report must draw no advice, got %v", adv)
	}
	t.Chdir(t.TempDir())
	if adv := QualityAdvisories(); adv != nil {
		t.Fatalf("a missing report must draw no advice, got %v", adv)
	}
}

// TestQualityArtifactNotRewritten: validating is a pure read. Nothing in the
// deletion path rewrites a user's existing artifact.
func TestQualityArtifactNotRewritten(t *testing.T) {
	body := qualityJSON("9", "3")
	r := writeQuality(t, body)
	if err := ValidateQuality(r); err != nil {
		t.Fatalf("validate: %v", err)
	}
	QualityAdvisories()
	after, _ := os.ReadFile(RoadmapQualityFile)
	if string(after) != body {
		t.Fatal("the quality report must be left byte-identical")
	}
}
