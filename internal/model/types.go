package model

import "github.com/PuerkitoBio/goquery"

// Page is raw fetched page data before HTML parsing.
type Page struct {
	URL         string
	StatusCode  int
	ContentType string
	Body        []byte
}

// ParsedPage is a Page converted into a queryable document.
type ParsedPage struct {
	URL string
	Doc *goquery.Document
}

// Hierarchy stores heading context for one indexed record.
//
// Levels map to document heading depth, not record type names:
//   - Lvl1: page title from h1#page-title
//   - Lvl2: current h2 heading
//   - Lvl3: current h3 heading
//   - Lvl4: current h4 heading
//   - Lvl5: current h5 heading
//   - Lvl6: current h6 heading
//
// Nil means level not set for this record.
type Hierarchy struct {
	Lvl0 *string `json:"lvl0,omitempty"`
	Lvl1 *string `json:"lvl1,omitempty"`
	Lvl2 *string `json:"lvl2,omitempty"`
	Lvl3 *string `json:"lvl3,omitempty"`
	Lvl4 *string `json:"lvl4,omitempty"`
	Lvl5 *string `json:"lvl5,omitempty"`
	Lvl6 *string `json:"lvl6,omitempty"`
}

// BreadcrumbHierarchy stores cumulative breadcrumb labels for hierarchical faceting.
type BreadcrumbHierarchy struct {
	Lvl0 *string `json:"lvl0,omitempty"`
	Lvl1 *string `json:"lvl1,omitempty"`
	Lvl2 *string `json:"lvl2,omitempty"`
	Lvl3 *string `json:"lvl3,omitempty"`
	Lvl4 *string `json:"lvl4,omitempty"`
	Lvl5 *string `json:"lvl5,omitempty"`
}

// RecordType identifies valid extracted record kinds.
type RecordType string

const (
	// RecordTypeContent stores paragraph or list item content.
	RecordTypeContent RecordType = "content"
	// RecordTypeField stores REST API parameter/request/response field records.
	RecordTypeField RecordType = "field"
	// RecordTypeLvl1 stores page-level record.
	RecordTypeLvl1 RecordType = "lvl1"
	// RecordTypeLvl2 stores h2 heading record.
	RecordTypeLvl2 RecordType = "lvl2"
	// RecordTypeLvl3 stores h3 heading record.
	RecordTypeLvl3 RecordType = "lvl3"
	// RecordTypeLvl4 stores h4 heading record.
	RecordTypeLvl4 RecordType = "lvl4"
	// RecordTypeLvl5 stores h5 heading record.
	RecordTypeLvl5 RecordType = "lvl5"
	// RecordTypeLvl6 stores h6 heading record.
	RecordTypeLvl6 RecordType = "lvl6"
)

// Record is one indexable unit extracted from a page.
type Record struct {
	// Canonical URL for this record, including #anchor
	URL string `json:"url"`
	// Page URL without any #anchor, useful for grouping/distinct in Algolia
	URLWithoutAnchor string `json:"urlWithoutAnchor"`
	// Human-readable ancestor breadcrumb labels for display; current page label comes from title/heading
	BreadcrumbSegments []string `json:"breadcrumbSegments,omitempty"`
	// Hierarchical ancestor breadcrumb labels for Algolia faceting
	BreadcrumbHierarchy *BreadcrumbHierarchy `json:"breadcrumbHierarchy,omitempty"`
	// High-level content classification inferred from URL path (for example: guide, api)
	ContentType string `json:"contentType,omitempty"`
	// Canonical API method name derived from final REST API URL slug, for example searchSingleIndex
	MethodName string `json:"methodName,omitempty"`
	// Record kind (content, field, lvl1, lvl2, ..., lvl6)
	RecordType RecordType `json:"recordType"`
	// Paragraph or list item content
	Content *string `json:"content,omitempty"`
	// Heading context inherited by this record
	Hierarchy Hierarchy `json:"hierarchy"`
	// Extraction order within page. Page title = 0, last match on the page = highest
	Position int `json:"position"`
	// Stable identifier for a record in Algolia
	ObjectID string `json:"objectID"`
}
