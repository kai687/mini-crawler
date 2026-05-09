package starlark

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/mini-crawler/pkg/model"
	"github.com/algolia/mini-crawler/pkg/script"
)

func TestProgramCanInspectDOM(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    title = doc.select_first("h1#page-title")
    description = doc.select_first("meta[name=description]")
    root = doc.select_first("#content")
    out = [{
        "url": doc.url,
        "title": text(title),
        "description": attr(description, "content"),
    }]
    for node in root.select("h2[id], span[data-as=p]"):
        out.append({
            "tag": node_name(node),
            "anchor": attr(node, "id"),
            "text": text(node),
        })
    return out

extract("^/doc", extract_docs)
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

	records, err := program.Extract(doc, script.Context{URL: "https://example.com/doc"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertRecordCount(t, records, 3)
	assertMapValue(t, records[0], "url", "https://example.com/doc")
	assertMapValue(t, records[0], "title", "Title")
	assertMapValue(t, records[0], "description", "Description")
	assertMapValue(t, records[1], "tag", "h2")
	assertMapValue(t, records[1], "anchor", "intro")
	assertMapValue(t, records[2], "text", "First paragraph")
}

func TestProgramCanUseDOMTraversalHelpers(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    item = doc.select_first("li")
    header = doc.select_first(".param-head")
    next_block = header.next("div.mt-4")
    return [{
        "has_parent": has_parent(doc.select_first("span"), "li"),
        "without_links": collapse_space(clone_without_text(item, "a")),
        "next_text": collapse_space(text(next_block.select_first("p"))),
    }]

extract("^/doc", extract_docs)
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

	records, err := program.Extract(doc, script.Context{URL: "https://example.com/doc"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertMapValue(t, records[0], "has_parent", true)
	assertMapValue(t, records[0], "without_links", "Keep me Child")
	assertMapValue(t, records[0], "next_text", "Description text")
}

func TestProgramCanUseDOMShortcutHelpers(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    root = doc.select_first("main")
    return [{
        "title": collapse_space(first_text(doc, "h1")),
        "nested": collapse_space(first_text(root, ".nested")),
        "description": first_attr(doc, "meta[name=description]", "content"),
        "missing_text": safe_text(doc.select_first(".missing")),
        "missing_attr": first_attr(doc, ".missing", "href"),
    }]

extract(".*", extract_docs)
`)

	program, err := Engine{}.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	doc := testDocument(t, `
<html>
  <head><meta name="description" content="Shortcut description"></head>
  <body><main><h1> Shortcut title </h1><p class="nested">Nested text</p></main></body>
</html>`)

	records, err := program.Extract(doc, script.Context{URL: "https://example.com/doc"})
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	assertMapValue(t, records[0], "title", "Shortcut title")
	assertMapValue(t, records[0], "nested", "Nested text")
	assertMapValue(t, records[0], "description", "Shortcut description")
	assertMapValue(t, records[0], "missing_text", "")
	assertMapValue(t, records[0], "missing_attr", nil)
}

func TestDOMHelpersRejectNonNode(t *testing.T) {
	path := writeScript(t, `
def extract_docs(pattern, doc, ctx):
    return [{"bad": text("not-node")}]

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

	if !strings.Contains(err.Error(), "text: want node, got string") {
		t.Fatalf("Extract() error = %q", err)
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
