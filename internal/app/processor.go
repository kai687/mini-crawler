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

	records, err := p.program.Extract(doc, scriptCtx)
	if err != nil {
		return nil, fmt.Errorf("script extract %s: %w", pageURL, err)
	}

	out := make([]any, 0, len(records))
	for _, record := range records {
		out = append(out, record)
	}

	return out, nil
}
