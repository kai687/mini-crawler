package source

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kai687/mini-crawler/pkg/httpheaders"
)

// Sitemap loads page URLs from a sitemap XML document.
//
// It resolves relative <loc> values against SitemapURL so downstream crawl code
// always receives absolute page URLs.
type Sitemap struct {
	// SitemapURL is the XML sitemap URL to fetch.
	SitemapURL string
	// Client performs sitemap HTTP requests. Nil uses http.DefaultClient.
	Client *http.Client
}

const maxSitemapIndexDepth = 10

type sitemapKind string

const (
	sitemapKindURLSet sitemapKind = "urlset"
	sitemapKindIndex  sitemapKind = "sitemapindex"
)

type parsedSitemap struct {
	kind sitemapKind
	locs []string
}

// URLs fetches the sitemap and returns resolved page URLs.
func (s Sitemap) URLs(ctx context.Context) ([]string, error) {
	if s.Client == nil {
		s.Client = http.DefaultClient
	}

	return s.urls(ctx, s.SitemapURL, map[string]bool{}, 0)
}

func (s Sitemap) urls(
	ctx context.Context,
	sitemapURL string,
	visited map[string]bool,
	depth int,
) ([]string, error) {
	if depth > maxSitemapIndexDepth {
		return nil, fmt.Errorf("sitemap index depth exceeded %d", maxSitemapIndexDepth)
	}

	if visited[sitemapURL] {
		return nil, nil
	}

	visited[sitemapURL] = true

	parsed, err := s.fetchAndParse(ctx, sitemapURL)
	if err != nil {
		return nil, err
	}

	switch parsed.kind {
	case sitemapKindURLSet:
		return parsed.locs, nil
	case sitemapKindIndex:
		// Expand child sitemaps below.
	default:
		return nil, fmt.Errorf("unsupported sitemap root %q", parsed.kind)
	}

	urls := []string{}
	for _, childURL := range parsed.locs {
		childURLs, err := s.urls(ctx, childURL, visited, depth+1)
		if err != nil {
			return nil, err
		}

		urls = append(urls, childURLs...)
	}

	return urls, nil
}

func (s Sitemap) fetchAndParse(ctx context.Context, sitemapURL string) (parsedSitemap, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sitemapURL, nil)
	if err != nil {
		return parsedSitemap{}, fmt.Errorf("build sitemap request: %w", err)
	}

	req.Header.Set("User-Agent", httpheaders.UserAgent)

	resp, err := s.Client.Do(req)
	if err != nil {
		return parsedSitemap{}, fmt.Errorf("fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	base, err := url.Parse(sitemapURL)
	if err != nil {
		return parsedSitemap{}, fmt.Errorf("parse sitemap url: %w", err)
	}

	parsed, err := parseSitemap(resp.Body, base)
	if err != nil {
		return parsedSitemap{}, fmt.Errorf("parse sitemap xml: %w", err)
	}

	return parsed, nil
}

func parseSitemap(r io.Reader, base *url.URL) (parsedSitemap, error) {
	decoder := xml.NewDecoder(r)
	parsed := parsedSitemap{}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return parsed, nil
		}

		if err != nil {
			return parsedSitemap{}, err
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		if parsed.kind == "" {
			parsed.kind = sitemapKind(start.Name.Local)
		}

		if start.Name.Local != "loc" {
			continue
		}

		var raw string
		if err := decoder.DecodeElement(&raw, &start); err != nil {
			return parsedSitemap{}, err
		}

		raw = strings.TrimSpace(raw)

		resolved, err := resolveURL(base, raw)
		if err != nil {
			return parsedSitemap{}, fmt.Errorf("resolve sitemap url %q: %w", raw, err)
		}

		parsed.locs = append(parsed.locs, resolved)
	}
}

// resolveURL resolves one sitemap <loc> value against the sitemap URL.
func resolveURL(base *url.URL, raw string) (string, error) {
	ref, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(ref).String(), nil
}
