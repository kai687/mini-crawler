package recordutil

import (
	"net/url"
	"strings"
	"unicode"

	"github.com/algolia/docs-crawler/internal/model"
)

var breadcrumbTokenCasing = map[string]string{
	"ai":           "AI",
	"algolia":      "Algolia",
	"api":          "API",
	"apis":         "APIs",
	"autocomplete": "Autocomplete",
	"cli":          "CLI",
	"cpu":          "CPU",
	"css":          "CSS",
	"csv":          "CSV",
	"dns":          "DNS",
	"faq":          "FAQ",
	"graphql":      "GraphQL",
	"html":         "HTML",
	"http":         "HTTP",
	"https":        "HTTPS",
	"id":           "ID",
	"ids":          "IDs",
	"ios":          "iOS",
	"ip":           "IP",
	"ipv4":         "IPv4",
	"ipv6":         "IPv6",
	"javascript":   "JavaScript",
	"json":         "JSON",
	"jwt":          "JWT",
	"ml":           "ML",
	"ocr":          "OCR",
	"pdf":          "PDF",
	"php":          "PHP",
	"ram":          "RAM",
	"rest":         "REST",
	"rpc":          "RPC",
	"sdk":          "SDK",
	"sql":          "SQL",
	"ssh":          "SSH",
	"ssl":          "SSL",
	"sso":          "SSO",
	"tcp":          "TCP",
	"tls":          "TLS",
	"typescript":   "TypeScript",
	"ui":           "UI",
	"uid":          "UID",
	"url":          "URL",
	"urls":         "URLs",
	"utc":          "UTC",
	"ux":           "UX",
	"xml":          "XML",
}

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

// BreadcrumbPathFromURL returns URL path with leading /doc removed.
func BreadcrumbPathFromURL(pageURL string) string {
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

// BreadcrumbSegmentsFromURL returns ancestor breadcrumb labels derived from URL path.
// Final path segment is excluded because current page label comes from title/heading fields.
func BreadcrumbSegmentsFromURL(pageURL string) []string {
	path := strings.Trim(BreadcrumbPathFromURL(pageURL), "/")
	if path == "" {
		return nil
	}

	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return nil
	}

	segments := make([]string, 0, len(parts)-1)
	for _, part := range parts[:len(parts)-1] {
		if part == "" {
			continue
		}

		segments = append(segments, humanizeSlug(part))
	}

	if len(segments) == 0 {
		return nil
	}

	return segments
}

// BreadcrumbHierarchyFromSegments builds cumulative breadcrumb hierarchy.
func BreadcrumbHierarchyFromSegments(segments []string) *model.BreadcrumbHierarchy {
	if len(segments) == 0 {
		return nil
	}

	h := &model.BreadcrumbHierarchy{}

	levels := make([]string, 0, len(segments))
	for i, segment := range segments {
		levels = append(levels, segment)
		value := strings.Join(levels, " > ")

		switch i {
		case 0:
			h.Lvl0 = stringPtr(value)
		case 1:
			h.Lvl1 = stringPtr(value)
		case 2:
			h.Lvl2 = stringPtr(value)
		case 3:
			h.Lvl3 = stringPtr(value)
		case 4:
			h.Lvl4 = stringPtr(value)
		case 5:
			h.Lvl5 = stringPtr(value)
		}
	}

	return h
}

func humanizeSlug(value string) string {
	parts := strings.Fields(strings.ReplaceAll(value, "-", " "))
	for i, part := range parts {
		normalized := strings.ToLower(part)
		if casing, ok := breadcrumbTokenCasing[normalized]; ok {
			parts[i] = casing

			continue
		}

		if i == 0 {
			parts[i] = sentenceCaseFirstWord(normalized)

			continue
		}

		parts[i] = normalized
	}

	return strings.Join(parts, " ")
}

func sentenceCaseFirstWord(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}

	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
