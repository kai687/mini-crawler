package starlark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algolia/mini-crawler/pkg/model"
	"github.com/algolia/mini-crawler/pkg/script"
)

func TestEngineLoadsAndRunsProgram(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return [
        {"url": ctx["url"], "pattern": pattern, "tags": ["a", "b"]},
    ]

extract("^/doc", extract_docs)
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	doc := script.NewDocument(model.ParsedPage{URL: "https://example.com/doc"})
	ctx := script.Context{URL: "https://example.com/doc"}

	records, err := program.Extract(doc, ctx)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertRecordCount(t, records, 1)
	assertMapValue(t, records[0], "url", "https://example.com/doc")
	assertMapValue(t, records[0], "pattern", "^/doc")
}

func TestProgramUsesFirstMatchingExtractor(t *testing.T) {
	path := writeScript(t, `
def extract_first(pattern, doc, ctx):
    return [{"name": "first"}]

def extract_second(pattern, doc, ctx):
    return [{"name": "second"}]

extract("^/doc", extract_first)
extract("^/doc/rest-api", extract_second)
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	records, err := program.Extract(
		script.Document{},
		script.Context{URL: "https://example.com/doc/rest-api/search"},
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertMapValue(t, records[0], "name", "first")
}

func TestProgramStopsAfterMaxExecutionSteps(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    for _ in range(1000000):
        pass
    return []

extract(".*", extract_docs)
`)

	program, err := Engine{MaxExecutionSteps: 1_000}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = program.Extract(script.Document{}, script.Context{URL: "https://example.com/doc"})
	if err == nil {
		t.Fatal("Extract() error = nil")
	}

	if !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("Extract() error = %q", err)
	}
}

func TestEngineRejectsMissingExtractor(t *testing.T) {
	path := writeScript(t, `
VALUE = 1
`)

	_, err := Engine{}.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil")
	}

	if !strings.Contains(err.Error(), "registered no extractors") {
		t.Fatalf("Load() error = %q", err)
	}
}

func TestEngineAcceptsExtractorWithoutPrefix(t *testing.T) {
	path := writeScript(t, `
def docs(pattern, doc, ctx):
    return []

extract(".*", docs)
`)

	_, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
}

func TestEngineRejectsExtractorWithWrongShape(t *testing.T) {
	path := writeScript(t, `
def extract_docs(doc):
    return []

extract(".*", extract_docs)
`)

	_, err := Engine{}.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil")
	}

	if !strings.Contains(err.Error(), "must accept exactly 3 positional arguments") {
		t.Fatalf("Load() error = %q", err)
	}
}

func TestEngineRejectsInvalidExtractorPattern(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return []

extract("[", extract_docs)
`)

	_, err := Engine{}.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil")
	}

	if !strings.Contains(err.Error(), "error parsing regexp") {
		t.Fatalf("Load() error = %q", err)
	}
}

func TestProgramErrorsWhenNoExtractorMatches(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return []

extract("^/doc", extract_docs)
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = program.Extract(script.Document{}, script.Context{URL: "https://example.com/blog"})
	if err == nil {
		t.Fatal("Extract() error = nil")
	}

	if !strings.Contains(err.Error(), "no extractor matches") {
		t.Fatalf("Extract() error = %q", err)
	}
}

func TestProgramRejectsInvalidReturnValue(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return [{"title": len}]

extract(".*", extract_docs)
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = program.Extract(script.Document{}, script.Context{URL: "https://example.com/doc"})
	if err == nil {
		t.Fatal("Extract() error = nil")
	}

	if !strings.Contains(
		err.Error(),
		"$[0].title: unsupported starlark value builtin_function_or_method",
	) {
		t.Fatalf("Extract() error = %q", err)
	}
}

func assertMapValue(t *testing.T, values map[string]any, key string, want any) {
	t.Helper()

	if values[key] != want {
		t.Fatalf("%s = %#v, want %#v in %#v", key, values[key], want, values)
	}
}

func assertRecordCount(t *testing.T, records []map[string]any, want int) {
	t.Helper()

	if len(records) != want {
		t.Fatalf("len(records) = %d, want %d: %#v", len(records), want, records)
	}
}

func writeScript(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "script.star")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
