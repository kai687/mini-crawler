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
	pageURL := recordutil.URLWithoutAnchor(doc.PageURL)
	breadcrumbSegments := recordutil.BreadcrumbSegmentsFromURL(pageURL)
	breadcrumbHierarchy := recordutil.BreadcrumbHierarchyFromSegments(breadcrumbSegments)
	contentType := contentTypeFromURL(pageURL)

	pageRecord := model.Record{
		URL:                 pageURL,
		URLWithoutAnchor:    pageURL,
		BreadcrumbSegments:  breadcrumbSegments,
		BreadcrumbHierarchy: breadcrumbHierarchy,
		ContentType:         contentType,
		RecordType:          model.RecordTypeLvl1,
		Content:             cloneStringPtr(doc.Description),
		Hierarchy: model.Hierarchy{
			Lvl1: cloneStringPtr(doc.PageHeading),
		},
		Position: 0,
		ObjectID: recordutil.ObjectIDFromURL(pageURL),
	}

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
	default:
		return ""
	}
}
