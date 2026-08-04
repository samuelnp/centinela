package roadmap

import "fmt"

// advisoryScore is the overall score at or above which a graded feature draws
// no comment. It gates NOTHING and never contributes to an error: the refusal
// that used to hang off it was deleted because a number assigned by the party
// it constrains is not evidence. It survives only to decide what
// `centinela roadmap validate` bothers mentioning on an otherwise valid roadmap.
const advisoryScore = 9

// QualityAdvisories returns one human-readable line per graded feature whose
// self-assigned overall score is below advisoryScore. It is a REPORT, not a
// check: a report that is missing, unreadable or structurally faulty yields no
// advisories at all, because ValidateQuality is the surface that names those
// faults and this one must never duplicate — or contradict — its verdict.
func QualityAdvisories() []string {
	features, err := loadQualityFeatures()
	if err != nil {
		return nil
	}
	var out []string
	for _, f := range features {
		if f.Scores.Overall < advisoryScore {
			out = append(out, fmt.Sprintf(
				"%s: overall %d (below %d) — advisory only, nothing is blocked",
				f.Name, f.Scores.Overall, advisoryScore))
		}
	}
	return out
}
