package recordutil

import (
	"fmt"
	"net/url"
	"strings"
)

// ObjectIDFromURL derives a stable record identifier from a page URL.
func ObjectIDFromURL(pageURL string) string {
	parsed, err := url.Parse(pageURL)
	if err == nil {
		pageURL = normalizedObjectIDURL(*parsed, pageURL)
	}

	id := strings.Trim(pageURL, "/")

	return strings.ReplaceAll(id, "/", "-")
}

func normalizedObjectIDURL(parsed url.URL, raw string) string {
	if parsed.Scheme != "" && parsed.Host != "" {
		value := parsed.EscapedPath()
		if parsed.RawFragment != "" {
			value += "#" + parsed.RawFragment
		} else if parsed.Fragment != "" {
			value += "#" + parsed.Fragment
		}

		if parsed.RawQuery != "" {
			value += "?" + parsed.RawQuery
		}

		return value
	}

	return raw
}

// ObjectIDWithPosition appends extraction position to an object ID for uniqueness.
func ObjectIDWithPosition(objectID string, position int) string {
	return fmt.Sprintf("%s-%d", objectID, position)
}
