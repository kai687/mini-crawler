package cmd

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/algolia/docs-crawler/internal/app"
	"github.com/algolia/docs-crawler/internal/config"
	"github.com/spf13/cobra"
)

func newCrawlCommand(ctx context.Context) *cobra.Command {
	cfg := config.Config{Workers: 1}

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl docs pages and emit records",
	}

	cmd.PersistentFlags().BoolVar(&cfg.Verbose, "verbose", false, "show crawl logs")
	cmd.PersistentFlags().
		StringVar(&cfg.Output, "output", "", "write records to file instead of stdout")
	cmd.PersistentFlags().
		StringVar(&cfg.Script, "script", "", "required Starlark script for site-specific extraction")

	cmd.AddCommand(newCrawlSitemapCommand(ctx, &cfg))
	cmd.AddCommand(newCrawlSingleCommand(ctx, &cfg))

	return cmd
}

func runCrawl(ctx context.Context, cfg config.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	configureLogger(cfg.Verbose)

	out, closeOut, err := openOutput(cfg.Output)
	if err != nil {
		return fmt.Errorf("open output: %w", err)
	}
	defer closeOut()

	if err := app.Run(ctx, cfg, out); err != nil {
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

func configureLogger(verbose bool) {
	logOut := io.Discard
	if verbose {
		logOut = os.Stderr
	}

	logger := slog.New(slog.NewTextHandler(logOut, nil))
	slog.SetDefault(logger)
}
