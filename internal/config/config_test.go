package config

import "testing"

func TestValidateSitemapMode(t *testing.T) {
	cfg := Config{
		Mode:   ModeSitemap,
		Target: "https://example.com/sitemap.xml",
		Script: "site.star",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() err = %v", err)
	}
}

func TestValidateSingleMode(t *testing.T) {
	cfg := Config{Mode: ModeSingle, Target: "https://example.com/page", Script: "site.star"}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() err = %v", err)
	}
}

func TestValidateMissingScript(t *testing.T) {
	cfg := Config{Mode: ModeSingle, Target: "https://example.com/page"}

	err := cfg.Validate()
	if err == nil || err.Error() != "script flag required" {
		t.Fatalf("err = %v, want missing script error", err)
	}
}

func TestValidateMissingSitemapTarget(t *testing.T) {
	cfg := Config{Mode: ModeSitemap, Script: "site.star"}

	err := cfg.Validate()
	if err == nil || err.Error() != "sitemap mode need sitemap URL argument" {
		t.Fatalf("err = %v, want missing sitemap URL error", err)
	}
}

func TestValidateMissingSingleTarget(t *testing.T) {
	cfg := Config{Mode: ModeSingle, Script: "site.star"}

	err := cfg.Validate()
	if err == nil || err.Error() != "single mode need URL argument" {
		t.Fatalf("err = %v, want missing single URL error", err)
	}
}

func TestValidateUnknownMode(t *testing.T) {
	cfg := Config{Mode: Mode("unknown"), Target: "https://example.com", Script: "site.star"}

	err := cfg.Validate()
	if err == nil || err.Error() != `unknown mode "unknown"` {
		t.Fatalf("err = %v, want unknown mode error", err)
	}
}
