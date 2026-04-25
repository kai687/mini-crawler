package model

// ExtractedKind identifies raw units observed on page before enrichment.
type ExtractedKind string

const (
	ExtractedHeading ExtractedKind = "heading"
	ExtractedContent ExtractedKind = "content"
	ExtractedField   ExtractedKind = "field"
)

// ExtractedDocument stores page-observable facts prior to record enrichment.
type ExtractedDocument struct {
	PageURL     string
	Description *string
	PageHeading *string
	Units       []ExtractedUnit
}

// ExtractedUnit is one ordered structural or content unit observed in page body.
type ExtractedUnit struct {
	Kind         ExtractedKind
	Anchor       string
	Text         string
	Description  string
	HeadingLevel int
	Position     int
}
