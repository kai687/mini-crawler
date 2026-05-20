package parse

import (
	"reflect"
	"testing"

	"github.com/kai687/mini-crawler/pkg/model"
)

func TestMarkdownParserExtractsTitleAndDescription(t *testing.T) {
	page := model.Page{
		Ref: "https://example.com/intro.md",
		URL: "https://example.com/intro.md",
		Body: []byte(`---
title: ignored
---

# Intro

Some text.

> First description line.
> Second description line.

## Next

### Included

#### Deep

## Last ##

> Later quote ignored.
`),
	}

	parsed, err := MarkdownParser{}.Parse(page)
	if err != nil {
		t.Fatalf("Parse() err = %v", err)
	}

	if parsed.Kind != "markdown" {
		t.Fatalf("Kind = %q, want markdown", parsed.Kind)
	}

	if parsed.Metadata["title"] != "Intro" {
		t.Fatalf("title = %#v, want Intro", parsed.Metadata["title"])
	}

	if parsed.Metadata["description"] != "First description line. Second description line." {
		t.Fatalf("description = %#v", parsed.Metadata["description"])
	}

	wantH2 := []any{"Next", "Last"}
	if !reflect.DeepEqual(parsed.Metadata["h2"], wantH2) {
		t.Fatalf("h2 = %#v, want %#v", parsed.Metadata["h2"], wantH2)
	}

	wantH3 := []any{"Included"}
	if !reflect.DeepEqual(parsed.Metadata["h3"], wantH3) {
		t.Fatalf("h3 = %#v, want %#v", parsed.Metadata["h3"], wantH3)
	}

	wantHeadings := []any{
		map[string]any{"level": 1, "text": "Intro"},
		map[string]any{"level": 2, "text": "Next"},
		map[string]any{"level": 3, "text": "Included"},
		map[string]any{"level": 4, "text": "Deep"},
		map[string]any{"level": 2, "text": "Last"},
	}
	if !reflect.DeepEqual(parsed.Metadata["headings"], wantHeadings) {
		t.Fatalf("headings = %#v, want %#v", parsed.Metadata["headings"], wantHeadings)
	}
}

func TestMarkdownParserDescriptionRequiresH1(t *testing.T) {
	parsed, err := MarkdownParser{}.Parse(model.Page{Body: []byte(`> Description

# Late`)})
	if err != nil {
		t.Fatalf("Parse() err = %v", err)
	}

	if parsed.Metadata["description"] != "" {
		t.Fatalf("description = %#v, want empty", parsed.Metadata["description"])
	}
}
