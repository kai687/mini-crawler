package crawler

import (
	"log/slog"
	"sync/atomic"
	"time"
)

type crawlMetrics struct {
	started  time.Time
	requests atomic.Int64
	pages    atomic.Int64
	records  atomic.Int64
}

func newCrawlMetrics() *crawlMetrics {
	return &crawlMetrics{started: time.Now()}
}

func (m *crawlMetrics) addRequest() {
	if m != nil {
		m.requests.Add(1)
	}
}

func (m *crawlMetrics) addPage() {
	if m != nil {
		m.pages.Add(1)
	}
}

func (m *crawlMetrics) addRecords(n int) {
	if m != nil && n > 0 {
		m.records.Add(int64(n))
	}
}

func (m *crawlMetrics) log(message string) {
	if m == nil {
		return
	}

	elapsed := time.Since(m.started).Seconds()
	if elapsed <= 0 {
		elapsed = 1
	}

	requests := m.requests.Load()
	pages := m.pages.Load()
	records := m.records.Load()

	slog.Info(
		message,
		"elapsed", time.Since(m.started).Round(time.Millisecond),
		"requests", requests,
		"pages", pages,
		"records", records,
		"requests_per_second", float64(requests)/elapsed,
		"pages_per_second", float64(pages)/elapsed,
		"records_per_second", float64(records)/elapsed,
	)
}

func startMetricsLogger(done <-chan struct{}, metrics *crawlMetrics, interval time.Duration) {
	if metrics == nil || interval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				metrics.log("crawl metrics")
			}
		}
	}()
}
