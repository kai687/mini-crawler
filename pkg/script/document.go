package script

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/kai687/mini-crawler/pkg/model"
)

// Document is the safe DOM handle passed to scripts.
//
// It keeps the parsed page URL with the DOM and lets script engine adapters add
// narrow wrappers instead of leaking mutable parser internals to user scripts.
type Document struct {
	url      string
	kind     string
	doc      *goquery.Document
	metadata map[string]any
}

// NewDocument wraps a parsed page for script execution.
func NewDocument(page model.ParsedPage) Document {
	return Document{
		url:      page.URL,
		kind:     page.Kind,
		doc:      page.Doc,
		metadata: page.Metadata,
	}
}

// URL returns the parsed page URL.
func (d Document) URL() string {
	return d.url
}

// Kind returns the parsed document kind, such as "html" or "markdown".
func (d Document) Kind() string {
	return d.kind
}

// MetadataValue returns one parser metadata value.
func (d Document) MetadataValue(key string) any {
	return d.metadata[key]
}

// MetadataString returns one parser metadata string, or empty when absent.
func (d Document) MetadataString(key string) string {
	value, ok := d.MetadataValue(key).(string)
	if !ok {
		return ""
	}

	return value
}

// GoqueryDocument returns the underlying DOM for engine adapter code.
//
// Callers must not expose this value directly to scripts.
func (d Document) GoqueryDocument() *goquery.Document {
	return d.doc
}
