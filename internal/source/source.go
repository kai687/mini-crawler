package source

import "context"

// Source discovers crawl target URLs.
type Source interface {
	URLs(ctx context.Context) ([]string, error)
}
