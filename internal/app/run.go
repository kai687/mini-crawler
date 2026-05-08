package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/algolia/docs-crawler/internal/config"
	"github.com/algolia/docs-crawler/internal/output"
	"github.com/algolia/docs-crawler/internal/script"
	starlarkengine "github.com/algolia/docs-crawler/internal/script/starlark"
	"github.com/algolia/docs-crawler/internal/source"
)

type pageResult struct {
	url     string
	records []any
	err     error
}

type runState struct {
	firstErr error
	failed   int
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

	processor, err := newPageProcessor(cfg, httpClient)
	if err != nil {
		return err
	}

	return runPages(ctx, cfg, processor, writer, urls, workers)
}

func newPageProcessor(cfg config.Config, httpClient *http.Client) (pageProcessor, error) {
	program, err := starlarkengine.Engine{}.Load(cfg.Script)
	if err != nil {
		return nil, fmt.Errorf("load script: %w", err)
	}

	return newScriptPageProcessor(httpClient, program), nil
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
	processor pageProcessor,
	writer *output.JSONLWriter,
	urls []string,
	workers int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string)
	results := make(chan pageResult)
	startWorkers(ctx, processor, workers, jobs, results)
	sendJobs(ctx, urls, jobs)

	failOnError := cfg.FailOnError || cfg.Mode == config.ModeSingle

	state := runState{}
	for result := range results {
		if err := handlePageResult(writer, result, failOnError, cancel, &state); err != nil {
			return err
		}
	}

	if state.firstErr != nil {
		return state.firstErr
	}

	if state.failed > 0 {
		fmt.Fprintf(os.Stderr, "crawl completed with %d failed URL(s)\n", state.failed)
	}

	return nil
}

func startWorkers(
	ctx context.Context,
	processor pageProcessor,
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
				records, err := processor.Process(ctx, pageURL)
				results <- pageResult{
					url:     pageURL,
					records: records,
					err:     err,
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
		if errors.Is(result.err, script.ErrNoExtractor) {
			slog.Warn("page skipped: no extractor matches", "url", result.url)

			return nil
		}

		state.failed++

		slog.Error("page failed", "url", result.url, "err", result.err)

		if failOnError && state.firstErr == nil {
			state.firstErr = result.err

			cancel()
		}

		return nil
	}

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

func newSource(cfg config.Config, client *http.Client) (source.Source, error) {
	switch cfg.Mode {
	case config.ModeSingle:
		return source.Single{URL: cfg.Target}, nil
	case config.ModeSitemap:
		return source.Sitemap{
			SitemapURL: cfg.Target,
			Client:     client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
}
