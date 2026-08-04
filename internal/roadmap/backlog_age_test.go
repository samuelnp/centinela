package roadmap

import (
	"testing"
	"time"
)

var ageNow = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

// backlogOf builds a roadmap whose Backlog holds one finding per age in days;
// a negative age means "no deferredAt at all".
func backlogOf(specs ...struct {
	name string
	at   string
}) *Roadmap {
	feats := make([]Feature, 0, len(specs))
	for _, s := range specs {
		feats = append(feats, Feature{Name: s.name, Summary: "s", DeferredAt: s.at})
	}
	return &Roadmap{Phases: []Phase{
		{Name: "Phase 1", Features: []Feature{{Name: "real"}}},
		{Name: BacklogPhaseName, Features: feats},
	}}
}

func daysAgo(n int) string {
	return ageNow.AddDate(0, 0, -n).Format(time.RFC3339)
}

type spec = struct {
	name string
	at   string
}

func TestAgeBacklogOrdersOldestFirst(t *testing.T) {
	r := backlogOf(spec{"young", daysAgo(2)}, spec{"old", daysAgo(40)}, spec{"mid", daysAgo(20)})
	got := AgeBacklog(r, ageNow, DefaultStaleDays)
	want := []string{"old", "mid", "young"}
	for i, name := range want {
		if got[i].Name != name {
			t.Fatalf("position %d = %q, want %q", i, got[i].Name, name)
		}
	}
	if got[0].AgeDays != 40 || got[2].AgeDays != 2 {
		t.Fatalf("ages = %d, %d", got[0].AgeDays, got[2].AgeDays)
	}
	if !got[0].Stale || !got[1].Stale || got[2].Stale {
		t.Fatal("only the two past the 14d threshold may be stale")
	}
}

// Strictly older than N: exactly N days is not yet stale.
func TestAgeBacklogThresholdBoundaryIsStrict(t *testing.T) {
	r := backlogOf(spec{"exact", daysAgo(14)}, spec{"past", daysAgo(15)})
	got := AgeBacklog(r, ageNow, DefaultStaleDays)
	byName := map[string]Aged{}
	for _, a := range got {
		byName[a.Name] = a
	}
	if byName["exact"].Stale {
		t.Fatal("exactly 14 days old is not older than 14 days")
	}
	if !byName["past"].Stale {
		t.Fatal("15 days old must be stale")
	}
}

// E9/E10: a missing or garbage clock is loud, never an error.
func TestAgeBacklogUnknownAgesSortFirstAndCountStale(t *testing.T) {
	r := backlogOf(spec{"fresh", daysAgo(1)}, spec{"garbage", "not-a-date"}, spec{"absent", ""})
	got := AgeBacklog(r, ageNow, DefaultStaleDays)
	if got[0].Name != "absent" || got[1].Name != "garbage" {
		t.Fatalf("unknown ages must sort first, got %q %q", got[0].Name, got[1].Name)
	}
	for _, a := range got[:2] {
		if a.KnownAge || !a.Stale {
			t.Fatalf("%q: KnownAge=%v Stale=%v", a.Name, a.KnownAge, a.Stale)
		}
	}
}

func TestAgeBacklogEmptyAndNil(t *testing.T) {
	if got := AgeBacklog(nil, ageNow, 0); len(got) != 0 {
		t.Fatalf("nil roadmap = %v", got)
	}
	if got := AgeBacklog(&Roadmap{Phases: []Phase{{Name: "Phase 1"}}}, ageNow, 0); len(got) != 0 {
		t.Fatalf("no Backlog phase = %v", got)
	}
}

func TestAgeBacklogHonorsAnOverriddenThreshold(t *testing.T) {
	r := backlogOf(spec{"old", daysAgo(40)}, spec{"mid", daysAgo(20)})
	stale := StaleOnly(AgeBacklog(r, ageNow, 30))
	if len(stale) != 1 || stale[0].Name != "old" {
		t.Fatalf("--older-than 30 = %v", stale)
	}
}
