package enrich

import (
	"testing"

	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/recordutil"
)

func TestRecordEnricherBuildsFinalRecords(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://example.com/page",
		Description: stringPtr("Line one line two"),
		PageHeading: stringPtr("Page Title"),
		Units: []model.ExtractedUnit{
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "first-section",
				Text:         "First section",
				HeadingLevel: 2,
				Position:     1,
			},
			{Kind: model.ExtractedContent, Text: "First paragraph", Position: 2},
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "first-detail",
				Text:         "First detail",
				HeadingLevel: 3,
				Position:     3,
			},
			{Kind: model.ExtractedContent, Text: "Second paragraph", Position: 4},
			{Kind: model.ExtractedContent, Text: "First bullet", Position: 5},
			{Kind: model.ExtractedContent, Text: "Second bullet", Position: 6},
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "first-subdetail",
				Text:         "First subdetail",
				HeadingLevel: 4,
				Position:     7,
			},
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "second-section",
				Text:         "Second section",
				HeadingLevel: 2,
				Position:     8,
			},
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "second-detail",
				Text:         "Second detail",
				HeadingLevel: 3,
				Position:     9,
			},
		},
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 10 {
		t.Fatalf("len(records) = %d, want 10", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      "https://example.com/page",
		typeName: model.RecordTypeLvl1,
		content:  "Line one line two",
		lvl1:     stringPtr("Page Title"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL("https://example.com/page"),
	})

	firstSectionURL := recordutil.URLWithAnchor("https://example.com/page", "first-section")
	assertRecord(t, records[1], recordExpectation{
		url:      firstSectionURL,
		typeName: model.RecordTypeLvl2,
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		position: 1,
		objectID: recordutil.ObjectIDFromURL(firstSectionURL),
	})

	assertRecord(t, records[2], recordExpectation{
		url:      firstSectionURL,
		typeName: model.RecordTypeContent,
		content:  "First paragraph",
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		position: 2,
		objectID: recordutil.ObjectIDWithPosition(
			recordutil.ObjectIDFromURL(firstSectionURL),
			2,
		),
	})

	firstDetailURL := recordutil.URLWithAnchor("https://example.com/page", "first-detail")
	assertRecord(t, records[3], recordExpectation{
		url:      firstDetailURL,
		typeName: model.RecordTypeLvl3,
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		lvl3:     stringPtr("First detail"),
		position: 3,
		objectID: recordutil.ObjectIDFromURL(firstDetailURL),
	})

	assertRecord(t, records[4], recordExpectation{
		url:      firstDetailURL,
		typeName: model.RecordTypeContent,
		content:  "Second paragraph",
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		lvl3:     stringPtr("First detail"),
		position: 4,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(firstDetailURL), 4),
	})

	assertRecord(t, records[5], recordExpectation{
		url:      firstDetailURL,
		typeName: model.RecordTypeContent,
		content:  "First bullet",
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		lvl3:     stringPtr("First detail"),
		position: 5,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(firstDetailURL), 5),
	})

	assertRecord(t, records[6], recordExpectation{
		url:      firstDetailURL,
		typeName: model.RecordTypeContent,
		content:  "Second bullet",
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		lvl3:     stringPtr("First detail"),
		position: 6,
		objectID: recordutil.ObjectIDWithPosition(recordutil.ObjectIDFromURL(firstDetailURL), 6),
	})

	assertRecord(t, records[7], recordExpectation{
		url:      recordutil.URLWithAnchor("https://example.com/page", "first-subdetail"),
		typeName: model.RecordTypeLvl4,
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("First section"),
		lvl3:     stringPtr("First detail"),
		lvl4:     stringPtr("First subdetail"),
		position: 7,
		objectID: recordutil.ObjectIDFromURL(
			recordutil.URLWithAnchor("https://example.com/page", "first-subdetail"),
		),
	})

	assertRecord(t, records[8], recordExpectation{
		url:      recordutil.URLWithAnchor("https://example.com/page", "second-section"),
		typeName: model.RecordTypeLvl2,
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("Second section"),
		position: 8,
		objectID: recordutil.ObjectIDFromURL(
			recordutil.URLWithAnchor("https://example.com/page", "second-section"),
		),
	})

	assertRecord(t, records[9], recordExpectation{
		url:      recordutil.URLWithAnchor("https://example.com/page", "second-detail"),
		typeName: model.RecordTypeLvl3,
		lvl1:     stringPtr("Page Title"),
		lvl2:     stringPtr("Second section"),
		lvl3:     stringPtr("Second detail"),
		position: 9,
		objectID: recordutil.ObjectIDFromURL(
			recordutil.URLWithAnchor("https://example.com/page", "second-detail"),
		),
	})
}

