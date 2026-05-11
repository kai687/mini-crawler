package starlark

import (
	"testing"

	"github.com/kai687/mini-crawler/pkg/script"
)

func TestProgramCanUseHostHelpers(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return [{
        "trimmed": trim("  hello  "),
        "collapsed": collapse_space("hello\n  docs"),
        "escaped_html": escape_html("<tag title=\"Tom & Jerry\">'x'</tag>"),
        "joined": url_join("https://example.com/docs/page", "../api#intro"),
        "without_anchor": url_without_anchor("https://example.com/docs/page#intro"),
        "path": path("https://example.com/docs/page?x=1#intro"),
        "sha1": sha1("abc"),
        "matched": regex_match("^a.+", "abc"),
        "replaced": regex_replace("[0-9]+", "N", "v123"),
    }]

extract(".*", extract_docs)
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	records, err := program.Extract(
		script.Document{},
		script.Context{URL: "https://example.com/docs/page"},
	)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertMapValue(t, records[0], "trimmed", "hello")
	assertMapValue(t, records[0], "collapsed", "hello docs")
	assertMapValue(
		t,
		records[0],
		"escaped_html",
		"&lt;tag title=&#34;Tom &amp; Jerry&#34;&gt;&#39;x&#39;&lt;/tag&gt;",
	)
	assertMapValue(t, records[0], "joined", "https://example.com/api#intro")
	assertMapValue(t, records[0], "without_anchor", "https://example.com/docs/page")
	assertMapValue(t, records[0], "path", "/docs/page")
	assertMapValue(t, records[0], "sha1", "a9993e364706816aba3e25717850c26c9cd0d89d")
	assertMapValue(t, records[0], "matched", true)
	assertMapValue(t, records[0], "replaced", "vN")
}
