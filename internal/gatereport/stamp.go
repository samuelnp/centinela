package gatereport

import (
	"encoding/json"
	"fmt"
	"strings"
)

// emptyCommands is what a block gains when it is created by a stamp: an
// ungrounded report must keep failing Assess until a verifier fills it.
const emptyCommands = "[]"

// Stamped returns report with revision and treeDigest set inside the
// verification block, creating the block when absent. The existing `commands`
// array is spliced back verbatim — stamping records tree state, never claims.
//
// Splicing verbatim is precisely why the shape is checked HERE: the stamp is
// the last moment a malformed record can be refused at write time, and a
// record only rejected by a later reader is a defect discovered by the wrong
// party. The rule itself lives in ValidateCommandsSchema, shared with the
// reader — one shape, one place.
func Stamped(report, revision, digest string) (string, error) {
	commands := emptyCommands
	start, end, ok := blockSpan(report)
	if ok {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(report[start:end]), &fields); err != nil {
			return "", fmt.Errorf("malformed centinela:verification block: %w", err)
		}
		if raw, has := fields["commands"]; has {
			commands = string(raw)
		}
	}
	// Checked on BOTH branches so the block-creating path can never become
	// the one that skips the rule.
	if err := ValidateCommandsSchema(json.RawMessage(commands)); err != nil {
		return "", fmt.Errorf("malformed centinela:verification block: %w", err)
	}
	if ok {
		return report[:start] + blockBodyText(revision, digest, commands) + report[end:], nil
	}
	return appendBlock(report, blockBodyText(revision, digest, commands)), nil
}

// blockBodyText renders the block payload with a stable key order. Built by
// hand rather than marshalled so `commands` survives byte-identically.
func blockBodyText(revision, digest, commands string) string {
	return fmt.Sprintf("{\n  \"revision\": %s,\n  \"treeDigest\": %s,\n  \"commands\": %s\n}\n",
		quote(revision), quote(digest), commands)
}

func appendBlock(report, body string) string {
	if report != "" && !strings.HasSuffix(report, "\n") {
		report += "\n"
	}
	return report + "\n" + FenceOpen + "\n" + body + "```\n"
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
