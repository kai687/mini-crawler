package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/antchfx/xmlquery"
)

// Sitemap loads page URLs from a sitemap XML document.
type Sitemap struct {
	SitemapURL string
	Filter     string
	Client     *http.Client
}

// URLs fetches the sitemap and returns resolved page URLs.
func (s Sitemap) URLs(ctx context.Context) ([]string, error) {
	if s.Client == nil {
		s.Client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.SitemapURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build sitemap request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	doc, err := xmlquery.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse sitemap xml: %w", err)
	}

	base, err := url.Parse(s.SitemapURL)
	if err != nil {
		return nil, fmt.Errorf("parse sitemap url: %w", err)
	}

	nodes := xmlquery.Find(doc, "//loc")

	urls := make([]string, 0, len(nodes))
	for _, node := range nodes {
		raw := node.InnerText()

		resolved, err := resolveURL(base, raw)
		if err != nil {
			return nil, fmt.Errorf("resolve sitemap url %q: %w", raw, err)
		}

		if s.Filter != "" && !strings.Contains(resolved, s.Filter) {
			continue
		}

		urls = append(urls, resolved)
	}

	return urls, nil
}

func resolveURL(base *url.URL, raw string) (string, error) {
	ref, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(ref).String(), nil
}
