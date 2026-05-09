package script

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCheckListsExtractors(t *testing.T) {
	path := writeTestScript(t, `
def extract_docs(pattern, doc, ctx):
    return []

def extract_api(pattern, doc, ctx):
    return []

extract("^/doc", extract_docs)
extract("^/api", extract_api)
`)

	out := bytes.Buffer{}
	if err := runCheck(&out, config{Script: path}); err != nil {
		t.Fatalf("runCheck() error = %v", err)
	}

	got := out.String()
	for _, want := range []string{
		"script: " + path,
		"extractors: 2",
		`1. extract_docs "^/doc"`,
		`2. extract_api "^/api"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
}

func TestRunCheckWritesJSON(t *testing.T) {
	path := writeTestScript(t, `
def extract_docs(pattern, doc, ctx):
    return []

extract("^/doc", extract_docs)
`)

	out := bytes.Buffer{}
	if err := runCheck(&out, config{Script: path, JSON: true}); err != nil {
		t.Fatalf("runCheck() error = %v", err)
	}

	var got struct {
		Path       string `json:"path"`
		Extractors []struct {
			Name    string `json:"name"`
			Pattern string `json:"pattern"`
		} `json:"extractors"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; output = %s", err, out.String())
	}

	if got.Path != path {
		t.Fatalf("path = %q, want %q", got.Path, path)
	}

	if len(got.Extractors) != 1 {
		t.Fatalf("len(extractors) = %d, want 1", len(got.Extractors))
	}

	if got.Extractors[0].Name != "extract_docs" || got.Extractors[0].Pattern != "^/doc" {
		t.Fatalf("extractor = %#v", got.Extractors[0])
	}
}

func TestRunCheckRequiresScript(t *testing.T) {
	out := bytes.Buffer{}

	err := runCheck(&out, config{})
	if err == nil {
		t.Fatal("runCheck() error = nil")
	}

	if !strings.Contains(err.Error(), "script flag required") {
		t.Fatalf("runCheck() error = %q", err)
	}
}

func writeTestScript(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "script.star")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
