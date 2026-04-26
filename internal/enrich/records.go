package enrich

import (
	"fmt"
	"strings"

	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/recordutil"
)

// RecordEnricher converts extracted page facts into final index records.
type RecordEnricher struct{}

// Enrich builds index-ready records from one extracted page document.
func (e RecordEnricher) Enrich(doc model.ExtractedDocument) ([]model.Record, error) {
	pageRecord := newPageRecord(doc)
	pageURL := pageRecord.URL

	records := []model.Record{pageRecord}
	currentHierarchy := cloneHierarchy(pageRecord.Hierarchy)
	currentURL := pageURL

	for _, unit := range doc.Units {
		switch unit.Kind {
		case model.ExtractedHeading:
			record, nextHierarchy, nextURL, err := headingRecord(
				pageRecord,
				currentHierarchy,
				doc.PageURL,
				unit,
			)
			if err != nil {
				return nil, err
			}

			records = append(records, record)
			currentHierarchy = nextHierarchy
			currentURL = nextURL
		case model.ExtractedField:
			record, nextHierarchy, nextURL := fieldRecord(
				pageRecord,
				currentHierarchy,
				doc.PageURL,
				unit,
			)
			records = append(records, record)
			currentHierarchy = nextHierarchy
			currentURL = nextURL
		case model.ExtractedContent:
			records = append(records, contentRecord(pageRecord, currentHierarchy, currentURL, unit))
		default:
			return nil, fmt.Errorf("unsupported extracted kind %q", unit.Kind)
		}
	}

	return records, nil
}

func newPageRecord(doc model.ExtractedDocument) model.Record {
	pageURL := recordutil.URLWithoutAnchor(doc.PageURL)
	breadcrumbSegments := recordutil.BreadcrumbSegmentsFromURL(pageURL)

	return model.Record{
		URL:                 pageURL,
		URLWithoutAnchor:    pageURL,
		BreadcrumbSegments:  breadcrumbSegments,
		BreadcrumbHierarchy: recordutil.BreadcrumbHierarchyFromSegments(breadcrumbSegments),
		ContentType:         contentTypeFromURL(pageURL),
		Product:             productFromURL(pageURL),
		MethodName:          recordutil.MethodNameFromURL(pageURL),
		RecordType:          model.RecordTypeLvl1,
		Content:             cloneStringPtr(doc.Description),
		Hierarchy: model.Hierarchy{
			Lvl1: cloneStringPtr(doc.PageHeading),
		},
		Position: 0,
		ObjectID: recordutil.ObjectIDFromURL(pageURL),
	}
}

func headingRecord(
	pageRecord model.Record,
	currentHierarchy model.Hierarchy,
	pageURL string,
	unit model.ExtractedUnit,
) (model.Record, model.Hierarchy, string, error) {
	recordType := typeFromLevel(unit.HeadingLevel)
	if recordType == "" {
		return model.Record{}, currentHierarchy, pageRecord.URL, fmt.Errorf(
			"unsupported heading level %d",
			unit.HeadingLevel,
		)
	}

	record := pageRecord
	record.RecordType = recordType
	record.Content = nil
	record.Hierarchy = cloneHierarchy(currentHierarchy)
	setHierarchyLevel(&record.Hierarchy, unit.HeadingLevel, stringPtr(unit.Text))
	clearHierarchyBelow(&record.Hierarchy, unit.HeadingLevel)
	record.Position = unit.Position
	record.URL = recordutil.URLWithAnchor(pageURL, unit.Anchor)
	record.ObjectID = recordutil.ObjectIDFromURL(record.URL)

	return record, cloneHierarchy(record.Hierarchy), record.URL, nil
}

func fieldRecord(
	pageRecord model.Record,
	currentHierarchy model.Hierarchy,
	pageURL string,
	unit model.ExtractedUnit,
) (model.Record, model.Hierarchy, string) {
	const fieldLevel = 3

	record := pageRecord
	record.RecordType = model.RecordTypeField
	record.Content = stringPtr(unit.Description)
	record.Hierarchy = cloneHierarchy(currentHierarchy)
	setHierarchyLevel(&record.Hierarchy, fieldLevel, stringPtr(unit.Text))
	clearHierarchyBelow(&record.Hierarchy, fieldLevel)
	record.Position = unit.Position
	record.URL = recordutil.URLWithAnchor(pageURL, unit.Anchor)
	record.ObjectID = recordutil.ObjectIDFromURL(record.URL)

	return record, cloneHierarchy(record.Hierarchy), record.URL
}

func contentRecord(
	pageRecord model.Record,
	currentHierarchy model.Hierarchy,
	currentURL string,
	unit model.ExtractedUnit,
) model.Record {
	record := pageRecord
	record.RecordType = model.RecordTypeContent
	record.Content = stringPtr(unit.Text)
	record.Hierarchy = cloneHierarchy(currentHierarchy)
	record.Position = unit.Position
	record.URL = currentURL
	record.ObjectID = recordutil.ObjectIDWithPosition(
		recordutil.ObjectIDFromURL(record.URL),
		unit.Position,
	)

	return record
}

