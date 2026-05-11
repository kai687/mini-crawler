package parse

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/kai687/mini-crawler/pkg/model"
)

// HTMLParser converts fetched page bytes into a goquery document.
//
// It is intentionally small: fetching already happened, and extraction engines
// decide how to query the resulting DOM.
type HTMLParser struct{}

// Parse reads HTML bytes from a Page and returns a ParsedPage.
func (p HTMLParser) Parse(page model.Page) (model.ParsedPage, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page.Body))
	if err != nil {
		return model.ParsedPage{}, fmt.Errorf("parse html: %w", err)
	}

	return model.ParsedPage{
		Ref:      page.Ref,
		URL:      page.URL,
		Kind:     "html",
		Document: doc,
		Doc:      doc,
		Metadata: page.Metadata,
	}, nil
}
