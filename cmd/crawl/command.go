package crawl

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/algolia/docs-crawler/pkg/crawler"
	"github.com/algolia/docs-crawler/pkg/extract"
	"github.com/algolia/docs-crawler/pkg/fetch"
	"github.com/algolia/docs-crawler/pkg/output"
	"github.com/algolia/docs-crawler/pkg/parse"
	"github.com/spf13/cobra"
)

type config struct {
	Verbose     bool
	DebugScript bool
	Output      string
	Script      string
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
		StringVar(&cfg.Output, "output", "", "write records to file instead of stdout")
	cmd.PersistentFlags().
		StringVar(&cfg.Script, "script", "", "required Starlark script for site-specific extraction")

	cmd.AddCommand(newSitemapCommand(ctx, &cfg))
	cmd.AddCommand(newSingleCommand(ctx, &cfg))

	return cmd
}

func runCrawl(ctx context.Context, cfg config, pipeline crawler.Pipeline) error {
	if cfg.Script == "" {
		return fmt.Errorf("invalid config: script flag required")
	}

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

	pipeline.Fetcher = fetch.HTTPFetcher{}
	pipeline.Parser = parse.HTMLParser{}
	pipeline.Extractor = extractor
	pipeline.Writer = output.NewJSONLWriter(out)

	if err := crawler.Run(ctx, pipeline); err != nil {
		return fmt.Errorf("crawl failed: %w", err)
	}

	return nil
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
