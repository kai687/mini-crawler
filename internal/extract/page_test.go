package extract

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/recordutil"
)

func TestPageExtractorBuildsPageHeadingParagraphAndListItemRecords(t *testing.T) {
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

	records, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 10 {
		t.Fatalf("len(records) = %d, want 10", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://example.com/page",
		typeName:    "lvl1",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		lvl1:        stringPtr("Page Title"),
		position:    0,
		objectID:    recordutil.ObjectIDFromURL("https://example.com/page"),
	})

	firstSectionURL := recordutil.URLWithAnchor("https://example.com/page", "first-section")
	assertRecord(t, records[1], recordExpectation{
		url:         firstSectionURL,
		typeName:    "lvl2",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		position:    1,
		objectID:    recordutil.ObjectIDFromURL(firstSectionURL),
	})

	assertRecord(t, records[2], recordExpectation{
		url:         firstSectionURL,
		typeName:    "content",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		content:     "First paragraph",
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		position:    2,
		objectID: recordutil.ObjectIDWithPosition(
			recordutil.ObjectIDFromURL(firstSectionURL),
			2,
		),
	})

	firstDetailURL := recordutil.URLWithAnchor("https://example.com/page", "first-detail")
	assertRecord(t, records[3], recordExpectation{
		url:         firstDetailURL,
		typeName:    "lvl3",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		lvl3:        stringPtr("First detail"),
		position:    3,
		objectID:    recordutil.ObjectIDFromURL(firstDetailURL),
	})

	assertRecord(t, records[4], recordExpectation{
		url:         firstDetailURL,
		typeName:    "content",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		content:     "Second paragraph",
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		lvl3:        stringPtr("First detail"),
		position:    4,
		objectID:    recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(firstDetailURL), 4),
	})

	assertRecord(t, records[5], recordExpectation{
		url:         firstDetailURL,
		typeName:    "content",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		content:     "First bullet",
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		lvl3:        stringPtr("First detail"),
		position:    5,
		objectID:    recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(firstDetailURL), 5),
	})

	assertRecord(t, records[6], recordExpectation{
		url:         firstDetailURL,
		typeName:    "content",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		content:     "Second bullet",
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		lvl3:        stringPtr("First detail"),
		position:    6,
		objectID:    recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(firstDetailURL), 6),
	})

	assertRecord(t, records[7], recordExpectation{
		url:         recordutil.URLWithAnchor("https://example.com/page", "first-subdetail"),
		typeName:    "lvl4",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("First section"),
		lvl3:        stringPtr("First detail"),
		lvl4:        stringPtr("First subdetail"),
		position:    7,
		objectID: recordutil.ObjectIDFromURL(
			recordutil.URLWithAnchor("https://example.com/page", "first-subdetail"),
		),
	})

	assertRecord(t, records[8], recordExpectation{
		url:         recordutil.URLWithAnchor("https://example.com/page", "second-section"),
		typeName:    "lvl2",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("Second section"),
		position:    8,
		objectID: recordutil.ObjectIDFromURL(
			recordutil.URLWithAnchor("https://example.com/page", "second-section"),
		),
	})

	assertRecord(t, records[9], recordExpectation{
		url:         recordutil.URLWithAnchor("https://example.com/page", "second-detail"),
		typeName:    "lvl3",
		title:       stringPtr("Page Title"),
		description: stringPtr("Line one line two"),
		lvl1:        stringPtr("Page Title"),
		lvl2:        stringPtr("Second section"),
		lvl3:        stringPtr("Second detail"),
		position:    9,
		objectID: recordutil.ObjectIDFromURL(
			recordutil.URLWithAnchor("https://example.com/page", "second-detail"),
		),
	})
}

func TestPageExtractorSetsGuideContentTypeFromURL(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title">Guide Title</h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<span data-as="p">Paragraph</span>`+
			`</div></body></html>`,
	)

	records, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://www.algolia.com/doc/guides/building-search/intro",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	for i, record := range records {
		if record.ContentType != "guide" {
			t.Fatalf("records[%d].ContentType = %q, want %q", i, record.ContentType, "guide")
		}

		if record.Breadcrumb != "/guides/building-search/intro" {
			t.Fatalf(
				"records[%d].Breadcrumb = %q, want %q",
				i,
				record.Breadcrumb,
				"/guides/building-search/intro",
			)
		}
	}
}

func TestPageExtractorSetsAPIContentTypeFromURL(t *testing.T) {
	doc := mustDocument(t, `<html><body><h1 id="page-title">API Title</h1></body></html>`)

	records, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://www.algolia.com/doc/rest-api/search/search-single-index",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if got := records[0].ContentType; got != "api" {
		t.Fatalf("records[0].ContentType = %q, want %q", got, "api")
	}

	if got := records[0].Breadcrumb; got != "/rest-api/search/search-single-index" {
		t.Fatalf("records[0].Breadcrumb = %q, want %q", got, "/rest-api/search/search-single-index")
	}
}

