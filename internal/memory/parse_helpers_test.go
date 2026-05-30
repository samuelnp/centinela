package memory

import "testing"

// decisionBullets: case-insensitive section match.
func TestDecisionBulletsCaseInsensitive(t *testing.T) {
	text := "## DECISIONS\n- item one\n"
	bullets := decisionBullets(text)
	if len(bullets) != 1 {
		t.Fatalf("expected 1 bullet (case-insensitive), got %d", len(bullets))
	}
}

// bulletText: star-style bullets work.
func TestBulletTextStarStyle(t *testing.T) {
	if bulletText("* star item") != "star item" {
		t.Fatal("expected star-style bullet to parse")
	}
}
