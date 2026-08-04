package roadmap

import "testing"

func TestNudgeFiresWhenOnlyTheBacklogRemains(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "real")
	r := backlogOf(spec{"old", daysAgo(40)}, spec{"mid", daysAgo(20)}, spec{"young", daysAgo(2)})
	n, ok := NudgeFor(r, ageNow, DefaultStaleDays)
	if !ok {
		t.Fatal("a finished roadmap with a non-empty Backlog must nudge")
	}
	if n.Total != 3 || n.Stale != 2 || n.ThresholdDays != DefaultStaleDays {
		t.Fatalf("nudge = %+v", n)
	}
}

// The Backlog itself must never read as unfinished schedulable work — without
// FirstIncompleteSchedulable every deferral would suppress the nudge forever.
func TestNudgeIsNotSuppressedByTheBacklogItself(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "real")
	if _, ok := FirstIncompleteSchedulable(backlogOf(spec{"old", daysAgo(40)})); ok {
		t.Fatal("a Backlog finding is not schedulable work")
	}
	if _, ok := FirstIncomplete(backlogOf(spec{"old", daysAgo(40)})); !ok {
		t.Fatal("FirstIncomplete keeps its all-phases meaning")
	}
}

func TestNudgeSilentWhileSchedulableWorkRemains(t *testing.T) {
	chdirRoadmapTemp(t)
	r := backlogOf(spec{"old", daysAgo(40)})
	if _, ok := NudgeFor(r, ageNow, DefaultStaleDays); ok {
		t.Fatal("an unfinished schedulable feature must suppress the nudge")
	}
}

func TestNudgeSilentOnAnEmptyBacklog(t *testing.T) {
	chdirRoadmapTemp(t)
	seedDone(t, "real")
	r := &Roadmap{Phases: []Phase{{Name: "Phase 1", Features: []Feature{{Name: "real"}}}}}
	if _, ok := NudgeFor(r, ageNow, DefaultStaleDays); ok {
		t.Fatal("an empty Backlog must not nudge")
	}
}

func TestNudgeSilentOnNilRoadmap(t *testing.T) {
	if _, ok := NudgeFor(nil, ageNow, DefaultStaleDays); ok {
		t.Fatal("a nil roadmap must not nudge")
	}
	if _, ok := FirstIncompleteSchedulable(nil); ok {
		t.Fatal("a nil roadmap has no incomplete feature")
	}
}
