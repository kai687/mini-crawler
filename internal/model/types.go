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
