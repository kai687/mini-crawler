package parse

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/internal/model"
)

// HTMLParser converts fetched page bytes into a goquery document.
type HTMLParser struct{}

// Parse reads HTML bytes from a Page and returns a ParsedPage.
func (p HTMLParser) Parse(page model.Page) (model.ParsedPage, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(page.Body))
	if err != nil {
		return model.ParsedPage{}, fmt.Errorf("parse html: %w", err)
	}

	return model.ParsedPage{
		URL: page.URL,
		Doc: doc,
	}, nil
}
