package crawl

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/algolia/mini-crawler/pkg/crawler"
	"github.com/schollz/progressbar/v3"
)

type terminalReporter struct {
	out      io.Writer
	path     string
	progress bool
	mu       sync.Mutex
	bar      *progressbar.ProgressBar
	started  bool
}

func newTerminalReporter(path string, progress bool) *terminalReporter {
	return &terminalReporter{out: os.Stderr, path: path, progress: progress}
}

func (r *terminalReporter) Start(snapshot crawler.MetricsSnapshot) {
	if !r.progress {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.started = true
	r.bar = progressbar.NewOptions(
		snapshot.Total,
		progressbar.OptionSetWriter(r.out),
		progressbar.OptionSetDescription("[cyan]Crawling[reset]"),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(32),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("pages"),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]█[reset]",
			SaucerHead:    "[green]█[reset]",
			SaucerPadding: "[green]░[reset]",
			BarStart:      "",
			BarEnd:        "",
		}),
	)
	r.render(snapshot)
}

func (r *terminalReporter) Update(snapshot crawler.MetricsSnapshot) {
	if !r.progress {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.render(snapshot)
}

func (r *terminalReporter) Finish(snapshot crawler.MetricsSnapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.progress && r.started {
		r.render(snapshot)

		if r.bar != nil {
			_ = r.bar.Exit()
		}

		fmt.Fprintln(r.out)
	}

	fmt.Fprintf(r.out, "Done in %s\n", formatDuration(snapshot.Elapsed))
	fmt.Fprintf(
		r.out,
		"Pages: %d ok, %d failed, %d skipped\n",
		snapshot.Pages,
		snapshot.Failed,
		snapshot.Skipped,
	)
	fmt.Fprintf(r.out, "Records: %d\n", snapshot.Records)
	fmt.Fprintf(r.out, "Output: %s\n", r.path)
}

func (r *terminalReporter) render(snapshot crawler.MetricsSnapshot) {
	if r.bar == nil {
		return
	}

	r.bar.Describe(fmt.Sprintf(
		"[cyan]Crawling[reset] %d failed  %d skipped  %d records",
		snapshot.Failed,
		snapshot.Skipped,
		snapshot.Records,
	))
	_ = r.bar.Set(completedPages(snapshot))
}

func completedPages(snapshot crawler.MetricsSnapshot) int {
	return int(snapshot.Pages + snapshot.Failed + snapshot.Skipped)
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}

	return d.Round(100 * time.Millisecond).String()
}

func stderrIsTerminal() bool {
	info, err := os.Stderr.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
