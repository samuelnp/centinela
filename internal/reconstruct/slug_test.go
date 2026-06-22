package reconstruct

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"internal/Handler":      "internal-handler",
		"src/api/v1":            "src-api-v1",
		"--leading--trailing--": "leading-trailing",
		"":                      "module",
		"///":                   "module",
		"a.b.c":                 "a-b-c",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDisambiguate(t *testing.T) {
	used := map[string]bool{}
	if disambiguate("svc", used) != "svc" {
		t.Fatal("first claimant keeps bare slug")
	}
	if got := disambiguate("svc", used); got != "svc-2" {
		t.Fatalf("second collision must suffix, got %q", got)
	}
	if got := disambiguate("svc", used); got != "svc-3" {
		t.Fatalf("third collision must suffix, got %q", got)
	}
}

func TestItoa(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 42: "42", 1000: "1000"}
	for in, want := range cases {
		if got := itoa(in); got != want {
			t.Errorf("itoa(%d) = %q, want %q", in, got, want)
		}
	}
}
