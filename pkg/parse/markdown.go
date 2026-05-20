package parse

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/kai687/mini-crawler/pkg/model"
)

// MarkdownParser extracts lightweight metadata from raw Markdown/MDX content.
type MarkdownParser struct{}

// MarkdownHeading is one ATX heading from Markdown/MDX source.
type MarkdownHeading struct {
	Level int
	Text  string
}

// MarkdownDocument is the parser output for Markdown/MDX pages.
type MarkdownDocument struct {
	Raw         string
	Title       string
	Description string
	Headings    []MarkdownHeading
}

// Parse reads Markdown bytes from a Page and returns a ParsedPage.
func (p MarkdownParser) Parse(page model.Page) (model.ParsedPage, error) {
	doc := parseMarkdownDocument(page.Body)
	metadata := copyMetadata(page.Metadata)
	metadata["title"] = doc.Title
	metadata["description"] = doc.Description
	metadata["headings"] = headingsToAny(doc.Headings)

	for level := 1; level <= 6; level++ {
		metadata[headingKey(level)] = headingTextsToAny(doc.Headings, level)
	}

	return model.ParsedPage{
		Ref:      page.Ref,
		URL:      page.URL,
		Kind:     "markdown",
		Document: doc,
		Metadata: metadata,
	}, nil
}

func parseMarkdownDocument(body []byte) MarkdownDocument {
	items := headings(body)
	title, h1Line := firstH1(body)

	return MarkdownDocument{
		Raw:         string(body),
		Title:       title,
		Description: firstBlockquoteAfterLine(body, h1Line),
		Headings:    items,
	}
}

func firstH1(body []byte) (string, int) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++

		if heading, ok := atxHeading(scanner.Text()); ok && heading.Level == 1 {
			return heading.Text, lineNumber
		}
	}

	return "", 0
}

func headings(body []byte) []MarkdownHeading {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	out := []MarkdownHeading{}

	for scanner.Scan() {
		if heading, ok := atxHeading(scanner.Text()); ok {
			out = append(out, heading)
		}
	}

	return out
}

func atxHeading(line string) (MarkdownHeading, bool) {
	line = strings.TrimSpace(line)
	level := 0

	for level < len(line) && level < 6 && line[level] == '#' {
		level++
	}

	if level == 0 || level == len(line) || line[level] != ' ' {
		return MarkdownHeading{}, false
	}

	heading := strings.TrimSpace(line[level+1:])

	heading = strings.TrimSpace(strings.TrimRight(heading, "#"))
	if heading == "" {
		return MarkdownHeading{}, false
	}

	return MarkdownHeading{Level: level, Text: heading}, true
}

func firstBlockquoteAfterLine(body []byte, afterLine int) string {
	if afterLine == 0 {
		return ""
	}

	scanner := bufio.NewScanner(bytes.NewReader(body))
	lineNumber := 0
	parts := []string{}
	collecting := false

	for scanner.Scan() {
		lineNumber++
		if lineNumber <= afterLine {
			continue
		}

		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, ">") {
			if collecting {
				break
			}

			continue
		}

		collecting = true

		text := strings.TrimSpace(strings.TrimPrefix(line, ">"))
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, " ")
}

func copyMetadata(metadata map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range metadata {
		out[key] = value
	}

	return out
}

func headingKey(level int) string {
	return "h" + string(rune('0'+level))
}

func headingTextsToAny(headings []MarkdownHeading, level int) []any {
	out := []any{}

	for _, heading := range headings {
		if heading.Level == level {
			out = append(out, heading.Text)
		}
	}

	return out
}

func headingsToAny(headings []MarkdownHeading) []any {
	out := make([]any, 0, len(headings))
	for _, heading := range headings {
		out = append(out, map[string]any{
			"level": heading.Level,
			"text":  heading.Text,
		})
	}

	return out
}
