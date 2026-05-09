package crawler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/algolia/docs-crawler/pkg/model"
)

// Source discovers crawl targets. A target can be a URL, file path, object key,
// or any other reference understood by the configured Fetcher.
type Source interface {
	URLs(ctx context.Context) ([]string, error)
}

// Fetcher loads raw content for one discovered reference.
type Fetcher interface {
	Fetch(ctx context.Context, ref string) (model.Page, error)
}

// Parser converts raw fetched content into an extractor-friendly document.
type Parser interface {
	Parse(page model.Page) (model.ParsedPage, error)
}

// Extractor turns one parsed page into JSON-like records.
type Extractor interface {
	Extract(ctx context.Context, page model.ParsedPage) ([]any, error)
}

// Writer writes extracted records and flushes/cleans up on Close.
type Writer interface {
	Write(record any) error
	Close() error
}

// PageProcessor is the unit Run workers execute for each discovered reference.
type PageProcessor interface {
	Process(ctx context.Context, ref string) ([]any, error)
}

type pipelineProcessor struct {
	fetcher   Fetcher
	parser    Parser
	extractor Extractor
}

func newPipelineProcessor(fetcher Fetcher, parser Parser, extractor Extractor) pipelineProcessor {
	return pipelineProcessor{fetcher: fetcher, parser: parser, extractor: extractor}
}

func (p pipelineProcessor) Process(ctx context.Context, ref string) ([]any, error) {
	slog.Info("processing", "ref", ref)

	page, err := p.fetcher.Fetch(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", ref, err)
	}

	parsed, err := p.parser.Parse(page)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", ref, err)
	}

	records, err := p.extractor.Extract(ctx, parsed)
	if err != nil {
		return nil, fmt.Errorf("extract %s: %w", ref, err)
	}

	return records, nil
}
