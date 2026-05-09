package source

import "context"

// Single returns exactly one target URL.
//
// Commands use this source for explicit one-page crawls so the app runner can
// treat single-page and multi-page crawls the same way.
type Single struct {
	// URL is the page URL to process.
	URL string
}

// URLs returns the configured single URL.
func (s Single) URLs(_ context.Context) ([]string, error) {
	return []string{s.URL}, nil
}
