package source

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/kai687/mini-crawler/pkg/httpheaders"
)

// LLMS loads page URLs from an llms.txt Markdown document.
//
// It discovers URLs from Markdown inline links and resolves relative targets
// against LLMSURL so downstream crawl code always receives absolute URLs.
type LLMS struct {
	// LLMSURL is the llms.txt URL to fetch.
	LLMSURL string
	// Client performs llms.txt HTTP requests. Nil uses http.DefaultClient.
	Client *http.Client
}

var markdownLinkRE = regexp.MustCompile(`\[[^\]]+\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

// URLs fetches llms.txt and returns resolved document URLs.
func (l LLMS) URLs(ctx context.Context) ([]string, error) {
	if l.Client == nil {
		l.Client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.LLMSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build llms.txt request: %w", err)
	}

	req.Header.Set("User-Agent", httpheaders.UserAgent)

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch llms.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch llms.txt: unexpected status %s", resp.Status)
	}

	base := resp.Request.URL
	if base == nil {
		base, err = url.Parse(l.LLMSURL)
		if err != nil {
			return nil, fmt.Errorf("parse llms.txt url: %w", err)
		}
	}

	urls, err := parseLLMS(resp.Body, base)
	if err != nil {
		return nil, fmt.Errorf("parse llms.txt: %w", err)
	}

	return urls, nil
}

func parseLLMS(r io.Reader, base *url.URL) ([]string, error) {
	scanner := bufio.NewScanner(r)
	seen := map[string]bool{}
	urls := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		matches := markdownLinkRE.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			resolved, ok, err := resolveLLMSURL(base, match[1])
			if err != nil {
				return nil, err
			}

			if !ok || seen[resolved] {
				continue
			}

			seen[resolved] = true
			urls = append(urls, resolved)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func resolveLLMSURL(base *url.URL, raw string) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false, nil
	}

	ref, err := url.Parse(raw)
	if err != nil {
		return "", false, fmt.Errorf("resolve llms.txt url %q: %w", raw, err)
	}

	switch ref.Scheme {
	case "", "http", "https":
		// supported below
	default:
		return "", false, nil
	}

	resolved := base.ResolveReference(ref)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return "", false, nil
	}

	return resolved.String(), true, nil
}
