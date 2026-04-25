package extract

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/recordutil"
)

// PageExtractor turns one parsed HTML page into page, heading, paragraph, and list item records.
type PageExtractor struct{}

// contentRootSelectors lists candidate content containers, in priority order.
var contentRootSelectors = []string{"#content", "#content-area"}

// headingSelectorRelative matches in-content headings (with id) relative to a content root.
const headingSelectorRelative = "h2[id], h3[id], h4[id], h5[id], h6[id]"

// paramFieldSelector matches API reference field headers rendered with the
// shared `param-head` component. IDs vary by section (`authorization-*`,
// `parameter-*`, `body-*`, `response-*`, ...), so match the component class
// instead of a specific prefix.
const paramFieldSelector = `div.param-head[id]`

const contentSelectorRelative = headingSelectorRelative + ", span[data-as='p'], li, " + paramFieldSelector

// ContentRoot returns the first matching content container in priority order, or nil.
func ContentRoot(doc *goquery.Document) *goquery.Selection {
	for _, selector := range contentRootSelectors {
		root := doc.Find(selector).First()
		if root.Length() > 0 {
			return root
		}
	}

	return nil
}

// CountExpectedHeadings reports how many in-content headings the doc contains.
func CountExpectedHeadings(doc *goquery.Document) int {
	root := ContentRoot(doc)
	if root == nil {
		return 0
	}

	return root.Find(headingSelectorRelative).Length() +
		root.Find(paramFieldSelector).Length()
}

// Extract builds records for page metadata, headings, paragraphs, and list items in document order.
func (e PageExtractor) Extract(page model.ParsedPage) ([]model.Record, error) {
	pageRecord := pageRecordFromPage(page)
	records := []model.Record{pageRecord}
	records = append(records, contentRecords(page, pageRecord)...)

	return records, nil
}

func pageRecordFromPage(page model.ParsedPage) model.Record {
	title := normalizeWhitespace(page.Doc.Find("h1").First().Text())
	if title == "" {
		title = normalizeWhitespace(page.Doc.Find("title").First().Text())
	}

	description := normalizeWhitespace(
		page.Doc.Find("meta[name='description']").First().AttrOr("content", ""),
	)

	lvl1 := stringPtr(normalizeWhitespace(page.Doc.Find("h1#page-title").First().Text()))

	pageURL := recordutil.URLWithoutAnchor(page.URL)

	return model.Record{
		URL:              pageURL,
		URLWithoutAnchor: pageURL,
		Breadcrumb:       recordutil.BreadcrumbFromURL(pageURL),
		ContentType:      contentTypeFromURL(pageURL),
		RecordType:       model.RecordTypeLvl1,
		Title:            stringPtr(title),
		Description:      stringPtr(description),
		Hierarchy: model.Hierarchy{
			Lvl1: lvl1,
		},
		Position: 0,
		ObjectID: recordutil.ObjectIDFromURL(pageURL),
	}
}

func contentRecords(page model.ParsedPage, pageRecord model.Record) []model.Record {
	var records []model.Record

	currentHierarchy := cloneHierarchy(pageRecord.Hierarchy)
	currentURL := page.URL
	position := 1

	root := ContentRoot(page.Doc)
	if root == nil {
		return records
	}

	root.Find(contentSelectorRelative).Each(func(_ int, selection *goquery.Selection) {
		emitted, nextHierarchy, nextURL := recordFromSelection(
			page,
			pageRecord,
			selection,
			currentHierarchy,
			currentURL,
			position,
		)

		records = append(records, emitted...)
		position += len(emitted)
		currentHierarchy = nextHierarchy
		currentURL = nextURL
	})

	return records
}

func recordFromSelection(
	page model.ParsedPage,
	pageRecord model.Record,
	selection *goquery.Selection,
	currentHierarchy model.Hierarchy,
	currentURL string,
	position int,
) ([]model.Record, model.Hierarchy, string) {
	switch goquery.NodeName(selection) {
	case "h2", "h3", "h4", "h5", "h6":
		record, ok := headingRecordFromSelection(
			page,
			pageRecord,
			selection,
			currentHierarchy,
			position,
		)
		if !ok {
			return nil, currentHierarchy, currentURL
		}

		return []model.Record{record}, cloneHierarchy(record.Hierarchy), record.URL
	case "span":
		record, ok, nextHierarchy, nextURL := contentRecordFromSelection(
			pageRecord,
			currentHierarchy,
			currentURL,
			selection,
			position,
		)
		if !ok {
			return nil, nextHierarchy, nextURL
		}

		return []model.Record{record}, nextHierarchy, nextURL
	case "li":
		record, ok, nextHierarchy, nextURL := listItemRecordFromSelection(
			pageRecord,
			currentHierarchy,
			currentURL,
			selection,
			position,
		)
		if !ok {
			return nil, nextHierarchy, nextURL
		}

		return []model.Record{record}, nextHierarchy, nextURL
	case "div":
		return paramFieldRecordsFromSelection(
			page,
			pageRecord,
			selection,
			currentHierarchy,
			currentURL,
			position,
		)
	default:
		return nil, currentHierarchy, currentURL
	}
}

