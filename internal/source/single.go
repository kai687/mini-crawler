package source

import "context"

// Single returns exactly one target URL.
type Single struct {
	URL string
}

// URLs returns the configured single URL.
func (s Single) URLs(_ context.Context) ([]string, error) {
	return []string{s.URL}, nil
}
