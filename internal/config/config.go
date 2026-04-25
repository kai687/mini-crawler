package config

import (
	"errors"
	"flag"
	"fmt"
)

// Mode selects how crawl target URLs are discovered.
type Mode string

const (
	// ModeSingle crawls one explicit page URL.
	ModeSingle Mode = "single"
	// ModeSitemap crawls all URLs discovered from a sitemap URL.
	ModeSitemap Mode = "sitemap"
)

// Config holds validated CLI configuration for one crawler run.
type Config struct {
	Mode        Mode
	Target      string
	Verbose     bool
	Workers     int
	FailOnError bool
	Filter      string
	Coverage    bool
	Output      string
}

// FromFlags parses CLI flags and positional arguments into Config.
func FromFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("docs-crawler", flag.ContinueOnError)

	var cfg Config

	fs.StringVar((*string)(&cfg.Mode), "mode", string(ModeSitemap), "crawl mode: single or sitemap")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "show crawl logs")
	fs.IntVar(&cfg.Workers, "workers", 1, "number of concurrent page workers")
	fs.BoolVar(
		&cfg.FailOnError,
		"fail-on-error",
		false,
		"fail whole run when one URL cannot be crawled",
	)
	fs.StringVar(&cfg.Filter, "filter", "", "substring filter for sitemap URLs")
	fs.BoolVar(&cfg.Coverage, "coverage", false, "print URL coverage summary to stderr after crawl")
	fs.StringVar(&cfg.Output, "output", "", "write records to file instead of stdout")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage:")
		fmt.Fprintln(fs.Output(), "  docs-crawler [flags] <sitemap-url>")
		fmt.Fprintln(fs.Output(), "  docs-crawler --mode single [flags] <url>")
		fmt.Fprintln(fs.Output())
		fmt.Fprintln(fs.Output(), "Flags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	rest := fs.Args()
	if len(rest) > 1 {
		return Config{}, errors.New("too many arguments")
	}

	if len(rest) == 1 {
		cfg.Target = rest[0]
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks that config fields are internally consistent.
func (c Config) Validate() error {
	switch c.Mode {
	case ModeSingle:
		if c.Target == "" {
			return errors.New("single mode need URL argument")
		}
	case ModeSitemap:
		if c.Target == "" {
			return errors.New("sitemap mode need sitemap URL argument")
		}
	default:
		return fmt.Errorf("unknown mode %q", c.Mode)
	}

	return nil
}
