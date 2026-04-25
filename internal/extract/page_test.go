package extract

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
)

func TestPageExtractorBuildsExtractedDocument(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><head><title>Head Title</title><meta name="description" content="  Line one

  line two  "></head><body>`+
			`<h1 id="page-title"> Page Title </h1>`+
			`<div id="content">`+
			`<h2 id="first-section"> First section </h2>`+
			`<h2>Ignored card title</h2>`+
			`<span data-as="p"> First
 paragraph </span>`+
			`<h3>Ignored card detail</h3>`+
			`<h3 id="first-detail"> First detail </h3>`+
			`<span data-as="p"> Second paragraph </span>`+
			`<ul><li> First bullet </li><li> Second bullet </li></ul>`+
			`<h4 id="first-subdetail"> First subdetail </h4>`+
			`<h2 id="second-section">Second section</h2>`+
			`<h3 id="second-detail">Second detail</h3>`+
			`</div>`+
			`<span data-as="p">Ignored outside paragraph</span>`+
			`<li>Ignored outside list item</li>`+
			`</body></html>`,
	)

	extracted, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	assertStringPtr(t, "Description", extracted.Description, stringPtr("Line one line two"))
	assertStringPtr(t, "PageHeading", extracted.PageHeading, stringPtr("Page Title"))

	if len(extracted.Units) != 9 {
		t.Fatalf("len(extracted.Units) = %d, want 9", len(extracted.Units))
	}

	assertUnit(
		t,
		extracted.Units[0],
		unitExpectation{
			kind:         model.ExtractedHeading,
			anchor:       "first-section",
			text:         "First section",
			headingLevel: 2,
			position:     1,
		},
	)
	assertUnit(
		t,
		extracted.Units[1],
		unitExpectation{kind: model.ExtractedContent, text: "First paragraph", position: 2},
	)
	assertUnit(
		t,
		extracted.Units[2],
		unitExpectation{
			kind:         model.ExtractedHeading,
			anchor:       "first-detail",
			text:         "First detail",
			headingLevel: 3,
			position:     3,
		},
	)
	assertUnit(
		t,
		extracted.Units[3],
		unitExpectation{kind: model.ExtractedContent, text: "Second paragraph", position: 4},
	)
	assertUnit(
		t,
		extracted.Units[4],
		unitExpectation{kind: model.ExtractedContent, text: "First bullet", position: 5},
	)
	assertUnit(
		t,
		extracted.Units[5],
		unitExpectation{kind: model.ExtractedContent, text: "Second bullet", position: 6},
	)
	assertUnit(
		t,
		extracted.Units[6],
		unitExpectation{
			kind:         model.ExtractedHeading,
			anchor:       "first-subdetail",
			text:         "First subdetail",
			headingLevel: 4,
			position:     7,
		},
	)
	assertUnit(
		t,
		extracted.Units[7],
		unitExpectation{
			kind:         model.ExtractedHeading,
			anchor:       "second-section",
			text:         "Second section",
			headingLevel: 2,
			position:     8,
		},
	)
	assertUnit(
		t,
		extracted.Units[8],
		unitExpectation{
			kind:         model.ExtractedHeading,
			anchor:       "second-detail",
			text:         "Second detail",
			headingLevel: 3,
			position:     9,
		},
	)
}

func TestPageExtractorWithoutPageHeading(t *testing.T) {
	doc := mustDocument(t, `<html><head><title> Doc Title </title></head><body></body></html>`)

	extracted, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	assertStringPtr(t, "Description", extracted.Description, nil)
	assertStringPtr(t, "PageHeading", extracted.PageHeading, nil)

	if len(extracted.Units) != 0 {
		t.Fatalf("len(extracted.Units) = %d, want 0", len(extracted.Units))
	}
}

func TestPageExtractorListItemWithParagraphIndexedOnce(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<ul><li><span data-as="p">Bullet text</span></li></ul>`+
			`</div></body></html>`,
	)

	extracted, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(extracted.Units) != 2 {
		t.Fatalf("len(extracted.Units) = %d, want 2", len(extracted.Units))
	}

	assertUnit(
		t,
		extracted.Units[1],
		unitExpectation{kind: model.ExtractedContent, text: "Bullet text", position: 2},
	)
}

func TestPageExtractorSkipsLinkOnlyListItem(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<ul><li><a href="/docs/ref">Reference</a></li></ul>`+
			`</div></body></html>`,
	)

	extracted, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(extracted.Units) != 1 {
		t.Fatalf("len(extracted.Units) = %d, want 1", len(extracted.Units))
	}
}

