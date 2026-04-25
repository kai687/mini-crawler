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

	out, closeOut, err := openOutput(cfg.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open output: %v\n", err)
		os.Exit(1)
	}
	defer closeOut()

	if err := app.Run(context.Background(), cfg, out); err != nil {
		fmt.Fprintf(os.Stderr, "crawl failed: %v\n", err)
		os.Exit(1)
	}
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
