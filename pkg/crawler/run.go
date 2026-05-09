package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/algolia/docs-crawler/pkg/script"
)

// Pipeline configures one crawl run from discovery through output.
type Pipeline struct {
	Source    Source
	Fetcher   Fetcher
	Parser    Parser
	Extractor Extractor
	Writer    Writer

	Workers     int
	FailOnError bool
}

// Run executes one crawl pipeline.
func Run(ctx context.Context, p Pipeline) error {
	if p.Source == nil {
		return errors.New("source required")
	}

	if p.Fetcher == nil {
		return errors.New("fetcher required")
	}

	if p.Parser == nil {
		return errors.New("parser required")
	}

	if p.Extractor == nil {
		return errors.New("extractor required")
	}

	if p.Writer == nil {
		return errors.New("writer required")
	}
	defer p.Writer.Close()

	urls, err := p.Source.URLs(ctx)
	if err != nil {
		return fmt.Errorf("load urls: %w", err)
	}

	workers := normalizedWorkers(p.Workers)
	slog.Info("crawl start", "urls", len(urls), "workers", workers)

	processor := newPipelineProcessor(p.Fetcher, p.Parser, p.Extractor)

	return runPages(ctx, p, processor, urls, workers)
}

// normalizedWorkers keeps worker count valid for the fan-out loop.
func normalizedWorkers(workers int) int {
	return max(workers, 1)
}

// runPages coordinates worker goroutines and applies crawl error policy.
func runPages(
	ctx context.Context,
	p Pipeline,
	processor PageProcessor,
	urls []string,
	workers int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan string)
	results := make(chan pageResult)
	startWorkers(ctx, processor, workers, jobs, results)
	sendJobs(ctx, urls, jobs)

	state := runState{}
	for result := range results {
		if err := handlePageResult(p.Writer, result, p.FailOnError, cancel, &state); err != nil {
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

// pageResult carries one worker result back to the coordinator.
type pageResult struct {
	url     string
	records []any
	err     error
}

// runState tracks crawl failures while results stream in from workers.
type runState struct {
	firstErr error
	failed   int
}

// startWorkers launches fixed worker pool for page processing.
func startWorkers(
	ctx context.Context,
	processor PageProcessor,
	workers int,
	jobs <-chan string,
	results chan<- pageResult,
) {
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for pageURL := range jobs {
				records, err := processor.Process(ctx, pageURL)
				results <- pageResult{url: pageURL, records: records, err: err}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()
}

// sendJobs streams discovered URLs to workers until done or cancelled.
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

// handlePageResult writes successful records and records recoverable failures.
func handlePageResult(
	writer Writer,
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
