package docgen

import (
	"os"
	"testing"
)

func TestLoadEvidenceAndStatesSkipsInvalid(t *testing.T) {
	d := t.TempDir()
	o, _ := os.Getwd()
	defer os.Chdir(o)                                                                                                                                                      //nolint:errcheck
	os.Chdir(d)                                                                                                                                                            //nolint:errcheck
	os.MkdirAll(".workflow", 0755)                                                                                                                                         //nolint:errcheck
	os.WriteFile(".workflow/roadmap.json", []byte(`{}`), 0644)                                                                                                             //nolint:errcheck
	os.WriteFile(".workflow/roadmap-analysis.json", []byte(`{}`), 0644)                                                                                                    //nolint:errcheck
	os.WriteFile(".workflow/x-qa-senior.json", []byte(`{"feature":"x","role":"qa-senior","step":"tests","handoffTo":"validate","outputs":["a"],"edgeCases":["e"]}`), 0644) //nolint:errcheck
	os.WriteFile(".workflow/y-bad.json", []byte(`{"feature":""}`), 0644)                                                                                                   //nolint:errcheck
	os.WriteFile(".workflow/x.json", []byte(`{"feature":"x","currentStep":"tests","steps":{"tests":{"status":"in-progress"}}}`), 0644)                                     //nolint:errcheck
	os.WriteFile(".workflow/readme.json", []byte(`{}`), 0644)                                                                                                              //nolint:errcheck
	ev := loadEvidence()
	if len(ev) != 1 || ev[0].Role != "qa-senior" {
		t.Fatalf("unexpected evidence: %#v", ev)
	}
	st := loadStates()
	if len(st) != 1 || st[0].Feature != "x" {
		t.Fatalf("unexpected states: %#v", st)
	}
}
