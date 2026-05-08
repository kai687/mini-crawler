package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/algolia/docs-crawler/internal/fetch"
	"github.com/algolia/docs-crawler/internal/parse"
	scriptapi "github.com/algolia/docs-crawler/internal/script"
)

type pageProcessor interface {
	Process(ctx context.Context, pageURL string) ([]any, error)
}

type scriptPageProcessor struct {
	httpClient *http.Client
	program    scriptapi.Program
}

func newScriptPageProcessor(
	httpClient *http.Client,
	program scriptapi.Program,
) scriptPageProcessor {
	return scriptPageProcessor{
		httpClient: httpClient,
		program:    program,
	}
}

func (p scriptPageProcessor) Process(ctx context.Context, pageURL string) ([]any, error) {
	slog.Info("processing", "url", pageURL)

	fetcher := fetch.HTTPFetcher{Client: p.httpClient}
	parser := parse.HTMLParser{}

	page, err := fetcher.Fetch(ctx, pageURL)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", pageURL, err)
	}

	parsed, err := parser.Parse(page)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", pageURL, err)
	}

	doc := scriptapi.NewDocument(parsed)
	scriptCtx := scriptapi.Context{URL: pageURL}

	meta, err := p.program.PageMeta(doc, scriptCtx)
	if err != nil {
		return nil, fmt.Errorf("script page_meta %s: %w", pageURL, err)
	}

	scriptCtx.Metadata = map[string]any{"pageMeta": meta}

	records, err := p.program.Records(doc, scriptCtx)
	if err != nil {
		return nil, fmt.Errorf("script records %s: %w", pageURL, err)
	}

	out := make([]any, 0, len(records))
	for i, record := range records {
		enrichCtx := scriptCtx
		enrichCtx.Position = i

		enriched, err := p.program.Enrich(record, enrichCtx)
		if err != nil {
			return nil, fmt.Errorf("script enrich %s record %d: %w", pageURL, i, err)
		}

		out = append(out, enriched)
	}

	return out, nil
}
