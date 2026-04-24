package coverage

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/recordutil"
)

// Report summarizes how completely one page is represented by extracted records.
type Report struct {
	URL               string   `json:"url"`
	ExpectedNodes     int      `json:"expectedNodes"`
	MatchedNodes      int      `json:"matchedNodes"`
	Coverage          float64  `json:"coverage"`
	HeadingCoverage   float64  `json:"headingCoverage"`
	MetadataCoverage  float64  `json:"metadataCoverage"`
	HierarchyCoverage float64  `json:"hierarchyCoverage"`
	Missing           []string `json:"missing"`
	Errors            []string `json:"errors"`
}

type expectedPage struct {
	Title       *string
	Description *string
	Lvl1        *string
	Headings    []expectedHeading
}

type expectedHeading struct {
	ID        string
	URL       string
	ObjectID  string
	Position  int
	Hierarchy model.Hierarchy
}

type metadataResult struct {
	expected          int
	matched           int
	hierarchyExpected int
	hierarchyMatched  int
}

type headingResult struct {
	matched           bool
	hierarchyExpected int
	hierarchyMatched  int
}

// Analyze compares a parsed page against extracted records and returns structural coverage metrics.
func Analyze(page model.ParsedPage, records []model.Record) Report {
	expected := buildExpectedPage(page)
	actualByURL := recordsByURL(records)
	pageRecord := findPageRecord(page.URL, records)

	report := Report{URL: page.URL}
	report.ExpectedNodes = expectedNodeCount(expected)

	metadata := compareMetadata(expected, pageRecord, &report)
	headingMatched := 0
	hierarchyExpected := metadata.hierarchyExpected
	hierarchyMatched := metadata.hierarchyMatched

	for _, heading := range expected.Headings {
		result := compareHeading(heading, actualByURL[heading.URL], &report)
		if result.matched {
			headingMatched++
		}

		hierarchyExpected += result.hierarchyExpected
		hierarchyMatched += result.hierarchyMatched
	}

	report.MatchedNodes = metadata.matched + headingMatched
	report.Coverage = ratio(report.MatchedNodes, report.ExpectedNodes)
	report.MetadataCoverage = ratio(metadata.matched, metadata.expected)
	report.HeadingCoverage = ratio(headingMatched, len(expected.Headings))
	report.HierarchyCoverage = ratio(hierarchyMatched, hierarchyExpected)

	return report
}

func buildExpectedPage(page model.ParsedPage) expectedPage {
	title := normalizeWhitespace(page.Doc.Find("h1").First().Text())
	if title == "" {
		title = normalizeWhitespace(page.Doc.Find("title").First().Text())
	}

	description := normalizeWhitespace(
		page.Doc.Find("meta[name='description']").First().AttrOr("content", ""),
	)

	return expectedPage{
		Title:       stringPtr(title),
		Description: stringPtr(description),
		Lvl1:        stringPtr(normalizeWhitespace(page.Doc.Find("h1#page-title").First().Text())),
		Headings:    expectedHeadingsFromPage(page),
	}
}

func expectedNodeCount(expected expectedPage) int {
	count := 1 + len(expected.Headings)
	if expected.Description != nil {
		count++
	}

	return count
}

func compareMetadata(
	expected expectedPage,
	pageRecord *model.Record,
	report *Report,
) metadataResult {
	result := metadataResult{expected: 1}
	if pageRecord != nil && sameStringPtr(pageRecord.Title, expected.Title) {
		result.matched++
	} else {
		report.Missing = append(report.Missing, "title")
	}

	if expected.Description != nil {
		result.expected++
		if pageRecord != nil && sameStringPtr(pageRecord.Description, expected.Description) {
			result.matched++
		} else {
			report.Missing = append(report.Missing, "description")
		}
	}

	if expected.Lvl1 != nil {
		result.hierarchyExpected++
		if pageRecord != nil && sameStringPtr(pageRecord.Hierarchy.Lvl1, expected.Lvl1) {
			result.hierarchyMatched++
		} else {
			report.Errors = append(report.Errors, "page hierarchy lvl1 mismatch")
		}
	}

	return result
}

