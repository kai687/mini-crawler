package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/algolia/docs-crawler/internal/app"
	"github.com/algolia/docs-crawler/internal/config"
)

func main() {
	cfg, err := config.FromFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

	configureLogger(cfg.Verbose)

	if err := app.Run(context.Background(), cfg, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "crawl failed: %v\n", err)
		os.Exit(1)
	}
}

func configureLogger(verbose bool) {
	logOut := io.Discard
	if verbose {
		logOut = os.Stderr
	}

	logger := slog.New(slog.NewTextHandler(logOut, nil))
	slog.SetDefault(logger)
}
