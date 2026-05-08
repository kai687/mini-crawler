package config

import (
	"errors"
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
	Output      string
	Script      string
}

// Validate checks that config fields are internally consistent.
func (c Config) Validate() error {
	if c.Script == "" {
		return errors.New("script flag required")
	}

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