func compareHeading(expected expectedHeading, actual model.Record, report *Report) headingResult {
	result := headingResult{hierarchyExpected: hierarchyCount(expected.Hierarchy)}
	if actual.URL == "" {
		report.Missing = append(report.Missing, expected.ID)

		return result
	}

	matched := true
	if actual.ObjectID != expected.ObjectID {
		matched = false

		report.Errors = append(report.Errors, fmt.Sprintf("%s objectID mismatch", expected.ID))
	}

	if actual.Position != expected.Position {
		matched = false

		report.Errors = append(report.Errors, fmt.Sprintf("%s position mismatch", expected.ID))
	}

	if !sameHierarchy(actual.Hierarchy, expected.Hierarchy) {
		matched = false

		report.Errors = append(report.Errors, fmt.Sprintf("%s hierarchy mismatch", expected.ID))
	}

	result.matched = matched
	result.hierarchyMatched = matchedHierarchyCount(actual.Hierarchy, expected.Hierarchy)

	return result
}

func expectedHeadingsFromPage(page model.ParsedPage) []expectedHeading {
	var headings []expectedHeading

	currentHierarchy := model.Hierarchy{
		Lvl1: stringPtr(normalizeWhitespace(page.Doc.Find("h1#page-title").First().Text())),
	}

	page.Doc.Find("#content h2[id], #content h3[id], #content h4[id], #content h5[id], #content h6[id]").
		Each(func(index int, selection *goquery.Selection) {
			id := normalizeWhitespace(selection.AttrOr("id", ""))

			text := normalizeWhitespace(selection.Text())
			if id == "" || text == "" {
				return
			}

			level := headingLevel(selection)
			setHierarchyLevel(&currentHierarchy, level, stringPtr(text))
			clearHierarchyBelow(&currentHierarchy, level)

			url := recordutil.URLWithAnchor(page.URL, id)
			headings = append(headings, expectedHeading{
				ID:        id,
				URL:       url,
				ObjectID:  recordutil.ObjectIDFromURL(url),
				Position:  index + 1,
				Hierarchy: cloneHierarchy(currentHierarchy),
			})
		})

	return headings
}

func recordsByURL(records []model.Record) map[string]model.Record {
	result := make(map[string]model.Record, len(records))
	for _, record := range records {
		if _, exists := result[record.URL]; exists {
			continue
		}

		result[record.URL] = record
	}

	return result
}

func findPageRecord(pageURL string, records []model.Record) *model.Record {
	for i := range records {
		if records[i].URL == pageURL {
			return &records[i]
		}
	}

	return nil
}

func sameHierarchy(got model.Hierarchy, want model.Hierarchy) bool {
	return sameStringPtr(got.Lvl0, want.Lvl0) &&
		sameStringPtr(got.Lvl1, want.Lvl1) &&
		sameStringPtr(got.Lvl2, want.Lvl2) &&
		sameStringPtr(got.Lvl3, want.Lvl3) &&
		sameStringPtr(got.Lvl4, want.Lvl4) &&
		sameStringPtr(got.Lvl5, want.Lvl5) &&
		sameStringPtr(got.Lvl6, want.Lvl6)
}

func sameStringPtr(got *string, want *string) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}

	return *got == *want
}

func hierarchyCount(h model.Hierarchy) int {
	count := 0

	for _, value := range []*string{h.Lvl1, h.Lvl2, h.Lvl3, h.Lvl4, h.Lvl5, h.Lvl6} {
		if value != nil {
			count++
		}
	}

	return count
}

func matchedHierarchyCount(got model.Hierarchy, want model.Hierarchy) int {
	count := 0

	pairs := []struct {
		got  *string
		want *string
	}{
		{got: got.Lvl1, want: want.Lvl1},
		{got: got.Lvl2, want: want.Lvl2},
		{got: got.Lvl3, want: want.Lvl3},
		{got: got.Lvl4, want: want.Lvl4},
		{got: got.Lvl5, want: want.Lvl5},
		{got: got.Lvl6, want: want.Lvl6},
	}

	for _, pair := range pairs {
		if pair.want != nil && sameStringPtr(pair.got, pair.want) {
			count++
		}
	}

	return count
}

func ratio(got int, total int) float64 {
	if total == 0 {
		return 1
	}

	return float64(got) / float64(total)
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

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
