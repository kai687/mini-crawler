package starlark

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/script"
)

func TestProgramCanInspectDOM(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    title = doc.select_one("h1#page-title")
    description = doc.select_one("meta[name=description]")
    return {
        "url": doc.url,
        "title": text(title),
        "description": attr(description, "content"),
    }

def records(doc, ctx):
    root = doc.select_one("#content")
    out = []
    for node in root.select("h2[id], span[data-as=p]"):
        out.append({
            "tag": node_name(node),
            "anchor": attr(node, "id"),
            "text": text(node),
        })
    return out

def enrich(record, ctx):
    return record
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	doc := testDocument(t, `
<html>
  <head><meta name="description" content="Description"></head>
  <body>
    <h1 id="page-title">Title</h1>
    <div id="content">
      <h2 id="intro">Intro</h2>
      <span data-as="p">First paragraph</span>
    </div>
  </body>
</html>`)

	meta, err := program.PageMeta(doc, script.Context{})
	if err != nil {
		t.Fatalf("PageMeta() error = %v", err)
	}

	assertMapValue(t, meta, "url", "https://example.com/doc")
	assertMapValue(t, meta, "title", "Title")
	assertMapValue(t, meta, "description", "Description")

	records, err := program.Records(doc, script.Context{})
	if err != nil {
		t.Fatalf("Records() error = %v", err)
	}

	assertRecordCount(t, records, 2)
	assertMapValue(t, records[0], "tag", "h2")
	assertMapValue(t, records[0], "anchor", "intro")
	assertMapValue(t, records[1], "text", "First paragraph")
}

func TestProgramCanUseDOMTraversalHelpers(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    item = doc.select_one("li")
    header = doc.select_one(".param-head")
    next_block = header.next("div.mt-4")
    return {
        "has_parent": has_parent(doc.select_one("span"), "li"),
        "without_links": collapse_space(clone_without_text(item, "a")),
        "next_text": collapse_space(text(next_block.select_one("p"))),
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

	doc := testDocument(t, `
<html><body>
  <ul><li><a href="/x">Link</a> Keep me <span>Child</span></li></ul>
  <div class="param-head" id="p"></div>
  <div class="mt-4"><p>Description text</p></div>
</body></html>`)

	meta, err := program.PageMeta(doc, script.Context{})
	if err != nil {
		t.Fatalf("PageMeta() error = %v", err)
	}

	assertMapValue(t, meta, "has_parent", true)
	assertMapValue(t, meta, "without_links", "Keep me Child")
	assertMapValue(t, meta, "next_text", "Description text")
}

func TestDOMHelpersRejectNonNode(t *testing.T) {
	path := writeScript(t, `
def page_meta(doc, ctx):
    return {"bad": text("not-node")}

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

	if !strings.Contains(err.Error(), "text: want node, got string") {
		t.Fatalf("PageMeta() error = %q", err)
	}
}

func testDocument(t *testing.T, html string) script.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}

	return script.NewDocument(model.ParsedPage{
		URL: "https://example.com/doc",
		Doc: doc,
	})
}
