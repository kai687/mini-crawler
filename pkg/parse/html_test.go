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

func TestHTMLParserStripsScriptAndStyle(t *testing.T) {
	page := model.Page{
		URL: "https://example.com/page",
		Body: []byte(`<html><head><style>h1 { color: red }</style></head>` +
			`<body><h1>Hello</h1><SCRIPT>var text = "noise";</SCRIPT></body></html>`),
	}

	parsed, err := HTMLParser{}.Parse(page)
	if err != nil {
		t.Fatalf("Parse() err = %v", err)
	}

	if got := parsed.Doc.Find("script, style").Length(); got != 0 {
		t.Fatalf("script/style count = %d, want 0", got)
	}

	if got := parsed.Doc.Find("h1").First().Text(); got != "Hello" {
		t.Fatalf("h1 = %q, want Hello", got)
	}
}

func TestStripRawTextElementsKeepsSimilarTags(t *testing.T) {
	input := []byte(`<scripture>keep</scripture><style-guide>keep</style-guide>`)

	if got := string(stripRawTextElements(input)); got != string(input) {
		t.Fatalf("stripRawTextElements() = %q, want %q", got, input)
	}
}
