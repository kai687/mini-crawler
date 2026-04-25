package extract

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
)

// PageExtractor turns one parsed HTML page into raw extracted page facts.
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

// Extract builds raw page facts in document order.
func (e PageExtractor) Extract(page model.ParsedPage) (model.ExtractedDocument, error) {
	doc := model.ExtractedDocument{
		PageURL: page.URL,
		Description: stringPtr(
			normalizeWhitespace(
				page.Doc.Find("meta[name='description']").First().AttrOr("content", ""),
			),
		),
		PageHeading: stringPtr(normalizeWhitespace(page.Doc.Find("h1#page-title").First().Text())),
	}

	doc.Units = contentUnits(page)

	return doc, nil
}

func contentUnits(page model.ParsedPage) []model.ExtractedUnit {
	var units []model.ExtractedUnit

	root := ContentRoot(page.Doc)
	if root == nil {
		return units
	}

	position := 1

	root.Find(contentSelectorRelative).Each(func(_ int, selection *goquery.Selection) {
		emitted := unitsFromSelection(selection, position)
		units = append(units, emitted...)
		position += len(emitted)
	})

	return units
}

func unitsFromSelection(selection *goquery.Selection, position int) []model.ExtractedUnit {
	switch goquery.NodeName(selection) {
	case "h2", "h3", "h4", "h5", "h6":
		unit, ok := headingUnitFromSelection(selection, position)
		if !ok {
			return nil
		}

		return []model.ExtractedUnit{unit}
	case "span":
		unit, ok := contentUnitFromSelection(selection, position)
		if !ok {
			return nil
		}

		return []model.ExtractedUnit{unit}
	case "li":
		unit, ok := listItemUnitFromSelection(selection, position)
		if !ok {
			return nil
		}

		return []model.ExtractedUnit{unit}
	case "div":
		return paramFieldUnitsFromSelection(selection, position)
	default:
		return nil
	}
}

func headingUnitFromSelection(
	selection *goquery.Selection,
	position int,
) (model.ExtractedUnit, bool) {
	text := normalizeWhitespace(selection.Text())
	if text == "" {
		return model.ExtractedUnit{}, false
	}

	level := headingLevel(selection)
	if level < 2 || level > 6 {
		return model.ExtractedUnit{}, false
	}

	return model.ExtractedUnit{
		Kind:         model.ExtractedHeading,
		Anchor:       normalizeWhitespace(selection.AttrOr("id", "")),
		Text:         text,
		HeadingLevel: level,
		Position:     position,
	}, true
}

func contentUnitFromSelection(
	selection *goquery.Selection,
	position int,
) (model.ExtractedUnit, bool) {
	if selection.ParentsFiltered("li").Length() > 0 {
		return model.ExtractedUnit{}, false
	}

	text := normalizeWhitespace(selection.Text())
	if text == "" {
		return model.ExtractedUnit{}, false
	}

	return model.ExtractedUnit{
		Kind:     model.ExtractedContent,
		Text:     text,
		Position: position,
	}, true
}

func listItemUnitFromSelection(
	selection *goquery.Selection,
	position int,
) (model.ExtractedUnit, bool) {
	text := normalizeWhitespace(selection.Text())
	if text == "" || !shouldIndexListItem(selection) {
		return model.ExtractedUnit{}, false
	}

	return model.ExtractedUnit{
		Kind:     model.ExtractedContent,
		Text:     text,
		Position: position,
	}, true
}

// paramFieldUnitsFromSelection emits one raw field unit, with optional
// description attached for indexing on same final record.
func paramFieldUnitsFromSelection(
	selection *goquery.Selection,
	position int,
) []model.ExtractedUnit {
	anchor := normalizeWhitespace(selection.AttrOr("id", ""))
	if anchor == "" {
		return nil
	}

	name := normalizeWhitespace(
		selection.Find("[data-component-part='field-name']").First().Text(),
	)
	if name == "" {
		return nil
	}

	return []model.ExtractedUnit{{
		Kind:        model.ExtractedField,
		Anchor:      anchor,
		Text:        name,
		Description: paramFieldDescription(selection),
		Position:    position,
	}}
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

func shouldIndexListItem(selection *goquery.Selection) bool {
	content := selection.Clone()
	content.Find("a").Each(func(_ int, anchor *goquery.Selection) {
		anchor.Remove()
	})

	return normalizeWhitespace(content.Text()) != ""
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

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
