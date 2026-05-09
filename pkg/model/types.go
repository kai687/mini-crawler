package model

import "github.com/PuerkitoBio/goquery"

// Page is raw fetched page data before HTML parsing.
//
// It is the contract between fetch and parse packages: fetch owns HTTP I/O,
// parse owns interpreting Body as HTML.
type Page struct {
	// Ref is the original crawl target reference: URL, file path, object key, etc.
	Ref string
	// URL is the canonical URL when content came from or maps to HTTP.
	URL string
	// StatusCode is the HTTP response status code returned by the server, when any.
	StatusCode int
	// ContentType is the response Content-Type header value or equivalent type hint.
	ContentType string
	// Body is the raw response body bytes.
	Body []byte
	// Metadata stores fetcher-provided values for custom pipelines.
	Metadata map[string]any
}

// ParsedPage is a Page converted into an extractor-friendly document.
type ParsedPage struct {
	// Ref is copied from the fetched page.
	Ref string
	// URL is copied from the fetched page so extractors can build canonical records.
	URL string
	// Kind identifies the parser output, for example "html" or "markdown".
	Kind string
	// Document stores parser-specific output. HTMLParser uses *goquery.Document.
	Document any
	// Doc is the parsed HTML document used by existing script adapters.
	// Deprecated: use Document with Kind == "html".
	Doc *goquery.Document
	// Metadata stores parser-provided values for custom extractors.
	Metadata map[string]any
}
