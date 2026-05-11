package crawl

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime/pprof"
	"time"

	"github.com/kai687/mini-crawler/pkg/crawler"
	"github.com/kai687/mini-crawler/pkg/extract"
	"github.com/kai687/mini-crawler/pkg/fetch"
	"github.com/kai687/mini-crawler/pkg/output"
	"github.com/kai687/mini-crawler/pkg/parse"
	"github.com/spf13/cobra"
)

type config struct {
	Verbose         bool
	DebugScript     bool
	Output          string
	Script          string
	RequestRate     float64
	MetricsInterval time.Duration
	CPUProfile      string
	IgnoreNoindex   bool
}

func NewCommand(ctx context.Context) *cobra.Command {
	cfg := config{}

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl docs pages and emit records",
	}

	cmd.PersistentFlags().BoolVar(&cfg.Verbose, "verbose", false, "show crawl logs")
	cmd.PersistentFlags().
		BoolVar(&cfg.DebugScript, "debug-script", false, "show script matching and extraction logs")
	cmd.PersistentFlags().
		StringVarP(&cfg.Output, "output", "o", "", "write records to file instead of stdout")
	cmd.PersistentFlags().
		StringVarP(&cfg.Script, "script", "s", "", "required Starlark script for site-specific extraction")
	cmd.PersistentFlags().
		Float64Var(&cfg.RequestRate, "rate", 0, "maximum page requests per second (0 disables limit)")
	cmd.PersistentFlags().DurationVar(
		&cfg.MetricsInterval,
		"metrics-interval",
		10*time.Second,
		"log crawl metrics at this interval when verbose (0 disables periodic logs)",
	)
	cmd.PersistentFlags().StringVar(&cfg.CPUProfile, "cpu-profile", "", "write CPU profile to file")
	cmd.PersistentFlags().BoolVar(
		&cfg.IgnoreNoindex,
		"ignore-noindex",
		false,
		"index pages even when they have robots noindex metadata",
	)

	cmd.AddCommand(newSitemapCommand(ctx, &cfg))
	cmd.AddCommand(newSingleCommand(ctx, &cfg))

	return cmd
}

func runCrawl(ctx context.Context, cfg config, pipeline crawler.Pipeline) error {
	if cfg.Script == "" {
		return fmt.Errorf("invalid config: script flag required")
	}

	stopProfile, err := startCPUProfile(cfg.CPUProfile)
	if err != nil {
		return err
	}
	defer stopProfile()

	configureLogger(cfg.Verbose, cfg.DebugScript)

	out, closeOut, err := openOutput(cfg.Output)
	if err != nil {
		return fmt.Errorf("open output: %w", err)
	}
	defer closeOut()

	extractor, err := extract.NewStarlarkExtractor(cfg.Script, cfg.DebugScript)
	if err != nil {
		return err
	}

	if cfg.RequestRate < 0 {
		return fmt.Errorf("invalid config: rate must be >= 0")
	}

	pipeline.Fetcher = fetch.HTTPFetcher{Client: newHTTPClient()}

	pipeline.Parser = parse.HTMLParser{}
	if !cfg.IgnoreNoindex {
		pipeline.ParsedPageFilter = crawler.RobotsNoindexFilter{}
	}

	pipeline.Extractor = extractor
	pipeline.Writer = output.NewJSONLWriter(out)
	pipeline.RequestRate = cfg.RequestRate

	pipeline.MetricsInterval = cfg.MetricsInterval
	if cfg.Output != "" {
		pipeline.Reporter = newTerminalReporter(cfg.Output, stderrIsTerminal())
	}

	if err := crawler.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	return nil
}

func startCPUProfile(path string) (func(), error) {
	if path == "" {
		return func() {}, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create cpu profile: %w", err)
	}

	if err := pprof.StartCPUProfile(file); err != nil {
		_ = file.Close()

		return nil, fmt.Errorf("start cpu profile: %w", err)
	}

	return func() {
		pprof.StopCPUProfile()

		_ = file.Close()
	}, nil
}

func openOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}

	return file, func() { _ = file.Close() }, nil
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   32,
			MaxConnsPerHost:       64,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func configureLogger(verbose bool, debugScript bool) {
	logOut := io.Discard
	if verbose || debugScript {
		logOut = os.Stderr
	}

	level := slog.LevelInfo
	if debugScript {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(logOut, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
}
