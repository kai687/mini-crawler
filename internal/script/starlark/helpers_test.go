package starlark

import (
	"testing"

	"github.com/algolia/docs-crawler/internal/script"
)

func TestProgramCanUseHostHelpers(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    return {
        "trimmed": trim("  hello  "),
        "collapsed": collapse_space("hello\n  docs"),
        "joined": url_join("https://example.com/docs/page", "../api#intro"),
        "without_anchor": url_without_anchor("https://example.com/docs/page#intro"),
        "path": path("https://example.com/docs/page?x=1#intro"),
        "sha1": sha1("abc"),
        "matched": regex_match("^a.+", "abc"),
        "replaced": regex_replace("[0-9]+", "N", "v123"),
    }

def records(doc, ctx):
    return []

def enrich(record, ctx):
    return record
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	meta, err := program.PageMeta(script.Document{}, script.Context{})
	if err != nil {
		t.Fatalf("PageMeta() error = %v", err)
	}

	assertMapValue(t, meta, "trimmed", "hello")
	assertMapValue(t, meta, "collapsed", "hello docs")
	assertMapValue(t, meta, "joined", "https://example.com/api#intro")
	assertMapValue(t, meta, "without_anchor", "https://example.com/docs/page")
	assertMapValue(t, meta, "path", "/docs/page")
	assertMapValue(t, meta, "sha1", "a9993e364706816aba3e25717850c26c9cd0d89d")
	assertMapValue(t, meta, "matched", true)
	assertMapValue(t, meta, "replaced", "vN")
}
