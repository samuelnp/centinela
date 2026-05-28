package main

import "testing"

func TestExtractPostwritePathBranches(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"malformed", "{not json", ""},
		{"snake", `{"tool_input":{"file_path":"/tmp/a"}}`, "/tmp/a"},
		{"camel", `{"tool_input":{"filePath":"/tmp/b"}}`, "/tmp/b"},
		{"missing", `{"tool_input":{}}`, ""},
	}
	for _, c := range cases {
		got := extractPostwritePath([]byte(c.in))
		if got != c.want {
			t.Fatalf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestReformatPostwriteSkipsMissingFile(t *testing.T) {
	// Path that does not exist should be a silent no-op (no panic).
	reformatPostwrite([]byte(`{"tool_input":{"file_path":"/does/not/exist.json"}}`))
}

func TestReformatPostwriteSkipsEmptyPath(t *testing.T) {
	reformatPostwrite([]byte(`{}`))
}
