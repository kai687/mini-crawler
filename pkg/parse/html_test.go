package parse

import (
	"testing"

	"github.com/kai687/mini-crawler/pkg/model"
)

func TestHTMLParserParse(t *testing.T) {
	page := model.Page{
		URL:  "https://example.com/page",
		Body: []byte(`<html><body><h1>Hello</h1></body></html>`),
	}

	parsed, err := HTMLParser{}.Parse(page)
	if err != nil {
		t.Fatalf("Parse() err = %v", err)
	}

	if parsed.URL != page.URL {
		t.Fatalf("URL = %q, want %q", parsed.URL, page.URL)
	}

	if got := parsed.Doc.Find("h1").First().Text(); got != "Hello" {
		t.Fatalf("h1 = %q, want Hello", got)
	}
}
