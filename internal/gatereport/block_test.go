package gatereport

import (
	"errors"
	"strings"
	"testing"
)

func blockReport(body string) string {
	return "### Report\n**Status:** SAFE\n\n" + FenceOpen + "\n" + body + "\n```\n"
}

func TestParseVerificationWellFormed(t *testing.T) {
	v, err := ParseVerification(blockReport(
		`{"revision":"9f2c1ab","treeDigest":"sha256:4e7d","commands":[{"argv":["centinela","validate"],"exitCode":0,"durationMs":84210}]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Revision != "9f2c1ab" || v.TreeDigest != "sha256:4e7d" || len(v.Commands) != 1 {
		t.Fatalf("bad parse: %+v", v)
	}
	if v.Commands[0].Line() != "centinela validate" || v.Commands[0].DurationMS != 84210 {
		t.Fatalf("bad command: %+v", v.Commands[0])
	}
}

func TestParseVerificationAbsentFence(t *testing.T) {
	if _, err := ParseVerification("### Report\n**Status:** SAFE\n"); !errors.Is(err, ErrNoBlock) {
		t.Fatalf("want ErrNoBlock, got %v", err)
	}
}

func TestParseVerificationIgnoresUntaggedFence(t *testing.T) {
	report := "### Report\n```json\n{\"revision\":\"x\"}\n```\n"
	if _, err := ParseVerification(report); !errors.Is(err, ErrNoBlock) {
		t.Fatalf("untagged ```json fence must be ignored, got %v", err)
	}
}

func TestParseVerificationUnterminatedFence(t *testing.T) {
	report := "### Report\n" + FenceOpen + "\n{\"revision\":\"x\"}\n"
	if _, err := ParseVerification(report); !errors.Is(err, ErrNoBlock) {
		t.Fatalf("unterminated fence must not parse, got %v", err)
	}
}

func TestParseVerificationMalformedJSON(t *testing.T) {
	_, err := ParseVerification(blockReport("{not json"))
	if err == nil || !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("want malformed error, got %v", err)
	}
}

func TestParseVerificationFirstFenceWins(t *testing.T) {
	report := blockReport(`{"revision":"first"}`) + blockReport(`{"revision":"second"}`)
	v, err := ParseVerification(report)
	if err != nil || v.Revision != "first" {
		t.Fatalf("first fence must win: %+v %v", v, err)
	}
}
