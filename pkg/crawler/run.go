package crawler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/kai687/mini-crawler/pkg/script"
)

// ProgressReporter receives crawl progress updates.
type ProgressReporter interface {
	Start(snapshot MetricsSnapshot)
	Update(snapshot MetricsSnapshot)
	Finish(snapshot MetricsSnapshot)
}

// Pipeline configures one crawl run from discovery through output.
type Pipeline struct {
	Source           Source
	RefFilter        RefFilter
	Fetcher          Fetcher
	PreParseFilter   PreParseFilter
	Parser           Parser
	ParsedPageFilter ParsedPageFilter
	Extractor        Extractor
	Writer           Writer
	Reporter         ProgressReporter

	Workers         int
	FailOnError     bool
	RequestRate     float64
	MetricsInterval time.Duration
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

	if p.RequestRate < 0 {
		return errors.New("request rate must be >= 0")
	}

	metrics := newCrawlMetrics()

	urls, err := p.Source.URLs(ctx)
	if err != nil {
		return fmt.Errorf("load urls: %w", err)
	}

	workers := normalizedWorkers(p.Workers)
	slog.Info("crawl start", "urls", len(urls), "workers", workers, "request_rate", p.RequestRate)

	if p.Reporter != nil {
		p.Reporter.Start(metrics.snapshot(len(urls)))
	}

	limiter := newRequestLimiter(p.RequestRate)
	defer limiter.stop()

	processor := newProcessor(p, limiter, metrics)

	err = runPages(ctx, p, processor, urls, workers, metrics)
	metrics.log("crawl metrics final")

	if p.Reporter != nil {
		p.Reporter.Finish(metrics.snapshot(len(urls)))
	}

	return err
}

func newProcessor(p Pipeline, limiter *requestLimiter, metrics *crawlMetrics) pipelineProcessor {
	return newPipelineProcessor(
		p.RefFilter,
		p.Fetcher,
		p.PreParseFilter,
		p.Parser,
		p.ParsedPageFilter,
		p.Extractor,
		limiter,
		metrics,
	)
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
	metrics *crawlMetrics,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	metricsDone := make(chan struct{})
	defer close(metricsDone)

	startMetricsLogger(metricsDone, metrics, p.MetricsInterval)

	jobs := make(chan string)
	results := make(chan pageResult)
	startWorkers(ctx, processor, workers, jobs, results)
	sendJobs(ctx, urls, jobs)

	state := runState{}
	for result := range results {
		if err := handlePageResult(
			p.Writer,
			result,
			p.FailOnError,
			cancel,
			&state,
			metrics,
		); err != nil {
			return err
		}

		if p.Reporter != nil {
			p.Reporter.Update(metrics.snapshot(len(urls)))
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
	metrics *crawlMetrics,
) error {
	if result.err != nil {
		if handled := handleSkippedPage(result, metrics); handled {
			return nil
		}

		state.failed++

		metrics.addFailed()

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

	written := 0

	for _, record := range result.records {
		if err := writer.Write(record); err != nil {
			cancel()

			return fmt.Errorf("write record for %s: %w", result.url, err)
		}

		written++
	}

	metrics.addRecords(written)

	return nil
}

func handleSkippedPage(result pageResult, metrics *crawlMetrics) bool {
	message, ok := skipMessage(result.err)
	if !ok {
		return false
	}

	metrics.addSkipped()
	slog.Warn(message, "url", result.url)

	return true
}

func skipMessage(err error) (string, bool) {
	switch {
	case errors.Is(err, script.ErrNoExtractor):
		return "page skipped: no extractor matches", true
	case errors.Is(err, ErrNoindex):
		return "page skipped: robots noindex", true
	case errors.Is(err, ErrFiltered):
		return "page skipped: filter", true
	default:
		return "", false
	}
}
