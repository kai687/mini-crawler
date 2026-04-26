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

		if record.Product != "" {
			t.Fatalf("records[%d].Product = %q, want empty", i, record.Product)
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

func TestRecordEnricherAddsProductForUILibrariesAutocompleteDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/ui-libraries/autocomplete/core-concepts/templates",
		Description: stringPtr("Autocomplete docs"),
		PageHeading: stringPtr("Templates"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      "https://www.algolia.com/doc/ui-libraries/autocomplete/core-concepts/templates",
		product:  "autocomplete",
		typeName: model.RecordTypeLvl1,
		content:  "Autocomplete docs",
		lvl1:     stringPtr("Templates"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/ui-libraries/autocomplete/core-concepts/templates",
		),
	})
}

func TestRecordEnricherAddsProductForCrawlerToolDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/tools/crawler/getting-started",
		Description: stringPtr("Crawler docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      "https://www.algolia.com/doc/tools/crawler/getting-started",
		product:  "crawler",
		typeName: model.RecordTypeLvl1,
		content:  "Crawler docs",
		lvl1:     stringPtr("Getting started"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/tools/crawler/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForCLIToolDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/tools/cli/get-started/install",
		Description: stringPtr("CLI docs"),
		PageHeading: stringPtr("Install"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      "https://www.algolia.com/doc/tools/cli/get-started/install",
		product:  "cli",
		typeName: model.RecordTypeLvl1,
		content:  "CLI docs",
		lvl1:     stringPtr("Install"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/tools/cli/get-started/install",
		),
	})
}

func TestRecordEnricherAddsProductForShopifyIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/shopify/getting-started",
		Description: stringPtr("Shopify docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/shopify/getting-started",
		contentType: "integration",
		product:     "shopify",
		typeName:    model.RecordTypeLvl1,
		content:     "Shopify docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/shopify/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForBigCommerceIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/bigcommerce/getting-started",
		Description: stringPtr("BigCommerce docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/bigcommerce/getting-started",
		contentType: "integration",
		product:     "bigcommerce",
		typeName:    model.RecordTypeLvl1,
		content:     "BigCommerce docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/bigcommerce/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForCommercetoolsIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/commercetools/getting-started",
		Description: stringPtr("Commercetools docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/commercetools/getting-started",
		contentType: "integration",
		product:     "commercetools",
		typeName:    model.RecordTypeLvl1,
		content:     "Commercetools docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/commercetools/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForMagentoIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/magento-2/getting-started",
		Description: stringPtr("Magento docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/magento-2/getting-started",
		contentType: "integration",
		product:     "magento",
		typeName:    model.RecordTypeLvl1,
		content:     "Magento docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/magento-2/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForSFCCIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/salesforce-commerce-cloud-b2c/getting-started",
		Description: stringPtr("SFCC docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/salesforce-commerce-cloud-b2c/getting-started",
		contentType: "integration",
		product:     "sfcc",
		typeName:    model.RecordTypeLvl1,
		content:     "SFCC docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/salesforce-commerce-cloud-b2c/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForZendeskIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/zendesk/getting-started",
		Description: stringPtr("Zendesk docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/zendesk/getting-started",
		contentType: "integration",
		product:     "zendesk",
		typeName:    model.RecordTypeLvl1,
		content:     "Zendesk docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/zendesk/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForDjangoFrameworkIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/framework-integration/django/getting-started",
		Description: stringPtr("Django docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/framework-integration/django/getting-started",
		contentType: "sdk",
		product:     "django",
		typeName:    model.RecordTypeLvl1,
		content:     "Django docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/framework-integration/django/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForRailsFrameworkIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/framework-integration/rails/getting-started",
		Description: stringPtr("Rails docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/framework-integration/rails/getting-started",
		contentType: "sdk",
		product:     "rails",
		typeName:    model.RecordTypeLvl1,
		content:     "Rails docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/framework-integration/rails/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForSymfonyFrameworkIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/framework-integration/symfony/getting-started",
		Description: stringPtr("Symfony docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/framework-integration/symfony/getting-started",
		contentType: "sdk",
		product:     "symfony",
		typeName:    model.RecordTypeLvl1,
		content:     "Symfony docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/framework-integration/symfony/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForLaravelFrameworkIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/framework-integration/laravel/getting-started",
		Description: stringPtr("Laravel docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/framework-integration/laravel/getting-started",
		contentType: "sdk",
		product:     "laravel",
		typeName:    model.RecordTypeLvl1,
		content:     "Laravel docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/framework-integration/laravel/getting-started",
		),
	})
}

func TestRecordEnricherAddsProductForBigQueryConnectorDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/sending-and-managing-data/send-and-update-your-data/" +
		"connectors/bigquery/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("BigQuery connector docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      pageURL,
		product:  "bigquery",
		typeName: model.RecordTypeLvl1,
		content:  "BigQuery connector docs",
		lvl1:     stringPtr("Getting started"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForElasticsearchConnectorDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/sending-and-managing-data/send-and-update-your-data/" +
		"connectors/elasticsearch/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("Elasticsearch connector docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      pageURL,
		product:  "elasticsearch",
		typeName: model.RecordTypeLvl1,
		content:  "Elasticsearch connector docs",
		lvl1:     stringPtr("Getting started"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForMySQLConnectorDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/sending-and-managing-data/send-and-update-your-data/" +
		"connectors/mysql/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("MySQL connector docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      pageURL,
		product:  "mysql",
		typeName: model.RecordTypeLvl1,
		content:  "MySQL connector docs",
		lvl1:     stringPtr("Getting started"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForSupabaseConnectorDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/sending-and-managing-data/send-and-update-your-data/" +
		"connectors/supabase/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("Supabase connector docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:      pageURL,
		product:  "supabase",
		typeName: model.RecordTypeLvl1,
		content:  "Supabase connector docs",
		lvl1:     stringPtr("Getting started"),
		position: 0,
		objectID: recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForTealiumGuideDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/guides/sending-events/connectors/tealium/" +
		"getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("Tealium guide docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         pageURL,
		contentType: "guide",
		product:     "tealium",
		typeName:    model.RecordTypeLvl1,
		content:     "Tealium guide docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID:    recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForSegmentGuideDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/guides/sending-events/connectors/segment/" +
		"getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("Segment guide docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         pageURL,
		contentType: "guide",
		product:     "segment",
		typeName:    model.RecordTypeLvl1,
		content:     "Segment guide docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID:    recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForGoogleTagManagerGuideDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/guides/sending-events/connectors/" +
		"google-tag-manager/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("Google Tag Manager guide docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         pageURL,
		contentType: "guide",
		product:     "google-tag-manager",
		typeName:    model.RecordTypeLvl1,
		content:     "Google Tag Manager guide docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID:    recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForAgentStudioGuideDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/guides/algolia-ai/agent-studio/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("Agent Studio guide docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         pageURL,
		contentType: "guide",
		product:     "agent-studio",
		typeName:    model.RecordTypeLvl1,
		content:     "Agent Studio guide docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID:    recordutil.ObjectIDFromURL(pageURL),
	})
}

func TestRecordEnricherAddsProductForAskAIGuideDocURLs(t *testing.T) {
	const pageURL = "https://www.algolia.com/doc/guides/algolia-ai/askai/getting-started"

	doc := model.ExtractedDocument{
		PageURL:     pageURL,
		Description: stringPtr("AskAI guide docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         pageURL,
		contentType: "guide",
		product:     "askai",
		typeName:    model.RecordTypeLvl1,
		content:     "AskAI guide docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID:    recordutil.ObjectIDFromURL(pageURL),
	})
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

func TestRecordEnricherAddsMethodNameForSDKMethodDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/libraries/sdk/methods/search/list-api-keys",
		Description: stringPtr("List API keys docs"),
		PageHeading: stringPtr("List API keys"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/libraries/sdk/methods/search/list-api-keys",
		contentType: "sdk",
		typeName:    model.RecordTypeLvl1,
		content:     "List API keys docs",
		lvl1:        stringPtr("List API keys"),
		methodName:  "listApiKeys",
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/libraries/sdk/methods/search/list-api-keys",
		),
	})
}

func TestRecordEnricherAddsAPIContentTypeForAPIParameterDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/api-reference/api-parameters/hitsPerPage",
		Description: stringPtr("API parameter docs"),
		PageHeading: stringPtr("hitsPerPage"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/api-reference/api-parameters/hitsPerPage",
		contentType: "api",
		typeName:    model.RecordTypeLvl1,
		content:     "API parameter docs",
		lvl1:        stringPtr("hitsPerPage"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/api-reference/api-parameters/hitsPerPage",
		),
	})
}

func TestRecordEnricherAddsIntegrationContentTypeForIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/integration/shopify/getting-started",
		Description: stringPtr("Integration docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/integration/shopify/getting-started",
		contentType: "integration",
		product:     "shopify",
		typeName:    model.RecordTypeLvl1,
		content:     "Integration docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/integration/shopify/getting-started",
		),
	})
}

func TestRecordEnricherAddsSDKContentTypeForFrameworkIntegrationDocURLs(t *testing.T) {
	doc := model.ExtractedDocument{
		PageURL:     "https://www.algolia.com/doc/framework-integration/react/getting-started",
		Description: stringPtr("Framework integration docs"),
		PageHeading: stringPtr("Getting started"),
	}

	records, err := RecordEnricher{}.Enrich(doc)
	if err != nil {
		t.Fatalf("Enrich() err = %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	assertRecord(t, records[0], recordExpectation{
		url:         "https://www.algolia.com/doc/framework-integration/react/getting-started",
		contentType: "sdk",
		typeName:    model.RecordTypeLvl1,
		content:     "Framework integration docs",
		lvl1:        stringPtr("Getting started"),
		position:    0,
		objectID: recordutil.ObjectIDFromURL(
			"https://www.algolia.com/doc/framework-integration/react/getting-started",
		),
	})
}

type recordExpectation struct {
	url         string
	contentType string
	product     string
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
	assertEqual(t, "Product", record.Product, want.product)

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
