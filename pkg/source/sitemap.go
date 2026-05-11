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

// URLs fetches the sitemap and returns resolved page URLs.
func (s Sitemap) URLs(ctx context.Context) ([]string, error) {
	if s.Client == nil {
		s.Client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.SitemapURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build sitemap request: %w", err)
	}

	req.Header.Set("User-Agent", httpheaders.UserAgent)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	base, err := url.Parse(s.SitemapURL)
	if err != nil {
		return nil, fmt.Errorf("parse sitemap url: %w", err)
	}

	urls, err := parseSitemapURLs(resp.Body, base)
	if err != nil {
		return nil, fmt.Errorf("parse sitemap xml: %w", err)
	}

	return urls, nil
}

func parseSitemapURLs(r io.Reader, base *url.URL) ([]string, error) {
	decoder := xml.NewDecoder(r)
	urls := []string{}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return urls, nil
		}

		if err != nil {
			return nil, err
		}

		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "loc" {
			continue
		}

		var raw string
		if err := decoder.DecodeElement(&raw, &start); err != nil {
			return nil, err
		}

		raw = strings.TrimSpace(raw)

		resolved, err := resolveURL(base, raw)
		if err != nil {
			return nil, fmt.Errorf("resolve sitemap url %q: %w", raw, err)
		}

		urls = append(urls, resolved)
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