func typeFromLevel(level int) model.RecordType {
	switch level {
	case 1:
		return model.RecordTypeLvl1
	case 2:
		return model.RecordTypeLvl2
	case 3:
		return model.RecordTypeLvl3
	case 4:
		return model.RecordTypeLvl4
	case 5:
		return model.RecordTypeLvl5
	case 6:
		return model.RecordTypeLvl6
	default:
		return ""
	}
}

func setHierarchyLevel(h *model.Hierarchy, level int, value *string) {
	switch level {
	case 2:
		h.Lvl2 = value
	case 3:
		h.Lvl3 = value
	case 4:
		h.Lvl4 = value
	case 5:
		h.Lvl5 = value
	case 6:
		h.Lvl6 = value
	}
}

func clearHierarchyBelow(h *model.Hierarchy, level int) {
	if level < 3 {
		h.Lvl3 = nil
	}

	if level < 4 {
		h.Lvl4 = nil
	}

	if level < 5 {
		h.Lvl5 = nil
	}

	if level < 6 {
		h.Lvl6 = nil
	}
}

func cloneHierarchy(h model.Hierarchy) model.Hierarchy {
	return model.Hierarchy{
		Lvl0: cloneStringPtr(h.Lvl0),
		Lvl1: cloneStringPtr(h.Lvl1),
		Lvl2: cloneStringPtr(h.Lvl2),
		Lvl3: cloneStringPtr(h.Lvl3),
		Lvl4: cloneStringPtr(h.Lvl4),
		Lvl5: cloneStringPtr(h.Lvl5),
		Lvl6: cloneStringPtr(h.Lvl6),
	}
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	copiedValue := *value

	return &copiedValue
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func contentTypeFromURL(pageURL string) string {
	path := recordutil.BreadcrumbPathFromURL(pageURL)

	switch {
	case strings.HasPrefix(path, "/guides"):
		return "guide"
	case strings.HasPrefix(path, "/rest-api"):
		return "api"
	case strings.HasPrefix(path, "/api-reference/api-parameters"):
		return "api"
	case strings.HasPrefix(path, "/integration"):
		return "integration"
	case strings.HasPrefix(path, "/libraries/sdk"):
		return "sdk"
	case strings.HasPrefix(path, "/framework-integration"):
		return "sdk"
	default:
		return ""
	}
}

var productPathPrefixes = []struct {
	prefix  string
	product string
}{
	{prefix: "/ui-libraries/autocomplete", product: "autocomplete"},
	{prefix: "/tools/crawler", product: "crawler"},
	{prefix: "/tools/cli", product: "cli"},
	{prefix: "/integration/shopify", product: "shopify"},
	{prefix: "/integration/bigcommerce", product: "bigcommerce"},
	{prefix: "/integration/commercetools", product: "commercetools"},
	{prefix: "/integration/magento-2", product: "magento"},
	{prefix: "/integration/salesforce-commerce-cloud-b2c", product: "sfcc"},
	{prefix: "/integration/zendesk", product: "zendesk"},
	{prefix: "/framework-integration/django", product: "django"},
	{prefix: "/framework-integration/rails", product: "rails"},
	{prefix: "/framework-integration/symfony", product: "symfony"},
	{prefix: "/framework-integration/laravel", product: "laravel"},
	{
		prefix:  "/sending-and-managing-data/send-and-update-your-data/connectors/bigquery",
		product: "bigquery",
	},
	{
		prefix:  "/sending-and-managing-data/send-and-update-your-data/connectors/elasticsearch",
		product: "elasticsearch",
	},
	{
		prefix:  "/sending-and-managing-data/send-and-update-your-data/connectors/mysql",
		product: "mysql",
	},
	{
		prefix:  "/sending-and-managing-data/send-and-update-your-data/connectors/supabase",
		product: "supabase",
	},
	{
		prefix:  "/guides/sending-events/connectors/tealium",
		product: "tealium",
	},
	{
		prefix:  "/guides/sending-events/connectors/segment",
		product: "segment",
	},
	{
		prefix:  "/guides/sending-events/connectors/google-tag-manager",
		product: "google-tag-manager",
	},
	{
		prefix:  "/guides/algolia-ai/agent-studio",
		product: "agent-studio",
	},
	{
		prefix:  "/guides/algolia-ai/askai",
		product: "askai",
	},
}

func productFromURL(pageURL string) string {
	path := recordutil.BreadcrumbPathFromURL(pageURL)

	for _, mapping := range productPathPrefixes {
		if strings.HasPrefix(path, mapping.prefix) {
			return mapping.product
		}
	}

	return ""
}