func TestPageExtractorFallsBackToTitleTag(t *testing.T) {
	doc := mustDocument(t, `<html><head><title> Doc Title </title></head><body></body></html>`)

	records, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://example.com/page",
		typeName:    "lvl1",
		title:       stringPtr("Doc Title"),
		description: nil,
		position:    0,
		objectID:    recordutil.ObjectIDFromURL("https://example.com/page"),
	})
}

func TestPageExtractorContentInheritsParentHierarchy(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="first-section">First section</h2>`+
			`<h3 id="first-detail">First detail</h3>`+
			`<span data-as="p">Paragraph</span>`+
			`</div></body></html>`,
	)

	records, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 4 {
		t.Fatalf("len(records) = %d, want 4", len(records))
	}

	assertStringPtr(t, "Hierarchy.Lvl2", records[3].Hierarchy.Lvl2, stringPtr("First section"))
	assertStringPtr(t, "Hierarchy.Lvl3", records[3].Hierarchy.Lvl3, stringPtr("First detail"))
	assertStringPtr(t, "Content", records[3].Content, stringPtr("Paragraph"))
}

func TestPageExtractorListItemIndexedAsContent(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<ul><li>First item</li></ul>`+
			`</div></body></html>`,
	)

	records, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}

	sectionURL := recordutil.URLWithAnchor("https://example.com/page", "section")
	assertRecord(t, records[2], recordExpectation{
		url:      sectionURL,
		typeName: "content",
		title:    stringPtr("Page Title"),
		content:  "First item",
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("Section"),
		position: 2,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(sectionURL), 2),
	})
}

func TestPageExtractorListItemWithParagraphIndexedOnce(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<ul><li><span data-as="p">Bullet text</span></li></ul>`+
			`</div></body></html>`,
	)

	records, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}

	if records[2].Content == nil || *records[2].Content != "Bullet text" {
		t.Fatalf("records[2].Content = %v, want %q", records[2].Content, "Bullet text")
	}
}

func TestPageExtractorSkipsLinkOnlyListItem(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title"> Page Title </h1><div id="content">`+
			`<h2 id="section">Section</h2>`+
			`<ul><li><a href="/docs/ref">Reference</a></li></ul>`+
			`</div></body></html>`,
	)

	records, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
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

	records, err := PageExtractor{}.Extract(
		model.ParsedPage{URL: "https://example.com/page", Doc: doc},
	)
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}

	sectionURL := recordutil.URLWithAnchor("https://example.com/page", "section")
	assertRecord(t, records[2], recordExpectation{
		url:      sectionURL,
		typeName: "content",
		title:    stringPtr("Page Title"),
		content:  "Read Reference first",
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("Section"),
		position: 2,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(sectionURL), 2),
	})
}

