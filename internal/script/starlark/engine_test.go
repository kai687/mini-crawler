package starlark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/script"
)

func TestEngineLoadsAndRunsProgram(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    return {"title": "Doc", "url": doc.url}

def records(doc, ctx):
    return [
        {"url": ctx["url"], "position": ctx["position"], "tags": ["a", "b"]},
    ]

def enrich(record, ctx):
    record["objectID"] = "id-1"
    record["metadata"] = ctx["metadata"]["source"]
    return record
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	doc := script.NewDocument(model.ParsedPage{URL: "https://example.com/doc"})
	ctx := script.Context{
		URL:      "https://example.com/doc",
		Position: 2,
		Metadata: map[string]any{"source": "test"},
	}

	meta, err := program.PageMeta(doc, ctx)
	if err != nil {
		t.Fatalf("PageMeta() error = %v", err)
	}

	assertMapValue(t, meta, "title", "Doc")
	assertMapValue(t, meta, "url", "https://example.com/doc")

	records, err := program.Records(doc, ctx)
	if err != nil {
		t.Fatalf("Records() error = %v", err)
	}

	assertRecordCount(t, records, 1)
	assertMapValue(t, records[0], "position", int64(2))

	enriched, err := program.Enrich(records[0], ctx)
	if err != nil {
		t.Fatalf("Enrich() error = %v", err)
	}

	assertMapValue(t, enriched, "objectID", "id-1")
	assertMapValue(t, enriched, "metadata", "test")
}

func TestProgramStopsAfterMaxExecutionSteps(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    for _ in range(1000000):
        pass
    return {}

def records(doc, ctx):
    return []

def enrich(record, ctx):
    return record
`)

	program, err := Engine{MaxExecutionSteps: 1_000}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = program.PageMeta(script.Document{}, script.Context{})
	if err == nil {
		t.Fatal("PageMeta() error = nil")
	}

	if !strings.Contains(err.Error(), "too many steps") {
		t.Fatalf("PageMeta() error = %q", err)
	}
}

func TestEngineRejectsMissingExport(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    return {}
`)

	_, err := Engine{}.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil")
	}

	if !strings.Contains(err.Error(), "missing records") {
		t.Fatalf("Load() error = %q", err)
	}
}

func TestProgramRejectsInvalidReturnValue(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    return {"title": len}

def records(doc, ctx):
    return []

def enrich(record, ctx):
    return record
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = program.PageMeta(script.Document{}, script.Context{})
	if err == nil {
		t.Fatal("PageMeta() error = nil")
	}

	if !strings.Contains(err.Error(), "unsupported starlark value builtin_function_or_method") {
		t.Fatalf("PageMeta() error = %q", err)
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
