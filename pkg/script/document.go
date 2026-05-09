package script

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/algolia/docs-crawler/pkg/model"
)

// Document is the safe DOM handle passed to scripts.
//
// It keeps the parsed page URL with the DOM and lets script engine adapters add
// narrow wrappers instead of leaking mutable parser internals to user scripts.
type Document struct {
	url string
	doc *goquery.Document
}

// NewDocument wraps a parsed page for script execution.
func NewDocument(page model.ParsedPage) Document {
	return Document{
		url: page.URL,
		doc: page.Doc,
	}
}

// URL returns the parsed page URL.
func (d Document) URL() string {
	return d.url
}

// GoqueryDocument returns the underlying DOM for engine adapter code.
//
// Callers must not expose this value directly to scripts.
func (d Document) GoqueryDocument() *goquery.Document {
	return d.doc
}