func TestRecordEnricherComputesBreadcrumbsAndContentType(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/guides/building-search/intro",
		Description: stringPtr("Guide summary"),
		PageHeading: stringPtr("Guide Title"),
		Units: []model.ExtractedUnit{
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "section",
				Text:         "Section",
				HeadingLevel: 2,
				Position:     1,
			},
			{Kind: model.ExtractedContent, Text: "Paragraph", Position: 2},
		},
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	for i, record := range records {
		if record.ContentType != "guide" {
			t.Fatalf("records[%d].ContentType = %q, want %q", i, record.ContentType, "guide")
		}

		if len(record.BreadcrumbSegments) != 2 {
			t.Fatalf(
				"len(records[%d].BreadcrumbSegments) = %d, want 2",
				i,
				len(record.BreadcrumbSegments),
			)
		}

		assertStringPtr(
			t,
			"BreadcrumbHierarchy.Lvl0",
			record.BreadcrumbHierarchy.Lvl0,
			stringPtr("Guides"),
		)
		assertStringPtr(
			t,
			"BreadcrumbHierarchy.Lvl1",
			record.BreadcrumbHierarchy.Lvl1,
			stringPtr("Guides > Building search"),
		)

		if record.BreadcrumbHierarchy.Lvl2 != nil {
			t.Fatalf("BreadcrumbHierarchy.Lvl2 = %v, want nil", record.BreadcrumbHierarchy.Lvl2)
		}
	}
}

func TestRecordEnricherIndexesAPIFieldRecords(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://example.com/page",
		Description: stringPtr("Search endpoint docs"),
		PageHeading: stringPtr("Search single index"),
		Units: []model.ExtractedUnit{
			{
				Kind:         model.ExtractedHeading,
				Anchor:       "path-parameters",
				Text:         "Path parameters",
				HeadingLevel: 2,
				Position:     1,
			},
			{
				Kind:        model.ExtractedField,
				Anchor:      "parameter-index-name",
				Text:        "indexName",
				Description: "string. required. Name of the index.",
				Position:    2,
			},
		},
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("len(records) = %d, want 3", len(records))
	}

	indexNameURL := recordutil.URLWithAnchor("https://example.com/page", "parameter-index-name")
	assertRecord(t, records[2], recordExpectation{
		url:      indexNameURL,
		typeName: model.RecordTypeField,
		content:  "string. required. Name of the index.",
		lvl1:     stringPtr("Search single index"),
		lvl2:     stringPtr("Path parameters"),
		lvl3:     stringPtr("indexName"),
		position: 2,
		objectID: recordutil.ObjectIDFromURL(indexNameURL),
	})
}

func TestRecordEnricherAddsMethodNameForRESTAPIDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/rest-api/search/search-single-index",
		Description: stringPtr("Search endpoint docs"),
		PageHeading: stringPtr("Search an index"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/rest-api/search/search-single-index",
		contentType: "api",
		typeName:    model.RecordTypeLvl1,
		content:     "Search endpoint docs",
		lvl1:        stringPtr("Search an index"),
		methodName:  "searchSingleIndex",
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/rest-api/search/search-single-index",
		),
	})
}

type recordExpectation struct {
	url         string
	contentType string
	typeName    model.RecordType
	content     string
	methodName  string
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
	assertEqual(t, "ContentType", record.ContentType, want.contentType)

	if record.RecordType != want.typeName {
		t.Fatalf("RecordType = %q, want %q", record.RecordType, want.typeName)
	}

	assertStringPtr(t, "Content", record.Content, stringPtr(want.content))
	assertEqual(t, "MethodName", record.MethodName, want.methodName)
	assertStringPtr(t, "Hierarchy.Lvl1", record.Hierarchy.Lvl1, want.lvl1)
	assertStringPtr(t, "Hierarchy.Lvl2", record.Hierarchy.Lvl2, want.lvl2)
	assertStringPtr(t, "Hierarchy.Lvl3", record.Hierarchy.Lvl3, want.lvl3)
	assertStringPtr(t, "Hierarchy.Lvl4", record.Hierarchy.Lvl4, want.lvl4)
	assertEqual(t, "ObjectID", record.ObjectID, want.objectID)

	if record.Position != want.position {
		t.Fatalf("Position = %d, want %d", record.Position, want.position)
	}
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
