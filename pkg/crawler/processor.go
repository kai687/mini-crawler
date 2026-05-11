package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/algolia/mini-crawler/pkg/model"
)

// Source discovers crawl targets. A target can be a URL, file path, object key,
// or any other reference understood by the configured Fetcher.
type Source interface {
	URLs(ctx context.Context) ([]string, error)
}

// RefFilter skips discovered references before fetching page content.
type RefFilter interface {
	FilterRef(ctx context.Context, ref string) error
}

// Fetcher loads raw content for one discovered reference.
type Fetcher interface {
	Fetch(ctx context.Context, ref string) (model.Page, error)
}

// PreParseFilter skips fetched pages before parsing.
type PreParseFilter interface {
	FilterPage(ctx context.Context, page model.Page) error
}

// Parser converts raw fetched content into an extractor-friendly document.
type Parser interface {
	Parse(page model.Page) (model.ParsedPage, error)
}

// ParsedPageFilter skips parsed pages before extraction.
type ParsedPageFilter interface {
	FilterParsedPage(ctx context.Context, page model.ParsedPage) error
}

// PageFilter is kept for compatibility. Use ParsedPageFilter for new code.
type PageFilter = ParsedPageFilter

// Extractor turns one parsed page into JSON-like records.
type Extractor interface {
	Extract(ctx context.Context, page model.ParsedPage) ([]any, error)
}

// Writer writes extracted records and flushes/cleans up on Close.
type Writer interface {
	Write(record any) error
	Close() error
}

// ErrFiltered means a crawl filter chose to skip the page.
var ErrFiltered = errors.New("filtered")

// PageProcessor is the unit Run workers execute for each discovered reference.
type PageProcessor interface {
	Process(ctx context.Context, ref string) ([]any, error)
}

type pipelineProcessor struct {
	refFilter        RefFilter
	fetcher          Fetcher
	preParseFilter   PreParseFilter
	parser           Parser
	parsedPageFilter ParsedPageFilter
	extractor        Extractor
	limiter          *requestLimiter
	metrics          *crawlMetrics
}

func newPipelineProcessor(
	refFilter RefFilter,
	fetcher Fetcher,
	preParseFilter PreParseFilter,
	parser Parser,
	parsedPageFilter ParsedPageFilter,
	extractor Extractor,
	limiter *requestLimiter,
	metrics *crawlMetrics,
) pipelineProcessor {
	return pipelineProcessor{
		refFilter:        refFilter,
		fetcher:          fetcher,
		preParseFilter:   preParseFilter,
		parser:           parser,
		parsedPageFilter: parsedPageFilter,
		extractor:        extractor,
		limiter:          limiter,
		metrics:          metrics,
	}
}

func (p pipelineProcessor) Process(ctx context.Context, ref string) ([]any, error) {
	slog.Info("processing", "ref", ref)

	if err := p.filterRef(ctx, ref); err != nil {
		return nil, err
	}

	if err := p.limiter.wait(ctx); err != nil {
		return nil, fmt.Errorf("wait for request slot: %w", err)
	}

	page, err := p.fetcher.Fetch(ctx, ref)
	p.metrics.addRequest()

	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", ref, err)
	}

	if err := p.filterPage(ctx, ref, page); err != nil {
		return nil, err
	}

	parsed, err := p.parser.Parse(page)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", ref, err)
	}

	if err := p.filterParsedPage(ctx, ref, parsed); err != nil {
		return nil, err
	}

	records, err := p.extractor.Extract(ctx, parsed)
	if err != nil {
		return nil, fmt.Errorf("extract %s: %w", ref, err)
	}

	p.metrics.addPage()

	return records, nil
}

func (p pipelineProcessor) filterRef(ctx context.Context, ref string) error {
	if p.refFilter == nil {
		return nil
	}

	if err := p.refFilter.FilterRef(ctx, ref); err != nil {
		return fmt.Errorf("filter ref %s: %w", ref, err)
	}

	return nil
}

func (p pipelineProcessor) filterPage(ctx context.Context, ref string, page model.Page) error {
	if p.preParseFilter == nil {
		return nil
	}

	if err := p.preParseFilter.FilterPage(ctx, page); err != nil {
		return fmt.Errorf("filter page %s: %w", ref, err)
	}

	return nil
}

func (p pipelineProcessor) filterParsedPage(
	ctx context.Context,
	ref string,
	page model.ParsedPage,
) error {
	if p.parsedPageFilter == nil {
		return nil
	}

	if err := p.parsedPageFilter.FilterParsedPage(ctx, page); err != nil {
		return fmt.Errorf("filter parsed page %s: %w", ref, err)
	}

	return nil
}
