package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/algolia/docs-crawler/internal/config"
	"github.com/algolia/docs-crawler/internal/coverage"
	"github.com/algolia/docs-crawler/internal/extract"
	"github.com/algolia/docs-crawler/internal/fetch"
	"github.com/algolia/docs-crawler/internal/model"
	"github.com/algolia/docs-crawler/internal/output"
	"github.com/algolia/docs-crawler/internal/parse"
	"github.com/algolia/docs-crawler/internal/source"
)

type pageResult struct {
	url               string
	records           []model.Record
	expectedHeadings  int
	extractedHeadings int
	err               error
}

type runState struct {
	firstErr error
	failed   int
	tracker  coverage.Tracker
}

// Run executes one crawler session and writes extracted records to out.
func Run(ctx context.Context, cfg config.Config, out io.Writer) error {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	src, err := newSource(cfg, httpClient)
	if err != nil {
		return err
	}

	urls, err := src.URLs(ctx)
	if err != nil {
		return fmt.Errorf("load urls: %w", err)
	}

	workers := normalizedWorkers(cfg)

	slog.Info("crawl start", "mode", cfg.Mode, "urls", len(urls), "workers", workers)

	writer := output.NewJSONLWriter(out)
	defer writer.Close()

	return runPages(ctx, cfg, httpClient, writer, urls, workers)
}

func normalizedWorkers(cfg config.Config) int {
	workers := cfg.Workers
	if workers < 1 {
		workers = 1
	}

	if cfg.Mode == config.ModeSingle {
		workers = 1
	}

	return workers
}

func runPages(
	ctx context.Context,
	cfg config.Config,
	httpClient *http.Client,
	writer *output.JSONLWriter,
	urls []string,
	workers int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string)
	results := make(chan pageResult)
	startWorkers(ctx, httpClient, workers, jobs, results)
	sendJobs(ctx, urls, jobs)

	state := runState{}
	for result := range results {
		if err := handlePageResult(writer, result, cfg.FailOnError, cancel, &state); err != nil {
			return err
		}
	}

	if state.firstErr != nil {
		return state.firstErr
	}

	if state.failed > 0 {
		fmt.Fprintf(os.Stderr, "crawl completed with %d failed URL(s)\n", state.failed)
	}

	if cfg.Coverage {
		fmt.Fprintln(os.Stderr, state.tracker.Report().Format())
	}

	return nil
}

func startWorkers(
	ctx context.Context,
	httpClient *http.Client,
	workers int,
	jobs <-chan string,
	results chan<- pageResult,
) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for pageURL := range jobs {
				records, expected, extracted, err := processPage(ctx, httpClient, pageURL)
				results <- pageResult{
					url:               pageURL,
					records:           records,
					expectedHeadings:  expected,
					extractedHeadings: extracted,
					err:               err,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()
}

func sendJobs(ctx context.Context, urls []string, jobs chan<- string) {
	go func() {
		defer close(jobs)

		for _, pageURL := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- pageURL:
			}
		}
	}()
}

func handlePageResult(
	writer *output.JSONLWriter,
	result pageResult,
	failOnError bool,
	cancel context.CancelFunc,
	state *runState,
) error {
	if result.err != nil {
		state.failed++
		state.tracker.Add(0, 0, 0)

		slog.Error("page failed", "url", result.url, "err", result.err)

		if failOnError && state.firstErr == nil {
			state.firstErr = result.err

			cancel()
		}

		return nil
	}

	state.tracker.Add(len(result.records), result.expectedHeadings, result.extractedHeadings)

	if state.firstErr != nil {
		return nil
	}

	for _, record := range result.records {
		if err := writer.Write(record); err != nil {
			cancel()

			return fmt.Errorf("write record for %s: %w", result.url, err)
		}
	}

	return nil
}

func processPage(
	ctx context.Context,
	httpClient *http.Client,
	pageURL string,
) ([]model.Record, int, int, error) {
	slog.Info("processing", "url", pageURL)

	fetcher := fetch.HTTPFetcher{Client: httpClient}
	parser := parse.HTMLParser{}
	extractor := extract.PageExtractor{}

	page, err := fetcher.Fetch(ctx, pageURL)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("fetch %s: %w", pageURL, err)
	}

	parsed, err := parser.Parse(page)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("parse %s: %w", pageURL, err)
	}

	expectedHeadings := extract.CountExpectedHeadings(parsed.Doc)

	records, err := extractor.Extract(parsed)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("extract %s: %w", pageURL, err)
	}

	return records, expectedHeadings, countHeadingRecords(records), nil
}

func countHeadingRecords(records []model.Record) int {
	count := 0

	for _, record := range records {
		switch record.RecordType {
		case model.RecordTypeLvl2,
			model.RecordTypeLvl3,
			model.RecordTypeLvl4,
			model.RecordTypeLvl5,
			model.RecordTypeLvl6:
			count++
		}
	}

	return count
}

func newSource(cfg config.Config, client *http.Client) (source.Source, error) {
	switch cfg.Mode {
	case config.ModeSingle:
		return source.Single{URL: cfg.Target}, nil
	case config.ModeSitemap:
		return source.Sitemap{
			SitemapURL: cfg.Target,
			Filter:     cfg.Filter,
			Client:     client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
}
