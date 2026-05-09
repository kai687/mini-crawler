package script

// Context carries crawl/page metadata exposed to scripts.
type Context struct {
	// URL is the canonical URL for the page or record currently being processed.
	URL string
	// Position is the record position within the current page when known.
	Position int
	// Metadata stores optional JSON-like host-provided values.
	Metadata map[string]any
}
