package recordutil

import (
	"net/url"
	"strings"
)

// URLWithAnchor returns pageURL with a fragment anchor when anchor is non-empty.
func URLWithAnchor(pageURL string, anchor string) string {
	if anchor == "" {
		return pageURL
	}

	return URLWithoutAnchor(pageURL) + "#" + anchor
}

// URLWithoutAnchor strips any fragment from pageURL.
func URLWithoutAnchor(pageURL string) string {
	before, _, _ := strings.Cut(pageURL, "#")

	return before
}

// BreadcrumbFromURL returns URL path with leading /doc removed.
func BreadcrumbFromURL(pageURL string) string {
	parsed, err := url.Parse(URLWithoutAnchor(pageURL))
	if err != nil {
		return ""
	}

	path := parsed.EscapedPath()
	if path == "" {
		path = parsed.Path
	}

	if path == "" || path == "/" {
		return ""
	}

	if strings.HasPrefix(path, "/doc") {
		path = strings.TrimPrefix(path, "/doc")
		if path == "" || path == "/" {
			return ""
		}
	}

	return path
}