func TestPageExtractorKeepsListItemWithNonLinkText(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<ul><li>Read <a href="/docs/ref">Reference</a> first</li></ul>`+
			`</div></body></html>`,
	)

	extracted, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(extracted.Units) != 2 {
		t.Fatalf("len(extracted.Units) = %d, want 2", len(extracted.Units))
	}

	assertUnit(
		t,
		extracted.Units[1],
		unitExpectation{kind: model.ExtractedContent, text: "Read Reference first", position: 2},
	)
}

func TestPageExtractorIndexesAPIParamHeadAcrossSections(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title">Search single index</h1><div id="content">`+
			`<h2 id="auth">Authorizations</h2>`+
			`<div class="param-head" id="authorization-x-algolia-api-key">`+
			`<div data-component-part="field-name">x-algolia-api-key</div>`+
			`<div data-component-part="field-info-pill">string</div>`+
			`<div data-component-part="field-required-pill">required</div>`+
			`</div>`+
			`<div class="mt-4"><div><p>API key.</p></div></div>`+
			`<h2 id="body">Body</h2>`+
			`<div class="param-head" id="body-one-of-0-params">`+
			`<div data-component-part="field-name">params</div>`+
			`<div data-component-part="field-info-pill">string</div>`+
			`</div>`+
			`<div class="mt-4"><div><p>Search parameters.</p></div></div>`+
			`<h2 id="response">Response</h2>`+
			`<div class="param-head" id="response-hits">`+
			`<div data-component-part="field-name">hits</div>`+
			`<div data-component-part="field-info-pill">object[]</div>`+
			`<div data-component-part="field-required-pill">required</div>`+
			`</div>`+
			`<div class="mt-4"><div><p>Search results.</p></div></div>`+
			`</div></body></html>`,
	)

	extracted, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(extracted.Units) != 6 {
		t.Fatalf("len(extracted.Units) = %d, want 6", len(extracted.Units))
	}

	assertUnit(
		t,
		extracted.Units[1],
		unitExpectation{
			kind:        model.ExtractedField,
			anchor:      "authorization-x-algolia-api-key",
			text:        "x-algolia-api-key",
			description: "string. required. API key.",
			position:    2,
		},
	)
	assertUnit(
		t,
		extracted.Units[3],
		unitExpectation{
			kind:        model.ExtractedField,
			anchor:      "body-one-of-0-params",
			text:        "params",
			description: "string. Search parameters.",
			position:    4,
		},
	)
	assertUnit(
		t,
		extracted.Units[5],
		unitExpectation{
			kind:        model.ExtractedField,
			anchor:      "response-hits",
			text:        "hits",
			description: "object[]. required. Search results.",
			position:    6,
		},
	)
}

func TestShouldIndexListItem(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><div id="content"><ul>`+
			`<li><a href="/x">Only link</a></li>`+
			`<li>Text <a href="/x">with link</a></li>`+
			`</ul></div></body></html>`,
	)
	items := doc.Find("#content li")

	if shouldIndexListItem(items.First()) {
		t.Fatal("shouldIndexListItem(link-only) = true, want false")
	}

	if !shouldIndexListItem(items.Last()) {
		t.Fatal("shouldIndexListItem(text+link) = false, want true")
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	got := normalizeWhitespace("  one\n\n two\t three  ")

	want := "one two three"
	if got != want {
		t.Fatalf("normalizeWhitespace() = %q, want %q", got, want)
	}
}

func TestStringPtr(t *testing.T) {
	if got := stringPtr(""); got != nil {
		t.Fatalf("stringPtr(\"\") = %v, want nil", got)
	}

	got := stringPtr("value")
	if got == nil || *got != "value" {
		t.Fatalf("stringPtr(\"value\") = %v", got)
	}
}

type unitExpectation struct {
	kind         model.ExtractedKind
	anchor       string
	text         string
	description  string
	headingLevel int
	position     int
}

func assertUnit(t *testing.T, got model.ExtractedUnit, want unitExpectation) {
	t.Helper()

	if got.Kind != want.kind {
		t.Fatalf("Kind = %q, want %q", got.Kind, want.kind)
	}

	if got.Anchor != want.anchor {
		t.Fatalf("Anchor = %q, want %q", got.Anchor, want.anchor)
	}

	if got.Text != want.text {
		t.Fatalf("Text = %q, want %q", got.Text, want.text)
	}

	if got.Description != want.description {
		t.Fatalf("Description = %q, want %q", got.Description, want.description)
	}

	if got.HeadingLevel != want.headingLevel {
		t.Fatalf("HeadingLevel = %d, want %d", got.HeadingLevel, want.headingLevel)
	}

	if got.Position != want.position {
		t.Fatalf("Position = %d, want %d", got.Position, want.position)
	}
}

func assertStringPtr(t *testing.T, name string, got *string, want *string) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Fatalf("%s = %v, want nil", name, *got)
		}

		return
	}

	if got == nil {
		t.Fatalf("%s = nil, want %q", name, *want)
	}

	if *got != *want {
		t.Fatalf("%s = %q, want %q", name, *got, *want)
	}
}

func mustDocument(t *testing.T, html string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() err = %v", err)
	}

	return doc
}