func TestPageExtractorIndexesAPIParameterField(t *testing.T) {
	doc := mustDocument(
		t,
		`<html><body><h1 id="page-title">Search single index</h1><div id="content">`+
			`<h2 id="path-parameters">Path parameters</h2>`+
			`<div class="primitive-param-field"><div class="py-6">`+
			`<div class="flex font-mono param-head" id="parameter-index-name">`+
			`<div><div><a href="#parameter-index-name">link</a></div></div>`+
			`<div data-component-part="field-name">indexName</div>`+
			`<div data-component-part="field-meta">`+
			`<div data-component-part="field-info-pill"><span>string</span></div>`+
			`<div data-component-part="field-required-pill">required</div>`+
			`</div>`+
			`</div>`+
			`<div class="mt-4">`+
			`<div class="prose"><p class="whitespace-pre-line">Name of the index.</p></div>`+
			`<div class="flex">Example: <code>"X"</code></div>`+
			`</div>`+
			`</div></div>`+
			`</div></body></html>`,
	)

	records, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 4 {
		t.Fatalf("len(records) = %d, want 4", len(records))
	}

	pathSectionURL := recordutil.URLWithAnchor("https://example.com/page", "path-parameters")
	indexNameURL := recordutil.URLWithAnchor("https://example.com/page", "parameter-index-name")

	assertRecord(t, records[2], recordExpectation{
		url:      indexNameURL,
		typeName: "lvl3",
		title:    stringPtr("Search single index"),
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Path parameters"),
		lvl3:     stringPtr("indexName"),
		position: 2,
		objectID: recordutil.ObjectIDFromURL(indexNameURL),
	})

	assertRecord(t, records[3], recordExpectation{
		url:      indexNameURL,
		typeName: "content",
		title:    stringPtr("Search single index"),
		content:  "string. required. Name of the index.",
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Path parameters"),
		lvl3:     stringPtr("indexName"),
		position: 3,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(indexNameURL), 3),
	})

	_ = pathSectionURL
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

	records, err := PageExtractor{}.Extract(model.ParsedPage{
		URL: "https://example.com/page",
		Doc: doc,
	})
	if err != nil {
		t.Fatalf("Extract() err = %v", err)
	}

	if len(records) != 10 {
		t.Fatalf("len(records) = %d, want 10", len(records))
	}

	authURL := recordutil.URLWithAnchor(
		"https://example.com/page",
		"authorization-x-algolia-api-key",
	)
	bodyURL := recordutil.URLWithAnchor("https://example.com/page", "body-one-of-0-params")
	responseURL := recordutil.URLWithAnchor("https://example.com/page", "response-hits")

	assertRecord(t, records[2], recordExpectation{
		url:      authURL,
		typeName: "lvl3",
		title:    stringPtr("Search single index"),
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Authorizations"),
		lvl3:     stringPtr("x-algolia-api-key"),
		position: 2,
		objectID: recordutil.ObjectIDFromURL(authURL),
	})

	assertRecord(t, records[3], recordExpectation{
		url:      authURL,
		typeName: "content",
		title:    stringPtr("Search single index"),
		content:  "string. required. API key.",
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Authorizations"),
		lvl3:     stringPtr("x-algolia-api-key"),
		position: 3,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(authURL), 3),
	})

	assertRecord(t, records[5], recordExpectation{
		url:      bodyURL,
		typeName: "lvl3",
		title:    stringPtr("Search single index"),
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Body"),
		lvl3:     stringPtr("params"),
		position: 5,
		objectID: recordutil.ObjectIDFromURL(bodyURL),
	})

	assertRecord(t, records[6], recordExpectation{
		url:      bodyURL,
		typeName: "content",
		title:    stringPtr("Search single index"),
		content:  "string. Search parameters.",
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Body"),
		lvl3:     stringPtr("params"),
		position: 6,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(bodyURL), 6),
	})

	assertRecord(t, records[8], recordExpectation{
		url:      responseURL,
		typeName: "lvl3",
		title:    stringPtr("Search single index"),
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Response"),
		lvl3:     stringPtr("hits"),
		position: 8,
		objectID: recordutil.ObjectIDFromURL(responseURL),
	})

	assertRecord(t, records[9], recordExpectation{
		url:      responseURL,
		typeName: "content",
		title:    stringPtr("Search single index"),
		content:  "object[]. required. Search results.",
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Response"),
		lvl3:     stringPtr("hits"),
		position: 9,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(responseURL), 9),
	})
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

func TestObjectIDWithPositionUniqueForParagraphs(t *testing.T) {
	base := recordutil.ObjectIDFromURL("https://example.com/page#section")
	if got := recordutil.ObjectIDWithPosition(base, 2); got != base+"-2" {
		t.Fatalf("ObjectIDWithPosition() = %q", got)
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

func TestCloneHierarchy(t *testing.T) {
	lvl1 := stringPtr("Page Title")

	h := cloneHierarchy(model.Hierarchy{Lvl1: lvl1})
	if h.Lvl1 == nil || *h.Lvl1 != "Page Title" {
		t.Fatalf("cloneHierarchy().Lvl1 = %v", h.Lvl1)
	}

	if h.Lvl1 == lvl1 {
		t.Fatal("cloneHierarchy() reused pointer")
	}
}

type recordExpectation struct {
	url         string
	breadcrumb  string
	contentType string
	typeName    model.RecordType
	title       *string
	description *string
	content     string
	lvl1        *string
	lvl2        *string
	lvl3        *string
	lvl4        *string
	position    int
	objectID    string
}

func assertRecord(t *testing.T, record model.Record, want recordExpectation) {
	t.Helper()

	assertEqual(t, "URL", record.URL, want.url)

	if want.breadcrumb != "" {
		assertEqual(t, "Breadcrumb", record.Breadcrumb, want.breadcrumb)
	}

	assertEqual(t, "ContentType", record.ContentType, want.contentType)
	assertEqual(t, "RecordType", string(record.RecordType), string(want.typeName))
	assertStringPtr(t, "Title", record.Title, want.title)
	assertStringPtr(t, "Description", record.Description, want.description)
	assertStringPtr(t, "Content", record.Content, stringPtr(want.content))
	assertStringPtr(t, "Hierarchy.Lvl1", record.Hierarchy.Lvl1, want.lvl1)
	assertStringPtr(t, "Hierarchy.Lvl2", record.Hierarchy.Lvl2, want.lvl2)
	assertStringPtr(t, "Hierarchy.Lvl3", record.Hierarchy.Lvl3, want.lvl3)
	assertStringPtr(t, "Hierarchy.Lvl4", record.Hierarchy.Lvl4, want.lvl4)
	assertEqual(t, "ObjectID", record.ObjectID, want.objectID)
	assertPosition(t, record.Position, want.position)
}

func assertEqual(t *testing.T, name string, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
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

func assertPosition(t *testing.T, got int, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("Position = %d, want %d", got, want)
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
