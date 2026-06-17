package insights

import "github.com/samuelnp/centinela/internal/telemetry"

// noneKey is the display token for an empty bucket field (reason/fileType/gate).
const noneKey = "<none>"

// blockKey is the display key for a block event: "<reason> · <fileType>", with
// each empty field shown as <none> so missing optional fields still bucket
// deterministically rather than collapsing into one another.
func blockKey(e telemetry.Event) string {
	return orNone(e.Reason) + " · " + orNone(e.FileType)
}

// orNone renders an empty field as the <none> token.
func orNone(s string) string {
	if s == "" {
		return noneKey
	}
	return s
}

// blocks ranks block events by bucket "<reason> · <fileType>", count desc then
// key asc, truncated to topN.
func blocks(events []telemetry.Event, topN int) []Count {
	m := make(map[string]int)
	for _, e := range events {
		if e.Type == telemetry.TypeBlock {
			m[blockKey(e)]++
		}
	}
	return rankTop(m, topN)
}
