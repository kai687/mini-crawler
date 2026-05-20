package parse

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/kai687/mini-crawler/pkg/model"
)

// HTMLParser converts fetched page bytes into a goquery document.
//
// It is intentionally small: fetching already happened, and extraction engines
// decide how to query the resulting DOM.
type HTMLParser struct{}

// Parse reads HTML bytes from a Page and returns a ParsedPage.
func (p HTMLParser) Parse(page model.Page) (model.ParsedPage, error) {
	body := stripRawTextElements(page.Body)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return model.ParsedPage{}, fmt.Errorf("parse html: %w", err)
	}

	return model.ParsedPage{
		Ref:      page.Ref,
		URL:      page.URL,
		Kind:     "html",
		Document: doc,
		Doc:      doc,
		Metadata: page.Metadata,
	}, nil
}

func stripRawTextElements(body []byte) []byte {
	out := []byte(nil)
	start := 0

	for i := 0; i < len(body); {
		next := bytes.IndexByte(body[i:], '<')
		if next < 0 {
			break
		}

		i += next

		tag, ok := rawTextTagAt(body, i)
		if !ok {
			i++

			continue
		}

		openEnd := bytes.IndexByte(body[i:], '>')
		if openEnd < 0 {
			return appendUntil(body, out, start, i)
		}

		closeStart := findClosingTag(body, i+openEnd+1, tag)
		if closeStart < 0 {
			return appendUntil(body, out, start, i)
		}

		closeEnd := bytes.IndexByte(body[closeStart:], '>')
		if closeEnd < 0 {
			return appendUntil(body, out, start, i)
		}

		if out == nil {
			out = make([]byte, 0, len(body)-(closeStart+closeEnd+1-i))
		}

		out = append(out, body[start:i]...)
		i = closeStart + closeEnd + 1
		start = i
	}

	if out == nil {
		return body
	}

	out = append(out, body[start:]...)

	return out
}

func appendUntil(body, out []byte, start, end int) []byte {
	if out == nil {
		return body[:end]
	}

	return append(out, body[start:end]...)
}

func rawTextTagAt(body []byte, i int) (string, bool) {
	if !hasASCIIPrefixFold(body[i:], []byte("<script")) {
		if !hasASCIIPrefixFold(body[i:], []byte("<style")) {
			return "", false
		}

		if !isTagBoundary(body, i+len("<style")) {
			return "", false
		}

		return "style", true
	}

	if !isTagBoundary(body, i+len("<script")) {
		return "", false
	}

	return "script", true
}

func findClosingTag(body []byte, start int, tag string) int {
	closing := []byte("</" + tag)

	for i := start; i < len(body); {
		next := bytes.IndexByte(body[i:], '<')
		if next < 0 {
			return -1
		}

		i += next

		if hasASCIIPrefixFold(body[i:], closing) && isTagBoundary(body, i+len(closing)) {
			return i
		}

		i++
	}

	return -1
}

func hasASCIIPrefixFold(value, prefix []byte) bool {
	if len(value) < len(prefix) {
		return false
	}

	for i, want := range prefix {
		if asciiLower(value[i]) != asciiLower(want) {
			return false
		}
	}

	return true
}

func asciiLower(value byte) byte {
	if 'A' <= value && value <= 'Z' {
		return value + ('a' - 'A')
	}

	return value
}

func isTagBoundary(body []byte, i int) bool {
	if i >= len(body) {
		return true
	}

	switch body[i] {
	case ' ', '\t', '\n', '\r', '\f', '/', '>':
		return true
	default:
		return false
	}
}
