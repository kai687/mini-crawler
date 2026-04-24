package config

import (
	"errors"
	"flag"
	"testing"
)

func TestFromFlagsSingleModeDefault(t *testing.T) {
	cfg, err := FromFlags([]string{"https://example.com/page"})
	if err != nil {
		t.Fatalf("FromFlags() err = %v", err)
	}

	if cfg.Mode != ModeSingle {
		t.Fatalf("Mode = %q, want %q", cfg.Mode, ModeSingle)
	}

	if cfg.Target != "https://example.com/page" {
		t.Fatalf("Target = %q", cfg.Target)
	}
}

func TestFromFlagsSitemapMode(t *testing.T) {
	cfg, err := FromFlags([]string{"--mode", "sitemap", "https://example.com/sitemap.xml"})
	if err != nil {
		t.Fatalf("FromFlags() err = %v", err)
	}

	if cfg.Mode != ModeSitemap {
		t.Fatalf("Mode = %q, want %q", cfg.Mode, ModeSitemap)
	}

	if cfg.Target != "https://example.com/sitemap.xml" {
		t.Fatalf("Target = %q", cfg.Target)
	}
}

func TestFromFlagsTooManyArgs(t *testing.T) {
	_, err := FromFlags([]string{"https://example.com/one", "https://example.com/two"})
	if err == nil || err.Error() != "too many arguments" {
		t.Fatalf("err = %v, want too many arguments", err)
	}
}

func TestFromFlagsMissingTarget(t *testing.T) {
	_, err := FromFlags(nil)
	if err == nil || err.Error() != "single mode need URL argument" {
		t.Fatalf("err = %v, want missing single URL error", err)
	}
}

func TestFromFlagsVerbose(t *testing.T) {
	cfg, err := FromFlags([]string{"--verbose", "https://example.com/page"})
	if err != nil {
		t.Fatalf("FromFlags() err = %v", err)
	}

	if !cfg.Verbose {
		t.Fatal("Verbose = false, want true")
	}
}

func TestFromFlagsWorkers(t *testing.T) {
	cfg, err := FromFlags([]string{"--workers", "4", "https://example.com/page"})
	if err != nil {
		t.Fatalf("FromFlags() err = %v", err)
	}

	if cfg.Workers != 4 {
		t.Fatalf("Workers = %d, want 4", cfg.Workers)
	}
}

func TestFromFlagsFailOnError(t *testing.T) {
	cfg, err := FromFlags([]string{"--fail-on-error", "https://example.com/page"})
	if err != nil {
		t.Fatalf("FromFlags() err = %v", err)
	}

	if !cfg.FailOnError {
		t.Fatal("FailOnError = false, want true")
	}
}

func TestFromFlagsFilter(t *testing.T) {
	cfg, err := FromFlags(
		[]string{"--mode", "sitemap", "--filter", "doc/guides", "https://example.com/sitemap.xml"},
	)
	if err != nil {
		t.Fatalf("FromFlags() err = %v", err)
	}

	if cfg.Filter != "doc/guides" {
		t.Fatalf("Filter = %q, want %q", cfg.Filter, "doc/guides")
	}
}

func TestFromFlagsHelp(t *testing.T) {
	_, err := FromFlags([]string{"-help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("err = %v, want flag.ErrHelp", err)
	}
}