// paramFieldRecordsFromSelection emits a heading record for an API parameter
// field plus an optional content record for its description block.
func paramFieldRecordsFromSelection(
	page model.ParsedPage,
	pageRecord model.Record,
	selection *goquery.Selection,
	currentHierarchy model.Hierarchy,
	currentURL string,
	position int,
) ([]model.Record, model.Hierarchy, string) {
	anchor := normalizeWhitespace(selection.AttrOr("id", ""))
	if anchor == "" {
		return nil, currentHierarchy, currentURL
	}

	name := normalizeWhitespace(
		selection.Find("[data-component-part='field-name']").First().Text(),
	)
	if name == "" {
		return nil, currentHierarchy, currentURL
	}

	const fieldLevel = 3

	heading := pageRecord
	heading.RecordType = typeFromLevel(fieldLevel)
	heading.Hierarchy = cloneHierarchy(currentHierarchy)
	setHierarchyLevel(&heading.Hierarchy, fieldLevel, stringPtr(name))
	clearHierarchyBelow(&heading.Hierarchy, fieldLevel)
	heading.Position = position
	heading.URL = recordutil.URLWithAnchor(page.URL, anchor)
	heading.ObjectID = recordutil.ObjectIDFromURL(heading.URL)

	records := []model.Record{heading}

	if desc := paramFieldDescription(selection); desc != "" {
		records = append(records, contentRecord(
			pageRecord,
			heading.Hierarchy,
			heading.URL,
			desc,
			position+1,
		))
	}

	return records, cloneHierarchy(heading.Hierarchy), heading.URL
}

// paramFieldDescription concatenates the type pill, required marker, and the
// first paragraph of the description sibling block.
func paramFieldDescription(header *goquery.Selection) string {
	var parts []string

	if pill := normalizeWhitespace(
		header.Find("[data-component-part='field-info-pill']").First().Text(),
	); pill != "" {
		parts = append(parts, pill)
	}

	if normalizeWhitespace(
		header.Find("[data-component-part='field-required-pill']").First().Text(),
	) != "" {
		parts = append(parts, "required")
	}

	descBlock := header.NextFiltered("div.mt-4").First()
	if descBlock.Length() > 0 {
		prose := descBlock.ChildrenFiltered("div").First()
		if text := normalizeWhitespace(prose.ChildrenFiltered("p").First().Text()); text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, ". ")
}

func headingRecordFromSelection(
	page model.ParsedPage,
	pageRecord model.Record,
	selection *goquery.Selection,
	currentHierarchy model.Hierarchy,
	position int,
) (model.Record, bool) {
	text := normalizeWhitespace(selection.Text())
	if text == "" {
		return model.Record{}, false
	}

	return headingRecord(
		pageRecord,
		page.URL,
		text,
		selection,
		currentHierarchy,
		position,
	)
}

func contentRecordFromSelection(
	pageRecord model.Record,
	currentHierarchy model.Hierarchy,
	currentURL string,
	selection *goquery.Selection,
	position int,
) (model.Record, bool, model.Hierarchy, string) {
	if selection.ParentsFiltered("li").Length() > 0 {
		return model.Record{}, false, currentHierarchy, currentURL
	}

	text := normalizeWhitespace(selection.Text())
	if text == "" {
		return model.Record{}, false, currentHierarchy, currentURL
	}

	return contentRecord(
		pageRecord,
		currentHierarchy,
		currentURL,
		text,
		position,
	), true, currentHierarchy, currentURL
}

func listItemRecordFromSelection(
	pageRecord model.Record,
	currentHierarchy model.Hierarchy,
	currentURL string,
	selection *goquery.Selection,
	position int,
) (model.Record, bool, model.Hierarchy, string) {
	text := normalizeWhitespace(selection.Text())
	if text == "" || !shouldIndexListItem(selection) {
		return model.Record{}, false, currentHierarchy, currentURL
	}

	return contentRecord(
		pageRecord,
		currentHierarchy,
		currentURL,
		text,
		position,
	), true, currentHierarchy, currentURL
}

func headingRecord(
	pageRecord model.Record,
	pageURL string,
	text string,
	selection *goquery.Selection,
	currentHierarchy model.Hierarchy,
	position int,
) (model.Record, bool) {
	level := headingLevel(selection)
	if level < 2 || level > 6 {
		return model.Record{}, false
	}

	record := pageRecord
	record.RecordType = typeFromLevel(level)
	record.Hierarchy = cloneHierarchy(currentHierarchy)
	setHierarchyLevel(&record.Hierarchy, level, stringPtr(text))
	clearHierarchyBelow(&record.Hierarchy, level)
	record.Position = position
	record.URL = recordutil.URLWithAnchor(pageURL, normalizeWhitespace(selection.AttrOr("id", "")))
	record.ObjectID = recordutil.ObjectIDFromURL(record.URL)

	return record, true
}

func shouldIndexListItem(selection *goquery.Selection) bool {
	content := selection.Clone()
	content.Find("a").Each(func(_ int, anchor *goquery.Selection) {
		anchor.Remove()
	})

	return normalizeWhitespace(content.Text()) != ""
}

func contentRecord(
	pageRecord model.Record,
	currentHierarchy model.Hierarchy,
	currentURL string,
	text string,
	position int,
) model.Record {
	record := pageRecord
	record.RecordType = model.RecordTypeContent
	record.Content = stringPtr(text)
	record.Hierarchy = cloneHierarchy(currentHierarchy)
	record.Position = position
	record.URL = currentURL
	record.ObjectID = recordutil.ObjectIDWithPosition(
		recordutil.ObjectIDFromURL(record.URL),
		position,
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

func headingLevel(selection *goquery.Selection) int {
	switch goquery.NodeName(selection) {
	case "h2":
		return 2
	case "h3":
		return 3
	case "h4":
		return 4
	case "h5":
		return 5
	case "h6":
		return 6
	default:
		return 0
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

func contentTypeFromURL(pageURL string) string {
	path := recordutil.BreadcrumbFromURL(pageURL)

	switch {
	case strings.HasPrefix(path, "/guides"):
		return "guide"
	case strings.HasPrefix(path, "/rest-api"):
		return "api"
	default:
		return ""
	}
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
