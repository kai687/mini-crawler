package output

import (
	"bytes"
	"testing"
)

func TestJSONLWriterWrite(t *testing.T) {
	var buf bytes.Buffer

	writer := NewJSONLWriter(&buf)

	err := writer.Write(map[string]any{
		"url":                "https://example.com/page",
		"urlWithoutAnchor":   "https://example.com/page",
		"breadcrumbSegments": []any{"Guides", "Getting Started"},
		"breadcrumbHierarchy": map[string]any{
			"lvl0": "Guides",
			"lvl1": "Guides > Getting Started",
		},
		"contentType": "guide",
		"variant":     "legacy",
		"methodName":  "searchSingleIndex",
		"recordType":  "content",
		"content":     "Paragraph content",
		"hierarchy": map[string]any{
			"lvl1": "Page Title",
			"lvl2": "Section",
			"lvl3": "Detail",
		},
		"position": 0,
		"objectID": "https:--example.com-page-0",
	})
	if err != nil {
		t.Fatalf("Write() err = %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}

	want := "{" +
		"\"breadcrumbHierarchy\":{\"lvl0\":\"Guides\",\"lvl1\":\"Guides \\u003e Getting Started\"}," +
		"\"breadcrumbSegments\":[\"Guides\",\"Getting Started\"]," +
		"\"content\":\"Paragraph content\"," +
		"\"contentType\":\"guide\"," +
		"\"hierarchy\":{\"lvl1\":\"Page Title\",\"lvl2\":\"Section\",\"lvl3\":\"Detail\"}," +
		"\"methodName\":\"searchSingleIndex\"," +
		"\"objectID\":\"https:--example.com-page-0\"," +
		"\"position\":0," +
		"\"recordType\":\"content\"," +
		"\"url\":\"https://example.com/page\"," +
		"\"urlWithoutAnchor\":\"https://example.com/page\"," +
		"\"variant\":\"legacy\"" +
		"}\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestJSONLWriterWriteMultiple(t *testing.T) {
	var buf bytes.Buffer

	writer := NewJSONLWriter(&buf)

	records := []map[string]any{
		{
			"url":              "https://example.com/one",
			"urlWithoutAnchor": "https://example.com/one",
			"recordType":       "content",
			"content":          "Body one",
			"hierarchy": map[string]any{
				"lvl1": "One",
				"lvl2": "Section One",
				"lvl3": "Detail One",
			},
			"position": 0,
			"objectID": "https:--example.com-one-0",
		},
		{
			"url":              "https://example.com/two",
			"urlWithoutAnchor": "https://example.com/two",
			"recordType":       "content",
			"content":          "Body two",
			"hierarchy": map[string]any{
				"lvl1": "Two",
				"lvl2": "Section Two",
				"lvl3": "Detail Two",
			},
			"position": 1,
			"objectID": "https:--example.com-two-1",
		},
	}
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			t.Fatalf("Write() err = %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}

	want := "{" +
		"\"content\":\"Body one\"," +
		"\"hierarchy\":{\"lvl1\":\"One\",\"lvl2\":\"Section One\",\"lvl3\":\"Detail One\"}," +
		"\"objectID\":\"https:--example.com-one-0\"," +
		"\"position\":0," +
		"\"recordType\":\"content\"," +
		"\"url\":\"https://example.com/one\"," +
		"\"urlWithoutAnchor\":\"https://example.com/one\"" +
		"}\n" +
		"{" +
		"\"content\":\"Body two\"," +
		"\"hierarchy\":{\"lvl1\":\"Two\",\"lvl2\":\"Section Two\",\"lvl3\":\"Detail Two\"}," +
		"\"objectID\":\"https:--example.com-two-1\"," +
		"\"position\":1," +
		"\"recordType\":\"content\"," +
		"\"url\":\"https://example.com/two\"," +
		"\"urlWithoutAnchor\":\"https://example.com/two\"" +
		"}\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}
