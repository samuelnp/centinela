package gates

import "fmt"

const fileSizeGate = "G1: File Size"

// fileSizeNothingInspected is deliberately not a pass claim: it states what was
// inspected (nothing), and the Skip status keeps it out of the passed tally.
const fileSizeNothingInspected = "No files in the diff scope — nothing inspected."

// fileSizePassMessage names the effective cap and where its per-file exceptions
// come from, so the operator can tell an earned pass from an absolute rule they
// have no visible lever on. maxLines is the single source of the number.
func fileSizePassMessage(justified []string) string {
	if len(justified) > 0 {
		return fmt.Sprintf(
			"All in-scope files meet the %d-line cap, including %d justified exception(s) "+
				"declared under [[gates.file_size_exceptions]].", maxLines, len(justified))
	}
	return fmt.Sprintf(
		"All in-scope files are within the %d-line cap "+
			"(per-file exceptions are configurable under [[gates.file_size_exceptions]]).", maxLines)
}

// fileSizeFailMessage keeps the Fail severity and the cap unchanged — this
// feature changes messages and the Pass/Skip axis only.
func fileSizeFailMessage() string {
	return fmt.Sprintf("Files exceeding %d lines must be split unless explicitly justified.", maxLines)
}
